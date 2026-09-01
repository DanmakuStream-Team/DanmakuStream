# CD 流水线设计与运维手册

> 对应工作流：`.github/workflows/deploy.yml`（由 `ci.yml` 的 `deploy` job 在 **PR 合并到 dev 且全部测试通过** 时调用）；服务器端脚本：`scripts/k3s-deploy.sh`。

## 1. 流程与触发规则

```text
功能分支 / 普通 PR            只跑 CI（deploy job 跳过）
直接 push 到 dev              跑完整 CI，合并来源校验不通过，CD 跳过
PR 合并到 dev                 CI 全绿 → 校验 merge commit 与 PR → CD：
                                构建单体前后端镜像（tag = commit SHA 前 12 位）
                                → 推送 GHCR（另推 dev 滚动标签）
                                → SSH 连接公网 k3s
                                → 应用并等待 nginx-gateway Ready
                                → 预检（节点/NS/Secret/PVC/基础服务/旧版本记录）
                                → kubectl set image 滚动更新
                                → rollout 等待 + Pod 状态 + 集群内三链路健康检查
                                → Runner 侧公网健康检查（/api/v1/health 需 db:up）
                                → 失败任意环节：自动 rollout undo 回滚旧版本并保留证据
main                          暂不自动部署（留给最终稳定版/人工批准）
```

- `concurrency: production-deploy`（排队不取消）保证**同一时间只有一次部署**。
- 数据库与视频 PVC 不随部署删除；**数据库不做自动回滚**（结构变更遵循向前兼容迁移）。
- CD 会在应用前自动创建/更新 `nginx-gateway` ConfigMap、Deployment 与 Service，避免 frontend 因网关 DNS 不存在而启动失败。
- 镜像规则：不使用裸 `latest`。当前 CD 部署单体前后端；三个微服务由各自独立 CI 使用完整 commit SHA 构建并推送，详见 [微服务独立 CI](microservice-ci.md)。

### 微服务 CI 阻断

`user-service-ci`、`content-service-ci`、`engagement-service-ci` 分别执行服务编译、测试和镜像构建。每条流水线中测试步骤位于镜像步骤之前，因此任一服务测试失败时都不会构建或推送该服务镜像。三个检查应同时配置为 `dev` 和 `main` 的必需状态检查。

## 2. 服务器准备（一次性，需要一台公网服务器）

**规格建议**：2 核 4G 以上（跑 k3s + MySQL + SRS + 前后端），Ubuntu 22.04/24.04，开放 80/443（Ingress）、22（SSH）；直播演示再放行 1935（RTMP）与 8081（HLS，或由 Ingress 代理）。

```bash
# 1) 安装 k3s（含内置 Ingress）
curl -sfL https://get.k3s.io | sh -s - --write-kubeconfig-mode 644

# 2) 应用单体部署骨架（deploy/k8s/monolith/）
git clone https://github.com/DanmakuStream-Team/DanmakuStream.git && cd DanmakuStream
sudo k3s kubectl apply -f deploy/k8s/monolith/namespace.yaml
sudo k3s kubectl -n danmakustream create secret generic danmakustream-secrets \
  --from-literal=DB_PASSWORD='<真实数据库密码>' \
  --from-literal=JWT_SECRET='<真实JWT密钥>' \
  --from-literal=DATABASE_DSN='root:<真实数据库密码>@tcp(mysql:3306)/danmakustream?charset=utf8mb4&parseTime=True&loc=Local'  # CD 永不覆盖这些值
sudo k3s kubectl apply -f deploy/k8s/monolith/        # 其余资源（PVC/MySQL/SRS/前后端/Ingress）

# 3) 首次部署后 CD 用 set image 接管版本；首次镜像可手动：
#    sudo k3s kubectl -n danmakustream set image deployment/backend backend=ghcr.io/.../backend:<sha>
```

> GHCR 私有镜像需在服务器登录拉取：`sudo docker login ghcr.io`（用 PAT）或创建 `regcred` Secret 并在 Deployment 补 `imagePullSecrets`。若仓库公开则免配置。

## 3. GitHub Environment 配置（一次性）

Settings → Environments → 新建 **`production`**，添加 Secrets：

| Secret | 内容 |
|---|---|
| `K3S_HOST` | 服务器公网 IP/域名 |
| `K3S_SSH_USER` | SSH 用户（需免密 sudo k3s kubectl） |
| `K3S_SSH_KEY` | SSH 私钥（ed25519） |
| `K3S_HOST_KEY` | `ssh-keyscan <host>` 输出（防中间人） |
| `PUBLIC_URL` | 公网访问地址，如 `http://<ip>` 或 `https://danmaku.example.com` |

**绝不入库/入 Secret**：明文数据库密码、JWT、k3s node token、完整 kubeconfig、个人密码（这些只存在于服务器上的 `danmakustream-secrets`）。

## 4. 部署验证清单（CD 自动执行）

| 层 | 检查 |
|---|---|
| K8s | Node Ready、PVC Bound、rollout 完成、无 `ImagePullBackOff`/`CrashLoopBackOff` |
| 集群内 | backend→frontend Service→Nginx→backend Service 的代理 `/api/v1/health`；backend→MySQL（health 含 `db:up`）；`/api/v1/livez` 存活 |
| 公网 | `$PUBLIC_URL/api/v1/health` 返回 200 且 `db:up`（10 次重试） |

每次部署产出 **`cd-evidence-<sha>` artifact**：CI 记录、镜像地址、rollout 输出、Pod/Service/Ingress/事件快照、backend 日志（失败时）——对应任务书"部署证据"要求。

## 5. 失败与回滚

触发回滚的条件：镜像拉取失败、rollout 超时（默认 180s，可用 `ROLLOUT_TIMEOUT` 覆盖）、集群内健康检查失败、公网检查失败。回滚由 `scripts/k3s-deploy.sh` 内部执行 `rollout undo`（backend+frontend）并等待旧版本就绪，CD 以非零退出并把全过程写入证据 artifact。

## 6. 验收演示（对应任务书两条流水线记录）

1. **成功记录**：合并一个全绿 PR 到 dev → 三个微服务独立 CI 与单体 CI 全部通过 → Actions 里 `ci` 工作流的 `CD 部署到公网 k3s` job 绿 → 下载 `cd-evidence-<sha>` artifact → 访问 `$PUBLIC_URL` 现场演示。
2. **失败阻断记录**：已有 `docs/testing/reports/ci-red-green-evidence-2026-08-27.md`（红灯时 api-test/docker-build 全部 skipped）——CD 位于 docker-build 之后，同样被阻断；如需 CD 专属记录，可临时推送一个会让集成测试失败的提交到 dev（取证后 revert，注意 dev 部署会停在旧版本）。

## 7. 尚未自动化 / 后续增强

- MySQL 使用 Recreate 策略已由 manifest 保证（避免双 Pod 挂同一卷）；SRS 直播链路（1935/8081）的公网验证待直播功能启用后补充；
- 数据库结构迁移目前依赖 GORM AutoMigrate 启动时执行（向前兼容），独立迁移工具后续引入；
- main 分支的人工批准发布流程在最终交付阶段补。
