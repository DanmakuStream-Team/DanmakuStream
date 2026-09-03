# content-service

DanmakuStream 内容域微服务，负责视频、投稿审核、媒体、协作者、动态、Banner、公告和创作者指标。它是独立 Go Module，只访问 `content_db`；用户与互动实体仅保存外部 ID。

## 本地运行

1. 创建 `content_db` 和最小权限账号，并执行 `migrations/001_init.sql`。
2. 设置运行环境：

```bash
export SERVICE_NAME=content-service
export SERVICE_VERSION=microservice-0.1.0
export COMMIT_SHA=0000000
export BUILD_TIME=2026-08-31T10:00:00+08:00
export PORT=8080
export DATABASE_DSN='content_user:change-me@tcp(127.0.0.1:3306)/content_db?charset=utf8mb4&parseTime=True&loc=Local'
export JWT_SECRET='local-only-secret'
export INTERNAL_API_TOKEN='local-only-token'
export STORAGE_DIR='./data'
go run ./cmd/server
```

`AUTO_MIGRATE=true` 仅供本地开发；共享环境应显式执行版本化 SQL migration。配置完整约定见 `etc/config.example.yaml`。

## 接口边界

- 公开：视频列表/详情、作者视频、动态、Banner、公告，以及统一 livez/health/version。
- 登录用户：投稿与媒体上传、编辑、封面、下载、删除、协作者、本人投稿、创作者指标和动态管理。
- `admin`/`moderator`：投稿审核；仅 `admin`：Banner 与公告管理。
- 点赞、收藏、评论、弹幕、历史和直播不属于本服务。

所有业务接口使用 `/api/v1` 前缀。上传文件通过 `/media/*` 读取，正式环境由 API Gateway 转发。作者响应保留兼容 DTO，但 D06 阶段只填充 `id`，不查询 `user_db`；用户摘要由 D07 内部 HTTP 联调补齐。

## 验证

```bash
go test ./...
go vet ./...
go build ./cmd/server
docker build -t content-service:local .
```
