# 测试说明

测试与被测包放在同一目录，以便覆盖包内边界并保持 `go test ./...` 可独立执行：

- `internal/handler/platform_test.go`：`livez`、`health`、`version`、内部 Token 和统一错误响应契约。
- `internal/handler/v1/**/**_test.go`：UC07、UC08、UC11 的接口与异常分支。
- `internal/logic/**/**_test.go`：UC01 鉴权逻辑和 UC11 会话逻辑。
- `internal/middleware/auth_test.go`：未认证与权限不足场景。

第 7 天新增 `internal/client/content_test.go`，验证内部 Token、Request ID、单个/批量视频摘要契约，以及 404、502、503、504 所需的错误分类。`integration/day7_integration_test.go` 使用 `USER_SERVICE_TEST_DSN` 连接真实 `user_db`，验证“内容接口校验 → 仅保存视频 ID → 返回视频及作者摘要”的完整链路；未配置真实数据库时明确跳过，不伪造通过状态。
