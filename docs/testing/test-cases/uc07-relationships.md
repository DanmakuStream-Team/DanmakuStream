# UC07 测试用例设计

执行日期：2026-08-29。汇总及最终复核见 `docs/testing/reports/UC07-UC13-completion-report-20260829.md`；原始输出和 HTML 报告保存在同目录。

## 单元测试（UNIT-TC07）

| 编号 | 测试目标 | 关键断言 | 当前状态 |
|---|---|---|---|
| UNIT-TC07-01 | 关注目标校验 | 不能关注自己，目标不存在返回规则错误 | 通过（自己：handler 单元；不存在：集成） |
| UNIT-TC07-02 | 关注关系状态转换 | 关注、取关后不产生重复有效关系 | 通过（MySQL 集成） |
| UNIT-TC07-03 | 分组参数校验 | 空名称、超长名称和重复名称被拒绝 | 通过（handler 单元 + MySQL 集成） |
| UNIT-TC07-04 | 屏蔽规则 | 自己不能被屏蔽，重复屏蔽可安全重试 | 通过（handler 单元 + MySQL 集成） |

## API 集成测试（INT-TC07）

| 编号 | 接口/场景 | 关键断言 | 当前状态 |
|---|---|---|---|
| INT-TC07-01 | POST `/users/:id/follow` | 关注/取关状态与关系表一致 | 通过 |
| INT-TC07-02 | POST `/users/:id/follow` 自己/不存在用户 | 返回 400/404，关系表不变 | 通过 |
| INT-TC07-03 | Follow group CRUD | 创建、修改、删除及计数结果一致 | 通过 |
| INT-TC07-04 | PUT `/users/:id/follow-settings` | 合法分组可绑定，非法分组返回 400 | 通过 |
| INT-TC07-05 | POST `/users/:id/block` | 屏蔽关系持久化，重复操作可解除且不重复计数 | 通过 |
| INT-TC07-06 | GET `/users/following` 与屏蔽列表 | 关注与屏蔽列表反映数据库状态 | 通过 |
| INT-TC07-07 | 关注创作者发布动态 | 关注者收到一条动态通知，未读数与内容正确 | 通过（MySQL 集成） |

## E2E 测试（E2E-TC07）

| 编号 | 场景 | 通过标准 | 当前状态 |
|---|---|---|---|
| E2E-TC07-01 | 关注创作者 | 关注后订阅页显示创作者 | 通过（`member-b-workflows.spec.ts`） |
| E2E-TC07-02 | 分组管理 | 新建并绑定分组后页面显示一致 | 通过（`member-b-workflows.spec.ts`） |
| E2E-TC07-03 | 屏蔽用户 | 屏蔽后黑名单可见，解除后状态恢复 | 通过（`member-b-workflows.spec.ts`） |
| E2E-TC07-04 | 内容发布通知 | 被关注创作者发布动态后，通知面板显示标题和动态正文 | 通过（`member-b-workflows.spec.ts`） |
