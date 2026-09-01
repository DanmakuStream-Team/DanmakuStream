# engagement-service

成员 D 负责的独立业务微服务，覆盖 UC05、UC06、UC09、UC10，以及 UC13 的弹幕管理。该目录是独立 Go Module，不导入 `backend/internal/*`，原 `backend/` 单体保持不变，可继续用于 `monolith-start` 基线和性能对比。

## 当前完成范围

- 独立入口、Go Module、Dockerfile、配置、迁移与测试；
- `livez`、`health`、`version` 统一契约；
- 视频点赞/收藏、评论、弹幕；
- 历史、播放进度、稍后再看与个人收藏；
- 直播间、直播计划/预约、礼物/Super Chat、直播点赞；
- 直播 WebSocket、浏览器推流 WebSocket、SRS Hook；
- UC13 弹幕查询与屏蔽；
- `user-service`、`content-service` 内部客户端，最大超时 2 秒；
- 仅迁移 `engagement_db` 自有表，不定义跨 Schema 外键或 JOIN。

第 6 天交付以“独立构建、启动和访问”为目标。开播通知、用户摘要/屏蔽/会员校验、视频摘要等真实跨服务主链路，要在 user/content 服务提供 `/internal/v1` 接口后联调；当前不会伪装为已经完成。

## 目录

```text
cmd/server/main.go       独立启动入口、HTTP 超时和优雅退出
internal/app/            路由、健康/版本与访问日志
internal/client/         user/content 内部 HTTP 适配器
internal/config/         环境变量优先的配置加载
internal/database/       engagement_db 连接与迁移
internal/handler/        API、WebSocket、SRS Hook
internal/logic/          纯领域规则扩展位置
internal/middleware/     JWT 与 staff 鉴权
internal/model/          仅本服务拥有的模型
internal/svc/            服务依赖边界说明
migrations/001_init.sql  独立 Schema 初始化
tests/                   集成回归说明
```

## 数据归属

本服务只拥有以下表：

- `video_likes`、`video_collections`、`comments`、`comment_likes`、`danmaku`；
- `watch_histories`、`watch_laters`；
- `live_rooms`、`live_schedules`、`live_reservations`；
- `live_likes`、`live_gifts`、`super_chats`、`live_interactions`。

表中只保存 `user_id`、`video_id` 等外部实体 ID。禁止查询或关联 `user_db`、`content_db`。

## 环境变量

| 变量 | 必填 | 示例/说明 |
| --- | --- | --- |
| `SERVICE_NAME` | 是 | `engagement-service` |
| `SERVICE_VERSION` | 是 | `microservice-0.1.0` |
| `COMMIT_SHA` | 是 | 至少 7 位 Git SHA |
| `BUILD_TIME` | 是 | RFC 3339 |
| `DATABASE_DSN` | 是 | `engagement_app:***@tcp(mysql:3306)/engagement_db?charset=utf8mb4&parseTime=True&loc=Local` |
| `JWT_SECRET` | 是 | 与 user-service 签发 JWT 使用的 Secret 相同 |
| `INTERNAL_API_TOKEN` | 是 | 内部接口凭证 |
| `USER_SERVICE_URL` | 联调时 | `http://user-service:8080` |
| `CONTENT_SERVICE_URL` | 联调时 | `http://content-service:8080` |
| `REQUEST_TIMEOUT` | 否 | 默认 `1500ms`，最大 `2s` |
| `SRS_RTMP_HOST` | 否 | 默认 `srs:1935` |
| `PORT` | 否 | 默认 `8080` |

仓库中的 `etc/config.example.yaml` 不含真实 Secret；容器环境必须通过环境变量注入敏感配置。

## 本地构建与测试

```bash
cd services/engagement-service
go test ./...
go build ./cmd/server
```

数据库准备好后可启动：

```bash
export SERVICE_NAME=engagement-service
export SERVICE_VERSION=microservice-0.1.0
export COMMIT_SHA=$(git rev-parse --short HEAD)
export BUILD_TIME=$(date --iso-8601=seconds)
export DATABASE_DSN='engagement_app:<local-password>@tcp(127.0.0.1:3306)/engagement_db?charset=utf8mb4&parseTime=True&loc=Local'
export JWT_SECRET='local-only-secret'
export INTERNAL_API_TOKEN='local-only-internal-token'
go run ./cmd/server -f etc/config.example.yaml
```

Windows PowerShell 使用 `$env:NAME='value'` 设置同名变量。

## 健康与版本

```text
GET /api/v1/livez   只检查进程，K8s livenessProbe
GET /api/v1/health  检查 engagement_db，readinessProbe
GET /api/v1/version 返回 service/version/commit/buildTime
```

正式流量只经过 API Gateway；服务容器监听 `0.0.0.0:8080`，正式 Compose/K8s 不映射宿主机业务端口。

## 内部依赖和失败语义

- user-service：`/internal/v1/users?id=...` 批量摘要、双向屏蔽关系、会员状态和关注状态；
- content-service：`/internal/v1/videos/:id`，返回存在性、可播放状态、时长和创作者 ID；

内部 GET 调用携带 `X-Internal-Token` 与入口请求的 `X-Request-ID`。连接、响应头和总请求超时均受
`REQUEST_TIMEOUT` 控制，且硬上限为 2 秒。下游 404、502、503、504 分别映射为资源不存在、无效响应、
暂不可用和调用超时；下游返回 HTTP 200 但业务码非零时也不会被误判为成功。
- SRS Hook 使用 `POST /internal/v1/live/hooks/srs` 并校验 `X-Internal-Token`，该路径不得暴露到公网网关；
- 内部请求携带 `X-Internal-Token` 与 `X-Request-ID`；
- 对象不存在返回 404；无效下游响应返回 502；依赖不可用返回 503；超时返回 504；
- 写请求不自动重试，禁止无限等待。

## Docker

```bash
docker build --build-arg SERVICE_VERSION=microservice-0.1.0 --build-arg COMMIT_SHA=$(git rev-parse HEAD) --build-arg BUILD_TIME=$(date --iso-8601=seconds) -t danmakustream/engagement-service:$(git rev-parse --short HEAD) .
```

镜像以非 root 用户运行，包含 FFmpeg 供 `/ws/live-publish/:id` 将浏览器 WebM 转推到 SRS。不得以裸 `latest` 作为正式版本证据。
