# UC05 / UC09 / UC10 测试用例设计

## 1. 单元测试

| 编号 | 目标 | 关键断言 | 状态 |
| --- | --- | --- | --- |
| UNIT-TC01 | SC 金额边界 | 49/50/200/500/1000 对应正确展示时长 | 通过 |
| UNIT-TC02 | 直播热度 | 在线、点赞、礼物按规则加权 | 通过 |
| UNIT-TC03 | 预约时间与状态 | 两种时间格式、非法时间及状态集合 | 通过 |
| UNIT-TC04 | 在线人数 | 登录用户去重、匿名逐连接、监控排除 | 通过 |
| UNIT-TC05 | Hub 事件与持久化错误 | 正常事件序列化；写入失败不广播 | 通过 |

## 2. API / MySQL 集成测试

| 编号 | 场景 | 关键断言 | 状态 |
| --- | --- | --- | --- |
| INT-TC01 | UC05 视频互动 | 弹幕/评论/回复/排序/删除、点赞收藏切换及数据库计数一致 | 通过 |
| INT-TC02 | UC05 权限异常 | 未登录、未审核、非法 ID、资源不存在均返回正确状态且不写脏数据 | 通过 |
| INT-TC03 | UC09 预约流程 | 创建、列表、预约切换、冲突、通知、本人预约及管理员取消 | 通过 |
| INT-TC04 | UC09 到时启动 | Worker 重复扫描只启动一次并生成一次统计与通知 | 通过 |
| INT-TC05 | UC10 直播管理 | 创建/复用/详情/监控/设置/下播/重启及持久化状态一致 | 通过 |
| INT-TC06 | UC10 实时互动 | 点赞、礼物、SC、排行榜、聊天权限、慢速模式和失败门控 | 通过 |
| INT-TC07 | WebSocket 往返 | 注册连接、在线人数、实时弹幕、默认样式、数据库失败错误事件 | 通过 |

## 3. E2E 测试

| 编号 | 场景 | 通过标准 | 状态 |
| --- | --- | --- | --- |
| E2E-TC05 | 发送弹幕和评论、点赞收藏并刷新 | 页面计数、评论内容及刷新后的持久化状态一致 | 2026-08-27 Chromium 通过 |
| E2E-TC09 | 创建预约并由另一用户预约/取消 | UI 状态刷新保持，API 可查，清理成功 | 2026-08-27 Chromium 通过 |
| E2E-TC10 | 创建直播、点赞、赠礼和结束 | 页面状态与互动 API、下播后 404 一致 | 2026-08-27 Chromium 通过 |
| E2E-MEDIA | FFmpeg→RTMP→SRS→HLS | 主清单、媒体清单和 TS 分片可访问 | 通过 |

执行命令：

```powershell
cd backend
go test -count=1 ./...
$env:DANMAKU_TEST_ADMIN_DSN='<测试管理员DSN>'
go test -count=1 -tags=integration ./internal/handler/v1/live ./internal/logic/danmaku ./integration

cd ../frontend
npm run test:e2e -- e2e/d-engagement.spec.ts（统一入口；全量则 npm run test:e2e）
```
