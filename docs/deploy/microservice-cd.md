# 三微服务自动部署与排障手册

> 对应工作流：`.github/workflows/microservice-cd.yml`；清单渲染：`scripts/render-microservice-manifests.sh`；集群验证与回滚：`scripts/k3s-microservice-deploy.sh`。

## 1. 发布链路

```text
push/merge 到 dev
  ├─ user-service-ci       ─┐
  ├─ content-service-ci    ─┼─ 同一完整 commit SHA 全绿
  ├─ engagement-service-ci ─┤
  └─ microservices-e2e     ─┘
              ↓
  渲染三个不可变 SHA 镜像 → apply 到 k3s → 等待 rollout
              ↓
  三服务 livez + health + version(commit=SHA) 验证
              ↓
  成功：保存状态、响应和三服务日志
  失败：四个工作负载 rollout undo，再保存同样证据
```

工作流只响应会影响微服务运行结果的路径。普通功能分支和 PR 运行测试但不连接生产集群。`microservice-production-deploy` concurrency 保证发布串行执行。

## 2. 一次性集群准备

生产集群必须已经安装 k3s 和 metrics-server，并在 `danmakustream-microservices` Namespace 中创建 `micro-secrets`。k3s 默认包含 metrics-server；若 `kubectl top node` 不可用，必须先修复指标组件，不能把 HPA 的 `TARGETS` 为 `<unknown>` 当作成功。仓库只保存键名模板，不保存真实值：

```bash
sudo k3s kubectl apply -f deploy/k8s/microservices/namespace.yaml
sudo k3s kubectl -n danmakustream-microservices create secret generic micro-secrets \
  --from-literal=MYSQL_ROOT_PASSWORD='<root-password>' \
  --from-literal=USER_DB_PASSWORD='<user-db-password>' \
  --from-literal=CONTENT_DB_PASSWORD='<content-db-password>' \
  --from-literal=ENGAGEMENT_DB_PASSWORD='<engagement-db-password>' \
  --from-literal=USER_DATABASE_DSN='<user_app DSN>' \
  --from-literal=CONTENT_DATABASE_DSN='<content_app DSN>' \
  --from-literal=ENGAGEMENT_DATABASE_DSN='<engagement_app DSN>' \
  --from-literal=JWT_SECRET='<jwt-secret>' \
  --from-literal=INTERNAL_API_TOKEN='<internal-token>'
```

GitHub `production` Environment 复用以下 Secrets：`K3S_HOST`、`K3S_SSH_USER`、`K3S_SSH_KEY`、`K3S_HOST_KEY`。SSH 用户需要能够免交互执行 `sudo k3s kubectl`。私有 GHCR 包还需要给 Deployment 配置 `imagePullSecrets`；当前公开包不需要仓库凭据进入集群。

CD 的 precheck 会验证 Node Ready、`metrics.k8s.io` 可用，并逐项检查 `micro-secrets` 的九个必需键是否存在，但不会输出任何 Secret 内容。

## 3. 自动验收内容

| 项目 | 自动检查 |
|---|---|
| 镜像 | 三个 Deployment 的镜像标签必须等于触发提交的完整 40 位 SHA |
| 滚动发布 | MySQL、三服务和 Gateway 均在超时内 Ready |
| 存活 | 每个服务容器内访问 `/api/v1/livez` 成功 |
| 就绪 | 每个服务容器内访问 `/api/v1/health` 成功，因而数据库也已可用 |
| 版本 | 每个 `/api/v1/version` 响应必须包含目标 commit SHA |
| 日志 | 保存 user/content/engagement/gateway 各 150 行日志 |
| 集群快照 | 保存 Deployment、Pod、Service、PVC、镜像和最近 50 条 Event |
| 自动扩缩容 | user/content/engagement/gateway 四个 HPA（autoscaling.yaml 统一管理，1～5 副本、CPU 60%）；保存目标值、副本数和容器资源使用率 |

每次运行都会上传 `microservice-cd-<sha>` artifact，保留 30 天。内容包括渲染后的非敏感清单、部署输出、可能的回滚输出和运行时证据。

## 4. 手动查看

```bash
NS=danmakustream-microservices
sudo k3s kubectl -n "$NS" get deploy,pod,svc
sudo k3s kubectl -n "$NS" get hpa --watch
sudo k3s kubectl -n "$NS" top pod --containers
sudo k3s kubectl -n "$NS" logs deploy/user-service --tail=100
sudo k3s kubectl -n "$NS" logs deploy/content-service --tail=100
sudo k3s kubectl -n "$NS" logs deploy/engagement-service --tail=100
sudo k3s kubectl -n "$NS" exec deploy/user-service -- wget -qO- http://127.0.0.1:8080/api/v1/version
sudo k3s kubectl -n "$NS" exec deploy/content-service -- wget -qO- http://127.0.0.1:8080/api/v1/version
sudo k3s kubectl -n "$NS" exec deploy/engagement-service -- wget -qO- http://127.0.0.1:8080/api/v1/version
```

## 5. 失败定位顺序

1. 在 Actions 中确认失败发生在质量门禁、apply、rollout、探针还是版本校验。
2. 下载该次 `microservice-cd-<sha>` artifact，先看 `deploy-output.txt` 和 `runtime-evidence.txt`。
3. `ImagePullBackOff` 优先核对完整 SHA 镜像是否存在及 GHCR 拉取权限。
4. `CrashLoopBackOff` 优先看对应服务日志；数据库连接失败再核对 MySQL Ready、DSN 的主机是否为 `mysql`、Secret 键是否匹配。
5. readiness 失败但 livez 成功，说明进程存活而依赖未就绪；不要放宽探针，应修复数据库或依赖配置。
6. version 不含目标 SHA，检查是否误用浮动标签或清单是否用环境变量覆盖了镜像内构建元数据。

自动回滚只回滚 Deployment revision，不回滚数据库结构和 PVC。数据库迁移必须保持向前、向后兼容。

## 6. 自动扩缩容

`deploy/k8s/microservices/autoscaling.yaml` 为三个业务服务和 API Gateway 配置 `autoscaling/v2` HPA：CPU 达到 request 的 60% 时扩容，副本范围为 1～5；升容不设稳定窗口，缩容等待 60 秒并逐步回落。四个 Deployment 均已有 CPU/内存 requests 与 limits，且不再在清单中声明固定 `replicas`，避免每次 CD apply 把 HPA 当前副本数重置为 1。

MySQL 不使用 HPA。它挂载 `ReadWriteOnce` PVC，直接复制普通 MySQL Pod 会产生多主写入和数据损坏风险；数据库扩展必须使用主从/集群方案或托管数据库，不属于本次无状态 Pod 扩缩容。

当前微服务 k3s/CD 清单没有部署前端和 SRS；本次 HPA 已覆盖该清单中的全部无状态 Deployment。它们后续迁入 Kubernetes 时，需要先定义 CPU/内存 requests 与 limits，再分别配置和压测 HPA。

完整的压力触发、时间线采集和验收步骤见 [HPA 扩缩容实验](hpa-experiment.md)。
