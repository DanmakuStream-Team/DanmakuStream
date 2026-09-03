# 互动与直播域追溯表（成员 D）

## 1. 追溯基线

本表把需求、用例、三层设计、代码模块、测试和结果放在同一处。直播回放不纳入本表。状态含义：

- 已实现：代码中存在且与文档规则一致；
- 部分实现：主流程存在，但仍有明确验收缺口；
- 待补测试：实现存在，但自动化测试不足；
- 已通过：所列自动化测试已实际执行通过。

## 2. 需求—设计—代码—测试追溯

| 需求 | 用例 | 系统级设计 | 组件级设计 | 对象级设计 | 代码模块 | 自动化测试 | 当前结果 |
| --- | --- | --- | --- | --- | --- | --- | --- |
| REQ01 视频弹幕、评论、点赞与收藏 | UC05 | `SYS-SEQ05` | `COMP-SEQ05` | `OBJ-SEQ05` | `danmaku_handler.go`、`comment_handler.go`、`video_handler.go`、`VideoDetailPage.vue`、`VideoPlayer.vue` | `UNIT-TC01`、`INT-TC01`、`INT-TC02`、`E2E-TC05` | 弹幕、评论、点赞/取消和收藏/取消的 MySQL 集成及浏览器持久化测试已通过 |
| REQ02 主播发布直播计划 | UC09 | `SYS-SEQ09` | `COMP-SEQ09` | `OBJ-SEQ09` | `schedule_handler.go`、`LiveListPage.vue`、`api/live.ts` | `UNIT-TC03`、`UNIT-TC04`、`INT-TC04`、`E2E-TC09` | 相同开播时间冲突校验与并发串行锁已实现，集成和浏览器测试通过 |
| REQ03 用户预约/取消预约 | UC09 | `SYS-SEQ09` | `COMP-SEQ09` | `OBJ-SEQ09` | `ReserveScheduleHandler`、`LiveReservation` | `INT-TC03`、`E2E-TC09` | 预约切换、刷新持久化、唯一关系和提醒人数测试已通过 |
| REQ04 主播创建、推流与结束直播 | UC10 | `SYS-SEQ10` | `COMP-SEQ10` | `OBJ-SEQ10` | `live_handler.go`、`live_publish_handler.go`、`LiveStudioPage.vue` | `INT-TC09`、`E2E-TC10`、`E2E-MEDIA` | 创建、重复创建、越权/正常下播和浏览器闭环通过；RTMP-SRS-HLS 媒体链路通过 |
| REQ05 实时弹幕、点赞、礼物与 SC | UC10 | `SYS-SEQ10` | `COMP-SEQ10` | `OBJ-SEQ10` | `interaction_handler.go`、`ws_handler.go`、`logic/danmaku/hub.go`、`LiveRoomPage.vue` | `UNIT-TC01`、`INT-TC06`、`INT-TC07`、`E2E-TC10` | 点赞切换、礼物/SC、持久化失败门控、真实 WebSocket 和浏览器流程已通过 |
| REQ06 在线人数、热度与主播监控 | UC10 | `SYS-SEQ10` | `COMP-SEQ10` | `OBJ-SEQ10` | `MonitorHandler`、`Hub.broadcastViewerCount`、`LiveStudioPage.vue` | `UNIT-TC02`、`UNIT-TC05` | 登录观众按用户去重、匿名按连接计数、监控连接排除，测试通过 |
| REQ07 实时连接与异常推流恢复 | UC10 | `SYS-SEQ10` | `COMP-SEQ10` | `OBJ-SEQ10` | `frontend/src/api/danmaku.ts`、`live_publish_handler.go`、`srs_hook_handler.go`、`deploy/srs.conf` | `UNIT-TC05`、`INT-TC-SRS`、`E2E-TC10` | 3 秒观看端重连及人数去重已通过；浏览器和 SRS/OBS 推流断开进入 15 秒恢复窗口，快速重连不误结束，超时后房间置为 ended |

## 3. 建议测试用例清单

| 测试 | 层级 | 前置条件/输入 | 预期结果 | 状态 |
| --- | --- | --- | --- | --- |
| UNIT-TC01 | 单元 | SC 价值位于 49/50/200/500/1000 边界 | 展示时长分别为 15/30/60/90/120 秒 | 已实现 |
| UNIT-TC02 | 单元 | viewer=12、like=7、gift=230 | 热度为 364 | 已实现 |
| UNIT-TC03 | 单元 | RFC3339、本地时间、非法时间 | 合法格式解析成功，非法格式失败 | 已实现 |
| UNIT-TC04 | 单元 | pending/canceled/live/ended | 仅前三种计划状态有效 | 已实现 |
| UNIT-TC05 | 单元 | 同一登录用户双连接、不同用户、匿名连接、监控连接 | 登录用户去重，匿名逐连接，监控不计数 | 已通过 |
| UNIT-TC06 | 单元 | MySQL 持久化函数返回错误 | 错误向调用者传播，ReadPump 不进入广播分支 | 已通过 |
| INT-TC01 | 集成 | 同一用户连续切换视频点赞/收藏 | 关系和计数同步增减，无重复关系 | 已通过 |
| INT-TC02 | 集成 | 合法及空弹幕和评论 | 合法数据保存，空内容拒绝 | 已通过 |
| INT-TC03 | 集成 | 用户预约同一计划后取消 | 唯一关系和 reminder_count 始终一致 | 已通过 |
| INT-TC04 | 集成 | 同一主播、相同开播时间创建两个 pending 计划 | 第二个请求返回 HTTP 409 | 已通过 |
| INT-TC05 | 集成 | 到期计划被 Worker 多次扫描 | 仅启动一次并产生正确通知 | 已通过 |
| INT-TC06 | 集成 | 直播点赞并再次点赞 | LiveLike 建立后删除，like_count 一致 | 已通过 |
| INT-TC07 | 集成 | 星光礼物、SC 留言和数量上界非法值 | 价值、展示时长和参数错误正确 | 已通过 |
| INT-TC08 | 集成 | everyone/followers/members 与慢速模式 | 仅符合权限者发言，主播可绕过 | 已通过 |
| INT-TC09 | 集成 | 创建/重复创建直播、越权/正常下播 | 房间复用正确，越权拒绝，正常下播并生成记录 | 已通过 |
| E2E-TC05 | E2E | 用户观看审核通过的视频，发送弹幕和评论并点赞收藏，随后刷新 | 计数、评论内容及互动状态刷新后保持一致 | 已通过；Chromium |
| E2E-TC09 | E2E | 主播创建预约，另一用户预约/取消并刷新页面 | 页面状态和 API 数据保持一致 | 已通过；Chromium |
| E2E-TC10 | E2E | 主播创建直播，观众点赞赠礼，主播结束直播 | 页面、互动 API 和结束状态保持一致 | 已通过；Chromium |
| E2E-MEDIA | E2E | FFmpeg 推送测试流至 SRS，读取 HLS 主清单、媒体清单和 TS 分片 | RTMP 接收正常，HLS 各级资源可下载 | 已通过 |
| E2E-TC02 | E2E | 观众断网后恢复，多标签页连接 | 自动重连且登录用户在线人数不重复 | 逻辑测试已通过，浏览器演示待执行 |
| CHAOS-TC01 | K3s 故障 | 将 content-service 缩容为 0 后调用视频互动接口，再恢复原副本数 | 2 秒内返回 503；engagement Pod UID/重启次数不变；其他工作负载 Ready；恢复后返回 200 | 自动化脚本已完成，待授权测试环境执行并归档证据 |

## 4. 剩余增强项

1. 在最终现场演示中执行浏览器 UI 操作和断线重连，并保存截图或录屏。
2. 补充推流中断后的状态恢复与超时兜底测试。
3. `engagement-service` 拆分时更新组件图、对象调用和故障降级测试，本次不提前标记为完成。
