# 微服务划分、服务接口与数据表归属方案

> 状态：**v1.1（2026-08-31）——本文件为仓库内权威版本**（收编自中期材料，中期提交版不再维护）。
> D06 拆分已落地：三个业务服务（`services/user-service`、`services/content-service`、`services/engagement-service`）已实现并合入 dev，微服务平台（网关路由表 / Compose / K8s 骨架 / 三库三账号）见 PR #103；单体 `backend/` 作为 v1.0 基线保留（main 分支）。

## 1. 服务划分

| 服务 | 覆盖用例 | 核心职责 | 目标 Schema |
|---|---|---|---|
| `user-service` | UC01、UC07、UC08、UC11 | 注册登录、资料、关注/分组/屏蔽、会员方案与订单、私信、通知 | `user_db` |
| `content-service` | UC02、UC03、UC04、UC12，以及 UC13 内容运营部分 | 视频发现、投稿/媒体、审核、创作者分析、横幅与公告 | `content_db` |
| `engagement-service` | UC05、UC06、UC09、UC10 | 点赞收藏评论弹幕、资料库、直播预约、直播互动与礼物 | `engagement_db` |

公共基础设施：前端、API Gateway、MySQL 实例、SRS、对象/文件存储不计入“三个业务微服务”。Gateway 负责统一入口、路由和身份透传，各服务仍需自行校验授权。

## 2. 服务接口清单

### 2.1 user-service

| 接口组 | 代表接口 | 用途 |
|---|---|---|
| 身份与资料 | `POST /auth/register`、`POST /auth/login`、`GET /auth/me`、`PUT /users/me` | 注册、登录、身份和资料维护 |
| 关注与屏蔽 | `POST /users/:id/follow`、`GET/POST/PUT/DELETE /users/follow-groups`、`POST /users/:id/block` | 关系、分组、特别关注和屏蔽 |
| 会员订阅 | `PUT /creator/membership-plan`、`POST /subscriptions/orders`、`POST /subscriptions/orders/:orderNo/demo-pay` | 方案、订单、演示支付与订阅状态 |
| 私信与通知 | `GET /messages/conversations`、`GET/POST /messages`、`PUT /messages/:userId/read`、`GET/PUT /notifications` | 会话、消息、已读和通知 |
| 内部接口 | `GET /internal/users/:id/summary`、`GET /internal/users/:id/relationship`、`GET /internal/users/:id/membership` | 供其他服务读取最小用户摘要、关系与会员状态 |

### 2.2 content-service

| 接口组 | 代表接口 | 用途 |
|---|---|---|
| 视频发现 | `GET /videos`、`GET /videos/:id`、`GET /search/users`、`GET /media/*` | 列表、搜索、详情和媒体访问 |
| 投稿与媒体 | `POST /videos/upload`、`POST /videos/:id/cover`、`PUT/DELETE /videos/:id` | 上传、封面、编辑和删除 |
| 审核与运营 | `GET /admin/videos`、`PUT /admin/videos/:id/status`、`/admin/banners`、`/admin/announcements` | 审核、横幅和公告管理 |
| 动态发布 | `GET/POST /dynamics`、`DELETE /dynamics/:id` | 创作者动态发布与删除（v1.1 由 engagement 移入） |
| 创作者分析 | `GET /creator/analytics?range=...` | 汇总指标、趋势和单作品分析 |
| 内部接口 | `GET /internal/videos/:id/summary`、`GET /internal/videos/:id/playability`、`GET /internal/videos/:id/owner` | 供互动服务校验视频存在性、可播放状态和所有权 |

### 2.3 engagement-service

| 接口组 | 代表接口 | 用途 |
|---|---|---|
| 视频互动 | `GET/POST /danmaku`、`GET/POST/DELETE /comments`、`POST /videos/:id/like`、`POST /videos/:id/collect` | 弹幕、评论、点赞和收藏 |
| 个人资料库 | `GET/PUT/DELETE /users/me/history`、`GET/POST/DELETE /users/me/watch-later`、`GET /users/me/collections` | 历史、进度、稍后再看和合集 |
| 直播预约 | `GET/POST/DELETE /live-schedules`、`POST /live-schedules/:id/reserve` | 创建、取消和预约直播 |
| 直播互动 | `POST/GET/PUT /live*`、`GET /ws/live/:id`、`GET /ws/live-publish/:id` | 直播管理、弹幕、在线人数、礼物和推流状态 |
| 内部接口 | `GET /internal/engagement/videos/:id/stats`、`POST /internal/events/content-published` | 提供互动聚合并接收内容发布事件 |

## 3. 数据表归属

| 目标 Schema | 归属表/数据 | 说明 |
|---|---|---|
| `user_db` | `users`、`user_infos`、`follows`、`follow_groups`、`user_blocks`、`creator_membership_plans`、`creator_subscriptions`、`subscription_orders`、`chat_messages`、`notifications` | 用户身份和关系是唯一权威来源；其他服务只保存 `user_id` |
| `content_db` | `videos`、`video_collaborators`、`creator_daily_stats`、`video_daily_stats`、`site_banners`、`site_announcements`、`traffic_stats`、`dynamic_posts`、`media_assets` | 视频与审核状态是唯一权威来源；媒体文件位于独立存储卷/对象存储 |
| `engagement_db` | `danmakus`、`comments`、`likes`、`collects`、`comment_likes`、`video_collections`、`video_collection_items`、`watch_histories`、`watch_laters`、`live_rooms`、`live_schedules`、`live_reservations`、`live_likes`、`live_gifts`、`live_replays` | 互动和直播数据只保存外部 `user_id/video_id`，不跨 Schema 联表 |

## 4. 跨服务规则

1. 每个服务使用独立数据库账号，只拥有所属 Schema 的读写权限。
2. 禁止跨 Schema 联表；跨域读取使用 `/internal/*` HTTP API 或事件消息。
3. 内部接口设置不超过 2 秒的超时，并返回明确错误；列表/页面允许使用受控降级结果。
4. 写操作使用幂等键或唯一约束，避免网络重试造成重复订单、重复关注和重复互动计数。
5. 内容发布后通过事件通知 user-service 生成关注通知；事件失败可重试，不阻断视频主事务。
6. Gateway 只负责路由和统一入口，不绕过服务自己的权限校验。

## 5. 检查时的说明口径

- 已完成：服务边界、覆盖用例、目标接口、表归属、跨服务约束和 Kubernetes 目标图。
- 尚未宣称完成：三个服务的独立代码仓、独立镜像、独立 Schema 权限和微服务版 E2E。
- 下一阶段验收：三个服务可独立构建/测试/部署，无跨表访问，并能通过网关完成跨服务主链路。

## 5. 版本记录

| 版本 | 日期 | 变更 |
|---|---|---|
| v1.0 | 2026-08-30 | 中期检查提交版（当时为设计目标） |
| v1.1 | 2026-08-31 | `dynamic_posts` 与 `/dynamics` 接口归属由 engagement_db/engagement-service 移至 content_db/content-service：动态属内容发布行为，且实际实现位于 content-service（PR #108，engagement-service 未含动态路由，单一所有者无运行时冲突）；同时确认 content 侧新增 `media_assets`。中期材料已提交不再修改，以本仓库版为准 |
