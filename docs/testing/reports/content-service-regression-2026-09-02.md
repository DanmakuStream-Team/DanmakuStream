# content-service 第 8 天回归报告（2026-09-02）

## 范围

- 负责人：成员 C
- 分支：`fix/p3-content-regression`
- 基线：`fe7d206`（2026-09-02 提交前重新同步后的最新 `origin/dev`）
- 服务：`services/content-service`
- 用例：UC02、UC03、UC04、UC13
- 关联任务：#50、#119、#122、#140

## 本次补强

1. 将生产路由注册从 `internal/svc` 提取为唯一的 `internal/server.Router`，生产启动与测试共用同一套路由。
2. 新增带 `integration` build tag 的真实 MySQL API 回归，维护 34 个公开/内部接口的显式清单并自动对比 Gin 实际注册结果。
3. 自动触达全部 34 个接口，统一断言无未解释 5xx，并断言入口 `X-Request-ID` 原样返回。
4. 对搜索、可播放详情、播放统计持久化、内部可播放性、审核发布和创作者分析增加真实数据成功断言。
5. content-service 独立 CI 启用 MySQL 集成测试；普通测试或集成回归失败时，不再继续构建 commit-SHA 镜像。

## 接口—测试映射

| 接口域 | 数量 | 自动化验证 |
|---|---:|---|
| 健康、版本 | 3 | 健康/版本契约测试；全路由真实 MySQL 触达 |
| 视频发现、详情、作者投稿 | 4 | 搜索与分页、公开视频过滤、详情和播放统计断言 |
| 投稿、媒体、编辑、下载、删除、协作者 | 8 | 上传/审核/所有权 SQLite 主流程；全路由真实 MySQL 触达和无 5xx 断言 |
| 动态 | 3 | 创建/查询/删除路由触达与权限回归 |
| 审核与分析 | 3 | 待审视频发布、重复审核、创作者统计真实数据断言 |
| Banner、公告 | 10 | public/admin CRUD 全路由触达与角色校验 |
| 内部视频接口 | 3 | 单条/批量摘要、可播放性、所有权字段、互动计数幂等同步和内部 Token 断言 |
| **合计** | **34** | 显式清单与 Gin 实际注册结果必须完全一致 |

## 主流程和异常流程

- 主流程：搜索 approved 视频、查看详情并增加播放统计、审核 pending 视频、查询创作者分析、单条和批量内部视频查询。
- 权限：缺少 JWT 返回 401，普通用户访问后台返回 403，moderator 只能审核，admin 才能维护运营内容。
- 内容状态：待审视频不公开；审核通过后公开；重复审核返回 409。
- 媒体：合法视频上传并落盘；伪造文件和超限文件分别返回 400/413；非所有者编辑返回 403。
- 数据边界：测试迁移只包含内容域表，不创建用户、评论、弹幕、点赞、收藏和直播表，也不建立跨库外键。
- 错误契约：资源不存在和未知路由使用业务错误码；错误响应携带 `requestId`。

## 本次发现并修复的缺陷

- **现象**：真实 MySQL 回归访问 `GET /api/v1/announcements` 时返回 500；SQLite 测试未复现。
- **原因**：公告有效期查询把布尔条件和两段 `CURRENT_TIMESTAMP` 表达式写在同一个 `Where` 中，GORM/MySQL 路径生成了未绑定的 `enabled = ?`，MySQL 返回 1064。
- **修复**：将启用状态、开始时间和结束时间拆分为独立查询条件，并使用跨 SQLite/MySQL 一致的 `enabled = 1`。
- **回归**：34 条公开/内部 API 的真实 MySQL 扫描会持续断言该接口及其他接口均不返回 5xx。

## 执行记录

| 检查 | 命令 | 当前结果 |
|---|---|---|
| 普通测试 | `go test -count=1 ./...` | 本地通过 |
| 集成套件编译 | `go test -count=1 -tags=integration ./integration` | 本地编译通过；未配置 DSN 时按约定跳过 |
| 构建 | `go build ./...` | 本地通过 |
| 静态检查 | `go vet ./...` | 本地通过 |
| 真实 MySQL API 回归 | `CONTENT_SERVICE_TEST_DSN=... go test -count=1 -tags=integration ./integration -v` | 通过：34/34 已注册公开/内部路由子测试通过，核心成功契约通过，未解释 5xx 为 0 |
| Docker 镜像与运行时 | `docker build ...`；容器内探针和版本请求 | 镜像构建通过；容器 `healthy`；以 `uid=100(app)` 非 root 运行；livez/health/version 通过 |
| 微服务 E2E | `cd frontend && npm run test:e2e:micro` | 通过：网关、服务目录、内部接口隔离及同一 JWT 跨三服务访问，Playwright 2/2 |

## CI 阻断

`.github/workflows/content-service-ci.yml` 已开启可复用流水线的 `mysql_integration`。流水线将向集成测试提供隔离 MySQL DSN，集成测试位于镜像构建之前；任何公开/内部接口出现 5xx、路由清单漂移或核心契约失败都会阻断镜像构建。

## 验收判定

代码结构、普通测试、静态检查、真实 MySQL 34/34 API 回归、Docker 运行时和微服务 E2E 已在本地验证。最终合并仍须以本分支 GitHub Actions 绿灯和另一名有写权限成员 Review 为准；远端证据完成前不提前标记 #140 Done。
