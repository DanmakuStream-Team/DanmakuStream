# 成员 D 交付索引

## 1. 负责范围

| 用例 | 业务场景 | 当前交付 |
| --- | --- | --- |
| UC05 | 视频观看互动 | 弹幕、评论、视频点赞与收藏的需求、设计、代码定位和测试计划 |
| UC09 | 直播预约与用户预约 | 直播计划创建/取消、用户预约/取消及到时启动的需求、设计、代码定位和测试计划 |
| UC10 | 直播发布、观看与实时互动 | 直播工作台、观看、弹幕、点赞、礼物、Super Chat、监控和重连的需求、设计、代码定位和测试计划 |

直播回放不在成员 D 的验收范围内。

## 2. 文档入口

- [需求说明书](./requirements.md)：需求编号、用例图、用例说明、概念类图，以及 UC05、UC09、UC10 各 1 张系统级顺序图。
- [概要设计说明书](./high-level-design.md)：组件图，以及三个用例各 1 张组件级顺序图。
- [详细设计说明书](./detailed-design.md)：实现类图，以及三个用例各 1 张对象级顺序图。
- [追溯表](./traceability.md)：需求、用例、三层模型、代码模块、测试编号和实际结果的统一追溯。
- [验收测试记录](./acceptance-test-report.md)：UC05、UC09、UC10 的自动化测试方法、执行环境和实际结果。
- [完整业务场景清单](../../project/use-case-catalog.md)：项目最终验收用例清单及详细用例说明。
- [PlantUML 模型索引](../../models/README.md)：正式图 `.puml` 源文件及 SVG/PNG 导出图。
- [部署设计说明](../../architecture/deployment-design.md)：改造前单体部署图和改造后 Kubernetes 部署图。

当前互动与直播域共交付 13 张图：用例图、概念类图、组件图和实现类图各 1 张，三类顺序图各 3 张；另有 2 张项目级部署图。15 张正式图均提供 PlantUML 源文件及 SVG/PNG。

## 3. 代码入口

| 范围 | 后端 | 前端 |
| --- | --- | --- |
| UC05 | [`danmaku_handler.go`](../../../backend/internal/handler/v1/danmaku/danmaku_handler.go)、[`comment_handler.go`](../../../backend/internal/handler/v1/comment/comment_handler.go)、[`video_handler.go`](../../../backend/internal/handler/v1/video/video_handler.go) | [`VideoDetailPage.vue`](../../../frontend/src/pages/video/VideoDetailPage.vue)、[`VideoPlayer.vue`](../../../frontend/src/components/common/VideoPlayer.vue) |
| UC09 | [`schedule_handler.go`](../../../backend/internal/handler/v1/live/schedule_handler.go) | [`LiveListPage.vue`](../../../frontend/src/pages/live/LiveListPage.vue)、[`live.ts`](../../../frontend/src/api/live.ts) |
| UC10 | [`live_handler.go`](../../../backend/internal/handler/v1/live/live_handler.go)、[`interaction_handler.go`](../../../backend/internal/handler/v1/live/interaction_handler.go)、[`ws_handler.go`](../../../backend/internal/handler/ws/ws_handler.go)、[`hub.go`](../../../backend/internal/logic/danmaku/hub.go) | [`LiveStudioPage.vue`](../../../frontend/src/pages/live/LiveStudioPage.vue)、[`LiveRoomPage.vue`](../../../frontend/src/pages/live/LiveRoomPage.vue)、[`danmaku.ts`](../../../frontend/src/api/danmaku.ts) |

## 4. 测试与验证

- [`interaction_handler_test.go`](../../../backend/internal/handler/v1/live/interaction_handler_test.go)：覆盖 Super Chat 展示时长边界和直播热度计算。
- [`schedule_handler_test.go`](../../../backend/internal/handler/v1/live/schedule_handler_test.go)：覆盖预约时间解析和预约状态校验。
- [`hub_test.go`](../../../backend/internal/logic/danmaku/hub_test.go)：覆盖登录用户在线人数去重、监控连接排除和弹幕持久化错误传播。
- [`schedule_worker_integration_test.go`](../../../backend/internal/handler/v1/live/schedule_worker_integration_test.go)：覆盖到期计划重复扫描幂等、通知与主播统计。
- [`hub_chat_integration_test.go`](../../../backend/internal/logic/danmaku/hub_chat_integration_test.go)：覆盖聊天身份模式、会员有效期、慢速模式及主播绕过。
- [`hub_websocket_integration_test.go`](../../../backend/internal/logic/danmaku/hub_websocket_integration_test.go)：覆盖真实 WebSocket 连接、事件往返、在线人数和持久化失败错误事件。
- [`engagement_integration_test.go`](../../../backend/integration/engagement_integration_test.go)：使用自动创建并删除的临时 MySQL 数据库，覆盖 UC05、UC09、UC10 的查询与写入、权限、参数边界、事务、通知、状态回读及结束后重启。
- [`d-engagement.spec.ts`](../../../frontend/e2e/d-engagement.spec.ts)：Playwright 覆盖 UC09 预约闭环和 UC10 直播创建、互动、结束闭环；HTML 报告见 [`index.html`](../../testing/reports/engagement-e2e/index.html)。
- 后端全量验证：`go test ./...` 已通过。
- 三用例及直播机制集成验证：`go test -tags=integration ./internal/handler/v1/live ./internal/logic/danmaku ./integration -v` 已通过。
- 浏览器 E2E：`npm run test:e2e:d`，Chromium 2/2 通过。
- 前端生产验证：`npm run build` 已通过；仅有既有的包体积提示，无编译错误。
- 真实媒体链路：FFmpeg 推送 30 秒测试流至 SRS，HLS 主清单、媒体清单和 TS 分片均返回 HTTP 200。
- 覆盖率：成员 D 用例函数口径 90.3%（924/1023），相关文件保守口径 58.1%（940/1618）。

## 5. 当前基线与状态

- 工作分支：`feature/floatingsoul423-live-features-20260825`。
- 开发基线：已合并 `dev` 的 `ea9a5ee`。
- 基线合并提交：`81359ed`。
- 文档、单元测试、三个用例的 MySQL/API 集成测试、真实 WebSocket 测试和两条 Playwright E2E：已完成。
- 已处理：同主播相同开播时间冲突、登录观众多连接去重、直播弹幕写入失败不广播并返回错误、UC05/UC09/UC10 集成测试证据。
- 后续增强：UC05 浏览器视觉演示、WebSocket 断网重连现场测试，以及异常推流状态恢复测试。
- `engagement-service` 仍是计划中的拆分目标；当前设计和代码追溯以可运行的单体实现为准。
