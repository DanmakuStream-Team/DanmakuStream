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

生产集群必须已经安装 k3s，并在 `danmakustream-microservices` Namespace 中创建 `micro-secrets`。仓库只保存键名模板，不保存真实值：

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

CD 的 precheck 会验证 Node Ready，并逐项检查 `micro-secrets` 的九个必需键是否存在，但不会输出任何 Secret 内容。

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
| HPA 策略 | 为三个服务应用 `autoscaling/v2`、1～4 副本、CPU 60% 的规范策略 |

每次运行都会上传 `microservice-cd-<sha>` artifact，保留 30 天。内容包括渲染后的非敏感清单、部署输出、可能的回滚输出和运行时证据。

## 4. 手动查看

```bash
NS=danmakustream-microservices
sudo k3s kubectl -n "$NS" get deploy,pod,svc
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

HPA 的实际扩容/缩容与依赖故障验证不在普通 push CD 中执行，避免自动制造生产压力或中断；必须使用带人工确认的 `microservice-resilience` 工作流，见 [HPA 扩缩容与依赖故障演练](hpa-chaos.md)。
