#!/usr/bin/env bash
# k3s 部署脚本（在服务器上执行，由 CD 通过 `ssh ... bash -s -- <args>` 调用）
#
# 用法：
#   k3s-deploy.sh <namespace> precheck
#   k3s-deploy.sh <namespace> deploy <backend_image> <frontend_image>
#   k3s-deploy.sh <namespace> rollback
#
# deploy 流程：预检 → 记录旧版本 → set image 滚动更新 → rollout 等待 →
#              Pod 就绪检查 → 集群内健康检查；任一步失败自动 rollout undo 回滚
#              并等待旧版本恢复，最终以非零退出（CD 据此判定部署失败）。
# 数据库/视频 PVC 不随应用部署删除；数据库不做自动回滚（向前兼容迁移原则）。
set -uo pipefail

NS="${1:?用法: k3s-deploy.sh <namespace> <precheck|deploy> [backend_image frontend_image]}"
MODE="${2:?缺少模式: precheck|deploy}"
KUBECTL="sudo k3s kubectl"
ROLLOUT_TIMEOUT="${ROLLOUT_TIMEOUT:-180s}"

log()  { printf '[k3s-deploy %s] %s\n' "$(date '+%T')" "$*"; }
die()  { log "错误: $*"; exit 1; }

precheck() {
  log "== 部署前预检 =="
  [ "$($KUBECTL get node -o jsonpath='{.items[0].status.conditions[?(@.type=="Ready")].status}')" = "True" ] || die "k3s 节点未 Ready"
  $KUBECTL get namespace "$NS" > /dev/null || die "Namespace $NS 不存在（首次请按 docs/deploy/cd-pipeline.md 初始化）"
  for secret in danmakustream-secrets; do
    $KUBECTL -n "$NS" get secret "$secret" > /dev/null || die "Secret $secret 不存在（含数据库密码与 JWT，须在服务器手动创建，CD 不管理）"
  done
  for key in DB_PASSWORD JWT_SECRET DATABASE_DSN; do
    value=$($KUBECTL -n "$NS" get secret danmakustream-secrets \
      -o "jsonpath={.data.$key}" 2>/dev/null) || die "无法读取 Secret 键 $key"
    [ -n "$value" ] || die "Secret danmakustream-secrets 缺少键 $key"
  done
  for pvc in mysql-data video-data; do
    phase=$($KUBECTL -n "$NS" get pvc "$pvc" -o jsonpath='{.status.phase}' 2>/dev/null) || die "PVC $pvc 不存在"
    [ "$phase" = "Bound" ] || die "PVC $pvc 状态为 $phase，非 Bound"
  done
  for dep in mysql srs nginx-gateway; do
    ready=$($KUBECTL -n "$NS" get deploy "$dep" -o jsonpath='{.status.readyReplicas}')
    [ "${ready:-0}" -ge 1 ] || die "基础服务 $dep 未就绪（readyReplicas=${ready:-0}）"
  done
  for dep in backend frontend; do
    $KUBECTL -n "$NS" get deploy "$dep" > /dev/null || die "Deployment $dep 不存在"
  done
  log "预检通过：节点/Namespace/Secret/PVC/基础服务/目标 Deployment 均正常"
}

current_image() { $KUBECTL -n "$NS" get deploy "$1" -o jsonpath='{.spec.template.spec.containers[0].image}'; }

capture_diagnostics() {
  log "== 失败现场诊断（回滚前） =="
  $KUBECTL -n "$NS" get pods -o wide || true
  $KUBECTL -n "$NS" describe pods -l app=frontend || true
  $KUBECTL -n "$NS" logs -l app=frontend --all-containers=true --tail=100 --prefix=true || true
  $KUBECTL -n "$NS" get events --sort-by=.lastTimestamp | tail -50 || true
}

rollback() {
  log "!! 部署失败，回滚到上一版本：$*"
  capture_diagnostics
  $KUBECTL -n "$NS" rollout undo deploy/backend   || true
  $KUBECTL -n "$NS" rollout undo deploy/frontend  || true
  $KUBECTL -n "$NS" rollout status deploy/backend   --timeout="$ROLLOUT_TIMEOUT" || true
  $KUBECTL -n "$NS" rollout status deploy/frontend  --timeout="$ROLLOUT_TIMEOUT" || true
  log "回滚完成；旧版本 Backend=$(current_image backend) Frontend=$(current_image frontend)"
}

deploy() {
  local backend_image="${1:?缺少 backend 镜像}"
  local frontend_image="${2:?缺少 frontend 镜像}"

  precheck

  local old_backend old_frontend old_rev
  old_backend=$(current_image backend)
  old_frontend=$(current_image frontend)
  old_rev=$($KUBECTL -n "$NS" get deploy backend -o jsonpath='{.metadata.annotations.deployment\.kubernetes\.io/revision}')
  log "旧版本记录：Backend=$old_backend Frontend=$old_frontend revision=$old_rev"
  log "目标版本：Backend=$backend_image Frontend=$frontend_image"

  # 滚动更新（失败即回滚）
  if ! $KUBECTL -n "$NS" set image deployment/backend backend="$backend_image" \
     || ! $KUBECTL -n "$NS" set image deployment/frontend frontend="$frontend_image"; then
    rollback "set image 失败"; exit 1
  fi
  if ! $KUBECTL -n "$NS" rollout status deploy/backend --timeout="$ROLLOUT_TIMEOUT" 2>&1; then
    rollback "backend rollout 超时/失败"; exit 1
  fi
  if ! $KUBECTL -n "$NS" rollout status deploy/frontend --timeout="$ROLLOUT_TIMEOUT" 2>&1; then
    rollback "frontend rollout 超时/失败"; exit 1
  fi

  # Pod 状态检查：不得有 ImagePullBackOff / CrashLoopBackOff
  sleep 5
  local bad
  bad=$($KUBECTL -n "$NS" get pods --field-selector=status.phase!=Succeeded \
        -o jsonpath='{range .items[*]}{.status.containerStatuses[0].state}{"\n"}{end}' | grep -E "Waiting|CrashLoop" || true)
  if [ -n "$bad" ]; then
    $KUBECTL -n "$NS" get pods -o wide
    rollback "存在未就绪容器: $bad"; exit 1
  fi

  # 集群内健康检查：backend -> mysql 连通（/health 含 DB ping）、livez 存活
  if ! $KUBECTL -n "$NS" exec deploy/backend -- wget -qO- http://localhost:8080/api/v1/health 2>/dev/null | grep -q '"db":"up"'; then
    rollback "集群内 /api/v1/health 检查失败（DB 连接异常）"; exit 1
  fi
  if ! $KUBECTL -n "$NS" exec deploy/backend -- wget -qO- http://localhost:8080/api/v1/livez > /dev/null 2>&1; then
    rollback "集群内 /api/v1/livez 检查失败"; exit 1
  fi
  # backend -> frontend Service -> nginx -> backend Service 完整代理通路。
  # 不在 frontend 容器内请求 localhost：Alpine 会优先解析到 ::1，
  # 而当前 nginx 仅监听 IPv4，容易产生 connection refused 的假失败；
  # 通过 Service 访问也可避免 rollout 后 exec 误选正在终止的旧 frontend Pod。
  if ! $KUBECTL -n "$NS" exec deploy/backend -- wget -qO- http://frontend/api/v1/health 2>/dev/null | grep -q '"status":"ok"'; then
    rollback "frontend Service/nginx 代理 /api/v1/health 失败"; exit 1
  fi

  log "== 部署成功并验证通过：Backend=$backend_image Frontend=$frontend_image =="
  $KUBECTL -n "$NS" get pods,svc,ingress -o wide
}

case "$MODE" in
  precheck) precheck ;;
  deploy)   deploy "${3:?}" "${4:?}" ;;
  rollback) rollback "公网健康检查失败" ;;
  *) die "未知模式: $MODE" ;;
esac
