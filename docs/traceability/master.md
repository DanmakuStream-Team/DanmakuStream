# DanmakuStream 总追溯表（UC01～UC13）

> 本表按课程任务书要求，把**需求、用例、三层模型、代码模块、测试编号与测试结果**汇总到一张表。
> 填写规则继承 `docs/traceability/uc13.md`："结果"一栏只能依据实际测试报告填写；缺图或未测的格子如实留空/标 ❌，不允许按代码存在就写通过。
> 状态统计（2026-08-27 第二次更新，UC06 图随 PR #68 合入）：三层图 **39/39 全部完成**；三层测试齐备且通过 **4/13**（UC05/09/10/13）；部分通过 1/13（UC04，与 UC13 共享审核链路）；未测 8/13（UC01/07/08/11 已有测试设计文档且集成测试代码已提交（PR #74），待合入执行）。

| 需求 | 用例 | 系统级图 | 组件级图 | 对象级图 | 主要代码模块 | 单元测试 | API/集成测试 | E2E 测试 | 结果 | 缺口（负责人） |
|---|---|---|---|---|---|---|---|---|---|---|
| REQ01 | UC01 注册/登录/资料维护 | ✅ | ✅ | ✅ | `auth/auth_handler.go`、`user/user_handler.go`、`LoginPage.vue`、`RegisterPage.vue`、`UserProfilePage.vue` | — | — | — | 未测（测试设计已入库） | 注册重名/密码错误/身份过期等异常测试（B），见 `docs/testing/test-cases/uc01-user-service.md` |
| REQ02 | UC02 视频发现/搜索/播放 | ✅ | ✅ | ✅ | `video/video_handler.go`、`logic/video/list_logic.go`、`detail_logic.go`、`HomePage.vue`、`VideoDetailPage.vue` | — | — | — | 未测 | 搜索命中/空结果、私有或不存在视频播放异常测试（C） |
| REQ03 | UC03 创作者投稿与状态跟踪 | ✅ | ✅ | ✅ | `media/`、`video` 上传与转码逻辑、`CreatorDashboardPage.vue` | — | — | — | 未测 | 上传中止、转码失败与恢复、状态流转测试（C） |
| REQ04 | UC04 视频审核与发布 | ✅ | ✅ | ✅ | `video/video_handler.go`（`AdminUpdateStatus`）、审核后台页面 | UNIT-TC13-06 | INT-TC13-02/03 | E2E-TC13-01 | 部分（复用 UC13 审核链路） | 专属用例说明级断言与独立报告（C） |
| REQ05 | UC05 视频观看互动 | ✅ | ✅ | ✅ | `danmaku/danmaku_handler.go`、`comment/`、`VideoDetailPage.vue`、`VideoPlayer.vue` | UNIT-TC01（`logic/danmaku/hub_test.go` 等） | INT-TC01/02（`integration/engagement_integration_test.go`） | `frontend/e2e/d-engagement.spec.ts` | 通过（报告见 `docs/testing/reports/engagement-e2e/`） | 断网重连现场证据（D，增强） |
| REQ06 | UC06 个人视频资料库管理 | ✅ | ✅ | ✅ | `collection/`、`user/`（历史/进度/稍后再看）、`UserLibraryPage.vue` | — | — | — | 未测（测试设计已入库） | 服务端历史/进度同步测试（E），见 `docs/testing/test-cases/uc06-personal-library.md`；图已随 PR #68 合入（Closes #73） |
| REQ07 | UC07 关注关系与内容通知 | ✅ | ✅ | ✅ | `user`（关注/分组/特别关注/屏蔽）、`notification/` | — | — | — | 未测（测试设计已入库） | 分组、特别关注、屏蔽规则测试（B），见 `docs/testing/test-cases/uc07-relationships.md` |
| REQ08 | UC08 创作者会员订阅 | ✅ | ✅ | ✅ | `membership/`（方案/订单/演示支付/订阅）、`SubscriptionPage.vue` | — | — | — | 未测（测试设计已入库） | 会员方案、订单状态机、订阅权限测试（B），见 `docs/testing/test-cases/uc08-membership.md` |
| REQ09 | UC09 直播预约与用户预约 | ✅ | ✅ | ✅ | `live/schedule_handler.go`、`live/schedule_worker`、`LiveListPage.vue` | UNIT-TC03/04（`schedule_handler_test.go`、`schedule_worker_integration_test.go`） | INT-TC03/04 | E2E-TC09（`d-engagement.spec.ts`） | 通过 | — |
| REQ10 | UC10 直播发布/观看/实时互动 | ✅ | ✅ | ✅ | `live/live_handler.go`、`live/interaction_handler.go`、`ws_handler.go`、`logic/chat/hub.go`、`LiveRoomPage.vue`、`LiveStudioPage.vue` | UNIT-TC01/02/05（`interaction_handler_test.go`、`chat/hub_test.go`） | INT-TC06/07/09（含 RTMP-SRS-HLS 媒体链路） | E2E-TC10/E2E-MEDIA | 通过 | 异常推流恢复测试（D，增强） |
| REQ11 | UC11 用户私信与媒体分享 | ✅ | ✅ | ✅ | `message/`、`ChatPage.vue`（WebSocket） | — | — | — | 未测（测试设计已入库） | 私信投递/离线消息/媒体分享测试（待认领），见 `docs/testing/test-cases/uc11-messaging.md` |
| REQ12 | UC12 创作者数据分析 | ✅ | ✅ | ✅ | `creator/`（汇总/趋势/权限）、`CreatorDashboardPage.vue` | — | — | — | 未测 | 汇总口径、趋势计算、越权访问测试（C） |
| REQ13 | UC13 平台审核/权限/运营/基础设施 | ✅ | ✅ | ✅ | `middleware/auth.go`、`admin/admin_handler.go`、后台六个 Vue 页面 | UNIT-TC13-01～06 | INT-TC13-01～10（`tests/api/uc13-admin-test.sh`） | E2E-TC13-01～05（`uc13-admin.spec.ts`） | 通过（`docs/testing/test-cases/uc13-admin.md` §6–9） | — |

## 汇总统计（截至 2026-08-26）

| 维度 | 数量 | 明细 |
|---|---|---|
| 三层模型 | 39/39 全部完成 | UC01/07/08/11 于 PR #66、UC06 于 PR #68 补齐（含总用例图 USECASE-OVERVIEW） |
| 单元测试 | 5/13 | UC05、UC09、UC10、UC13、UC04（共享）；测试文件 12 个 `*_test.go` |
| API/集成测试 | 4/13 | UC05、UC09、UC10、UC13；脚本+Go 集成测试 |
| E2E 测试 | 3/13 | UC05、UC09、UC10（`d-engagement.spec.ts`）、UC13（`uc13-admin.spec.ts`） |
| 全层通过 | 4/13 | UC05、UC09、UC10、UC13 |

## 维护规则

1. 每次补齐一张图或一项测试后，同一次 PR 内更新本表与"结果"，并在 PR 描述中注明对应 UC 编号。
2. "结果"必须能链接到 `docs/testing/reports/` 下的原始报告或 CI 运行记录。
3. 用例最终验收范围以教师确认的 `docs/project/use-case-catalog.md` 为准；范围变更需同步本表。
