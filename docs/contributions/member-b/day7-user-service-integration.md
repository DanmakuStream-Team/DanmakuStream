# 成员 B 第 7 天：user-service 跨服务联调

## 对应用例

- UC01：用户身份和资料由 user-service 提供。
- UC07：关注、分组与双向屏蔽关系通过 user-service 内部接口提供。
- UC08：会员有效状态通过 user-service 内部接口提供。
- UC11：私信分享视频调用 content-service，不访问 `content_db`。
- UC13：用户摘要与角色仍归 user-service。

## 已实现链路

发送 `video_share` 私信时，user-service 调用：

```text
GET {CONTENT_SERVICE_URL}/internal/v1/videos/{videoId}
X-Internal-Token: <INTERNAL_API_TOKEN>
X-Request-ID: <入口请求编号>
```

仅当返回视频可播放时才写入 `chat_messages.shared_video_id`。会话和历史列表使用 `/internal/v1/videos/batch?ids=...` 一次性回填标题、封面和时长，避免 N+1 请求；content-service 暂时不可用时，读取链路降级为只返回已保存的视频 ID。

下游失败映射如下：

| 情况 | user-service 响应 |
| --- | --- |
| 视频不存在或不可播放 | 404 |
| 下游响应损坏或内部凭证被拒绝 | 502 |
| content-service 不可用 | 503 |
| 调用超过 `REQUEST_TIMEOUT` | 504 |

## 对外提供的内部接口

- `GET /internal/v1/users/:id`
- `GET /internal/v1/users?id=1&id=2`
- `GET /internal/v1/users/:id/exists`
- `GET /internal/v1/relationships/blocked?blockerId=1&blockedId=2`
- `GET /internal/v1/relationships/following?followerId=1&followeeId=2`
- `GET /internal/v1/memberships/status?userId=1&creatorId=2`

所有接口都要求 `X-Internal-Token`，并保留 `X-Request-ID`。

## 验证

```bash
cd services/user-service
go test -count=1 ./...
go build ./...
go vet ./...
```

真实 Schema 集成测试：

```bash
USER_SERVICE_TEST_DSN='<user_db DSN>' go test -tags=integration ./integration -count=1 -v
```

当前本机 Docker 引擎未运行，已有 MySQL 实例也不接受项目示例账号，因此真实 `user_db` 测试尚未取得绿色证据。测试代码已入库，必须在提供有效 `USER_SERVICE_TEST_DSN` 后执行。

## 联调依赖

content-service 需要实现以下受内部 Token 保护的契约：

- `GET /internal/v1/videos/:id`
- `GET /internal/v1/videos/batch?ids=1,2`

本次不修改 content-service 的业务代码；在上述接口合入前，视频分享写请求会按真实状态返回下游不可用或资源不存在，不会绕过校验直接写库。
