# UC01、UC07、UC08、UC11 自动化测试用例

## 测试范围

| 编号 | 用例 | 主要验证内容 |
| --- | --- | --- |
| UNIT/INT-TC01 | UC01 注册、登录与资料维护 | 参数校验、重复注册、密码哈希、JWT 鉴权、登录兼容字段、资料与头像持久化、数据库异常 |
| UNIT/INT-TC07 | UC07 关注与关系管理 | 关注/取关、分组、通知、双向关注解除、拉黑/解除、付费关注保护、数据库异常 |
| UNIT/INT-TC08 | UC08 创作者会员订阅 | 套餐、订单、演示支付、幂等支付、续费、自动续费限制、过期降级、拉黑限制、数据库异常 |
| UNIT/INT-TC11 | UC11 私信 | 文本/图片/视频/站内视频分享、校验、会话、历史、未读/已读、拉黑限制、数据库异常、WebSocket 双端广播 |
| E2E-TC01 | UC01 页面全流程 | 页面注册、重新登录、编辑资料、刷新后持久化、API 结果核对 |
| E2E-TC07 | UC07 页面全流程 | 页面关注、拉黑、黑名单 API 核对、解除拉黑 |
| E2E-TC08 | UC08 页面全流程 | 页面订阅、演示支付、特别关注展示、订阅状态 API 核对 |
| E2E-TC11 | UC11 页面全流程 | 页面发送、接收方读取、消息展示、未读数归零 |

## 执行命令

```powershell
# MySQL/API/实时私信与覆盖率
cd backend
$env:DANMAKU_TEST_ADMIN_DSN='root:<password>@tcp(127.0.0.1:3306)/danmakustream?charset=utf8mb4&parseTime=True&loc=Local'
go test -tags=integration ./integration ./internal/handler/v1/membership ./internal/logic/chat -run 'TestUserDomain|TestSubscriptionExpiration|TestValidateMessageInput|TestNormalizedMessageType|TestHubBackpressure|TestChatHubWebSocketRoundTripPersistenceAndError' -count=1 -coverprofile=user-domain.cover.out

# 浏览器端到端测试；MYSQL_CMD 可替换为本机 mysql 客户端命令
cd ../frontend
$env:MYSQL_CMD='mysql -h127.0.0.1 -P3306 -uroot -p<password> danmakustream'
npm run test:e2e:user-domain
```

覆盖率由 `backend/cmd/userdomaincoverage` 按源码块去重，并按四个用例实际涉及函数计算，避免多个测试包重复统计，也避免同一 `user` 包内 UC06 视频资料库等无关代码将结果错误拉低。CI 下限为 90%。
