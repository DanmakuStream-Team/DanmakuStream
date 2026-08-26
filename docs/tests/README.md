# 自动化测试目录

测试组织参考 `origin/feature/uc13-tests`：每个用例分别维护单元、API 集成和 E2E 测试设计，使用统一追溯编号。

| 用例 | 测试设计 |
|---|---|
| UC01 | [UC01-test-cases.md](./UC01-test-cases.md) |
| UC07 | [UC07-test-cases.md](./UC07-test-cases.md) |
| UC08 | [UC08-test-cases.md](./UC08-test-cases.md) |
| UC11 | [UC11-test-cases.md](./UC11-test-cases.md) |

执行约束：

1. 测试数据必须带用例前缀和时间戳，执行前后可清理。
2. API 测试失败以非零状态退出，并保存原始输出。
3. E2E 测试使用 Playwright，优先 API 换 token，再注入浏览器会话。
4. 未执行不能写“通过”；缺陷需要关联 Issue，修复后保留回归报告。
