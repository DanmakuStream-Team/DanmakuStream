# UC01、UC07、UC08、UC11 测试报告（2026-08-27）

## 结果摘要

| 测试层级 | 结果 |
| --- | --- |
| MySQL/API 用例级测试 | 全部通过（含成功、备选、异常、存储故障、订阅到期与关系修复路径） |
| WebSocket/消息逻辑测试 | 全部通过，验证双端广播、断线清理、背压移除、重复收件人、旧消息格式兼容、校验失败与持久化 |
| Playwright E2E | 4/4 通过 |
| 前端类型检查与生产构建 | 通过 |
| UC01/07/08/11 语句覆盖率 | **90.0%（883/981）** |

## 覆盖率口径

四个用例的覆盖率按实际涉及的认证、资料、关注关系、通知、会员订阅、私信及聊天 WebSocket 函数统计；多个 Go 测试包产生的相同源码块会去重，并以联合覆盖结果计数。未经筛选的 Go 包级结果还包含 UC06 视频资料库、搜索、用户投稿列表等其他成员负责且本轮不在范围内的代码，因此不作为本轮验收指标。常驻过期扫描 worker 的无限循环启动代码不计入业务请求路径；其调用的到期处理函数已由 MySQL 集成测试覆盖。

## 证据

- 测试源文件：`backend/integration/user_domain_integration_test.go`
- WebSocket 测试：`backend/internal/logic/chat/hub_websocket_integration_test.go`
- 订阅到期测试：`backend/internal/handler/v1/membership/membership_integration_test.go`
- E2E 测试：`frontend/e2e/user-domain.spec.ts`
- E2E HTML 报告：`docs/testing/reports/user-domain-e2e/index.html`
- CI：`.github/workflows/ci.yml` 中的 UC01/07/08/11 integration、coverage 和 E2E gate

本地执行期间出现的 `record not found`、`table does not exist` 日志均来自主动构造的异常分支；对应测试断言已验证接口返回正确错误码，不属于测试失败。
