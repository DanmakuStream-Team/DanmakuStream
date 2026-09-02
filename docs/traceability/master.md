# DanmakuStream 总追溯表（UC01～UC13）

> 本表按课程任务书要求，把**需求、用例、三层模型、代码模块、测试编号与测试结果**汇总到一张表。
> 填写规则继承 `docs/traceability/uc13.md`："结果"一栏只能依据实际测试报告填写；缺图或未测的格子如实留空/标 ❌，不允许按代码存在就写通过。
> 状态统计（2026-09-02，Day8 微服务 E2E 迁移完成后）：三层图 **39/39 全部完成**；**单体 E2E 13/13 全通过**；**微服务 E2E 基线 2/2 + 业务 5/13 非 skip + 8/13 骨架待跨服务打通**。统一 CI 会在镜像构建前执行全部测试，任一失败即阻断后续构建与部署。

| 需求 | 用例 | 系统级图 | 组件级图 | 对象级图 | 主要代码模块 | 单元测试 | API/集成测试 | 单体 E2E | 微服务 E2E（Day8） | 单体结果 | 微服务结果（骨架/待打通） | 缺口（负责人） |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| REQ01 | UC01 注册/登录/资料维护 | ✅ | ✅ | ✅ | `auth/auth_handler.go`、`user/user_handler.go`、`LoginPage.vue`、`RegisterPage.vue`、`UserProfilePage.vue` | UNIT-TC01-01～03 | 注册/重复注册/错误密码/无效token/资料更新 | E2E-TC01-01～05（5/5） | `uc01-user.spec.ts` 微服务版（5 case 非 skip） | 通过（2026-08-29） | ✅ 已实现骨架（user-service） | — |
| REQ02 | UC02 视频发现/搜索/播放 | ✅ | ✅ | ✅ | `video/video_handler.go`、`list_logic.go`、`detail_logic.go`、`HomePage.vue` | UNIT-TC02-01/02 | INT-TC02-01～04 | E2E-TC02-01 | `uc02-video-search.spec.ts` 微服务版（1 case 非 skip） | 通过 | ✅ 已实现骨架（content-service） | — |
| REQ03 | UC03 创作者投稿与状态跟踪 | ✅ | ✅ | ✅ | `media/`、`video` 上传与转码、`CreatorDashboardPage.vue` | UNIT-TC03-01～03 | INT-TC03-01～04 | E2E-TC03-01 | `uc03-video-upload.spec.ts`（skip：上传路由+媒体卷） | 通过 | ⏭️ skip（文件上传+媒体共享） | 成员C：content-service 上传+媒体卷 |
| REQ04 | UC04 视频审核与发布 | ✅ | ✅ | ✅ | `AdminUpdateStatus`、审核后台页面 | UNIT-TC04-01～03 + TC13-06 | INT-TC04-01～04 | E2E-TC04-01 + TC13-01 | `uc04-video-review.spec.ts` 微服务版（1 case 非 skip） | 通过 | ✅ 已实现骨架（content admin） | — |
| REQ05 | UC05 视频观看互动 | ✅ | ✅ | ✅ | `danmaku/`、`comment/`、`VideoDetailPage.vue` | UNIT-TC01（hub_test 等） | engagement_integration_test.go | `d-engagement.spec.ts` | `uc05-danmaku-comment.spec.ts`（skip：跨服务视频 join） | 通过 | ⏭️ skip（engagement+content 跨服务） | 成员D+C：视频标题跨服务 internal API |
| REQ06 | UC06 个人视频资料库管理 | ✅ | ✅ | ✅ | `collection/`、`user/`（历史/进度/稍后再看）、`UserLibraryPage.vue` | UNIT-TC06-01～05 | INT-TC06-01～12（30/30） | E2E-TC06-01～04（4/4） | `uc06-library.spec.ts` 微服务版（4 case 非 skip） | 通过（2026-08-29） | ✅ 已实现骨架（engagement-service） | — |
| REQ07 | UC07 关注关系与内容通知 | ✅ | ✅ | ✅ | `user/relationship_handler.go`、`notification_handler.go`、`SubscriptionPage.vue` | relationship_handler_test.go | 关注/取关/分组/屏蔽/动态通知 | E2E-TC07-01～04 | `uc07-follow-group-block.spec.ts`（skip：动态+通知跨服务） | 通过 | ⏭️ skip（user+content 动态/通知推送） | 成员B：通知推送+动态发布联调 |
| REQ08 | UC08 创作者会员订阅 | ✅ | ✅ | ✅ | `membership_handler.go`、`SubscriptionPage.vue` | membership_handler_test.go | 方案/订单/支付幂等/续期 | E2E-TC08-01～03 | `uc08-membership-subscription.spec.ts`（skip：支付 mock） | 通过 | ⏭️ skip（支付回调 mock 服务） | 成员B：demo-pay 端点+mock回调 |
| REQ09 | UC09 直播预约与用户预约 | ✅ | ✅ | ✅ | `schedule_handler.go`、`LiveListPage.vue` | schedule_handler_test.go | 预约冲突/通知/Worker幂等 | E2E-TC09 | `uc09-live-schedule.spec.ts`（skip：作者信息跨服务） | 通过 | ⏭️ skip（engagement+user 作者昵称 join） | 成员D：跨服务作者信息 API |
| REQ10 | UC10 直播发布/观看/实时互动 | ✅ | ✅ | ✅ | `live_handler.go`、`srs_hook_handler.go`、`LiveRoomPage.vue` | Hub/SRS Hook 分支测试 | 直播管理/互动/SRS 断流重连 | E2E-TC10（WS+礼物+结束） | `uc10-live-streaming.spec.ts`（skip：SRS webhook+推流） | 通过 | ⏭️ skip（SRS webhook+OBS 串流） | 成员D：SRS webhook 回传 engagement |
| REQ11 | UC11 用户私信与媒体分享 | ✅ | ✅ | ✅ | `message_handler.go`、`chat hub`、`ws_handler.go`、`ChatPage.vue` | 参数/附件/幂等全部测试 | 文本/图片/视频分享/未读/幂等 | E2E-TC11-01～04（实时/断线/离线） | `uc11-chat-message.spec.ts`（skip：WS+媒体上传） | 通过 | ⏭️ skip（/ws/chat 网关+媒体上传路径） | 成员B：网关 WS 路由+媒体卷映射 |
| REQ12 | UC12 创作者数据分析 | ✅ | ✅ | ✅ | `analytics_handler.go`、`creator_stat.go`、`CreatorDashboardPage.vue` | UNIT-TC12-01/02（边界+鉴权） | INT-TC12-01～04 | E2E-TC12-01 | `uc12-creator-analytics.spec.ts`（skip：跨服务聚合） | 通过 | ⏭️ skip（content+engagement 聚合） | 成员C：analytics 跨服务 internal 查询 |
| REQ13 | UC13 平台审核/权限/运营/基础设施 | ✅ | ✅ | ✅ | `middleware/auth.go`、`admin_handler.go`、后台页面 | UNIT-TC13-01～06 | INT-TC13-01～10（33/33） | E2E-TC13-01～05（5/5） | `uc13-admin.spec.ts`（skip：三服务 admin 路由） | 通过（CI 阻断） | ⏭️ skip（网关 /admin/* 路由） | 成员E：网关 admin 路由+权限边界 |

## 汇总统计（Day8 微服务 E2E 迁移后，含单体 & 微服务）

### 单体（稳定）
| 维度 | 数量 | 明细 |
|---|---|---|
| 三层模型 | 39/39 全部完成 | UC01/07/08/11 于 PR #66、UC06 于 PR #68 补齐（含总用例图 USECASE-OVERVIEW） |
| 单元测试 | 13/13 有覆盖 | PR #99 补 UC01/06；PR #100 补 UC07/08/11，UC04 与 UC13 共享审核链路测试 |
| API/集成测试 | 13/13 有执行证据 | UC06 真实 MySQL API 套件 30/30；其余 UC 真实 MySQL 回归均通过 |
| 单体 E2E | 13/13 有通过报告 | UC01 5/5、UC06 4/4、其余全通过 |
| 全层通过 | 13/13 | 统一 CI 镜像构建前阻断机制生效 |

### 微服务（Day8 迁移基线）
| 维度 | 数量 | 明细 |
|---|---|---|
| 微服务基线冒烟 | 2/2 ✅ | 网关健康 + 三服务目录 + JWT 跨三服务连通（gateway-smoke.spec.ts） |
| 微服务业务已实现骨架（非 skip） | 5/13 | UC01、UC02、UC04、UC06 （4个纯单服务或 admin 简单路径） |
| 微服务业务骨架待跨服务打通（skip） | 8/13 | UC03/05/07/08/09/10/11/12/13 = 8 项，阻塞原因已在上方表格标注 |
| 微服务 E2E 文件数 | 15 个文件 | gateway-smoke（2 case） + uc01～uc13 共 13 个独立 spec + 共享 fixtures + test-data |
| 一键运行脚本 | 2 种 | `bash scripts/run-microservices-e2e.sh`（基线）；`MICRO_E2E_FULL_SUITE=1 bash scripts/run-microservices-e2e.sh`（全量 13UC） |
| 报告模板 | 1 份 | `docs/testing/microservices-e2e-full-report.md`（环境信息、13UC 表、失败分析、行动项、执行痕迹） |

## 维护规则

1. 每次补齐一张图或一项测试后，同一次 PR 内更新本表与"结果"，并在 PR 描述中注明对应 UC 编号。
2. "结果"必须能链接到 `docs/testing/reports/` 下的原始报告或 CI 运行记录。
3. 用例最终验收范围以教师确认的 `docs/project/use-case-catalog.md` 为准；范围变更需同步本表。
