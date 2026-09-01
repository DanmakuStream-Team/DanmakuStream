#!/usr/bin/env bash
# 微服务 E2E 一键入口（#49 E 项）：
#   1. 确保微服务栈运行（docker-compose.microservices.yml，未起则构建并启动）
#   2. 等待网关健康（http://localhost:8888/gateway/health）
#   3. 以微服务模式运行全部 Playwright 用例（E2E_MICROSERVICES=1 + E2E_USE_GATEWAY=1）
# 依赖：docker compose、mysql 客户端、ffmpeg；数据库 root 口令默认 password（微服务 compose）。
set -euo pipefail
cd "$(dirname "$0")/../.."   # 仓库根

GW="http://localhost:8888"
export MYSQL_ROOT_CMD="${MYSQL_ROOT_CMD:-mysql -h127.0.0.1 -P3306 -uroot -ppassword}"
export COMPOSE_MICRO="${COMPOSE_MICRO:-docker compose -f docker-compose.microservices.yml}"

log() { printf '[e2e-micro] %s\n' "$*"; }

if ! curl -sf "$GW/gateway/health" > /dev/null 2>&1; then
  log "网关不可达，启动微服务栈（首次会构建镜像，耗时较长）…"
  $COMPOSE_MICRO up -d --build
else
  log "微服务栈已在运行"
fi

for i in $(seq 1 60); do
  if curl -sf "$GW/gateway/health" > /dev/null 2>&1; then break; fi
  [ "$i" = 60 ] && { log "网关 60 次重试后仍不可达"; $COMPOSE_MICRO ps; exit 1; }
  sleep 3
done
log "网关健康：$GW/gateway/health"

# 三个业务服务各自 livez 就绪（否则用例会在 setup 阶段失败且难定位）
for svc in user-service content-service engagement-service; do
  for i in $(seq 1 30); do
    if curl -sf "$GW/api/v1/livez" -H "X-Probe-Origin: $svc" > /dev/null 2>&1; then break; fi
    [ "$i" = 30 ] && { log "$svc 健康检查超时"; $COMPOSE_MICRO ps; exit 1; }
    sleep 2
  done
done
log "业务服务 livez 通过（经网关转发）"

cd frontend
export E2E_MICROSERVICES=1
export E2E_USE_GATEWAY=1
export E2E_BASE_URL="${E2E_BASE_URL:-http://127.0.0.1}"
log "运行全部 E2E（数据按三库准备，媒体夹具经 compose cp 注入）…"
exec ./node_modules/.bin/playwright test "$@"
