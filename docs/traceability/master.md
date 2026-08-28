# DanmakuStream 总追溯表（UC01～UC13）

> 本表按课程任务书要求，把**需求、用例、三层模型、代码模块、测试编号与测试结果**汇总到一张表。
> 填写规则继承 `docs/traceability/uc13.md`："结果"一栏只能依据实际测试报告填写；缺图或未测的格子如实留空/标 ❌，不允许按代码存在就写通过。
> 状态统计（2026-08-27，PR #74/#75 合入后）：三层图 **39/39 全部完成**；全层测试通过 **8/13**（UC02/03/04/12 含 GAP，Issue #80 跟踪）；部分通过 **4/13**（UC01/07/08/11，集成层已过）；未测 **1/13**（UC06，设计已入库）。

| 需求 | 用例 | 系统级图 | 组件级图 | 对象级图 | 主要代码模块 | 单元测试 | API/集成测试 | E2E 测试 | 结果 | 缺口（负责人） |
|---|---|---|---|---|---|---|---|---|---|---|
| REQ01 | UC01 注册/登录/资料维护 | ✅ | ✅ | ✅ | `auth/auth_handler.go`、`user/user_handler.go`、`LoginPage.vue`、`RegisterPage.vue`、`UserProfilePage.vue` | — | `member_b_integration_test.go`（注册/重复注册/错误密码/无效 token/资料更新，PR #74，CI 绿） | — | 部分通过（集成层） | 单元与 E2E 层补齐（B），设计见 `docs/testing/test-cases/uc01-user-service.md` |
| REQ02 | UC02 视频发现/搜索/播放 | ✅ | ✅ | ✅ | `video/video_handler.go`、`logic/video/list_logic.go`、`detail_logic.go`、`HomePage.vue`、`VideoDetailPage.vue` | `list_logic_test.go`（PR #75） | `member-c-content-test.sh`（PR #75，46 断言通过） | `member-c-content.spec.ts`（PR #75，CI 绿） | 通过（GAP1/2 见 Issue #80） | GAP 修复（C，#80） |
| REQ03 | UC03 创作者投稿与状态跟踪 | ✅ | ✅ | ✅ | `media/`、`video` 上传与转码逻辑、`CreatorDashboardPage.vue` | `video_handler_test.go` 扩展（PR #75） | `member-c-content-test.sh`（上传/取消/转码路径） | `member-c-content.spec.ts` | 通过（GAP3 见 Issue #80） | 转码失败状态回写（C，#80） |
| REQ04 | UC04 视频审核与发布 | ✅ | ✅ | ✅ | `video/video_handler.go`（`AdminUpdateStatus`）、审核后台页面 | UNIT-TC13-06 | INT-TC13-02/03 + `member-c-content-test.sh` | E2E-TC13-01 + `member-c-content.spec.ts` | 通过（GAP4 见 Issue #80） | 审核终态保护（C，#80） |
| REQ05 | UC05 视频观看互动 | ✅ | ✅ | ✅ | `danmaku/danmaku_handler.go`、`comment/`、`VideoDetailPage.vue`、`VideoPlayer.vue` | UNIT-TC01（`logic/danmaku/hub_test.go` 等） | INT-TC01/02（`integration/engagement_integration_test.go`） | `frontend/e2e/d-engagement.spec.ts` | 通过（报告见 `docs/testing/reports/engagement-e2e/`） | 断网重连现场证据（D，增强） |
| REQ06 | UC06 个人视频资料库管理 | ✅ | ✅ | ✅ | `collection/`、`user/`（历史/进度/稍后再看）、`UserLibraryPage.vue` | — | — | — | 未测（测试设计已入库） | 服务端历史/进度同步测试（E），见 `docs/testing/test-cases/uc06-personal-library.md`；图已随 PR #68 合入（Closes #73） |
| REQ07 | UC07 关注关系与内容通知 | ✅ | ✅ | ✅ | `user`（关注/分组/特别关注/屏蔽）、`notification/` | — | `member_b_integration_test.go`（关注/分组/绑定/屏蔽/幂等，PR #74，CI 绿） | — | 部分通过（集成层） | 单元与 E2E 层补齐（B），设计见 `docs/testing/test-cases/uc07-relationships.md` |
| REQ08 | UC08 创作者会员订阅 | ✅ | ✅ | ✅ | `membership/`（方案/订单/演示支付/订阅）、`SubscriptionPage.vue` | — | `member_b_integration_test.go`（方案/下单/支付幂等/自动特别关注，PR #74，CI 绿） | — | 部分通过（集成层） | 单元与 E2E 层补齐（B），设计见 `docs/testing/test-cases/uc08-membership.md` |
| REQ09 | UC09 直播预约与用户预约 | ✅ | ✅ | ✅ | `live/schedule_handler.go`、`live/schedule_worker`、`LiveListPage.vue` | UNIT-TC03/04（`schedule_handler_test.go`、`schedule_worker_integration_test.go`） | INT-TC03/04 | E2E-TC09（`d-engagement.spec.ts`） | 通过 | — |
| REQ10 | UC10 直播发布/观看/实时互动 | ✅ | ✅ | ✅ | `live/live_handler.go`、`live/interaction_handler.go`、`ws_handler.go`、`logic/chat/hub.go`、`LiveRoomPage.vue`、`LiveStudioPage.vue` | UNIT-TC01/02/05（`interaction_handler_test.go`、`chat/hub_test.go`） | INT-TC06/07/09（含 RTMP-SRS-HLS 媒体链路） | E2E-TC10/E2E-MEDIA | 通过 | 异常推流恢复测试（D，增强） |
| REQ11 | UC11 用户私信与媒体分享 | ✅ | ✅ | ✅ | `message/`、`ChatPage.vue`（WebSocket） | — | `member_b_integration_test.go`（发送/未读计数/已读标记/自发拒绝，PR #74，CI 绿） | — | 部分通过（集成层） | 单元与 E2E 层补齐（待认领），设计见 `docs/testing/test-cases/uc11-messaging.md` |
| REQ12 | UC12 创作者数据分析 | ✅ | ✅ | ✅ | `creator/`（汇总/趋势/权限）、`CreatorDashboardPage.vue` | `analytics_handler_test.go`（鉴权与参数校验，PR #75） | `member-c-content-test.sh` | `member-c-content.spec.ts` | 通过（GAP5 见 Issue #80） | topVideos 口径修复（C，#80） |
| REQ13 | UC13 平台审核/权限/运营/基础设施 | ✅ | ✅ | ✅ | `middleware/auth.go`、`admin/admin_handler.go`、后台六个 Vue 页面 | UNIT-TC13-01～06 | INT-TC13-01～10（`tests/api/uc13-admin-test.sh`） | E2E-TC13-01～05（`uc13-admin.spec.ts`） | 通过（`docs/testing/test-cases/uc13-admin.md` §6–9） | — |

## 汇总统计（截至 2026-08-27，PR #74/#75 合入后）

| 维度 | 数量 | 明细 |
|---|---|---|
| 三层模型 | 39/39 全部完成 | UC01/07/08/11 于 PR #66、UC06 于 PR #68 补齐（含总用例图 USECASE-OVERVIEW） |
| 单元测试 | 7/13 | UC02、UC03、UC05、UC09、UC10、UC12、UC13（UC04 共享 UC13 链路） |
| API/集成测试 | 12/13 | 仅缺 UC06；成员 B（#74）与成员 C（#75）接入 CI，任一失败阻断镜像构建 |
| E2E 测试 | 6/13 | UC02/03/04/12（#75）、UC05/09/10（d-engagement）、UC13（uc13-admin） |
| 全层通过 | 8/13 | UC02/03/04/05/09/10/12/13（UC02/03/04/12 含 GAP，Issue #80 跟踪） |
| 部分通过 | 4/13 | UC01/07/08/11（集成层已过，单元/E2E 待补） |
| 未测 | 1/13 | UC06（测试设计已入库待实现） |

## 维护规则

1. 每次补齐一张图或一项测试后，同一次 PR 内更新本表与"结果"，并在 PR 描述中注明对应 UC 编号。
2. "结果"必须能链接到 `docs/testing/reports/` 下的原始报告或 CI 运行记录。
3. 用例最终验收范围以教师确认的 `docs/project/use-case-catalog.md` 为准；范围变更需同步本表。
