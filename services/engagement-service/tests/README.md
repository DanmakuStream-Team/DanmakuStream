# engagement-service tests

当前可直接执行的契约与单元测试位于各 `internal/*_test.go`，运行：

```bash
go test ./...
```

接入 `user-service`、`content-service` 与 `engagement_db` 后，API 集成回归至少覆盖：

1. 重复点赞/收藏切换后仅保留一条有效关系；
2. 评论、弹幕的非法内容与不可播放视频；
3. 播放进度幂等覆盖、历史/稍后再看清空；
4. 重复预约、取消预约和预约自己直播；
5. 礼物/Super Chat 事务一致性；
6. WebSocket 重连和在线人数；
7. user/content 超时分别映射为 504，服务不可用映射为 503；
8. 数据库不可用时 `/api/v1/health` 返回 503。

不得连接 `user_db` 或 `content_db` 执行测试。
