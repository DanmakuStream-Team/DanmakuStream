# engagement-service tests

当前可直接执行的契约与单元测试位于各 `internal/*_test.go`，运行：

```bash
go test ./...
```

真实 MySQL API/WebSocket 回归位于 `tests/integration_test.go`。测试使用假 user/content HTTP 服务，
因此只访问 `engagement_db`，可以独立验证 D 负责的跨服务调用契约、数据库事务和 WebSocket 行为：

```bash
ENGAGEMENT_INTEGRATION_DSN='engagement_test:***@tcp(127.0.0.1:3306)/engagement_db?charset=utf8mb4&parseTime=True&loc=Local' \
  go test ./tests -run TestEngagementAPIAndWebSocketRegression -v -count=1
```

当前自动回归覆盖：平台探针/版本、鉴权与后台权限、视频点赞与收藏切换、评论创建/点赞/删除、
弹幕发送/查询/屏蔽、播放进度幂等、历史/稍后再看/收藏列表及清空、直播创建/管理/监控/设置、
点赞/礼物/Super Chat、SRS Hook 内部鉴权、WebSocket 受限聊天鉴权、消息与同用户重连计数、
浏览器推流 WebSocket 鉴权，以及直播预约冲突、取消和重复预约。

故障隔离回归会注入 content-service 宕机、250ms 超时、畸形响应及“审核通过但媒体不可播放”四种状态，
分别断言 503、504、502、409；每次故障后继续断言 `/api/v1/livez` 与 `/api/v1/health` 为 200。
这与 K8s 中只检查进程的 livenessProbe、只检查自有数据库的 readinessProbe 配合，防止下游故障触发
engagement-service 重启或形成级联 Pod 故障。

与真实 `user-service`、`content-service` 联调时还必须核对：

1. 重复点赞/收藏切换后仅保留一条有效关系；
2. 评论、弹幕的非法内容与不可播放视频；
3. 播放进度幂等覆盖、历史/稍后再看清空；
4. 重复预约、取消预约和预约自己直播；
5. 礼物/Super Chat 事务一致性；
6. WebSocket 重连和在线人数；浏览器 WebM 经 FFmpeg 向真实 SRS 转推仍需在部署环境做媒体流验收；
7. user/content 超时分别映射为 504，服务不可用映射为 503；当前客户端契约测试已覆盖；
8. 数据库不可用时 `/api/v1/health` 返回 503。

真实依赖必须提供以下接口，缺少任意一项均应作为联调阻断记录，不允许回退为跨 Schema 查询：

- user-service：`GET /internal/v1/users?id=...`、`relationships/blocked`、`memberships/status`、`relationships/following`；
- content-service：`GET /internal/v1/videos/:id`、`GET /internal/v1/videos/batch?ids=...`。

不得连接 `user_db` 或 `content_db` 执行测试。
