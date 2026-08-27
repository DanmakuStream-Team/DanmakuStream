# DanmakuStream 13 个用例需求—设计—代码—测试追溯矩阵（汇总总表）

> 本总表把 UC01～UC13 的追溯信息放在同一视图，便于一次性核对覆盖率和验证完成度。
> 图编号严格与 UC 编号一一对应（SYS-SEQ01～13、COMP-SEQ01～13、OBJ-SEQ01～13）。
> 「主要代码模块」「单元测试」「API/集成测试」「E2E 测试」四列中，已从代码或测试设计文档中确认的信息直接填实；尚未确认或未生成正式编号的填"待补充"。
> 「结果」列只能按实际执行报告填写，不允许根据代码存在就填通过。

| 需求 | 用例 | 系统级模型 | 组件级模型 | 对象级模型 | 主要代码模块 | 单元测试 | API/集成测试 | E2E 测试 | 结果 |
|---|---|---|---|---|---|---|---|---|---|
| REQ01 | UC01 用户注册、登录与资料维护 | SYS-SEQ01 | COMP-SEQ01 | OBJ-SEQ01 | `backend/internal/logic/auth/auth_logic.go`；`backend/internal/handler/v1/auth/auth_handler.go`；`backend/internal/handler/v1/user/user_handler.go`（UpdateMeHandler/UploadAvatarHandler/ProfileHandler）；`frontend/src/pages/home/LoginPage.vue`；`frontend/src/pages/home/RegisterPage.vue`；`frontend/src/pages/user/UserProfilePage.vue`；`frontend/src/api/auth.ts`；`frontend/src/api/user.ts` | UNIT-TC01-01～05 | INT-TC01-01～06 | E2E-TC01-01～03 | 待补充 |
| REQ02 | UC02 视频发现、搜索与播放 | SYS-SEQ02 | COMP-SEQ02 | OBJ-SEQ02 | `backend/internal/handler/v1/video/video_handler.go`（ListHandler/DetailHandler）；`backend/internal/logic/video/list_logic.go`；`backend/internal/logic/video/detail_logic.go`；`frontend/src/pages/home/HomePage.vue`；`frontend/src/pages/video/VideoDetailPage.vue`；`frontend/src/components/common/VideoPlayer.vue`；`frontend/src/api/video.ts` | 待补充 | 待补充 | 待补充 | 待补充 |
| REQ03 | UC03 创作者投稿与状态跟踪 | SYS-SEQ03 | COMP-SEQ03 | OBJ-SEQ03 | `backend/internal/handler/v1/video/video_handler.go`（UploadHandler/UpdateHandler/UpdateCoverHandler/DeleteHandler/MeVideosHandler）；`backend/internal/handler/v1/media/media_handler.go`（UploadImageHandler）；`frontend/src/pages/video/VideoUploadPage.vue`；`frontend/src/pages/user/CreatorDashboardPage.vue`；`frontend/src/api/video.ts`；`frontend/src/api/media.ts` | 待补充 | 待补充 | 待补充 | 待补充 |
| REQ04 | UC04 视频审核与发布 | SYS-SEQ04 | COMP-SEQ04 | OBJ-SEQ04 | `backend/internal/handler/v1/video/video_handler.go`（AdminListHandler/AdminUpdateStatusHandler）；`backend/internal/handler/v1/admin/admin_handler.go` 权限中间件；`frontend/src/pages/admin/AdminVideosPage.vue`；`frontend/src/pages/user/CreatorDashboardPage.vue`；`frontend/src/api/video.ts`；`frontend/src/api/admin.ts` | 待补充 | 待补充 | 待补充 | 待补充 |
| REQ05 | UC05 视频观看互动（弹幕/评论/点赞/收藏） | SYS-SEQ05 | COMP-SEQ05 | OBJ-SEQ05 | `backend/internal/handler/v1/danmaku/danmaku_handler.go`；`backend/internal/handler/v1/comment/comment_handler.go`；`backend/internal/handler/v1/video/video_handler.go`（LikeHandler/CollectHandler）；`backend/internal/logic/danmaku/hub.go`；`frontend/src/pages/video/VideoDetailPage.vue`；`frontend/src/components/common/VideoPlayer.vue`；`frontend/src/api/danmaku.ts`；`frontend/src/api/comment.ts` | UNIT-TC05-01～（参考成员 D 表） | INT-TC05-01～02（参考成员 D 表） | 待补充 | 弹幕/评论/点赞/收藏集成测试通过（参考成员 D 验收报告） |
| REQ06 | UC06 个人视频资料库管理 | SYS-SEQ06 | COMP-SEQ06 | OBJ-SEQ06 | `backend/internal/handler/v1/user/library_handler.go`；`backend/internal/model/mysql/models.go`（WatchHistory/WatchProgress/WatchLater）；`frontend/src/api/library.ts`；`frontend/src/pages/user/UserLibraryPage.vue`；`frontend/src/utils/playQueue.ts`；`frontend/src/utils/userLibrary.ts`；`frontend/src/store/video.ts` | UNIT-TC06-01～04 | INT-TC06-01～05 | E2E-TC06-01～03 | 待补充 |
| REQ07 | UC07 关注关系、分组、屏蔽与内容通知 | SYS-SEQ07 | COMP-SEQ07 | OBJ-SEQ07 | `backend/internal/handler/v1/user/user_handler.go`（FollowHandler/FollowingListHandler）；`backend/internal/handler/v1/user/relationship_handler.go`；`backend/internal/handler/v1/notification/notification_handler.go`；`frontend/src/pages/user/UserProfilePage.vue`；`frontend/src/pages/user/SubscriptionPage.vue`；`frontend/src/api/user.ts`；`frontend/src/api/notification.ts` | UNIT-TC07-01～04 | INT-TC07-01～06 | E2E-TC07-01～03 | 待补充 |
| REQ08 | UC08 创作者会员订阅 | SYS-SEQ08 | COMP-SEQ08 | OBJ-SEQ08 | `backend/internal/handler/v1/membership/membership_handler.go`；`backend/internal/model/mysql/models.go`（MembershipPlan/Order/Subscription）；`frontend/src/api/membership.ts`；`frontend/src/pages/user/SubscriptionPage.vue`；`frontend/src/pages/user/UserProfilePage.vue` | UNIT-TC08-01～05 | INT-TC08-01～07 | E2E-TC08-01～03 | 待补充 |
| REQ09 | UC09 直播预约与用户预约 | SYS-SEQ09 | COMP-SEQ09 | OBJ-SEQ09 | `backend/internal/handler/v1/live/schedule_handler.go`（ScheduleListHandler/CreateScheduleHandler/CancelScheduleHandler/ReserveScheduleHandler）；`frontend/src/api/live.ts`；`frontend/src/pages/live/LiveListPage.vue` | UNIT-TC09-03～04（参考成员 D 表） | INT-TC09-03～04（参考成员 D 表） | E2E-TC09 | 已通过；Chromium（2026-08-26 参考成员 D 验收报告） |
| REQ10 | UC10 直播发布、观看与实时互动 | SYS-SEQ10 | COMP-SEQ10 | OBJ-SEQ10 | `backend/internal/handler/v1/live/live_handler.go`（ListHandler/DetailHandler/CreateHandler/EndHandler）；`backend/internal/handler/v1/live/interaction_handler.go`；`backend/internal/handler/ws/live_publish_handler.go`；`backend/internal/handler/ws/ws_handler.go`（LiveWebSocketHandler）；`backend/internal/logic/danmaku/hub.go`；`frontend/src/pages/live/LiveStudioPage.vue`；`frontend/src/pages/live/LiveRoomPage.vue` | UNIT-TC10-01～06（参考成员 D 表） | INT-TC10-06～09（参考成员 D 表） | E2E-TC10；E2E-MEDIA | 已通过；Chromium + RTMP-SRS-HLS（2026-08-26 参考成员 D 验收报告） |
| REQ11 | UC11 用户私信与媒体分享 | SYS-SEQ11 | COMP-SEQ11 | OBJ-SEQ11 | `backend/internal/handler/v1/message/message_handler.go`；`backend/internal/logic/chat/hub.go`；`backend/internal/handler/ws/ws_handler.go`（ChatWebSocketHandler）；`backend/internal/handler/v1/media/media_handler.go`（UploadMessageMediaHandler）；`frontend/src/api/message.ts`；`frontend/src/pages/user/ChatPage.vue` | UNIT-TC11-01～05 | INT-TC11-01～07 | E2E-TC11-01～03 | 待补充 |
| REQ12 | UC12 创作者数据分析 | SYS-SEQ12 | COMP-SEQ12 | OBJ-SEQ12 | `backend/internal/handler/v1/creator/analytics_handler.go`（AnalyticsHandler）；`backend/internal/logic/analytics/creator_stat.go`；`frontend/src/pages/user/CreatorDashboardPage.vue`；`frontend/src/components/creator/MetricLineChart.vue` | 待补充 | 待补充 | 待补充 | 待补充 |
| REQ13 | UC13 平台审核、权限、运营与基础设施管理 | SYS-SEQ13 | COMP-SEQ13 | OBJ-SEQ13 | `backend/api/main.go`；`backend/internal/middleware/auth.go`；`backend/internal/handler/v1/admin/admin_handler.go`；`backend/internal/handler/v1/video/video_handler.go`（AdminListHandler/AdminUpdateStatusHandler）；`backend/internal/handler/v1/danmaku/danmaku_handler.go`；`frontend/src/pages/admin/Admin*.vue`；`frontend/src/api/admin.ts` | UNIT-TC13-01～06 | INT-TC13-01～10 | E2E-TC13-01～05 | 通过（2026-08-26 uc13-admin.md 与 uc13-api-2026-08-26.txt） |

---

## 单用例追溯表索引

| 用例 | 单用例文件 | 用例说明 | 测试设计文档 |
|---|---|---|---|
| UC01 | `docs/traceability/uc01.md` | `docs/models/usecase/UC01.md` | `docs/testing/test-cases/uc01-user-service.md` |
| UC02 | `docs/traceability/uc02.md` | `docs/project/use-case-catalog.md §UC02` | 待补充 |
| UC03 | `docs/traceability/uc03.md` | `docs/project/use-case-catalog.md §UC03` | 待补充 |
| UC04 | `docs/traceability/uc04.md` | `docs/project/use-case-catalog.md §UC04` | 待补充 |
| UC05 | 见 `docs/contributions/member-d/traceability.md` | `docs/project/use-case-catalog.md §UC05` | `docs/testing/test-cases/engagement-uc05-uc09-uc10.md` |
| UC06 | `docs/traceability/uc06.md` | `docs/contributions/member-e/uc06-personal-library.md` | `docs/testing/test-cases/uc06-personal-library.md` |
| UC07 | `docs/traceability/uc07.md` | `docs/models/usecase/UC07.md` | `docs/testing/test-cases/uc07-relationships.md` |
| UC08 | `docs/traceability/uc08.md` | `docs/models/usecase/UC08.md` | `docs/testing/test-cases/uc08-membership.md` |
| UC09 | 见 `docs/contributions/member-d/traceability.md` | `docs/project/use-case-catalog.md §UC09` | `docs/testing/test-cases/engagement-uc05-uc09-uc10.md` |
| UC10 | 见 `docs/contributions/member-d/traceability.md` | `docs/project/use-case-catalog.md §UC10` | `docs/testing/test-cases/engagement-uc05-uc09-uc10.md` |
| UC11 | `docs/traceability/uc11.md` | `docs/models/usecase/UC11.md` | `docs/testing/test-cases/uc11-messaging.md` |
| UC12 | `docs/traceability/uc12.md` | `docs/project/use-case-catalog.md §UC12` | 待补充 |
| UC13 | `docs/traceability/uc13.md` | `docs/models/usecase/UC13.md` | `docs/testing/test-cases/uc13-admin.md` |

---

## 图档索引（SEQ 编号与 UC 一一对应）

| 层级 | 图档目录 | 命名规范 |
|---|---|---|
| 系统级顺序图 | `docs/models/system/` | `SYS-SEQ{NN}.puml`，NN = 01～13 |
| 组件级顺序图 | `docs/models/component/` | `COMP-SEQ{NN}.puml`，NN = 01～13 |
| 对象级顺序图 | `docs/models/object/` | `OBJ-SEQ{NN}.puml`，NN = 01～13 |
| 总用例图 | `docs/models/usecase/` | `USECASE-OVERVIEW.puml` + `USECASE-B.puml` |
| 部署图 | `docs/models/deployment/` | `DEPLOY-MONO*.puml` / `DEPLOY-K8S.puml` |
| 类图（概念+实现） | `docs/models/class/`、`docs/models/object/` | `CLASS-*.puml` / `DOMAIN-CLASS-*.puml` / `IMPLEMENTATION-CLASS-*.puml` |

---

## 覆盖率统计（按追溯表字段填实度，不含测试执行结果）

- 需求 + 用例 + 三层模型编号：13/13 ✅ 100%（按 UC 编号一一映射）
- 主要代码模块（确认存在并填入 Handler/Service/Repo/前端页面 API）：
  - 完整填实：UC01、UC05、UC06、UC07、UC08、UC09、UC10、UC11、UC13 → **9/13**
  - 部分填实：UC02、UC03、UC04、UC12（无独立测试编号，需后续补齐）
- 单元测试编号：UC01、UC05、UC06、UC07、UC08、UC09、UC10、UC11、UC13 → 9/13
- API/集成测试编号：UC01、UC05、UC06、UC07、UC08、UC09、UC10、UC11、UC13 → 9/13
- E2E 测试编号：UC01、UC06、UC07、UC08、UC09、UC10、UC11、UC13 → 8/13
- 结果（真实执行通过）：UC05、UC09、UC10、UC13 → **4/13**

---

## 填写规则（与单用例文件一致）

1. **结果只能据实填写**：不允许根据代码存在、图档存在就写通过。
2. **编号一致性**：SEQ 编号严格等于 UC 编号；若某 UC 产生多张图，在主图基础上以 `-sub1` 后缀区分，主图仍取同号。
3. **测试执行后同步**：每次跑过 go test / playwright test / 集成测试脚本后，需把结果列从「待补充」改为 `通过（YYYY-MM-DD 证据文件路径）`或 `失败（YYYY-MM-DD Issue 链接）`。
4. **代码重构同步**：Handler/Service/Repo 路径变更后，同步更新本矩阵和对应单用例追溯表的「主要代码模块」列。
