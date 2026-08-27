# 自动化测试目录

测试组织参考 `origin/feature/uc13-tests`：每个用例分别维护单元、API 集成和 E2E 测试设计，使用统一追溯编号。

| 用例 | 测试设计 |
|---|---|
| UC01 | [uc01-user-service.md](./test-cases/uc01-user-service.md) |
| UC06 | [uc06-personal-library.md](./test-cases/uc06-personal-library.md) |
| UC07 | [uc07-relationships.md](./test-cases/uc07-relationships.md) |
| UC08 | [uc08-membership.md](./test-cases/uc08-membership.md) |
| UC11 | [uc11-messaging.md](./test-cases/uc11-messaging.md) |

执行约束：

1. 测试数据必须带用例前缀和时间戳，执行前后可清理。
2. API 测试失败以非零状态退出，并保存原始输出。
3. E2E 测试使用 Playwright，优先 API 换 token，再注入浏览器会话。
4. 未执行不能写“通过”；缺陷需要关联 Issue，修复后保留回归报告。
