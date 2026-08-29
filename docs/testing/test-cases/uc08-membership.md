# UC08 测试用例设计

执行日期：2026-08-29。汇总及最终复核见 `docs/testing/reports/UC07-UC13-completion-report-20260829.md`；原始输出和 HTML 报告保存在同目录。

## 单元测试（UNIT-TC08）

| 编号 | 测试目标 | 关键断言 | 当前状态 |
|---|---|---|---|
| UNIT-TC08-01 | 会员方案金额校验 | 金额低于 100 或超过 100000 分拒绝 | 通过（handler 单元） |
| UNIT-TC08-02 | 订阅时长校验 | 仅 1、3、12 个月通过 | 通过（handler 单元 + 集成） |
| UNIT-TC08-03 | 订阅自己规则 | buyerId 等于 creatorId 时拒绝 | 通过（handler 单元 + 集成） |
| UNIT-TC08-04 | 支付状态转换 | pending→paid，重复支付不重复激活 | 通过（MySQL 集成） |
| UNIT-TC08-05 | 续期计算 | 有效订阅从原 expiresAt 累加；已过期订阅从当前时间重算 | 通过（`TestRenewalExpiry`） |

## API 集成测试（INT-TC08）

| 编号 | 接口/场景 | 关键断言 | 当前状态 |
|---|---|---|---|
| INT-TC08-01 | PUT `/creator/membership-plan` | creator 可保存方案，非法金额返回 400 | 通过 |
| INT-TC08-02 | GET `/creators/:id/membership-plan` | 返回方案和 enabled 状态 | 通过 |
| INT-TC08-03 | POST `/subscriptions/orders` | 有效方案生成 pending 订单，金额正确 | 通过（3 个月金额 1800 分） |
| INT-TC08-04 | 创建自己的订阅/无效月份 | 返回 400，不创建订单 | 通过 |
| INT-TC08-05 | POST `/subscriptions/orders/:orderNo/demo-pay` | 支付后订单 paid、订阅 active、特别关注生效 | 通过 |
| INT-TC08-06 | 重复演示支付 | 不重复创建订阅，返回相同有效状态 | 通过 |
| INT-TC08-07 | 跨订单续期、取消自动续费与到期识别 | 第二订单延长 expiresAt；自动续费关闭；到期后 status=expired 且特别关注解除 | 通过 |
| INT-TC08-08 | GET `/subscriptions` 与 `/subscriptions/orders` | 订单和订阅列表数据一致 | 通过 |

## E2E 测试（E2E-TC08）

| 编号 | 场景 | 通过标准 | 当前状态 |
|---|---|---|---|
| E2E-TC08-01 | 创作者配置会员方案 | 保存后用户侧能看到方案价格 | 通过（`member-b-workflows.spec.ts`） |
| E2E-TC08-02 | 用户购买会员 | 创建订单、演示支付后页面显示特别关注 | 通过（`member-b-workflows.spec.ts`） |
| E2E-TC08-03 | 刷新与重复支付 | 刷新状态保持，重复支付不产生重复订阅 | 通过（`member-b-workflows.spec.ts`） |
