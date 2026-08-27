# DanmakuStream 最终追溯表（骨架）

> **目标**：确保 13 个用例（UC01~UC10、UC12~UC14）需求→设计→代码→测试→交付物**追溯率 100%**。
>
> **填写规则**：
> 1. 每个单元格真实可链接，不得使用「待补充」作为长期状态；未落地需写「未实现」并在**缺口清单**列出风险。
> 2. 结果列只能填 **通过 / 失败 / 阻塞 / 未执行** 四种，通过必须有测试报告与截图证据；严禁根据「代码存在」直接写通过。
> 3. UC11（私信与媒体分享）属于**空缺待确认**，单独放在第 4 节，不计入 13 UC 的 100% 追溯率。
> 4. 交付时必须保证：**13 行的「设计/代码/单元/API/E2E/结果」7 列，全部为非空有效值**。
>
> **版本**：v1.0（骨架）
> **负责人**：_____________
> **日期**：YYYY-MM-DD

---

## 1. 13 用例追溯矩阵（追溯率 100% 目标行）

| 用例编号 | 用例名称 | 需求来源（REQs/PRD） | 系统级设计 | 组件级设计 | 对象级设计 | 主要代码模块 | 单元测试 | API/集成测试 | E2E 测试 | 结果（通过/失败/阻塞/未执行） |
|---|---|---|---|---|---|---|---|---|---|---|
| UC01 | 用户注册、登录与资料维护 | [业务场景用例清单.md](file:///d:/DanmakuStream/docs/业务场景用例清单.md) 第 3 节 | SYS-SEQ01（待落图） | COMP-SEQ01（待落图） | OBJ-SEQ01（待落图） | auth_handler.go、user_handler.go、LoginPage.vue、RegisterPage.vue、UserProfilePage.vue | UNIT-UC01-* | INT-UC01-* | E2E-UC01-01~03 | |
| UC02 | 视频发现、搜索与播放 | 同上第 UC02 节 | SYS-SEQ02 | COMP-SEQ02 | OBJ-SEQ02 | video_handler.go、list_logic.go、detail_logic.go、HomePage.vue、VideoDetailPage.vue | UNIT-UC02-* | INT-UC02-* | E2E-UC02-01~03 | |
| UC03 | 创作者投稿与状态跟踪 | 同上 UC03 节 | SYS-SEQ03 | COMP-SEQ03 | OBJ-SEQ03 | video_handler.go、media_handler.go、VideoUploadPage.vue、CreatorDashboardPage.vue | UNIT-UC03-* | INT-UC03-* | E2E-UC03-01~03 | |
| UC04 | 视频审核与发布 | 同上 UC04 节 | SYS-SEQ04 | COMP-SEQ04 | OBJ-SEQ04 | admin_handler.go、video_handler.go、AdminVideosPage.vue | UNIT-UC04-* | INT-UC04-* | E2E-UC04-01~03 | |
| UC05 | 视频观看互动 | 同上 UC05 节 | [SYS-SEQ05.puml](file:///d:/DanmakuStream/docs/models/system/SYS-SEQ05.puml) | [COMP-SEQ05.puml](file:///d:/DanmakuStream/docs/models/component/COMP-SEQ05.puml) | [OBJ-SEQ05.puml](file:///d:/DanmakuStream/docs/models/object/OBJ-SEQ05.puml) | danmaku_handler.go、comment_handler.go、VideoDetailPage.vue、VideoPlayer.vue、DanmakuLayer.vue | UNIT-UC05-* | INT-UC05-* | E2E-UC05-01~03 | |
| UC06 | 个人视频资料库管理 | 同上 UC06 节 | [SYS-SEQ06.puml](file:///d:/DanmakuStream/docs/models/system/SYS-SEQ06.puml) | [COMP-SEQ06.puml](file:///d:/DanmakuStream/docs/models/component/COMP-SEQ06.puml) | [OBJ-SEQ06.puml](file:///d:/DanmakuStream/docs/models/object/OBJ-SEQ06.puml) | library_handler.go、UserLibraryPage.vue、library.ts、userLibrary.ts、VideoPlayer.vue（timeupdate） | UNIT-UC06-* | INT-UC06-* | E2E-UC06-01~03 | |
| UC07 | 关注关系与内容通知 | 同上 UC07 节 | SYS-SEQ07 | COMP-SEQ07 | OBJ-SEQ07 | relationship_handler.go、notification_handler.go、dynamic_handler.go、UserProfilePage.vue、SubscriptionPage.vue | UNIT-UC07-* | INT-UC07-* | E2E-UC07-01~03 | |
| UC08 | 创作者会员订阅 | 同上 UC08 节 | SYS-SEQ08 | COMP-SEQ08 | OBJ-SEQ08 | membership_handler.go、membership/*Worker、SubscriptionPage.vue、membership.ts | UNIT-UC08-* | INT-UC08-* | E2E-UC08-01~03 | |
| UC09 | 直播预约与用户预约 | 同上 UC09 节 | [SYS-SEQ09.puml](file:///d:/DanmakuStream/docs/models/system/SYS-SEQ09.puml) | [COMP-SEQ09.puml](file:///d:/DanmakuStream/docs/models/component/COMP-SEQ09.puml) | [OBJ-SEQ09.puml](file:///d:/DanmakuStream/docs/models/object/OBJ-SEQ09.puml) | schedule_handler.go、LiveListPage.vue、live.ts | UNIT-UC09-* | INT-UC09-* | E2E-UC09-01~03 | |
| UC10 | 直播发布、观看与实时互动 | 同上 UC10 节 | [SYS-SEQ10.puml](file:///d:/DanmakuStream/docs/models/system/SYS-SEQ10.puml) | [COMP-SEQ10.puml](file:///d:/DanmakuStream/docs/models/component/COMP-SEQ10.puml) | [OBJ-SEQ10.puml](file:///d:/DanmakuStream/docs/models/object/OBJ-SEQ10.puml) | live_handler.go、interaction_handler.go、ws_handler.go、live_publish_handler.go、logic/danmaku/hub.go、LiveStudioPage.vue、LiveRoomPage.vue | UNIT-UC10-* | INT-UC10-* | E2E-UC10-01~03 | |
| UC12 | 创作者数据分析 | 同上 UC12 节 | SYS-SEQ12 | COMP-SEQ12 | OBJ-SEQ12 | analytics_handler.go、creator_stat.go、CreatorDashboardPage.vue、MetricLineChart.vue | UNIT-UC12-* | INT-UC12-* | E2E-UC12-01~03 | |
| UC13 | 平台审核、权限、运营与基础设施管理 | 同上 UC13 节 | [SYS-SEQ13.puml](file:///d:/DanmakuStream/docs/models/system/SYS-SEQ13.puml) | [COMP-SEQ13.puml](file:///d:/DanmakuStream/docs/models/component/COMP-SEQ13.puml) | [OBJ-SEQ13.puml](file:///d:/DanmakuStream/docs/models/object/OBJ-SEQ13.puml) | admin_handler.go、middleware/auth.go、AdminDashboardPage.vue、AdminVideosPage.vue、AdminDanmakuPage.vue、AdminUsersPage.vue、AdminOperationsPage.vue、AdminInfrastructurePage.vue | UNIT-TC13-01~06 | INT-TC13-01~10 | E2E-TC13-01~05 | |
| UC14 | 用户标签偏好与个性化推荐 | 同上尾端 + recommendation/README | SYS-SEQ14 | COMP-SEQ14 | OBJ-SEQ14 | recommendation/（itemcf.py、baselines.py、evaluate.py）、TagAffinityPage.vue、首页推荐区（预留） | UNIT-UC14-* | INT-UC14-* | E2E-UC14-01~03 | |

**13 用例追溯检查**：
- [ ] 每行 「需求来源/系统级/组件级/对象级/代码/单元/API/E2E/结果」 9 列均非空
- [ ] 13 行「结果」中不含「未执行」（允许「阻塞」，但需在 §3 给出解阻塞计划）
- [ ] 追溯率 = 13 / 13 = **100%** → 通过 / 不通过

---

## 2. 用例→页面/API/E2E 编号映射子表（与 13 UC 追溯表一一对应）

> 此表与 [all-uc-traceability.md](file:///d:/DanmakuStream/docs/traceability/all-uc-traceability.md) 对齐，作为最终交付的精简快照。

| 用例 | 前端页面（路由） | 后端 API（关键路由） | E2E 编号 |
|---|---|---|---|
| UC01 | /login、/register、/user/:id（本人） | /auth/*、/users/me、/users/me/avatar | E2E-UC01-01~03 |
| UC02 | /、/video/:id | /videos*、/videos/:id、/search/users | E2E-UC02-01~03 |
| UC03 | /creator/upload、/creator（我的视频） | /videos/upload、/videos/:id、/users/me/videos | E2E-UC03-01~03 |
| UC04 | /admin/videos + /creator（状态展示） | /admin/videos、/admin/videos/:id/status | E2E-UC04-01~03 |
| UC05 | /video/:id（弹幕/评论/点赞/收藏） | /danmaku*、/comments*、/videos/:id/like、/videos/:id/collect | E2E-UC05-01~03 |
| UC06 | /me/:kind ∈ {history,watchlater,liked,collections,downloads} | /users/me/history*、/users/me/watch-later* + localStorage 本地 | E2E-UC06-01~03 |
| UC07 | /user/:id、/subscriptions、/me/blocked | /users/:id/follow、/users/following、/users/follow-groups*、/users/:id/block、/dynamics、/notifications* | E2E-UC07-01~03 |
| UC08 | /subscriptions（端）+ 创作中心会员配置 | /creator/membership-plan、/subscriptions/*、/subscriptions/orders/*/demo-pay | E2E-UC08-01~03 |
| UC09 | /live（预约列表） | /live-schedules*、/live-schedules/:id/reserve | E2E-UC09-01~03 |
| UC10 | /live/studio/:id、/live/:id | /live*、/live/:id/like、/ws/*、SRS RTMP/HLS | E2E-UC10-01~03 |
| UC12 | /creator（数据中心） | /creator/analytics | E2E-UC12-01~03 |
| UC13 | /admin、/admin/videos、/admin/danmaku、/admin/users、/admin/operations、/admin/infrastructure | /admin/*（staff/admin 守卫） | E2E-TC13-01~05 |
| UC14 | /me/tags + /（首页推荐区） | recommendation/ 离线脚本 + 前端标签聚合 | E2E-UC14-01~03 |

---

## 3. 缺口与风险清单（交付前必须清零或签核）

> 说明：追溯率 100% 的硬约束是「13 行所有列非空 + 结果非未执行」。如交付时仍存在缺口，需在此列出并让项目经理/客户签核。

| 编号 | 用例 | 缺口列 | 具体说明 | 责任方 | 计划完成日 | 签核人 |
|---|---|---|---|---|---|---|
| GAP-01 | UC01/02/03/04/07/08/12/14 | 系统级/组件级/对象级设计图 | 目前只有 UC05/06/09/10/13 的三层顺序图，其余 8 UC 未建图 | 各 UC 负责人 | YYYY-MM-DD | |
| GAP-02 | UC01~03、07~08、12、14 | 单元/API/E2E 自动化 | 只有 UC13 有落地；其余尚未编写 `e2e/*.spec.ts`、`*_test.go`、`tests/api/*.sh` | 各 UC 负责人 | YYYY-MM-DD | |
| GAP-03 | UC10-01 媒体链路 | E2E 环境（SRS） | 报告中若 SRS 不可用则标记 [MEDIA-REQUIRED] 阻塞 | DevOps | YYYY-MM-DD | |
| GAP-04 | UC14 | 线上推荐 API | recommendation 仅离线脚本，尚未接入后端 `/recommendations` 路由；当前仅前端本地标签聚合 | 推荐负责人 | YYYY-MM-DD | |

---

## 4. UC11 占位（空缺待确认，不计入 13 UC 追溯率）

| 项 | 内容 |
|---|---|
| UC11 名称 | 用户私信与媒体分享 |
| 关联页面/代码（已存在） | ChatPage.vue、message_handler.go、logic/chat/hub.go、ws_handler.go、messages/* API 已注册 |
| 当前状态 | 空缺待确认：PRD 尚未冻结、业务边界与附件合规未签核 |
| 是否纳入 13 UC 追溯率 | 否 |
| 后续接入计划 | 待定；确认后新增第 14 行到 §1，补三层图+测试，总追溯率目标改为 14/14 = 100% |

---

## 5. 交付物索引（链接到所有正式产出）

| 类别 | 交付物 | 路径 |
|---|---|---|
| 总用例图 | UC-OVERVIEW.puml | [UC-OVERVIEW.puml](file:///d:/DanmakuStream/docs/models/usecase/UC-OVERVIEW.puml) |
| UC06 三层顺序图 | SYS/COMP/OBJ-SEQ06 | §1 第 6 行已链接 |
| 部署图（改造前/改造后） | DEPLOY-MONO / DEPLOY-K8S | [DEPLOY-MONO.puml](file:///d:/DanmakuStream/docs/models/deployment/DEPLOY-MONO.puml) / [DEPLOY-K8S.puml](file:///d:/DanmakuStream/docs/models/deployment/DEPLOY-K8S.puml) |
| UC06 用例说明 | UC06.md | [UC06.md](file:///d:/DanmakuStream/docs/models/usercase/UC06.md) |
| 13 用例追溯（详细） | all-uc-traceability.md | [all-uc-traceability.md](file:///d:/DanmakuStream/docs/traceability/all-uc-traceability.md) |
| E2E 测试计划 | e2e-test-plan.md | [e2e-test-plan.md](file:///d:/DanmakuStream/docs/tests/e2e-test-plan.md) |
| E2E 测试报告 | e2e-test-report-skeleton.md（填实版） | [e2e-test-report-skeleton.md](file:///d:/DanmakuStream/docs/tests/reports/e2e-test-report-skeleton.md) |
| 技术报告（骨架） | technical-report-skeleton.md | [technical-report-skeleton.md](file:///d:/DanmakuStream/docs/technical-report-skeleton.md) |

---

## 6. 审批

| 角色 | 姓名 | 签名 | 日期 | 意见 |
|---|---|---|---|---|
| 开发负责人 | | | | |
| 测试负责人 | | | | |
| 项目经理 | | | | |
| 客户/产品（如需） | | | | |
