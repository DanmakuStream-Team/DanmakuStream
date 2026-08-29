# UC07–UC13 补测与回归报告

- 执行日期：2026-08-29
- 基线：`origin/dev@dc6a4cf`
- 分支：`codex/dev-followup-20260829`
- 环境：Windows 11、Go、Node.js、Chromium、MySQL 8（独立测试容器）、SRS 5.0.213
- 结论：UC07–UC13 的单元、API/集成和 E2E 证据已复核；本轮新增项全部通过。

## 执行结果

| 层级 | 命令/套件 | 结果 |
|---|---|---|
| 后端单元回归 | `cd backend && go test ./... -count=1` | 通过；所有含测试的 package 均为 `ok`，失败 0 |
| UC07/08/11 单元 | relationship、membership、message、media、chat package | 通过；含续期边界、附件文件头及 clientMessageId 校验 |
| MySQL 集成 | `go test -tags=integration ./internal/handler/v1/live ./internal/handler/ws ./integration -count=1` | 3 个 package 全部通过，失败 0 |
| UC07/08/11 E2E | `npm run test:e2e:member-b` | 3/3 通过 |
| UC05/09/10 E2E | `npm run test:e2e:d` | 3/3 通过；UC10 含等待流、断开重连和人数去重 |
| UC02/03/04/12 E2E | `npm run test:e2e:member-c` | 4/4 通过；UC12 在完整串行套件中通过 |
| UC13 E2E | `npm run test:e2e:uc13` | 5/5 通过 |
| 前端生产构建 | member C E2E 启动前执行 `npm run build` | 通过（类型检查与 Vite 构建成功） |
| SRS 配置校验 | `srs -t -c conf/danmakustream.conf`（SRS 5 容器） | 配置解析成功 |

## 用例补齐摘要

| 用例 | 本轮补齐或复核的关键路径 | 结果 |
|---|---|---|
| UC07 | 关注/取关、分组、屏蔽/解除；被关注创作者发布动态后生成通知并在页面显示 | 通过 |
| UC08 | 方案配置、下单支付、重复支付幂等、跨订单续期、关闭自动续费、到期识别与特别关注解除 | 通过 |
| UC09 | 冲突预约、提醒预约/取消、状态刷新持久化 | 通过 |
| UC10 | 直播等待态、WebSocket 断线自动重连且人数不重复；浏览器推流和 SRS/OBS 推流异常断开后恢复房间状态，快速重连不误结束 | 通过 |
| UC11 | 文本、图片、短视频、站内视频分享；未读/已读、断线期间历史恢复、clientMessageId 重试幂等 | 通过 |
| UC12 | 7 天趋势和单作品范围切换、越权作品过滤 | 通过 |
| UC13 | 审核、弹幕屏蔽、角色权限、横幅/公告、基础设施指标、普通用户越权拒绝 | 通过（5/5） |

## 修复记录

1. 打开私信会话时，前端原先只清本地红点，没有调用服务端已读接口；已在 `ChatPage.vue` 补调用并由 E2E 验证服务端未读归零。
2. 私信新增 `clientMessageId` 复合唯一约束和竞争重试回查，避免 HTTP/WebSocket 重试产生重复消息。
3. SRS 增加 `on_publish`/`on_unpublish` 回调；断流保留 15 秒重连窗口，超时后把直播间置为 ended，未知 stream key 拒绝发布。
4. E2E 测试夹具统一遵循 `VIDEO_DIR`，修复独立环境下 UC02/UC13 媒体文件路径不一致导致的误失败。
5. 删除已被可运行综合套件替代的 UC07/08/11/12 跳过式草稿 spec，避免默认 Playwright 扫描出现虚假“未执行”。

## 证据位置

- UC07/08/11 浏览器报告：`docs/testing/reports/UC07-UC08-UC11-e2e-report/index.html`
- UC05/09/10 浏览器报告：`docs/testing/reports/engagement-e2e/index.html`
- UC02/03/04/12 浏览器报告：`docs/testing/reports/UC02-03-04-12-e2e-report/index.html`
- UC13 浏览器报告：`docs/testing/reports/uc13-e2e/index.html`
- 既有原始输出：`docs/testing/reports/UC07-UC08-UC11-unit-test-report-20260829.txt`、`UC05-UC07-UC08-UC09-UC10-UC11-integration-report-20260829.txt`、`UC07-UC08-UC11-e2e-test-report-20260829.txt`

## CI 门禁

`ci.yml` 已将 UC07/08/11 与 UC13 浏览器测试纳入 `e2e-member-b` Job；镜像构建依赖该 Job。任一测试失败时，Docker 镜像构建不会执行。
