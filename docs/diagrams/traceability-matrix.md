# DanmakuStream 文档追溯表

> 版本：v1.0
> 编制日期：2026-08-26
> 编制人：E
> 范围：UC06（个人视频资料库）三层顺序图、总用例图、部署图、架构组件图

---

## 一、图文件总览

| 序号 | 图名称 | 图文件名 (.puml) | SVG 文件 | PNG 文件 | 输出目录 |
|------|--------|-------------------|----------|----------|----------|
| 1 | UC06 系统级顺序图 | `SYS-SEQ06.puml` | `SYS-SEQ06.svg` | `SYS-SEQ06.png` | `docs/diagrams/` |
| 2 | UC06 组件级顺序图 | `COMP-SEQ06.puml` | `COMP-SEQ06.svg` | `COMP-SEQ06.png` | `docs/diagrams/` |
| 3 | UC06 对象级顺序图 | `OBJ-SEQ06.puml` | `OBJ-SEQ06.svg` | `OBJ-SEQ06.png` | `docs/diagrams/` |
| 4 | 总用例图 (UC01~UC14) | `usecase-overview.puml` | `usecase-overview.svg` | `usecase-overview.png` | `docs/diagrams/` |
| 5 | 改造前单体部署图 | `deployment-monolith.puml` | `deployment-monolith.svg` | `deployment-monolith.png` | `docs/diagrams/` |
| 6 | 改造后K8s部署图 | `deployment-k8s.puml` | `deployment-k8s.svg` | `deployment-k8s.png` | `docs/diagrams/` |
| 7 | 架构组件图 | `component-architecture.puml` | `component-architecture.svg` | `component-architecture.png` | `docs/diagrams/` |

---

## 二、UC06 个人视频资料库追溯

### 2.1 UC06 三层顺序图 → 代码追溯表

| 功能点 | 涉及图编号 | 对应用例场景 | 对应前端代码 | 对应后端代码 | 数据库模型 (Entity) | 测试文件 (如果有) | 实现状态 |
|--------|------------|--------------|-------------|-------------|--------------------|-------------------|---------|
| **查看历史记录** | SYS-SEQ06, COMP-SEQ06, OBJ-SEQ06 | 用户进入"我的资料库"→历史记录页 | [UserLibraryPage.vue](file:///d:/DanmakuStream/frontend/src/pages/user/UserLibraryPage.vue#L115-L117) `loadRecords()` <br> [userLibrary.ts](file:///d:/DanmakuStream/frontend/src/utils/userLibrary.ts#L33-L35) `getUserLibraryRecords()` | ⚠️ // TODO: 代码中未找到后端 History API <br> 后端未提供 `/api/v1/library/history` 端点 | ⚠️ 缺失 `history_records` 表 (models.go 未定义) | 暂未找到 | **前端本地 localStorage 已实现，后端未实现** |
| **保存播放进度** | SYS-SEQ06, COMP-SEQ06, OBJ-SEQ06 | 视频页面 onUnmounted 时自动保存 | [VideoDetailPage.vue](file:///d:/DanmakuStream/frontend/src/pages/video/VideoDetailPage.vue#L591-L596) `saveHistory()` <br> [userLibrary.ts](file:///d:/DanmakuStream/frontend/src/utils/userLibrary.ts#L37-L45) `upsertUserLibraryRecord()` | ⚠️ // TODO: 代码中未找到后端 Progress API <br> 后端未提供 `/api/v1/library/progress` 端点 | ⚠️ 缺失 `play_progress` 表 | 暂未找到 | **前端本地 localStorage 已实现，后端未实现** |
| **添加/移除稍后再看** | SYS-SEQ06, COMP-SEQ06, OBJ-SEQ06 | 用户在列表中管理稍后再看 | ⚠️ // TODO: 前端 UserLibraryPage 实际提供了 `removeRecord()` 单条删除 + `clearRecords()` 清空，但 WatchLater 独立入口在前端未单独区分（与历史记录、赞过、收藏、下载并列4类） | ⚠️ // TODO: 代码中未找到 WatchLater 后端 API | ⚠️ 缺失 `watch_later` 表 | 暂未找到 | **前端4类本地库(history/liked/collections/downloads)已实现，WatchLater 独立后端未实现** |
| **清空历史记录** | SYS-SEQ06, COMP-SEQ06, OBJ-SEQ06 | 用户点击"清空"按钮二次确认后清空 | [UserLibraryPage.vue](file:///d:/DanmakuStream/frontend/src/pages/user/UserLibraryPage.vue#L124-L132) `clearRecords()` <br> [userLibrary.ts](file:///d:/DanmakuStream/frontend/src/utils/userLibrary.ts#L51-L53) `clearUserLibraryRecords()` | ⚠️ // TODO: 代码中未找到后端清空历史 API | ⚠️ 依赖 history_records / play_progress 表 | 暂未找到 | **前端本地 localStorage 已实现，后端未实现** |
| **点赞视频 (进入资料库-赞过)** | SYS-SEQ06, COMP-SEQ06, OBJ-SEQ06 | 详情页点赞后自动进入"赞过的视频"库 | [VideoDetailPage.vue](file:///d:/DanmakuStream/frontend/src/pages/video/VideoDetailPage.vue#L450-L461) `toggleLike()` <br> [video.ts](file:///d:/DanmakuStream/frontend/src/api/video.ts#L38-L40) `videoApi.like()` | [video_handler.go](file:///d:/DanmakuStream/backend/internal/handler/v1/video/video_handler.go#L376-L438) `LikeHandler()` | [models.go](file:///d:/DanmakuStream/backend/internal/model/mysql/models.go#L181-L185) `Like{UserID, VideoID}` | `scripts/test_core_apis.sh` 有提及 | **后端已完整实现（GORM事务 + 计数原子增减）** |
| **收藏视频 (进入资料库-收藏)** | SYS-SEQ06, COMP-SEQ06, OBJ-SEQ06 | 详情页收藏后自动进入"收藏内容"库 | [VideoDetailPage.vue](file:///d:/DanmakuStream/frontend/src/pages/video/VideoDetailPage.vue#L463-L474) `toggleCollect()` <br> [video.ts](file:///d:/DanmakuStream/frontend/src/api/video.ts#L41-L43) `videoApi.collect()` | [video_handler.go](file:///d:/DanmakuStream/backend/internal/handler/v1/video/video_handler.go#L440-L502) `CollectHandler()` | [models.go](file:///d:/DanmakuStream/backend/internal/model/mysql/models.go#L187-L191) `Collect{UserID, VideoID}` | `scripts/test_core_apis.sh` 有提及 | **后端已完整实现（GORM事务 + 计数原子增减）** |
| **下载视频 (进入资料库-下载)** | SYS-SEQ06 | 详情页下载后本地记录 | [UserLibraryPage.vue](file:///d:/DanmakuStream/frontend/src/pages/user/UserLibraryPage.vue#L134-L147) <br> [video_handler.go](file:///d:/DanmakuStream/backend/internal/handler/v1/video/video_handler.go#L198-L265) `DownloadHandler()` | [video_handler.go](file:///d:/DanmakuStream/backend/internal/handler/v1/video/video_handler.go#L198-L265) | 下载记录仅存前端 localStorage，后端无 downloads 表 | 暂未找到 | **下载接口已实现，资料库记录仅本地** |

### 2.2 UC06 前端-后端 API 路径对齐情况

> 关键点：UC06 当前为"前端本地存储优先"架构，点赞/收藏有后端API但列表查询无后端分页。

| 资料库子分类 (UserLibraryKind) | 前端本地存储 key | 对应后端 API 是否存在 | 规划中的后端 API 路径 |
|--------------------------------|----------------|---------------------|----------------------|
| `history` (历史记录+进度) | `danmaku:user-library:history` | ❌ 缺失 | `GET /api/v1/library/history?page=1&pageSize=20` <br> `DELETE /api/v1/library/history` (清空) |
| `liked` (赞过的视频) | `danmaku:user-library:liked` | ⚠️ 只有点赞动作 API，无列表查询 API <br> `POST /videos/:id/like` 已实现 | `GET /api/v1/library/liked` (分页查已点赞视频) |
| `collections` (收藏内容) | `danmaku:user-library:collections` | ⚠️ 只有收藏动作 API，无列表查询 API <br> `POST /videos/:id/collect` 已实现 | `GET /api/v1/library/collected` (分页查已收藏视频) |
| `downloads` (下载内容) | `danmaku:user-library:downloads` | ⚠️ 只有下载接口，无列表 API | `GET /api/v1/library/downloads` (可选) |

---

## 三、总用例图 (UC01~UC14) → 代码追溯表

| 用例编号 | 用例名称 | 对应参与者 | 对应用图 | 对应前端代码 | 对应后端代码 (Controller/Handler) | 数据库实体 | 备注 |
|----------|---------|-----------|---------|------------|-------------------------------|-----------|------|
| **UC01** | 用户注册与登录 | 全部4类 | usecase-overview | [LoginPage.vue](file:///d:/DanmakuStream/frontend/src/pages/home/LoginPage.vue) <br> [RegisterPage.vue](file:///d:/DanmakuStream/frontend/src/pages/home/RegisterPage.vue) | [auth_handler.go](file:///d:/DanmakuStream/backend/internal/handler/v1/auth/auth_handler.go) `LoginHandler/RegisterHandler/MeHandler` | `User` 实体 | ✅ 已实现 JWT Token 鉴权 |
| **UC02** | 视频浏览与搜索 | 普通用户/创作者 | usecase-overview | [HomePage.vue](file:///d:/DanmakuStream/frontend/src/pages/home/HomePage.vue) | [video_handler.go](file:///d:/DanmakuStream/backend/internal/handler/v1/video/video_handler.go#L47-L63) `ListHandler` <br> [list_logic.go](file:///d:/DanmakuStream/backend/internal/logic/video/list_logic.go) | `Video` 实体 | ✅ 支持热/日期/点赞/收藏4种排序 + 关键词搜索 + 分类标签过滤 |
| **UC03** | 视频上传与管理 | 创作者 | usecase-overview | [VideoUploadPage.vue](file:///d:/DanmakuStream/frontend/src/pages/video/VideoUploadPage.vue) | [video_handler.go](file:///d:/DanmakuStream/backend/internal/handler/v1/video/video_handler.go#L83-L196) `UploadHandler / UpdateHandler / DeleteHandler` | `Video, VideoCollaborator` | ✅ 已实现：封面设置、FFmpeg HLS 转码异步、共创成员管理 |
| **UC04** | 视频播放 | 普通用户/创作者 | usecase-overview, SYS-SEQ06 | [VideoDetailPage.vue](file:///d:/DanmakuStream/frontend/src/pages/video/VideoDetailPage.vue#L5-L14) `VideoPlayer 组件` | [video_handler.go](file:///d:/DanmakuStream/backend/internal/handler/v1/video/video_handler.go#L65-L81) `DetailHandler` <br> [detail_logic.go](file:///d:/DanmakuStream/backend/internal/logic/video/detail_logic.go) | `Video` (ViewCount 原子 +1) | ✅ 已实现 HLS m3u8 播放、观看量统计 |
| **UC05** | 弹幕互动 | 普通用户/创作者 | usecase-overview | [VideoPlayer.vue](file:///d:/DanmakuStream/frontend/src/components/common/VideoPlayer.vue) <br> [VideoDetailPage.vue](file:///d:/DanmakuStream/frontend/src/pages/video/VideoDetailPage.vue#L109-L166) | [danmaku_handler.go](file:///d:/DanmakuStream/backend/internal/handler/v1/danmaku/danmaku_handler.go) `SendHandler, UploadAdvancedHandler, ListHandler` | `Danmaku` | ✅ 已实现普通/高级弹幕、.danmaku 批量上传 |
| **UC06** | 个人视频资料库 | 普通用户/创作者 | SYS-SEQ06, COMP-SEQ06, OBJ-SEQ06, usecase-overview | [UserLibraryPage.vue](file:///d:/DanmakuStream/frontend/src/pages/user/UserLibraryPage.vue) <br> [userLibrary.ts](file:///d:/DanmakuStream/frontend/src/utils/userLibrary.ts) | ⚠️ 后端 API 大部分缺失，仅 LikeHandler/CollectHandler/DownloadHandler 有动作接口 | 缺失 `history_records, play_progress, watch_later` | 详见 2.1 表 |
| **UC07** | 评论与互动 | 普通用户/创作者 | usecase-overview, SYS-SEQ06 (点赞收藏) | [CommentItem.vue](file:///d:/DanmakuStream/frontend/src/components/common/CommentItem.vue) + `comment_store.ts` | [comment_handler.go](file:///d:/DanmakuStream/backend/internal/handler/v1/comment/comment_handler.go) `CreateHandler, DeleteHandler, LikeHandler, ListHandler` | `Comment, CommentLike, Like, Collect` | ✅ 评论/回复/删除/评论点赞已实现 |
| **UC08** | 关注与社交关系 | 普通用户/创作者 | usecase-overview | [UserProfilePage.vue](file:///d:/DanmakuStream/frontend/src/pages/user/UserProfilePage.vue) <br> [SubscriptionPage.vue](file:///d:/DanmakuStream/frontend/src/pages/user/SubscriptionPage.vue) | [user_handler.go](file:///d:/DanmakuStream/backend/internal/handler/v1/user/user_handler.go#L221-L299) `FollowHandler, FollowingListHandler, ProfileHandler` | `Follow, User` (FanCount/FollowCount 原子) | ✅ 已实现关注双向计数 |
| **UC09** | 直播与预约 | 创作者/普通用户 | usecase-overview, deployment-monolith, deployment-k8s | [LiveRoomPage.vue](file:///d:/DanmakuStream/frontend/src/pages/live/LiveRoomPage.vue) <br> [LiveListPage.vue](file:///d:/DanmakuStream/frontend/src/pages/live/LiveListPage.vue) | [live_handler.go](file:///d:/DanmakuStream/backend/internal/handler/v1/live/live_handler.go) <br> [schedule_handler.go](file:///d:/DanmakuStream/backend/internal/handler/v1/live/schedule_handler.go) <br> [ws_handler.go](file:///d:/DanmakuStream/backend/internal/handler/ws/ws_handler.go) WebSocket | `LiveRoom, LiveSchedule, LiveReservation` | ✅ 已实现 SRS 对接 + WS 实时弹幕 + 预约提醒 |
| **UC10** | 视频合集管理 | 创作者/普通用户 | usecase-overview | [VideoDetailPage.vue](file:///d:/DanmakuStream/frontend/src/pages/video/VideoDetailPage.vue#L168-L211) (合集面板) | [collection_handler.go](file:///d:/DanmakuStream/backend/internal/handler/v1/collection/collection_handler.go) `CreateHandler, AddVideoHandler, RemoveVideoHandler, MineHandler, VideoCollectionsHandler, AddCollaboratorHandler, RemoveCollaboratorHandler` | `VideoCollection, VideoCollectionItem` | ✅ 已实现创建/增删视频/详情/我的合集 |
| **UC12** | 创作者中心 | 创作者 | usecase-overview | [CreatorDashboardPage.vue](file:///d:/DanmakuStream/frontend/src/pages/user/CreatorDashboardPage.vue) | [user_handler.go](file:///d:/DanmakuStream/backend/internal/handler/v1/user/user_handler.go#L525-L606) `MeVideosHandler` (按 status=pending/approved/rejected 筛选) | `Video` (Status 字段) | ✅ 已实现按视频状态筛选列表 |
| **UC13** | 内容审核 | 审核员 (moderator) | usecase-overview | [AdminVideosPage.vue](file:///d:/DanmakuStream/frontend/src/pages/admin/AdminVideosPage.vue) + `AdminDanmakuPage.vue` | [video_handler.go](file:///d:/DanmakuStream/backend/internal/handler/v1/video/video_handler.go#L504-L646) `AdminListHandler, AdminUpdateStatusHandler` <br> [danmaku_handler.go](file:///d:/DanmakuStream/backend/internal/handler/v1/danmaku/danmaku_handler.go) `AdminListHandler, BlockHandler` | `Video.Status` ('pending','approved','rejected'), `Danmaku.Blocked` | ✅ 已实现 StaffMiddleware 审核员 + 管理员角色准入 |
| **UC14** | 系统后台管理 | 管理员 (admin) | usecase-overview, deployment-k8s | [AdminDashboardPage.vue](file:///d:/DanmakuStream/frontend/src/pages/admin/AdminDashboardPage.vue) <br> [AdminUsersPage.vue](file:///d:/DanmakuStream/frontend/src/pages/admin/AdminUsersPage.vue) <br> [AdminInfrastructurePage.vue](file:///d:/DanmakuStream/frontend/src/pages/admin/AdminInfrastructurePage.vue) | [admin_handler.go](file:///d:/DanmakuStream/backend/internal/handler/v1/admin/admin_handler.go) `InfrastructureHandler, UserListHandler, UpdateUserRoleHandler, BannerListHandler, Banner CRUD, Announcement CRUD` | `User.Role, SiteBanner, SiteAnnouncement, TrafficStat` | ✅ 已实现 AdminMiddleware 准入 + 用户角色切换 + Banner/公告 CRUD + 基础设施指标 |

---

## 四、部署图 → 代码追溯表

| 图编号 | 图名称 | 对应配置文件 | 核心端口/协议 | 关键资源引用 | 说明 |
|--------|--------|------------|--------------|------------|------|
| deployment-monolith | **改造前单体部署图** | [docker-compose.yml](file:///d:/DanmakuStream/docker-compose.yml) | Nginx前端:80, Go:8888, MySQL:3306, SRS:1935/1985/8081 | `backend/Dockerfile` <br> `frontend/Dockerfile` <br> `scripts/init.sql` 初始化脚本 | 单体 Go Gin 包含全部业务逻辑，单数据库 `danmakustream` 共用所有表 |
| deployment-k8s | **改造后 K8s 部署图** | ⚠️ // TODO: 代码仓库中暂未提供 K8s YAML 文件（本次为设计规划） | Ingress:443, Gateway:80, user-svc:8081, content-svc:8082, engagement-svc:8083, SRS:1935(NodePort)/1985/8080, MySQL:3306 | 设计引用: ConfigMap, Secret, StatefulSet(mysql), Deployment×6, Service×7, Ingress×1, PV×2, liveness/readiness探针已按最佳实践配置 | 三个微服务独立 Deployment，MySQL 内部拆 3 个 Schema 做数据隔离 |

---

## 五、架构组件图 → 代码追溯表

| 图编号 | 架构模块 | 对应包/代码位置 | 通信方式 | 数据库 Schema | 关键发现/备注 |
|--------|---------|--------------|---------|--------------|------------|
| component-architecture | **user-service :8081** | `handler/v1/auth`, `handler/v1/user` + 中间件 `AuthMiddleware`, `StaffMiddleware` | 对外: 经 Gateway HTTP/JSON <br> 对内: content/engagement 同步 HTTP 调用 | `user_schema` (users, follows) | ✅ 鉴权、用户资料、关注关系、角色检查可独立部署 |
| component-architecture | **content-service :8082** | `handler/v1/video`, `handler/v1/collection`, `handler/v1/media`, `logic/video/*` | 对外: Gateway HTTP 路由 <br> 内部: 调用 user-service 获取作者信息 | `content_schema` (videos, video_collections*, video_collaborators, dynamic_posts, site_banners, site_announcements) | ✅ 包含 FFmpeg 异步转码逻辑，需挂载 RWX 共享卷存视频 |
| component-architecture | **engagement-service :8083** | `handler/v1/danmaku`, `handler/v1/comment`, `handler/v1/notification`, `handler/v1/live`, `handler/v1/dynamic`, `video_handler.go Like/Collect 部分` | 对外: Gateway HTTP + WebSocket (/ws/live/:id) <br> 对内: 调用 user-service 校验用户身份 | `engagement_schema` (danmakus, comments, comment_likes, likes, collects, **UC06三表-待建**, live_*, notifications, traffic_stats) | ⚠️ UC06 history/progress/watchLater 归属此服务，但代码未实现 API 与表 |
| component-architecture | **API Gateway 统一入口** | 对应单体 `main.go` 路由层可改造为 Gateway | HTTPS 终止、JWT 验证、路由 `/api/v1/auth/*` → user-svc <br> `/api/v1/videos*` → content-svc <br> `/api/v1/danmaku*` → engagement-svc | N/A | ✅ 可按现有 `main.go` 路由前缀直接拆分 |
| component-architecture | **数据隔离 (三独立 Schema)** | 当前 `svc.NewServiceContext()` 仅 1 个 gorm DB，拆分后需 3 个 DB 连接池，各自使用不同账号限定 Schema 权限 | JDBC/GORM 连接串区分 user_schema / content_schema / engagement_schema | MySQL 8.0 | 🔧 设计建议：为每个服务创建独立 MySQL 用户，通过 GRANT 限制只能访问自己的 Schema，避免跨服务越权写 |

---

## 六、缺失功能汇总 (TODO 清单)

> UC06 后端补齐建议优先级

| 优先级 | 缺失项 | 涉及文件补全位置 | 需要新增的表 | 需要新增的 API 端点 |
|--------|-------|----------------|------------|-------------------|
| P0 | 历史记录后端 API + 播放进度 | 新建 `handler/v1/engagement/library_handler.go` + service/repo 层 | `history_records(id, user_id, video_id, progress, saved_at, deleted_at)` <br> `play_progress(id, user_id, video_id, progress, saved_at)` | GET `/library/history?page=&pageSize=` <br> POST `/library/progress {videoId, progress}` <br> DELETE `/library/history` (清空) <br> DELETE `/library/history/:videoId` |
| P0 | 稍后再看后端 API | 同上 service/repo 层扩展 | `watch_later(id, user_id, video_id, added_at)` | POST `/library/watch-later {videoId}` <br> DELETE `/library/watch-later/:videoId` <br> GET `/library/watch-later?page=` |
| P1 | 点赞/收藏的"我已点赞/已收藏视频列表"分页 API | 在 LikeHandler / CollectHandler 旁追加 List 方法 | 复用 likes, collects 表，Join videos 表做分页 | GET `/library/liked?page=&pageSize=` <br> GET `/library/collected?page=&pageSize=` |
| P1 | K8s 部署 YAML 文件集 | 新建 `k8s/` 目录 | N/A | deployment/service/ingress/configmap/secret/pvc 共 15+ 个 YAML |
| P2 | UC06 单元测试 | 新建 `backend/internal/logic/engagement/*_test.go` | N/A | 针对 HistoryService / ProgressService 做 mock DB 测试 |

---

## 七、测试文件引用 (当前仓库存在)

| 测试工具/脚本 | 路径 | 覆盖范围 |
|--------------|------|---------|
| Shell E2E 脚本 | [scripts/test_core_apis.sh](file:///d:/DanmakuStream/scripts/test_core_apis.sh) | UC01 登录注册、UC02 视频列表、UC03 上传、UC05 弹幕、UC07 评论点赞等核心 API |
| SQL 初始化脚本 | [scripts/init.sql](file:///d:/DanmakuStream/scripts/init.sql) | 建表 + 初始数据（若已定义） |
| GORM AutoMigrate | [service_context.go](file:///d:/DanmakuStream/backend/internal/svc/service_context.go#L24-L44) | 启动时自动建表（当前 20+ 个实体已在此声明） |

---

## 八、图文件快速链接

| 图 | PUML (可编辑源) | SVG (可缩放矢量) | PNG (截图) |
|----|----------------|-----------------|-----------|
| SYS-SEQ06 (UC06系统顺序图) | [SYS-SEQ06.puml](file:///d:/DanmakuStream/docs/diagrams/SYS-SEQ06.puml) | [SYS-SEQ06.svg](file:///d:/DanmakuStream/docs/diagrams/SYS-SEQ06.svg) | [SYS-SEQ06.png](file:///d:/DanmakuStream/docs/diagrams/SYS-SEQ06.png) |
| COMP-SEQ06 (UC06组件顺序图) | [COMP-SEQ06.puml](file:///d:/DanmakuStream/docs/diagrams/COMP-SEQ06.puml) | [COMP-SEQ06.svg](file:///d:/DanmakuStream/docs/diagrams/COMP-SEQ06.svg) | [COMP-SEQ06.png](file:///d:/DanmakuStream/docs/diagrams/COMP-SEQ06.png) |
| OBJ-SEQ06 (UC06对象顺序图) | [OBJ-SEQ06.puml](file:///d:/DanmakuStream/docs/diagrams/OBJ-SEQ06.puml) | [OBJ-SEQ06.svg](file:///d:/DanmakuStream/docs/diagrams/OBJ-SEQ06.svg) | [OBJ-SEQ06.png](file:///d:/DanmakuStream/docs/diagrams/OBJ-SEQ06.png) |
| 总用例图 (13个UC) | [usecase-overview.puml](file:///d:/DanmakuStream/docs/diagrams/usecase-overview.puml) | [usecase-overview.svg](file:///d:/DanmakuStream/docs/diagrams/usecase-overview.svg) | [usecase-overview.png](file:///d:/DanmakuStream/docs/diagrams/usecase-overview.png) |
| 改造前单体部署图 | [deployment-monolith.puml](file:///d:/DanmakuStream/docs/diagrams/deployment-monolith.puml) | [deployment-monolith.svg](file:///d:/DanmakuStream/docs/diagrams/deployment-monolith.svg) | [deployment-monolith.png](file:///d:/DanmakuStream/docs/diagrams/deployment-monolith.png) |
| 改造后K8s部署图 | [deployment-k8s.puml](file:///d:/DanmakuStream/docs/diagrams/deployment-k8s.puml) | [deployment-k8s.svg](file:///d:/DanmakuStream/docs/diagrams/deployment-k8s.svg) | [deployment-k8s.png](file:///d:/DanmakuStream/docs/diagrams/deployment-k8s.png) |
| 架构组件图 | [component-architecture.puml](file:///d:/DanmakuStream/docs/diagrams/component-architecture.puml) | [component-architecture.svg](file:///d:/DanmakuStream/docs/diagrams/component-architecture.svg) | [component-architecture.png](file:///d:/DanmakuStream/docs/diagrams/component-architecture.png) |
