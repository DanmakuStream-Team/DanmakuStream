# DanmakuStream 总追溯表（UC01～UC13）

> 本表按课程任务书要求，把**需求、用例、三层模型、代码模块、测试编号与测试结果**汇总到一张表。
> 填写规则继承 `docs/traceability/uc13.md`："结果"一栏只能依据实际测试报告填写；缺图或未测的格子如实留空/标 ❌，不允许按代码存在就写通过。
> 状态统计（2026-08-29，合并 PR #99 与 PR #100 的测试证据）：三层图 **39/39 全部完成**；UC01～UC13 均具备单元、API/集成与 E2E 覆盖。统一 CI 会在镜像构建前执行全部测试，任一失败即阻断后续构建与部署。

| 需求 | 用例 | 系统级图 | 组件级图 | 对象级图 | 主要代码模块 | 单元测试 | API/集成测试 | E2E 测试 | 结果 | 缺口（负责人） |
|---|---|---|---|---|---|---|---|---|---|---|
| REQ01 | UC01 注册/登录/资料维护 | ✅ | ✅ | ✅ | `auth/auth_handler.go`、`user/user_handler.go`、`LoginPage.vue`、`RegisterPage.vue`、`UserProfilePage.vue` | UNIT-TC01-01～03（`auth_logic_test.go`） | `member_b_integration_test.go`（注册/重复注册/错误密码/无效 token/资料更新） | E2E-TC01-01～05（`uc01-user.spec.ts`，5/5） | 通过（2026-08-29 回归） | — |
| REQ02 | UC02 视频发现/搜索/播放 | ✅ | ✅ | ✅ | `video/video_handler.go`、`logic/video/list_logic.go`、`detail_logic.go`、`HomePage.vue`、`VideoDetailPage.vue` | UNIT-TC02-01/02（20260828 全量通过） | INT-TC02-01～04（20260828，57 项套件全绿） | E2E-TC02-01（20260828） | 通过（非法排序与媒体缺失已验证） | —（Issue #80 修复已合入 `dev`） |
| REQ03 | UC03 创作者投稿与状态跟踪 | ✅ | ✅ | ✅ | `media/`、`video` 上传与转码逻辑、`CreatorDashboardPage.vue` | UNIT-TC03-01～03（20260828） | INT-TC03-01～04（含 failed 状态与原因） | E2E-TC03-01（失败状态可见） | 通过 | —（Issue #80 修复已合入 `dev`） |
| REQ04 | UC04 视频审核与发布 | ✅ | ✅ | ✅ | `video/video_handler.go`（`AdminUpdateStatus`）、审核后台页面 | UNIT-TC04-01～03 + UNIT-TC13-06 | INT-TC04-01～04 + INT-TC13-02/03 | E2E-TC04-01 + E2E-TC13-01 | 通过（终态保护与媒体检查已验证） | —（Issue #80 修复已合入 `dev`） |
| REQ05 | UC05 视频观看互动 | ✅ | ✅ | ✅ | `danmaku/danmaku_handler.go`、`comment/`、`VideoDetailPage.vue`、`VideoPlayer.vue` | UNIT-TC01（`logic/danmaku/hub_test.go` 等） | INT-TC01/02（`integration/engagement_integration_test.go`） | `frontend/e2e/d-engagement.spec.ts` | 通过（报告见 `docs/testing/reports/engagement-e2e/`） | 断网重连现场证据（D，增强） |
| REQ06 | UC06 个人视频资料库管理 | ✅ | ✅ | ✅ | `collection/`、`user/`（历史/进度/稍后再看）、`UserLibraryPage.vue` | UNIT-TC06-01～05 | INT-TC06-01～12（30/30） | E2E-TC06-01～04（4/4） | 通过（2026-08-29 回归） | — |
| REQ07 | UC07 关注关系与内容通知 | ✅ `SYS-SEQ07` | ✅ `COMP-SEQ07` | ✅ `OBJ-SEQ07` | `user/user_handler.go`、`user/relationship_handler.go`、`notification/notification_handler.go`、`UserProfilePage.vue`、`SubscriptionPage.vue` | `relationship_handler_test.go`：自身操作、分组名称、设置参数校验 | `member_b_integration_test.go`：非法目标、关注/取关、分组 CRUD/绑定、屏蔽及关注后动态通知 | E2E-TC07-01～04：关注、分组、黑名单、内容发布通知 | 通过（2026-08-29 三层实跑） | — |
| REQ08 | UC08 创作者会员订阅 | ✅ `SYS-SEQ08` | ✅ `COMP-SEQ08` | ✅ `OBJ-SEQ08` | `membership/membership_handler.go`、`model/mysql/models.go`、`membership.ts`、`SubscriptionPage.vue`、`UserProfilePage.vue` | `membership_handler_test.go`：参数规则及有效/过期订阅续期计算 | `member_b_integration_test.go`：方案、订单、支付幂等、跨订单续期、自动续费关闭、到期识别 | E2E-TC08-01～03：方案展示、购买、重复支付和刷新持久化 | 通过（2026-08-29 三层实跑） | — |
| REQ09 | UC09 直播预约与用户预约 | ✅ `SYS-SEQ09` | ✅ `COMP-SEQ09` | ✅ `OBJ-SEQ09` | `live/schedule_handler.go`、`live/live_handler.go`、`schedule_worker_integration_test.go`、`LiveListPage.vue` | UNIT-TC03/04：`schedule_handler_test.go`、`schedule_worker_integration_test.go` | INT-TC03/04：`backend/integration/engagement_integration_test.go`（预约冲突、通知、取消与 Worker 幂等启动） | E2E-TC09：`frontend/e2e/d-engagement.spec.ts`；2026-08-27 Chromium 报告见 `docs/testing/reports/engagement-e2e/` | 通过（单元、集成、E2E 均有证据） | — |
| REQ10 | UC10 直播发布/观看/实时互动 | ✅ `SYS-SEQ10` | ✅ `COMP-SEQ10` | ✅ `OBJ-SEQ10` | `live/live_handler.go`、`live/srs_hook_handler.go`、`ws/live_publish_handler.go`、`logic/danmaku/hub.go`、`LiveRoomPage.vue`、`LiveStudioPage.vue` | 互动规则、Hub、浏览器推流及 SRS Hook 分支测试 | 直播管理/互动/人数去重/数据库故障；浏览器推流与 SRS/OBS 断流后结束、快速重连不误结束 | E2E-TC10：等待流、WebSocket 断线重连、人数保持、点赞赠礼和结束直播；另有 RTMP-SRS-HLS 证据 | 通过（2026-08-29 异常路径补齐） | — |
| REQ11 | UC11 用户私信与媒体分享 | ✅ `SYS-SEQ11` | ✅ `COMP-SEQ11` | ✅ `OBJ-SEQ11` | `message/message_handler.go`、`logic/chat/hub.go`、`ws/ws_handler.go`、`media/media_handler.go`、`message.ts`、`ChatPage.vue` | 参数、类型、长度、附件文件头、媒体所有权、clientMessageId 校验和错误映射 | 文本/图片/视频/视频分享、历史、未读/已读、屏蔽及 clientMessageId 幂等落库 | E2E-TC11-01～04：实时收发、断线重连、离线历史、图片/视频/站内视频、服务端已读和重复请求 | 通过（2026-08-29 三层实跑） | — |
| REQ12 | UC12 创作者数据分析 | ✅ `SYS-SEQ12` | ✅ `COMP-SEQ12` | ✅ `OBJ-SEQ12` | `creator/analytics_handler.go`、`logic/analytics/creator_stat.go`、`CreatorDashboardPage.vue`、`MetricLineChart.vue` | UNIT-TC12-01/02：鉴权/参数与自然日边界 | INT-TC12-01～04：权限、7/30 日趋势、单作品收窄与空数据 | E2E-TC12-01：完整 member C 串行套件 4/4 通过 | 通过（2026-08-29 回归） | — |
| REQ13 | UC13 平台审核/权限/运营/基础设施 | ✅ `SYS-SEQ13` | ✅ `COMP-SEQ13` | ✅ `OBJ-SEQ13` | `middleware/auth.go`、`admin/admin_handler.go`、`video/video_handler.go`、`danmaku/danmaku_handler.go`、后台六个 Vue 页面 | UNIT-TC13-01～06：鉴权、角色、运营参数和审核状态 | INT-TC13-01～10：API 套件 33/33 通过 | E2E-TC13-01～05：Chromium 5/5 通过，并加入镜像构建前 CI 门禁 | 通过（2026-08-29 回归） | — |

## 汇总统计（截至 2026-08-29，PR #99 + PR #100 合并证据）

| 维度 | 数量 | 明细 |
|---|---|---|
| 三层模型 | 39/39 全部完成 | UC01/07/08/11 于 PR #66、UC06 于 PR #68 补齐（含总用例图 USECASE-OVERVIEW） |
| 单元测试 | 13/13 有覆盖 | PR #99 补 UC01/06；PR #100 补 UC07/08/11，UC04 与 UC13 共享审核链路测试 |
| API/集成测试 | 13/13 有执行证据 | UC06 真实 MySQL API 套件 30/30；UC05/07/08/09/10/11 完成真实 MySQL 回归 |
| E2E 测试 | 13/13 有通过报告 | UC01 5/5、UC06 4/4、UC07/08/11 完整套件，其余用例报告均已入库 |
| 全层通过 | 13/13 | UC01～UC13；统一 CI 在镜像构建前执行全量测试 |
| 部分通过 | 0/13 | — |
| 未测 | 0/13 | — |

## 维护规则

1. 每次补齐一张图或一项测试后，同一次 PR 内更新本表与"结果"，并在 PR 描述中注明对应 UC 编号。
2. "结果"必须能链接到 `docs/testing/reports/` 下的原始报告或 CI 运行记录。
3. 用例最终验收范围以教师确认的 `docs/project/use-case-catalog.md` 为准；范围变更需同步本表。
