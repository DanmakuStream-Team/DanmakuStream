# UC06 个人视频资料库管理（Personal Library）用例说明

> 版本：v1.0
> 关联图档：
> - 系统级：[sys-seq06-library.puml](../models/system/sys-seq06-library.puml)
> - 组件级：[comp-seq06-library.puml](../models/component/comp-seq06-library.puml)
> - 对象级（进度专项）：[obj-seq06-progress.puml](../models/object/obj-seq06-progress.puml)
> - 总用例图：[usecase-overview.puml](../models/usecase/usecase-overview.puml)

## 1. 业务目标

为登录用户提供统一的个人视频资料库入口，将**观看历史与进度**、**稍后再看**、**赞过的视频**、**收藏内容**、**下载内容**五大类聚合收纳，支持跨设备同步观看进度、随时继续观看、单项移除与整库清空，并在数据来源上明确区分服务端持久化（history / watchLater）与浏览器本地存储（liked / collections / downloads），确保：

1. 观看体验不中断：进入视频详情页即可从上次离开的位置恢复播放。
2. 稍后再看内容跨设备可用：服务端持久化，任意登录设备均可访问。
3. 互动与下载本地留存：点赞、收藏、下载优先以本机为准，避免污染其他设备状态。
4. 分类内容可被用户自主清理：单项删除与整库清空有二次确认，操作结果刷新后保持一致。
5. 权限与隔离严格：未登录用户不能访问他人资料库，视频被删除或未通过审核时自动在列表中隐藏。

## 2. 前置条件

| 编号 | 前置条件 | 失败时的表现 |
|---|---|---|
| P01 | 用户已成功登录并持有有效 JWT Token | 路由守卫将用户重定向至 `/login?redirect=<原路径>`，接口返回 HTTP 401 |
| P02 | 请求访问的历史/稍后再看关联视频存在、已通过审核、未软删除 | MySQL JOIN 时自动过滤；返回列表不包含该记录；单独查该 videoId 返回 404 |
| P03 | 稍后再看切换/保存进度的目标视频状态为 `approved` | Handler 中先查 Video，未通过则返回 404「视频不存在」，不写库 |
| P04 | 清空操作有用户主动的二次确认 | 前端 `ElMessageBox.confirm` 取消后不发起任何请求或本地写操作 |
| P05 | 本地存储（liked/collections/downloads）使用同域 `localStorage` | 隐私模式或跨域下降级为空数组，读不到历史但不报错 |

## 3. 主成功流程

### 3.1 子流程 A：记录与恢复观看进度

1. 用户在任意入口点击视频，进入 `/video/:id`。
2. 若 URL 带 `?t=<秒>`，播放器在 `loadedmetadata` 后跳转到指定秒。
3. 播放过程中，`VideoPlayer` 通过 `@timeupdate` 事件按节流（典型 5~10 秒）调用 `libraryApi.saveHistory(videoId, currentTime)`。
4. 后端 `SaveHistoryHandler` 校验 token、videoId、position ≥ 0，按 `video.Duration` 裁剪 position。
5. 使用 `OnConflict(user_id, video_id).DoUpdates(position, updated_at)` 做 UPSERT 写入 `watch_histories`。
6. 用户下次进入 `/me/history` 时看到按 `updated_at DESC` 的卡片，展示进度条与「继续观看」按钮。
7. 用户点击「继续观看」→ 路由跳转 `/video/:id?t=<position>` → 从断点继续播放。

### 3.2 子流程 B：加入/移除稍后再看

1. 用户在视频详情或任意入口点击「稍后再看」按钮。
2. 前端调 `libraryApi.toggleWatchLater(videoId)`。
3. 后端先查 Video（status=approved），不存在返回 404。
4. 查询 `watch_laters` 是否存在 (user, video)：
   - 存在则 `Unscoped Delete` → 返回 `{saved: false}`；
   - 不存在则 `Create` → 返回 `{saved: true}`。
5. 前端按钮 UI 同步切换为实心/空心状态；资料库 `/me/watchlater` 中按 `created_at DESC` 列出。

### 3.3 子流程 C：资料库列表浏览与单项删除

1. 用户访问 `/me/:kind`，`kind ∈ {history, watchlater, liked, collections, downloads}`。
2. 路由守卫验证登录后，`UserLibraryPage.loadRecords()` 按来源分派：
   - `history` / `watchlater` → HTTP 拉取服务端分页（pageSize=100）。
   - `liked` / `collections` / `downloads` → `getUserLibraryRecords(kind)` 读 localStorage。
3. 列表展示：封面、标题、简介、作者昵称、浏览/点赞、保存时间；history 额外展示 `el-progress` 进度条。
4. 对任一条目点击「×」按钮 → `removeRecord(videoId)`：
   - history/watchlater → DELETE 接口单条删除 → 重拉列表。
   - liked/collections/downloads → 本地 filter 后写回 → 重读本地渲染。

### 3.4 子流程 D：整库清空

1. 用户在页面头部点击「清空」按钮（无记录时 disabled）。
2. `ElMessageBox.confirm("确定清空<分类名>吗？")` 给出二次确认。
3. 用户点「清空」后按 kind 分派：
   - history → `libraryApi.clearHistory()` → 后端 `Unscoped().Where("user_id=?").Delete(&WatchHistory{})` → 返回 `{cleared:true}`。
   - watchlater → `libraryApi.clearWatchLater()` → 同上处理 `WatchLater` 表。
   - liked/collections/downloads → `clearUserLibraryRecords(kind)` → `localStorage.setItem(key, "[]")`。
4. 前端再次 `loadRecords()`，展示空态页（`el-empty`）。

## 4. 异常或备选流程

| 编号 | 场景 | 处理策略 | 面向用户的表现 |
|---|---|---|---|
| E01 | 保存进度时视频已下架/未审核 | 后端 First Video 失败 → 返回 404 | 前端 `skipErrorMessage`，静默忽略，不阻断播放 |
| E02 | position 超过 duration | 后端 `position = min(position, duration)` | 进度百分比按 duration 封顶（≤100%） |
| E03 | 清空接口网络失败或 DB 失败 | 后端 5xx + message「历史记录清空失败」 | 前端保留列表，弹出 `ElMessage.error`，不置空 |
| E04 | localStorage 被禁用或额度超限 | `try/catch` 包裹读写 | 返回空数组或跳过写入；页面展示空态且不弹错 |
| E05 | 多端并发写入同一视频进度 | 唯一索引 (user_id, video_id) + updated_at 以服务端最新覆盖 | 以最后一次写入的 position 为准，不做合并 |
| E06 | 用户未登录却通过 URL 直连 `/me/history` | Vue Router beforeEach 中 `requiresAuth=true` | 重定向 `/login?redirect=/me/history`，登录后自动回到原页 |
| E07 | 用户尝试读他人资料库（攻击接口） | 所有接口从 `c.GetUint(CtxKeyUserID)` 拿 uid，不接受 path/query uid | 返回 401 或只看自己数据，绝不越权 |
| E08 | 重复点击稍后再看按钮（防抖失效） | 后端先查再写，UNIQUE 索引兜底重复 Create 报错 | 前端 UI 状态以前后端最后一次 `saved` 为准 |
| E09 | 下载接口 Blob 失败 | `downloadVideo()` catch 分支弹消息 | 不写入 downloads 本地列表，避免脏数据 |
| E10 | 非法 page/pageSize | `libraryPage()` 默认 1 和 20，pageSize 上限 100 | 接口不报错，按修正值返回 |

## 5. 验收标准（可验证结果）

### 5.1 功能正确性

| AC 编号 | 断言 | 验证方式 |
|---|---|---|
| AC01 | 播放视频 ≥ 10s 后刷新页面再进入「继续观看」，自动从上次 position ± 3s 内恢复 | Playwright E2E：进入视频 → 模拟跳 120s → 等待节流 save → 关页重开 → 断言 `video.currentTime` 在 117~123s |
| AC02 | 进度不超过 duration；保存 position=9999s，duration=600s，实际持久化为 600s | API 测试：PUT history → 再 GET history/:id → 断言 position=600，progress=100 |
| AC03 | 同一视频反复加入稍后再看不产生重复 DB 行 | 连续 3 次 toggle → 直查 MySQL COUNT(user_id, video_id) = 0 或 1 |
| AC04 | 清空历史后 /users/me/history 返回 `list=[]` 且 `total=0` | DELETE history → GET history 断言空列表；同样覆盖 watchLater |
| AC05 | 删除单条历史后刷新列表，该 videoId 不再出现 | DELETE history/:id → GET history list → `!list.find(r => r.video.id===id)` |
| AC06 | 未审核/已删除视频不出现在历史与稍后再看列表中 | DB 直插一条 Video.status='rejected' 的 WatchHistory → GET history 断言不出现在 list |
| AC07 | 权限隔离：A 的 token 不能读 B 的 history；无 token 请求返回 401 | 两组 API 请求：他人 uid 间接尝试（实际上 uid 从 token 解出）；缺 Authorization 头 → 断言 401 |
| AC08 | 本地 liked/collections/downloads：关闭浏览器再打开，条目与排序保持 | Playwright：写入 localStorage → `context.close()` 重启浏览器 → 再进 `/me/liked` → 断言内容一致 |
| AC09 | 清空按钮二次确认取消：列表不变 | 模拟 ElMessageBox cancel → 断言列表长度与 GET 前一致，无额外 DELETE 请求 |
| AC10 | 进度百分比正确：position/duration * 100，整数截断 | position=180, duration=600 → progress=30；position>duration → progress=100 |

### 5.2 性能与体验

1. 列表首屏（history/watchLater 100 条）接口响应 ≤ 500ms，P95 ≤ 1s。
2. `saveHistory` 采用前端节流 + 后端 UPSERT，单视频 1 小时播放不产生 > 720 次 DB 写（建议节流 5s，约 720 次；10s 约 360 次）。
3. 清空操作需为全量 Unscoped DELETE，不走逐行软删除，确保行数多时响应稳定。
4. 空态页展示对应文案与插图，不出现白屏或报错堆栈。

### 5.3 安全与合规

1. 所有写接口（PUT/POST/DELETE）校验 Bearer Token，未登录 401。
2. 写库仅用 token 解析出的 user_id，禁止接收外部 user_id 参数。
3. SQL 通过 GORM 参数化拼接，history 与 watchLater 不出现原始字符串注入入口。
4. 本地存储 key 加 `danmaku:user-library:` 前缀，避免与其他应用冲突。

## 6. 代码对应

| 层级 | 模块 | 绝对路径 |
|---|---|---|
| 后端 Handler（历史记录与稍后再看全部 CRUD） | `library_handler.go` | [library_handler.go](../backend/internal/handler/v1/user/library_handler.go) |
| 后端路由注册 | `api/main.go` 中 `/users/me/history*` 与 `/users/me/watch-later*` | [main.go](../backend/api/main.go#L125-L135) |
| 后端模型（WatchHistory / WatchLater） | `models.go` | [models.go](../backend/internal/model/mysql/models.go#L335-L348) |
| 前端 API 封装 | `library.ts` | [library.ts](../frontend/src/api/library.ts) |
| 前端页面（列表 + 删除 + 清空 + 继续观看跳转） | `UserLibraryPage.vue` | [UserLibraryPage.vue](../frontend/src/pages/user/UserLibraryPage.vue) |
| 前端路由守卫（requiresAuth） | `router/index.ts` | [index.ts](../frontend/src/router/index.ts#L19) |
| 本地存储工具（liked/collections/downloads） | `userLibrary.ts` | [userLibrary.ts](../frontend/src/utils/userLibrary.ts) |
| 播放器 timeupdate 触发进度保存位置 | `VideoPlayer.vue` 的 `@timeupdate="emitTime"` 与上层 VideoDetailPage 调用链 | [VideoPlayer.vue](../frontend/src/components/common/VideoPlayer.vue#L18) |
