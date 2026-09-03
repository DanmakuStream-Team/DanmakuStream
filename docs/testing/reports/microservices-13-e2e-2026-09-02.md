# 微服务版 UC01–UC13 E2E 验收报告

- 执行日期：2026-09-02
- 部署形态：MySQL + user-service + content-service + engagement-service + nginx gateway + frontend + SRS
- 测试入口：nginx gateway（API）与独立 frontend 容器（页面）
- 浏览器：Playwright Chromium，单 worker 串行执行业务流程
- 最终结果：**24 passed / 0 failed / 0 skipped（UC01–UC13：13/13）**
- 执行耗时：1.6 分钟

## 用例覆盖

| 用例 | 自动化验收内容 | 结果 |
| --- | --- | --- |
| UC01 | 注册、重复昵称、登录、错误密码、资料持久化 | 5/5 通过 |
| UC02 | 搜索公开视频并进入播放页 | 通过 |
| UC03 | 取消上传、重新投稿、非法媒体转码失败状态 | 通过 |
| UC04 | 审核发布/拒绝及普通用户越权阻断 | 通过 |
| UC05 | 弹幕、评论、点赞收藏及刷新持久化 | 通过 |
| UC06 | 历史、进度、删除/清空、稍后再看、未登录阻断 | 4/4 通过 |
| UC07 | 关注、分组、黑名单与动态通知 | 通过 |
| UC08 | 套餐、订阅、幂等支付与特别关注 | 通过 |
| UC09 | 创建预约、预约/取消及刷新持久化 | 通过 |
| UC10 | 创建直播、点赞、赠礼、结束及 API 状态一致 | 通过 |
| UC11 | 会话、WebSocket、历史、已读、未读和媒体分享 | 通过 |
| UC12 | 7 天/30 天统计、全部作品/单作品分析与权限 | 通过 |
| UC13 | 视频/弹幕审核、角色、运营配置、基础设施指标、403 | 5/5 通过 |

## 执行命令

```powershell
cd frontend
$env:E2E_MICROSERVICES='1'
$env:E2E_USE_GATEWAY='1'
$env:E2E_API_BASE='http://127.0.0.1:18888/api/v1'
$env:E2E_BASE_URL='http://127.0.0.1:18080'
npm run test:e2e:microservices:full
```

完整 HTML 报告见同目录的 `e2e/index.html`。CI 的 `microservices-e2e` 工作流已设置 `MICRO_E2E_FULL_SUITE=1`，PR 合并前会执行同一完整套件；失败时上传 Compose 状态、服务日志与 Playwright 证据。

## 额外回归

- `user-service`: `go test ./... -cover` 通过。
- `content-service`: `go test ./... -cover` 通过。
- `engagement-service`: `go test ./... -cover` 通过；核心包覆盖率包括 app 94.2%、client 83.3%、config 73.8%、middleware 71.1%、model 100%。
- `frontend`: `npm run build` 通过。

覆盖率数字是 Go 按包统计值；浏览器 E2E 作为独立黑盒验收，不计入 Go 语句覆盖率。
