# DanmakuStream 总追溯表（UC01～UC13）

> 本表按课程任务书要求，把**需求、用例、三层模型、代码模块、测试编号与测试结果**汇总到一张表。
> 填写规则继承 `docs/traceability/uc13.md`："结果"一栏只能依据实际测试报告填写；缺图或未测的格子如实留空/标 ❌，不允许按代码存在就写通过。
> 状态统计（2026-08-29，本地回归）：三层图 **39/39 全部完成**；全层测试通过 **10/13**；部分通过 **3/13**；未测 **0/13**。UC01/UC06 的远程状态待本 PR 的 CI 复验。

| 需求 | 用例 | 系统级图 | 组件级图 | 对象级图 | 主要代码模块 | 单元测试 | API/集成测试 | E2E 测试 | 结果 | 缺口（负责人） |
|---|---|---|---|---|---|---|---|---|---|---|
| REQ01 | UC01 注册/登录/资料维护 | ✅ | ✅ | ✅ | `auth/auth_handler.go`、`user/user_handler.go`、`LoginPage.vue`、`RegisterPage.vue`、`UserProfilePage.vue` | UNIT-TC01-01～03（`auth_logic_test.go`） | `member_b_integration_test.go`（注册/重复注册/错误密码/无效 token/资料更新） | E2E-TC01-01～05（`uc01-user.spec.ts`，5/5） | 通过（2026-08-29 本地） | 等待本 PR CI 复验 |
| REQ02 | UC02 视频发现/搜索/播放 | ✅ | ✅ | ✅ | `video/video_handler.go`、`logic/video/list_logic.go`、`detail_logic.go`、`HomePage.vue`、`VideoDetailPage.vue` | UNIT-TC02-01/02（20260828 全量通过） | INT-TC02-01～04（20260828，57 项套件全绿） | E2E-TC02-01（20260828） | 通过（非法排序与媒体缺失已验证） | —（GAP 已由 PR #91 修复合入并 CI 验证） |
| REQ03 | UC03 创作者投稿与状态跟踪 | ✅ | ✅ | ✅ | `media/`、`video` 上传与转码逻辑、`CreatorDashboardPage.vue` | UNIT-TC03-01～03（20260828） | INT-TC03-01～04（含 failed 状态与原因） | E2E-TC03-01（失败状态可见） | 通过 | —（GAP 已由 PR #91 修复合入并 CI 验证） |
| REQ04 | UC04 视频审核与发布 | ✅ | ✅ | ✅ | `video/video_handler.go`（`AdminUpdateStatus`）、审核后台页面 | UNIT-TC04-01～03 + UNIT-TC13-06 | INT-TC04-01～04 + INT-TC13-02/03 | E2E-TC04-01 + E2E-TC13-01 | 通过（终态保护与媒体检查已验证） | —（GAP 已由 PR #91 修复合入并 CI 验证） |
| REQ05 | UC05 视频观看互动 | ✅ | ✅ | ✅ | `danmaku/danmaku_handler.go`、`comment/`、`VideoDetailPage.vue`、`VideoPlayer.vue` | UNIT-TC01（`logic/danmaku/hub_test.go` 等） | INT-TC01/02（`integration/engagement_integration_test.go`） | `frontend/e2e/d-engagement.spec.ts` | 通过（报告见 `docs/testing/reports/engagement-e2e/`） | 断网重连现场证据（D，增强） |
| REQ06 | UC06 个人视频资料库管理 | ✅ | ✅ | ✅ | `collection/`、`user/`（历史/进度/稍后再看）、`UserLibraryPage.vue` | UNIT-TC06-01～05 | INT-TC06-01～12（30/30） | E2E-TC06-01～04（4/4） | 通过（2026-08-29 本地） | 等待本 PR CI 复验 |
| REQ07 | UC07 关注关系与内容通知 | ✅ | ✅ | ✅ | `user`（关注/分组/特别关注/屏蔽）、`notification/` | — | `member_b_integration_test.go`（关注/分组/绑定/屏蔽/幂等，PR #74，CI 绿） | — | 部分通过（集成层） | 单元与 E2E 层补齐（B），设计见 `docs/testing/test-cases/uc07-relationships.md` |
| REQ08 | UC08 创作者会员订阅 | ✅ | ✅ | ✅ | `membership/`（方案/订单/演示支付/订阅）、`SubscriptionPage.vue` | — | `member_b_integration_test.go`（方案/下单/支付幂等/自动特别关注，PR #74，CI 绿） | — | 部分通过（集成层） | 单元与 E2E 层补齐（B），设计见 `docs/testing/test-cases/uc08-membership.md` |
| REQ09 | UC09 直播预约与用户预约 | ✅ | ✅ | ✅ | `live/schedule_handler.go`、`live/schedule_worker`、`LiveListPage.vue` | UNIT-TC03/04（`schedule_handler_test.go`、`schedule_worker_integration_test.go`） | INT-TC03/04 | E2E-TC09（`d-engagement.spec.ts`） | 通过 | — |
| REQ10 | UC10 直播发布/观看/实时互动 | ✅ | ✅ | ✅ | `live/live_handler.go`、`live/interaction_handler.go`、`ws_handler.go`、`logic/chat/hub.go`、`LiveRoomPage.vue`、`LiveStudioPage.vue` | UNIT-TC01/02/05（`interaction_handler_test.go`、`chat/hub_test.go`） | INT-TC06/07/09（含 RTMP-SRS-HLS 媒体链路） | E2E-TC10/E2E-MEDIA | 通过 | 异常推流恢复测试（D，增强） |
| REQ11 | UC11 用户私信与媒体分享 | ✅ | ✅ | ✅ | `message/`、`ChatPage.vue`（WebSocket） | — | `member_b_integration_test.go`（发送/未读计数/已读标记/自发拒绝，PR #74，CI 绿） | — | 部分通过（集成层） | 单元与 E2E 层补齐（待认领），设计见 `docs/testing/test-cases/uc11-messaging.md` |
| REQ12 | UC12 创作者数据分析 | ✅ | ✅ | ✅ | `creator/`（汇总/趋势/权限）、`CreatorDashboardPage.vue` | UNIT-TC12-01/02（20260828） | INT-TC12-01～04（单作品排行收窄） | E2E-TC12-01（20260828） | 通过 | —（GAP 已由 PR #91 修复合入） |
| REQ13 | UC13 平台审核/权限/运营/基础设施 | ✅ | ✅ | ✅ | `middleware/auth.go`、`admin/admin_handler.go`、后台六个 Vue 页面 | UNIT-TC13-01～06 | INT-TC13-01～10（`tests/api/uc13-admin-test.sh`） | E2E-TC13-01～05（`uc13-admin.spec.ts`） | 通过（`docs/testing/test-cases/uc13-admin.md` §6–9） | — |

## 汇总统计（截至 2026-08-29，本地回归）

| 维度 | 数量 | 明细 |
|---|---|---|
| 三层模型 | 39/39 全部完成 | UC01/07/08/11 于 PR #66、UC06 于 PR #68 补齐（含总用例图 USECASE-OVERVIEW） |
| 单元测试 | 9/13 | 新增 UC01、UC06；其余为 UC02/03/05/09/10/12/13（UC04 共享 UC13 链路） |
| API/集成测试 | 13/13 | UC06 新增真实 MySQL API 套件 30/30；任一失败阻断镜像构建 |
| E2E 测试 | 10/13 | 新增 UC01 5/5、UC06 4/4；其余为 UC02/03/04/12、UC05/09/10、UC13 |
| 全层通过 | 10/13 | UC01/02/03/04/05/06/09/10/12/13 |
| 部分通过 | 3/13 | UC07/08/11（集成层已过，单元/E2E 待补） |
| 未测 | 0/13 | — |

## 维护规则

1. 每次补齐一张图或一项测试后，同一次 PR 内更新本表与"结果"，并在 PR 描述中注明对应 UC 编号。
2. "结果"必须能链接到 `docs/testing/reports/` 下的原始报告或 CI 运行记录。
3. 用例最终验收范围以教师确认的 `docs/project/use-case-catalog.md` 为准；范围变更需同步本表。
