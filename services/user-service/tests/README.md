# 测试说明

测试与被测包放在同一目录，以便覆盖包内边界并保持 `go test ./...` 可独立执行：

- `internal/handler/platform_test.go`：`livez`、`health`、`version`、内部 Token 和统一错误响应契约。
- `internal/handler/v1/**/**_test.go`：UC07、UC08、UC11 的接口与异常分支。
- `internal/logic/**/**_test.go`：UC01 鉴权逻辑和 UC11 会话逻辑。
- `internal/middleware/auth_test.go`：未认证与权限不足场景。

第 6 天完成独立单元/契约测试；使用真实 `user_db` 的完整 API 集成验证和跨服务 E2E 按计划继续执行，不伪造通过状态。
