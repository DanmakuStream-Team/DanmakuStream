# user-service

成员 B 在第 6 天拆出的独立用户域服务，覆盖 UC01、UC07、UC08、UC11，以及 UC13 的用户列表和角色管理能力。服务拥有独立 Go Module，只迁移用户域模型，并继续保留原 `backend/` 单体实现作为回退路径。

## 数据归属

本服务只读写 `user_db`：用户、关注/分组/屏蔽、通知、私信、创作者会员方案、订阅和订单。私信分享视频只保存 `shared_video_id` 外部标识，不迁移视频表，也不直接查询内容库；内容有效性与展示信息由后续 content-service 内部 API 补齐。

## 平台接口

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/v1/livez` | 进程存活探针 |
| GET | `/api/v1/health` | 数据库就绪探针 |
| GET | `/api/v1/version` | 服务、版本、提交和构建时间 |

业务成功响应统一为 `{"code":0,"message":"ok","data":{...}}`；失败响应带非零业务码和 `requestId`。调用方可传 `X-Request-ID`，未传时服务自动生成。

## 内部接口

内部接口统一使用 `/internal/v1` 前缀和 `X-Internal-Token` 鉴权：

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/internal/v1/users?id=1&id=2` | 批量查询用户摘要 |
| GET | `/internal/v1/users/:id/exists` | 判断用户是否存在 |
| GET | `/internal/v1/relationships/blocked?firstId=1&secondId=2` | 判断双向屏蔽关系 |
| GET | `/internal/v1/memberships/status?subscriberId=1&creatorId=2` | 判断有效会员关系 |

## 配置和运行

必填环境变量：`DATABASE_DSN`、`JWT_SECRET`、`INTERNAL_API_TOKEN`。标准元数据变量为 `SERVICE_NAME`、`SERVICE_VERSION`、`COMMIT_SHA`、`BUILD_TIME`、`PORT`。本地配置示例见 `etc/config.example.yaml`。

```bash
cd services/user-service
go test ./...
go run ./cmd/server -f etc/config.example.yaml
docker build -t danmakustream/user-service:day6 .
```

数据库建表脚本位于 `migrations/001_init.sql`。容器使用非 root 用户运行并暴露 8080 端口。
