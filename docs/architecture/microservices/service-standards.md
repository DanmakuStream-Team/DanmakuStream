# DanmakuStream 微服务统一规范

> 状态：第 6 天微服务拆分的强制基线
>
> 适用范围：API Gateway、`user-service`、`content-service`、`engagement-service` 及调用它们的前端、Compose、Kubernetes 和 CI/CD
>
> 基线原则：保留现有 `backend/`、`docker-compose.yml` 和 `deploy/k8s/monolith/` 作为单体版本，不在拆分过程中搬移或覆盖。

## 1. 目标与责任边界

本规范用于保证三个业务服务能够独立构建、测试、运行和部署，同时保持现有前端公开 API 尽量兼容。所有成员在新增服务代码前必须遵守本文约定；需要改变公开路径、数据归属或错误格式时，必须先更新本文并经 Review。

| 组件 | 主责 | 用例/职责 | 代码所有权 |
| --- | --- | --- | --- |
| API Gateway 与部署平台 | A | 路由、配置、Secret、Compose、K8s、健康与版本规范 | `docker-compose.microservices.yml`、`deploy/k8s/microservices/`、`deploy/nginx-gateway-microservices.conf` |
| `user-service` | B | UC01、UC07、UC08、UC11；UC13 用户与角色管理 | `services/user-service/` |
| `content-service` | C | UC02、UC03、UC04、UC12；UC13 内容运营 | `services/content-service/` |
| `engagement-service` | D | UC05、UC06、UC09、UC10；UC13 弹幕管理 | `services/engagement-service/` |
| 前端与微服务 E2E | E | 统一网关入口、登录态、错误处理和 E2E 环境 | `frontend/` 中的公共请求层与 E2E 配置 |

任何服务不得直接读取其他服务的 Schema，也不得直接导入其他服务的 `internal` 包。跨服务数据通过内部 HTTP API 获取，只在本地保存对方实体的 ID。

## 2. 仓库目录约定

```text
backend/                              # 已冻结的单体后端
services/
  user-service/
  content-service/
  engagement-service/
docker-compose.yml                    # 单体 Compose
docker-compose.microservices.yml      # 微服务 Compose
deploy/
  nginx-gateway.conf                  # 单体网关
  nginx-gateway-microservices.conf    # 微服务网关
  k8s/
    monolith/                         # 单体清单
    microservices/                    # 微服务清单
```

每个业务服务是独立 Go Module，并至少包含：

```text
services/<service>/
  cmd/server/main.go
  internal/config/
  internal/handler/
  internal/logic/
  internal/middleware/
  internal/model/
  internal/svc/
  migrations/
  tests/
  etc/config.example.yaml
  Dockerfile
  go.mod
  go.sum
  README.md
```

服务可以从单体复制必要代码后逐步裁剪，但不得复制整个 `backend/` 后长期保留无关 Handler、模型或表。

## 3. 服务命名、端口与网络

| 项目 | 规范 |
| --- | --- |
| Compose/K8s 服务名 | `user-service`、`content-service`、`engagement-service` |
| 容器监听地址 | `0.0.0.0:8080` |
| 宿主机端口 | 三个业务服务不直接暴露；开发调试需临时映射时不得写入正式 Compose |
| 对外入口 | HTTP、上传和 WebSocket 全部经过 API Gateway |
| 服务发现 | Compose 使用服务名；K8s 使用 Service DNS |
| API 前缀 | 公开 API 保持 `/api/v1` |
| 时区 | `Asia/Shanghai`；存储和接口时间优先使用带时区的 RFC 3339 |

前端不得引用某个业务服务的主机名或端口。浏览器只访问同源 `/api/v1/*`、`/ws/*` 和 `/media/*`。

## 4. 统一响应与错误规范

### 4.1 成功响应

```json
{
  "code": 0,
  "message": "ok",
  "data": {}
}
```

列表接口的分页信息放入 `data`：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "items": [],
    "page": 1,
    "pageSize": 20,
    "total": 0
  }
}
```

### 4.2 失败响应

```json
{
  "code": 40401,
  "message": "video not found",
  "requestId": "01J6..."
}
```

| HTTP 状态 | 使用场景 | 业务码范围 |
| --- | --- | --- |
| 400 | 参数、格式或状态转换不合法 | `40000`～`40099` |
| 401 | 未携带凭证、Token 无效或过期 | `40100`～`40199` |
| 403 | 角色、资源所有权或内部调用权限不足 | `40300`～`40399` |
| 404 | 用户、视频、直播或其他资源不存在 | `40400`～`40499` |
| 409 | 幂等键冲突、重复操作或业务状态冲突 | `40900`～`40999` |
| 413 | 上传文件超过限制 | `41300`～`41399` |
| 429 | 触发限流 | `42900`～`42999` |
| 500 | 当前服务内部错误 | `50000`～`50099` |
| 502 | 下游返回无效响应 | `50200`～`50299` |
| 503 | 数据库或下游服务暂时不可用 | `50300`～`50399` |
| 504 | 下游服务调用超时 | `50400`～`50499` |

不得把 SQL、堆栈、DSN、Token 或内部地址返回给客户端。业务错误必须同时具有正确的 HTTP 状态，不能全部返回 HTTP 200。

## 5. 健康与版本接口

三个业务服务必须实现相同路径和结构。

### 5.1 存活检查

`GET /api/v1/livez`

- 只验证进程能够处理请求。
- 不访问数据库或下游服务。
- 正常返回 HTTP 200。
- Kubernetes `livenessProbe` 使用此接口。

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "status": "up"
  }
}
```

### 5.2 就绪检查

`GET /api/v1/health`

- 至少检查本服务拥有的数据库连接。
- 数据库不可用时返回 HTTP 503。
- 下游状态可放在 `dependencies` 中；只有核心能力完全不可用时才将服务标记为未就绪，避免下游故障造成所有 Pod 连锁重启。
- Docker Compose 健康检查和 Kubernetes `readinessProbe` 使用此接口。

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "status": "up",
    "database": "up",
    "dependencies": {
      "user-service": "up"
    }
  }
}
```

### 5.3 版本查询

`GET /api/v1/version`

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "service": "user-service",
    "version": "microservice-0.1.0",
    "commit": "abc1234",
    "buildTime": "2026-08-31T10:00:00+08:00"
  }
}
```

`service`、`version`、`commit` 和 `buildTime` 均不得为空。正式镜像的 `commit` 使用完整或至少 7 位 Git SHA，不使用 `latest` 代替版本证据。

API Gateway 自身提供 `GET /gateway/health`；Kubernetes 探针直接检查各业务 Service，不通过网关检查业务 Pod。

## 6. 配置与 Secret

服务必须优先从环境变量读取容器环境配置；本地示例 YAML 只能提供非敏感默认值。

### 6.1 公共配置

| 环境变量 | 必填 | 说明 |
| --- | --- | --- |
| `SERVICE_NAME` | 是 | 固定为服务名 |
| `SERVICE_VERSION` | 是 | 发布版本 |
| `COMMIT_SHA` | 是 | 构建提交 |
| `BUILD_TIME` | 是 | RFC 3339 构建时间 |
| `PORT` | 否 | 默认 `8080` |
| `DATABASE_DSN` | 是 | 当前服务自己的数据库连接 |
| `JWT_SECRET` | 是 | 公开请求 Token 校验密钥 |
| `INTERNAL_API_TOKEN` | 是 | 内部 HTTP 调用凭证 |
| `LOG_LEVEL` | 否 | 默认 `info` |
| `REQUEST_TIMEOUT` | 否 | 下游调用默认不超过 `2s` |

按依赖关系使用 `USER_SERVICE_URL`、`CONTENT_SERVICE_URL` 或 `ENGAGEMENT_SERVICE_URL`，没有该依赖的服务不得配置无关地址。

### 6.2 管理规则

- ConfigMap 只保存端口、服务地址、日志级别和超时等非敏感项。
- Secret 保存 DSN、JWT 密钥和内部调用 Token。
- 仓库只提交 `.env.microservices.example` 和 `secret.example.yaml.template`。
- 禁止提交真实密码、Token、私钥和带密码的生产 DSN。
- 日志和错误响应不得打印 Secret。

## 7. 身份认证与内部调用

### 7.1 公开请求

- `user-service` 负责签发 JWT。
- 三个服务使用同一签名算法和密钥验证 JWT，并统一解析 `userId`、`role` 和过期时间。
- 网关只负责转发，不代替业务服务执行资源所有权和角色判断。
- 不信任客户端直接传入的 `X-User-ID` 或 `X-Role`。

### 7.2 内部请求

- 内部接口使用 `/internal/v1` 前缀，不从公网网关暴露。
- 调用方携带 `X-Internal-Token` 和 `X-Request-ID`。
- 内部接口仍需校验参数和调用凭证。
- 默认连接和总请求超时均不得超过 2 秒。
- 只允许对幂等 GET 做最多一次受控重试；创建订单、发送消息、点赞、支付等写请求不得自动重试。
- 下游失败必须映射为明确的 502、503 或 504，不得无限等待。

建议优先提供批量摘要接口，避免列表页面产生逐条跨服务请求。

## 8. 数据所有权

三个服务可以共用一个 MySQL 实例，但必须使用独立 Schema 和最小权限账号。

| 服务 | Schema | 主要数据 |
| --- | --- | --- |
| `user-service` | `user_db` | 用户、关系/分组/屏蔽、会员方案/订单/订阅、会话/消息/已读、通知 |
| `content-service` | `content_db` | 视频、媒体、封面、投稿/审核、协作者、动态、横幅/公告、创作者指标 |
| `engagement-service` | `engagement_db` | 点赞/收藏/评论/弹幕、观看历史/进度/稍后再看、直播/预约/礼物/Super Chat |

强制规则：

1. 每个服务只迁移和维护自己的表。
2. 数据库账号只拥有本 Schema 的权限。
3. 禁止跨 Schema `JOIN`、外键和直接查询。
4. 跨域引用只存不可变 ID，例如 `user_id`、`video_id`。
5. 删除跨域实体时通过内部 API 或后续事件机制处理，不使用跨库级联删除。
6. 数据库迁移文件随服务代码放在 `services/<service>/migrations/`。

## 9. API Gateway 路由归属

公开路径尽量保持单体版本兼容；最具体的路由必须写在通用路由之前。

| 路由 | 目标服务 |
| --- | --- |
| `/api/v1/auth/*` | `user-service` |
| `/api/v1/users/*`、`/api/v1/search/users` | `user-service` |
| `/api/v1/subscriptions/*`、会员方案接口 | `user-service` |
| `/api/v1/messages/*`、`/api/v1/notifications/*` | `user-service` |
| `/ws/chat` | `user-service` |
| 视频列表、详情、上传、编辑、媒体和创作者分析 | `content-service` |
| `/api/v1/dynamics/*` | `content-service` |
| `/api/v1/admin/videos/*`、横幅和公告接口 | `content-service` |
| `/api/v1/videos/:id/like`、`/api/v1/videos/:id/collect` | `engagement-service` |
| `/api/v1/comments/*`、`/api/v1/danmaku/*` | `engagement-service` |
| 观看历史、稍后再看和个人收藏 | `engagement-service` |
| `/api/v1/live*`、`/ws/live/*`、`/ws/live-publish/*` | `engagement-service` |
| `/api/v1/admin/danmaku/*` | `engagement-service` |

`/api/v1/admin/infrastructure` 属于平台聚合能力，不归任何业务 Schema。第 6 天保留为待联调项；在实现聚合端点前，不得随意放入某个业务服务并读取其他 Schema。

网关必须：

- 转发 `Host`、`X-Real-IP`、`X-Forwarded-For`、`X-Forwarded-Proto` 和 `X-Request-ID`。
- 为 WebSocket 转发 `Upgrade` 和 `Connection`。
- 上传请求关闭不必要的请求缓冲，并设置明确的大小与超时上限。
- 未知 API 返回 404，不设置“全部转给某个服务”的兜底业务路由。

## 10. 日志、超时与可观测性

应用日志输出到标准输出/标准错误，不在容器内维护滚动日志文件。每条访问日志至少包含：

```text
timestamp level service requestId method path status latencyMs userId
```

- `requestId` 优先沿用网关传入值，没有时由入口服务生成。
- 不记录密码、JWT、内部 Token、完整 Cookie 或私信正文。
- 下游调用日志记录目标服务、耗时和结果，不记录完整敏感请求体。
- HTTP Server 应设置读取、写入和空闲超时。
- 优雅退出时停止接收新请求，并为正在处理的请求保留有限完成时间。

## 11. Docker、Compose 与 Kubernetes

### 11.1 Docker

- 使用多阶段构建。
- 运行镜像使用非 root 用户。
- 只复制运行必需文件。
- 正式镜像标签使用 Git SHA。
- Dockerfile 中声明 `EXPOSE 8080`。
- 镜像内健康检查使用 `/api/v1/health`。

### 11.2 Compose

- 微服务使用 `docker-compose.microservices.yml`，不得覆盖单体 Compose。
- 三个业务服务只使用 `expose: 8080`。
- 宿主机只暴露前端、API Gateway、SRS 所需端口。
- 服务通过容器服务名互相访问。
- 健康依赖使用 `condition: service_healthy`，同时避免形成循环启动依赖。

### 11.3 Kubernetes

每个服务必须提供 Deployment 和 ClusterIP Service，并至少配置：

```yaml
livenessProbe:
  httpGet:
    path: /api/v1/livez
    port: 8080
readinessProbe:
  httpGet:
    path: /api/v1/health
    port: 8080
resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 500m
    memory: 512Mi
```

第 6 天先使用单副本；HPA、故障演练和资源参数调优在后续计划中完成。

## 12. 测试与完成标准

每个服务 PR 至少通过：

1. 独立 `go test ./...`。
2. 使用自己的 Schema 运行 API/集成测试。
3. `livez`、`health` 和 `version` 契约测试。
4. 未认证、权限不足、资源不存在和数据库不可用测试。
5. Docker 镜像独立构建。
6. Compose 配置解析和 Kubernetes `--dry-run=client` 校验。

第 6 天平台阶段完成标准：

- 三个服务目录、构建入口、镜像和健康/版本接口齐全。
- 网关路由表没有兜底错投和已知路径冲突。
- 三个 Schema 与账号权限方案明确。
- 微服务 Compose 和 K8s 骨架可解析。
- 前端只经网关访问。
- 单体目录和部署文件仍可用于 `v1.0` 对比。

第 7 天继续完成：真实跨服务主链路、移除全部跨 Schema 查询、内部接口鉴权、超时/降级和最小权限账号实测。第 8～9 天再完成独立流水线、全量回归、HPA、故障和性能对比，不把这些内容伪装成第 6 天已完成。

## 13. 变更与 Review 规则

- A 维护本规范、网关和部署模板；B/C/D 分别维护自己的服务目录；E 维护公共前端和 E2E 环境。
- 修改公开 API、路由归属、Schema 归属或公共错误格式时，PR 必须同步修改本文。
- PR 描述必须列出对应 UC、测试命令和部署验证结果。
- 任何 Secret 泄漏、跨 Schema 查询、绕过网关的前端请求或使用 `latest` 的正式部署均视为阻断项。
