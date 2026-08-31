# UC11 测试用例设计

执行日期：2026-08-29。REST、MySQL 和浏览器 WebSocket 汇总见 `docs/testing/reports/UC07-UC13-completion-report-20260829.md`；原始输出和 HTML 报告保存在同目录。

## 单元测试（UNIT-TC11）

| 编号 | 测试目标 | 关键断言 | 当前状态 |
|---|---|---|---|
| UNIT-TC11-01 | 消息参数校验 | receiverId 缺失、空内容和非法类型被拒绝 | 通过（handler/chat 单元） |
| UNIT-TC11-02 | 会话权限校验 | 自发与被屏蔽用户不能发送 | 通过（单元 + MySQL 集成） |
| UNIT-TC11-03 | 消息幂等 | clientMessageId 长度受限；同一发送者重试返回同一消息 | 通过（单元 + MySQL 集成） |
| UNIT-TC11-04 | 媒体分享校验 | 无效 videoId、非法路径或媒体类型返回业务错误 | 通过（chat 单元 + MySQL 集成） |
| UNIT-TC11-05 | 未读数与已读 | 读取会话后接收方未读数归零 | 通过（MySQL 集成） |
| UNIT-TC11-06 | 附件文件头 | 伪造扩展名被拒绝，合法 MP4/WebM 文件头通过 | 通过（`media_handler_test.go`） |

## API 集成测试（INT-TC11）

| 编号 | 接口/场景 | 关键断言 | 当前状态 |
|---|---|---|---|
| INT-TC11-01 | GET `/messages/conversations` | 返回当前用户会话和正确未读数 | 通过 |
| INT-TC11-02 | GET `/messages/:userId` | 双向历史按时间返回并标记已读 | 通过 |
| INT-TC11-03 | POST `/messages` 文本 | 合法消息持久化并返回消息对象 | 通过 |
| INT-TC11-04 | POST `/messages` 非法接收方 | 404/400，数据库无消息 | 通过 |
| INT-TC11-05 | POST `/messages` 视频分享 | 有效 videoId 返回关联视频，非法引用失败 | 通过 |
| INT-TC11-06 | PUT `/messages/:userId/read` 与 unread | 已读后未读数量正确变化 | 通过 |
| INT-TC11-07 | clientMessageId 重试 | 两次请求返回同一 ID，数据库只有一行 | 通过 |
| INT-TC11-08 | WebSocket `/ws/chat` | 登录连接实时接收 message 事件，断开后自动重连 | 通过（浏览器 E2E） |

## E2E 测试（E2E-TC11）

| 编号 | 场景 | 通过标准 | 当前状态 |
|---|---|---|---|
| E2E-TC11-01 | 两个用户私信与重连 | 实时收发；主动断开后显示离线并在 8 秒内自动恢复 | 通过（`member-b-workflows.spec.ts`） |
| E2E-TC11-02 | 图片、视频与站内视频分享 | 上传图片/MP4 后显示媒体气泡；分享公开视频后显示卡片 | 通过（`member-b-workflows.spec.ts`） |
| E2E-TC11-03 | 离线历史与已读 | 断线期间写入的消息在重连刷新后恢复；打开会话后服务端未读归零 | 通过（`member-b-workflows.spec.ts`） |
| E2E-TC11-04 | 重复请求 | 相同 clientMessageId 两次发送返回相同消息 ID | 通过（`member-b-workflows.spec.ts`） |
