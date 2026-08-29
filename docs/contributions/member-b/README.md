# 成员 B 交付索引

成员 B（用户域）负责 UC01 注册/登录/资料维护、UC07 关注关系与内容通知、UC08 创作者会员订阅、UC11 用户私信与媒体分享。

## 文档入口

- [四用例顺序图索引](./sequence-diagrams-index.md)：指向正式 PlantUML 图（USECASE-B/CLASS-B/COMPONENT-B 及各层顺序图）
- [UC01 用户服务设计与测试](./uc01-user-service-design-and-test.md)：设计与测试说明

## 其他归属

- 用例说明：`docs/project/use-case-catalog.md` §UC01/07/08/11
- 测试设计：`docs/testing/test-cases/uc01-user-service.md`、`uc07-relationships.md`、`uc08-membership.md`、`uc11-messaging.md`
- 集成测试代码：`backend/integration/member_b_integration_test.go`（PR #74，CI job `member-b-integration`）
- 追溯：`docs/traceability/master.md` 与 `uc01/07/08/11.md`
