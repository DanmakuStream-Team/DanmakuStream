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

当前自动回归覆盖：平台探针/版本、视频点赞与收藏切换、评论与弹幕校验、播放进度幂等、稍后再看、
直播创建/设置/点赞/礼物、WebSocket 消息与同用户重连计数、预约冲突/取消及重复预约。

与真实 `user-service`、`content-service` 联调时还必须核对：

1. 重复点赞/收藏切换后仅保留一条有效关系；
2. 评论、弹幕的非法内容与不可播放视频；
3. 播放进度幂等覆盖、历史/稍后再看清空；
4. 重复预约、取消预约和预约自己直播；
5. 礼物/Super Chat 事务一致性；
6. WebSocket 重连和在线人数；
7. user/content 超时分别映射为 504，服务不可用映射为 503；当前客户端契约测试已覆盖；
8. 数据库不可用时 `/api/v1/health` 返回 503。

真实依赖必须提供以下接口，缺少任意一项均应作为联调阻断记录，不允许回退为跨 Schema 查询：

- user-service：`GET /internal/v1/users?id=...`、`relationships/blocked`、`memberships/status`、`relationships/following`；
- content-service：`GET /internal/v1/videos/:id`、`GET /internal/v1/videos/batch?ids=...`。

不得连接 `user_db` 或 `content_db` 执行测试。
