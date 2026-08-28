# 成员 C 内容域测试用例（UC02 / UC03 / UC04 / UC12）

## 1. 范围与口径

- 修复基线：`83984544425103f74be8b27f69c194533d4bc84f`，本地分支 `lbh-revise`。
- 单元测试：Go `testing`、表驱动、Gin `httptest`，不连接数据库。
- API 集成测试：真实后端与 MySQL，独立测试用户和 `RUN_TAG`，结束后清理数据库及上传文件。
- E2E：Playwright Chromium，使用真实前后端、MySQL 和 FFmpeg；四个有状态场景串行执行。
- `PASS` 表示断言满足；`FAIL` 表示能力回归失败。修复后 API 套件不再容许 `GAP` 作为成功结果。

## 2. 单元测试

| 编号 | UC | 测试点 | 预期结果 | 实现位置 |
|---|---|---|---|---|
| UNIT-TC02-01 | UC02 | hot/date/like/collect 及非法排序 | 合法排序生成确定表达式，非法值报错 | `backend/internal/logic/video/list_logic_test.go` |
| UNIT-TC02-02 | UC02 | 非数字详情 ID | 访问数据库前返回 400 | `backend/internal/handler/v1/video/video_handler_test.go` |
| UNIT-TC03-01 | UC03 | 缺标题、缺视频文件 | 返回 400，不创建投稿 | 同上 |
| UNIT-TC03-02 | UC03 | 下载文件名清理 | 路径和保留字符被替换，空标题有默认名 | 同上 |
| UNIT-TC03-03 | UC03 | 站内媒体路径转换 | 仅接受 `/media/` 路径，拒绝远程和非法路径 | 同上 |
| UNIT-TC04-01 | UC04 | 上传源文件判断 | 仅 `videos/<id>/upload.*` 被识别为源文件 | 同上 |
| UNIT-TC04-02 | UC04 | 非法审核 ID、JSON、状态 | 访问数据库前返回 400 | 同上 |
| UNIT-TC04-03 | UC04 | staff 权限和审核状态枚举 | user 被拒绝，moderator/admin 放行 | 复用 `auth_test.go`、`video_handler_test.go` |
| UNIT-TC12-01 | UC12 | 未登录、非法 days/videoId | 分别返回 401/400 | `backend/internal/handler/v1/creator/analytics_handler_test.go` |
| UNIT-TC12-02 | UC12 | 自然日边界 | 保留时区并归零到当天 00:00:00 | 同上 |

## 3. API 集成测试

统一入口：`tests/api/member-c-content-test.sh`。

| 编号 | UC | 场景 | 核心断言 |
|---|---|---|---|
| INT-TC02-01 | UC02 | 搜索、分页、公开过滤 | 仅返回 approved，分页参数正确 |
| INT-TC02-02 | UC02 | 访问详情 | 播放量、创作者日统计和作品日统计同步增加 |
| INT-TC02-03 | UC02 | 待审和非法 ID | 待审 404，非法 ID 400 |
| INT-TC02-04 | UC02 | 非法排序、媒体缺失 | 非法排序 400；媒体缺失 503 且播放量不增加 |
| INT-TC03-01 | UC03 | 未登录、缺字段、错误请求类型 | 返回 401/400 |
| INT-TC03-02 | UC03 | 转码失败 | 审核状态保持 pending；转码状态回写 failed 和安全失败原因 |
| INT-TC03-03 | UC03 | 客户端中止上传 | 请求被中止，不产生数据库记录 |
| INT-TC03-04 | UC03 | 投稿状态跟踪 | 创作者只能看到自己的投稿及 pending 状态 |
| INT-TC04-01 | UC04 | 审核列表权限 | 未登录 401、普通创作者 403、审核员 200 |
| INT-TC04-02 | UC04 | 审核通过 | 状态持久化且视频变为公开可访问 |
| INT-TC04-03 | UC04 | 审核拒绝 | 状态持久化且视频不可公开访问 |
| INT-TC04-04 | UC04 | 重复审核、缺媒体、非法状态、不存在视频 | 分别返回 409/403/400/404，终态不被覆盖 |
| INT-TC12-01 | UC12 | 鉴权、参数、作品所有权 | 返回 401/400/404 |
| INT-TC12-02 | UC12 | 全部作品 7 天统计 | 补齐 7 日、汇总与数据库一致、排行不超过 5 条 |
| INT-TC12-03 | UC12 | 单作品统计 | 趋势、汇总和排行均只包含指定作品 |
| INT-TC12-04 | UC12 | 空数据 | 返回 30 个零值点和零汇总 |

## 4. E2E 测试

入口：`npm run test:e2e:member-c`。

| 编号 | UC | 场景 | 验收点 |
|---|---|---|---|
| E2E-TC02-01 | UC02 | 搜索公开视频并进入详情 | 搜索结果、详情标题、播放器和空结果提示可见 |
| E2E-TC03-01 | UC03 | 取消、重新投稿和转码失败 | 取消无记录；投稿可跟踪；失败状态和安全原因可见 |
| E2E-TC04-01 | UC04 | 审核通过、拒绝、重复审核及越权 | UI 状态更新并禁用终态操作；重复审核 409；普通用户 403 |
| E2E-TC12-01 | UC12 | 7 天与单作品切换 | 指标同步刷新；排行只返回选中作品；跨用户查询 404 |

## 5. 测试数据与清理

- API 使用 `mc-<RUN_TAG>-*` 用户和 `MC-<RUN_TAG>-*` 视频。
- E2E 使用 `e2e-mc-*` 用户和 `E2E-MC-*` 视频。
- API `trap` 和 Playwright `globalTeardown` 删除统计行、视频行和本次生成的媒体目录。
- CI 通过 `MYSQL_CMD`、`VIDEO_DIR`、`E2E_BACKEND_CONFIG` 和 `FFMPEG_BIN` 适配运行环境。

## 6. Issue #80 修复验证

2026-08-28 本地回归结果：Go 全量测试通过，API `57 PASS / 0 FAIL / 0 GAP`，Playwright `4/4 PASS`。原五项 GAP 均已转为阻断断言；证据见 `docs/testing/reports/UC02-03-04-12-*-20260828.*`。
