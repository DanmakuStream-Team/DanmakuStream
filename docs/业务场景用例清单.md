# DanmakuStream 业务场景（用例）清单

## UC01 用户注册、登录与资料维护

- **参与者：** 游客、普通用户。
- **触发条件：** 游客希望创建账号或已有用户希望登录并维护个人资料。
- **前置条件：** 用户能够访问系统；注册昵称未被占用；登录账号有效。
- **代码对应：** `backend/internal/handler/v1/auth/auth_handler.go`、`backend/internal/handler/v1/user/user_handler.go`、`frontend/src/pages/home/LoginPage.vue`、`frontend/src/pages/home/RegisterPage.vue`、`frontend/src/pages/user/UserProfilePage.vue`。

### 主成功流程

1. 游客填写昵称和密码并提交注册。
2. 系统校验输入并创建普通用户账号。
3. 用户使用账号登录，系统验证凭据并签发登录身份。
4. 用户进入个人主页修改昵称、头像或简介。
5. 系统保存资料，并在后续页面展示最新信息。

### 备选或异常流程

- 昵称已存在或输入不合法时，系统拒绝注册并提示原因。
- 密码错误时，系统拒绝登录且不签发身份凭据。
- 登录身份无效或过期时，系统要求用户重新登录。
- 头像或资料不符合要求时，系统拒绝保存。

### 可验证结果

- 注册后存在唯一用户记录，正确密码可以登录，错误密码不能登录。
- 登录用户可以访问受保护功能，未登录用户不能访问。
- 修改后的个人资料在刷新和重新登录后保持一致。

## UC02 视频发现、搜索与播放

- **参与者：** 游客、普通用户。
- **触发条件：** 用户希望浏览、搜索并观看视频。
- **前置条件：** 系统中存在审核通过且媒体资源可访问的视频。
- **代码对应：** `backend/internal/handler/v1/video/video_handler.go`、`backend/internal/logic/video/list_logic.go`、`backend/internal/logic/video/detail_logic.go`、`frontend/src/pages/home/HomePage.vue`、`frontend/src/pages/video/VideoDetailPage.vue`、`frontend/src/components/common/VideoPlayer.vue`。

### 主成功流程

1. 用户进入首页浏览视频，或输入关键词搜索视频和创作者。
2. 系统返回符合条件且允许公开播放的视频。
3. 用户选择目标视频进入详情页。
4. 系统加载视频、作者、互动统计和播放地址。
5. 播放器加载媒体并开始播放。

### 备选或异常流程

- 没有搜索结果时，系统显示空结果。
- 视频不存在、未通过审核或已删除时，系统拒绝播放。
- 视频仍在处理或媒体加载失败时，系统显示明确状态并有限重试。

### 可验证结果

- 用户能够从列表或搜索结果进入正确的视频详情页。
- 审核通过的视频可以播放，其他视频不能公开播放。
- 播放失败时系统返回明确、稳定且可重复验证的结果。

## UC03 创作者投稿与状态跟踪

- **参与者：** 创作者。
- **触发条件：** 创作者希望提交新视频供平台审核和发布。
- **前置条件：** 创作者已登录；视频文件满足上传要求；存储和转码环境可用。
- **代码对应：** `backend/internal/handler/v1/video/video_handler.go`、`backend/internal/handler/v1/media/media_handler.go`、`frontend/src/pages/video/VideoUploadPage.vue`、`frontend/src/pages/user/CreatorDashboardPage.vue`。

### 主成功流程

1. 创作者选择视频文件并填写标题、简介等投稿信息。
2. 创作者上传封面，或选择由系统生成封面。
3. 系统接收文件并建立投稿记录。
4. 系统处理视频并生成可播放媒体资源。
5. 投稿进入待审核状态，创作者在创作中心查看状态。

### 备选或异常流程

- 未上传封面时，系统尝试截取视频首帧。
- 创作者取消上传时，系统终止流程且不保留无效待审记录。
- 文件格式或大小不合法时，系统拒绝上传。
- 转码失败或网络中断时，系统记录失败且不能显示上传成功。

### 可验证结果

- 成功投稿后存在对应视频、媒体文件和审核状态。
- 处理完成前视频不会被错误公开。
- 取消和失败的上传不会产生可播放的残留投稿。

## UC04 视频审核与发布

- **参与者：** 审核员、创作者、普通用户。
- **触发条件：** 系统中存在等待审核的视频。
- **前置条件：** 审核员已登录并具有内容审核权限。
- **代码对应：** `backend/internal/handler/v1/video/video_handler.go`、`frontend/src/pages/admin/AdminVideosPage.vue`、`frontend/src/pages/user/CreatorDashboardPage.vue`。

### 主成功流程

1. 审核员进入后台查看待审核视频。
2. 审核员查看投稿信息和媒体内容。
3. 审核员选择通过或拒绝并提交审核结果。
4. 系统保存审核状态。
5. 通过的视频进入公开列表；拒绝的视频向创作者展示拒绝状态。

### 备选或异常流程

- 普通用户访问审核功能时，系统拒绝操作。
- 视频已被处理时，系统返回最新状态并避免重复覆盖。
- 视频媒体不可访问时，系统不能将其发布为可正常播放状态。

### 可验证结果

- 通过审核的视频可以公开查询和播放。
- 待审核或拒绝的视频不能公开播放。
- 审核结果和权限控制在刷新后保持一致。

## UC05 视频观看互动

- **参与者：** 普通用户。
- **触发条件：** 已登录用户正在观看一个可播放视频并希望参与互动。
- **前置条件：** 用户已登录；目标视频存在且审核通过。
- **代码对应：** `backend/internal/handler/v1/danmaku/danmaku_handler.go`、`backend/internal/handler/v1/comment/comment_handler.go`、`backend/internal/handler/v1/video/video_handler.go`、`backend/internal/logic/danmaku/hub.go`、`frontend/src/pages/video/VideoDetailPage.vue`、`frontend/src/components/common/VideoPlayer.vue`。

### 主成功流程

1. 用户进入视频详情页并开始播放。
2. 用户在指定播放时间发送弹幕，系统保存并展示弹幕。
3. 用户发表评论，系统将评论加入评论区。
4. 用户点赞或收藏视频。
5. 系统更新互动状态和统计数据。
6. 用户刷新页面后仍能看到一致的互动结果。

### 备选或异常流程

- 用户可以取消点赞或取消收藏。
- 空内容、超长内容或非法参数被系统拒绝。
- 被管理员屏蔽的弹幕不能继续公开展示。
- 重复请求或网络重试不能导致统计重复增加。

### 可验证结果

- 弹幕时间点、评论内容、点赞和收藏状态正确保存。
- 页面状态和后端统计一致。
- 取消操作、非法输入及重复请求均得到正确处理。

## UC06 个人视频资料库管理

- **参与者：** 普通用户。
- **触发条件：** 用户希望查看或整理与自己相关的视频记录。
- **前置条件：** 用户已登录，并产生过观看或收藏等记录。
- **代码对应：** `backend/internal/handler/v1/user/library_handler.go`、`frontend/src/api/library.ts`、`frontend/src/pages/user/UserLibraryPage.vue`、`frontend/src/utils/playQueue.ts`、`frontend/src/utils/userLibrary.ts`。

### 主成功流程

1. 用户播放视频，系统记录观看进度和时间。
2. 用户将视频加入稍后再看、收藏或个人合集。
3. 用户进入个人资料库查看分类内容。
4. 系统返回观看历史、稍后再看、点赞、收藏等记录。
5. 用户继续播放时，系统恢复已保存的进度。
6. 用户移除单项记录或清空允许清空的列表。

### 备选或异常流程

- 同一视频重复加入时，系统不生成重复有效记录。
- 视频已删除或不可播放时，系统显示不可用状态。
- 未登录用户不能读取其他用户的资料库。
- 多设备更新进度时，以服务端记录为准。

### 可验证结果

- 各分类列表内容和数量正确。
- 观看进度能够保存和恢复。
- 添加、移除及清空操作在刷新后保持一致。

## UC07 关注关系与内容通知

- **参与者：** 普通用户、创作者。
- **触发条件：** 用户希望持续关注某位创作者并接收相关内容。
- **前置条件：** 用户已登录；目标创作者存在。
- **代码对应：** `backend/internal/handler/v1/user/relationship_handler.go`、`backend/internal/handler/v1/notification/notification_handler.go`、`frontend/src/pages/user/UserProfilePage.vue`、`frontend/src/pages/user/SubscriptionPage.vue`。

### 主成功流程

1. 用户进入创作者主页并发起关注。
2. 系统建立关注关系并更新状态。
3. 用户将关注对象加入分组或设置为特别关注。
4. 创作者产生相关内容后，系统生成动态或通知。
5. 用户在订阅页或通知中心查看相关内容。

### 备选或异常流程

- 用户可以取消关注或调整分组。
- 用户不能关注自己，重复关注不能产生重复关系。
- 用户可以屏蔽其他用户，屏蔽后相关交互受限。
- 目标账号不存在时，系统拒绝建立关系。

### 可验证结果

- 关注、分组、特别关注和取消关注状态正确保存。
- 订阅页及通知内容与关注关系一致。
- 非法目标和重复请求不会产生错误关系数据。

## UC08 创作者会员订阅

- **参与者：** 普通用户、创作者。
- **触发条件：** 创作者希望提供会员方案，用户希望订阅该方案。
- **前置条件：** 双方已登录；创作者已配置可用会员方案。
- **代码对应：** `backend/internal/handler/v1/membership/membership_handler.go`、`frontend/src/api/membership.ts`、`frontend/src/pages/user/SubscriptionPage.vue`、`frontend/src/pages/user/UserProfilePage.vue`。

### 主成功流程

1. 创作者配置会员价格和方案说明。
2. 用户查看创作者会员方案并创建订阅订单。
3. 用户完成演示支付。
4. 系统更新订单状态并创建有效订阅。
5. 用户查看订阅状态并管理自动续订。

### 备选或异常流程

- 创作者未开放方案时，用户不能创建订单。
- 用户不能订阅自己。
- 重复支付不能创建重复订阅。
- 订单不存在、已失效或金额不一致时，系统拒绝支付。
- 订阅到期后，系统更新订阅状态。

### 可验证结果

- 订单状态、订阅有效期和自动续订设置正确。
- 重复请求不会产生重复订单结果或重复有效订阅。
- 到期订阅能够被系统识别并更新。

## UC09 直播预约与用户预约

- **参与者：** 主播、普通用户。
- **触发条件：** 主播计划未来开播，用户希望提前预约。
- **前置条件：** 主播和用户已登录；预约时间合法。
- **代码对应：** `backend/internal/handler/v1/live/schedule_handler.go`、`frontend/src/api/live.ts`、`frontend/src/pages/live/LiveListPage.vue`。

### 主成功流程

1. 主播填写直播主题和计划时间并发布预约。
2. 系统保存预约并在直播页面展示。
3. 用户查看预约信息并提交预约。
4. 系统保存用户预约关系并更新预约人数。
5. 用户再次进入页面时查看已预约状态。

### 备选或异常流程

- 主播可以取消尚未开始的预约。
- 用户可以取消自己的预约。
- 过去时间、非法时间或冲突预约不能提交。
- 重复预约不能产生重复记录。

### 可验证结果

- 直播预约、用户预约关系和人数统计正确。
- 取消、重复请求和非法时间均得到正确处理。

## UC10 直播发布、观看与实时互动

- **参与者：** 主播、普通用户、SRS 直播服务。
- **触发条件：** 主播希望开始直播，观众希望进入直播间观看。
- **前置条件：** 主播和观众已登录；SRS 服务可用；主播拥有直播间。
- **代码对应：** `backend/internal/handler/v1/live/live_handler.go`、`backend/internal/handler/v1/live/interaction_handler.go`、`backend/internal/handler/ws/live_publish_handler.go`、`backend/internal/handler/ws/ws_handler.go`、`frontend/src/pages/live/LiveStudioPage.vue`、`frontend/src/pages/live/LiveRoomPage.vue`。

### 主成功流程

1. 主播进入直播工作台并配置直播信息。
2. 系统生成推流和播放信息。
3. 主播开始发布直播流，SRS 生成观众可访问的播放资源。
4. 观众进入直播间观看直播。
5. 观众发送实时弹幕、点赞、礼物或 Super Chat。
6. 主播查看互动、观看人数和直播状态。
7. 主播结束直播，系统更新并清理本次直播状态。

### 备选或异常流程

- 推流未就绪时，观众端显示等待状态并有限重试。
- 非房主不能修改直播设置或结束他人直播。
- 空弹幕、非法礼物和重复点赞请求被系统拒绝。
- WebSocket 中断后页面恢复连接，不能重复计算在线人数。
- 推流异常结束时，系统最终恢复直播间状态。

### 可验证结果

- 直播状态、播放地址、观看人数和互动数据正确。
- 非房主无法执行主播操作。
- 下播后直播列表和直播间状态及时更新。

## UC11 用户私信与媒体分享

- **参与者：** 普通用户。
- **触发条件：** 用户希望与另一名用户进行私下沟通或内容分享。
- **前置条件：** 发送者已登录；发送者和接收者均为有效用户。
- **代码对应：** `backend/internal/handler/v1/message/message_handler.go`、`backend/internal/logic/chat/hub.go`、`backend/internal/handler/ws/ws_handler.go`、`frontend/src/api/message.ts`、`frontend/src/pages/user/ChatPage.vue`。

### 主成功流程

1. 用户进入目标用户会话。
2. 用户发送文字、图片、短视频或平台视频分享。
3. 系统验证内容及媒体归属并保存消息。
4. 在线接收者实时收到消息，离线接收者稍后读取历史消息。
5. 接收者进入会话后，系统更新已读状态和未读数量。

### 备选或异常流程

- 空消息、未知类型或非法媒体地址被系统拒绝。
- 用户不能引用其他用户目录中的私信附件。
- 目标用户不存在或受到屏蔽规则限制时，系统拒绝发送。
- 实时连接中断时，已保存消息仍能从历史记录恢复。

### 可验证结果

- 消息内容、双方用户、时间和已读状态正确。
- 未读数量随发送和读取正确变化。
- 非法附件、无效用户和重复请求不会生成错误消息。

## UC12 创作者数据分析

- **参与者：** 创作者。
- **触发条件：** 创作者希望了解自己的内容和粉丝表现。
- **前置条件：** 创作者已登录，并拥有可统计的内容或历史数据。
- **代码对应：** `backend/internal/handler/v1/creator/analytics_handler.go`、`backend/internal/logic/analytics/creator_stat.go`、`frontend/src/pages/user/CreatorDashboardPage.vue`、`frontend/src/components/creator/MetricLineChart.vue`。

### 主成功流程

1. 创作者进入创作中心数据分析页面。
2. 系统验证身份和数据访问范围。
3. 系统汇总作品、播放、互动和粉丝数据。
4. 页面展示核心指标和按日趋势。
5. 创作者切换统计范围并查看对应结果。

### 备选或异常流程

- 没有历史数据时，系统返回零值和空趋势。
- 普通用户不能读取其他创作者的内部数据。
- 部分统计暂不可用时，系统显示明确的降级结果。

### 可验证结果

- 汇总结果与原始业务数据一致。
- 趋势日期和数值对应正确。
- 用户只能查看自己有权限访问的数据。

## UC13 平台审核、权限、运营与基础设施管理

- **参与者：** 审核员、管理员。
- **触发条件：** 管理人员需要治理内容、调整权限、配置运营信息或检查系统状态。
- **前置条件：** 操作者已登录并具有审核员或管理员角色。
- **代码对应：** `backend/internal/handler/v1/admin/admin_handler.go`、`backend/internal/handler/v1/danmaku/danmaku_handler.go`、`backend/internal/handler/v1/video/video_handler.go`、`frontend/src/pages/admin/AdminUsersPage.vue`、`frontend/src/pages/admin/AdminOperationsPage.vue`、`frontend/src/pages/admin/AdminInfrastructurePage.vue`、`frontend/src/pages/admin/AdminDanmakuPage.vue`。

### 主成功流程

1. 审核员处理待审核视频或违规弹幕。
2. 管理员查看用户并调整角色。
3. 管理员创建、修改或删除首页横幅和系统公告。
4. 管理员查看存储、流量、CPU、在线人数和直播数量等指标。
5. 系统保存管理操作并向相关页面展示最新状态。

### 备选或异常流程

- 审核员不能调整管理员权限或平台级运营配置。
- 普通用户不能进入后台或调用后台接口。
- 非法角色值或不完整运营参数被系统拒绝。
- 指标暂不可用时，系统返回明确的不可用状态。

### 可验证结果

- 审核员、管理员和普通用户的权限边界正确。
- 审核结果、用户角色和运营内容正确保存并生效。
- 基础设施指标具有明确的数据来源和采集结果。
