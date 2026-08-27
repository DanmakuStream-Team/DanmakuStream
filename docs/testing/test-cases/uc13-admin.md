# UC13 测试用例设计

> 状态说明：本文只定义测试范围和断言。未执行的用例统一标记为“待实现/待执行”，不能提前填写通过。

## 1. 测试范围

- 视频审核与弹幕屏蔽。
- 用户角色查询和修改。
- 横幅、公告的查询与增删改。
- 存储、流量、CPU、在线和直播指标查询。
- 普通用户、审核员、管理员之间的权限边界。

## 2. 单元测试（UNIT-TC13）

| 编号 | 测试目标 | 输入/条件 | 关键断言 | 当前状态 |
|---|---|---|---|---|
| UNIT-TC13-01 | AdminMiddleware 权限 | role=user/moderator/admin | 仅 admin 放行，其余返回 403 | 通过（2026-08-26） |
| UNIT-TC13-02 | StaffMiddleware 权限 | role=user/moderator/admin | moderator/admin 放行，user 返回 403 | 通过（2026-08-26） |
| UNIT-TC13-03 | 用户角色参数校验 | user/moderator/admin/非法值 | 三种合法值通过，非法值返回 400 | 通过（2026-08-26） |
| UNIT-TC13-04 | 横幅参数校验 | 空标题/合法标题 | 空标题返回 400 | 通过（2026-08-26） |
| UNIT-TC13-05 | 公告参数校验 | 空内容、非法时间、合法时间 | 对应返回 400 或构造成功 | 通过（2026-08-26） |
| UNIT-TC13-06 | 审核状态校验 | 合法/非法视频状态 | 非法状态返回 400 | 通过（2026-08-26） |

## 3. API/集成测试（INT-TC13）

| 编号 | 接口/场景 | 前置条件 | 关键断言 | 当前状态 |
|---|---|---|---|---|
| INT-TC13-01 | GET /admin/videos | moderator token | 200，返回分页待审视频 | 通过（2026-08-26） |
| INT-TC13-02 | PUT /admin/videos/:id/status | 待审且媒体就绪 | 审核状态持久化；公开查询结果同步变化 | 通过（2026-08-26） |
| INT-TC13-03 | PUT /admin/videos/:id/status | 视频未转码/不存在 | 返回 403/404，状态不被错误修改 | 通过（2026-08-26） |
| INT-TC13-04 | PUT /admin/danmaku/:id/block | moderator token | 200，blocked=true | 通过（2026-08-26） |
| INT-TC13-05 | PUT /admin/users/:id/role | admin token | 200，数据库角色与响应一致 | 通过（2026-08-26） |
| INT-TC13-06 | PUT /admin/users/:id/role | moderator/user token | 403，数据库不变 | 通过（2026-08-26） |
| INT-TC13-07 | Banner CRUD | admin token | 创建、更新、查询、删除结果一致 | 通过（2026-08-26，见缺陷 D1） |
| INT-TC13-08 | Announcement CRUD | admin token | 内容和时间持久化正确 | 通过（2026-08-26，见缺陷 D1） |
| INT-TC13-09 | GET /admin/infrastructure | admin token | 返回 storage/traffic/cpu/online 及来源 | 通过（2026-08-26） |
| INT-TC13-10 | 任意后台接口 | 无 token/无效 token | 返回 401，无数据修改 | 通过（2026-08-26） |

## 4. 端到端测试（E2E-TC13）

| 编号 | 端到端场景 | 操作步骤摘要 | 通过标准 | 当前状态 |
|---|---|---|---|---|
| E2E-TC13-01 | 审核员处理内容 | 登录审核员→审核视频→屏蔽弹幕→刷新页面 | 状态保存且普通页面结果同步 | 通过（2026-08-26） |
| E2E-TC13-02 | 管理员修改角色 | 登录管理员→搜索用户→修改角色→刷新 | 新角色保持，权限随之生效 | 通过（2026-08-26） |
| E2E-TC13-03 | 管理员维护运营内容 | 新建/修改/删除横幅和公告 | 后台与展示端结果一致 | 通过（2026-08-26，触发 D1 修复） |
| E2E-TC13-04 | 管理员查看基础设施 | 进入监控页→刷新指标 | 指标、单位、阈值状态和来源可见 | 通过（2026-08-26） |
| E2E-TC13-05 | 普通用户越权 | 普通用户访问后台页面/API | 页面拦截或接口返回 403，无数据修改 | 通过（2026-08-26） |

## 5. 测试证据要求

- 保存测试命令、环境版本、测试数据初始化方式和原始输出。
- 每个失败结果关联 GitHub Issue，修复后保留重新执行结果。
- CI 中的失败测试必须阻断镜像和部署步骤。
- 测试执行后再把“当前状态”改为通过或失败，并在追溯表填写报告链接。

## 6. 测试执行记录（2026-08-26，INT-TC13 全量）

**环境**：后端 `go run api/main.go -f etc/config.yaml`（main 分支检出），API 前缀 `/api/v1`；MySQL 8.0.46（用户态实例，socket `/home/haoyue/dms-mysql.sock`，库 `danmakustream`，表由 GORM 自动迁移生成）；认证方式 `Authorization: Bearer <JWT>`。

**测试数据**：API 注册三个用户（昵称 tuser/tmod/tadmin，密码 `Test1234!`，数据库直改 role 为 user/moderator/admin）后重新登录取 token；视频 1（pending + video_url 就绪）、视频 2（pending + 无 video_url，模拟转码中）、弹幕 1 均为数据库直插。

**执行方式**：curl 逐条调用，断言 HTTP 状态码 + 响应体 + 数据库实际值（`mysql` 直查）+ 公开接口同步结果。

**结果**：INT-TC13-01～10 全部通过。要点：

- 01：moderator 请求 200，返回 `data.list` 待审视频列表。
- 02：视频 1 审核通过后数据库 `status=approved`，公开视频列表立即包含该视频。
- 03：转码中视频返回 403（“视频文件未上传成功”），不存在 ID 返回 404，两者状态均未被修改。
- 04：屏蔽后数据库 `blocked=1`。
- 05/06：admin 改角色成功且持久化；moderator 改角色 403 且数据库不变。
- 07/08：横幅、公告 CRUD 全链路一致；公告非法时间返回 400“开始时间格式错误”（覆盖 UNIT-TC13-05 的 API 层路径）；删除为软删除（`deleted_at`），列表查询正确排除。
- 09：返回 storage（路径/总量/用量/告警位）、traffic（当月/当日下行字节及来源）、cpu（/proc/stat 使用率）、online（当前在线/最高并发/直播间与观看数）。
- 10：无 token 与伪造 token 均返回 401。

**发现的问题**：

- **D1（缺陷，待修）**：`POST /admin/banners`、`POST /admin/announcements` 的创建响应使用 Go 结构体默认字段名（`ID`、`ImageURL`、`StartedAt`），与项目其余接口统一的 camelCase（更新/删除响应中的 `id`）不一致，前端对接易踩坑。建议在 Banner/Announcement 模型或响应结构上补 `json` tag。
- **D2（备忘，非缺陷）**：登录按昵称+密码校验（`LoginReq` 虽含 `username` 字段但匹配逻辑使用昵称），新注册用户 username 为自动生成的数字。测试与联调时统一使用昵称登录。

**未执行**：E2E-TC13-01～05（待前端联调后执行）。

## 7. 单元测试执行记录（2026-08-26）

**测试代码**（新增）：

- `backend/internal/middleware/auth_test.go` — UNIT-TC13-01/02，httptest 直接驱动两个中间件，断言状态码与后续 handler 是否被触达（含角色缺失场景）。
- `backend/internal/handler/v1/admin/admin_handler_test.go` — UNIT-TC13-03/04/05：`isValidRole` 表驱动测试 + `UpdateUserRoleHandler`/`CreateBannerHandler` 的 httptest 400 分支 + `buildAnnouncement` 合法/非法输入表驱动测试（RFC3339 与 `2006-01-02 15:04:05` 两种时间格式）。
- `backend/internal/handler/v1/video/video_handler_test.go` — UNIT-TC13-06：`isValidVideoStatus` 表驱动测试。

**可测试性重构**（行为等价）：`admin_handler.go` 将内联的三元角色判断提取为 `isValidRole(role string) bool`，handler 改为调用该函数；其余校验（`buildAnnouncement`、banner 标题 TrimSpace 判空）原已具备独立可测结构，未改动。

**运行命令与结果**：

```bash
cd backend
GOCACHE=/tmp/go-build go test ./...
```

6 个测试函数（13 个子测试）全部通过，全后端无编译/回归问题。原始报告：`docs/testing/reports/uc13-unit-2026-08-26.txt`（总数 6、通过 6、失败 0）。

**结论**：单元测试未发现新的业务缺陷；此前 API 层发现的 D1（JSON 字段命名不一致）不属单元层断言范围，仍待修。

## 8. API 自动化测试（2026-08-26）

原手工 curl 序列已固化为自动脚本 **`tests/api/uc13-admin-test.sh`**：

- 覆盖 INT-TC13-01～10，共 33 条断言（HTTP 状态码 + 响应字段 + 数据库实际值 + 公开接口同步 + 软删除后列表排除）。
- 自带幂等测试数据准备（注册/重置三个测试账号角色、按时间戳清理并重建视频与弹幕），可用 `API_BASE`、`MYSQL_CMD` 环境变量适配不同环境（本地用户态实例或 Docker）。
- 任一断言失败即以非零状态退出，可直接接入 CI 作为阻断步骤。
- 本次运行 33/33 通过，exit 0。原始报告：`docs/testing/reports/uc13-api-2026-08-26.txt`。

**前端就绪情况**：UC13 六个管理页面（Dashboard/Videos/Danmaku/Users/Operations/Infrastructure）均已实现并注册路由，带 `requiresStaff`/`requiresAdmin` 守卫，可直接进入 Playwright E2E 编写阶段。

## 9. E2E 自动化测试（2026-08-26，Playwright）

**框架与目录**：Playwright（`@playwright/test` + Chromium headless），代码位于 `frontend/e2e/`：

- `uc13-admin.spec.ts` — E2E-TC13-01～05 五条用例
- `fixtures/auth.ts` — `loginViaApi`（API 换 token）与 `openAs`（注入 localStorage 会话后打开页面）
- `test-data.ts` — 固定测试账号（tuser/tmod/tadmin/e2eplain，密码 `Test1234!`）
- `global-setup.ts` — 每次运行前注册账号、固化角色、按 `E2E-UC13-` 前缀清理并重建待审视频/弹幕
- `frontend/playwright.config.ts` — 单 worker 串行、失败自动截图/录屏/trace、自动拉起 Vite dev server（代理到后端 8080）

**运行命令**：`cd frontend && npx playwright test`（HTML 报告输出到 `docs/testing/reports/uc13-e2e/`，用 `npx playwright show-report <该目录>` 查看）

**结果**：5/5 通过。要点：

- 01：真实 UI 登录 → 下拉切换审核状态为“通过” → 屏蔽弹幕 → 刷新后“已屏蔽”保持。
- 02：搜索用户 → 角色改为“内容审核员/版主” → 刷新重搜保持 → 重新登录取到的角色为 moderator（权限生效）。
- 03：横幅与公告的新建/编辑/删除全链路，最终以列表接口断言测试数据已清理。
- 04：存储/流量/CPU/在线四个区块、已用/剩余容量与流量来源说明均可见。
- 05：普通用户访问 `/admin/videos` 被路由守卫重定向回首页；直连接口 403。

**触发并修复的缺陷（D1 关闭）**：E2E-TC13-03 首次运行时发现横幅/公告在页面上不显示标题和图片——后端 `SiteBanner`/`SiteAnnouncement` 模型缺 json tag，接口返回 Go 默认字段名（`ID`/`Title`/`ImageURL`），前端模板读 `item.title` 为 undefined。修复：两个模型改为显式字段并补 camelCase json tag（`backend/internal/model/mysql/models.go`），gorm 行为不变。修复后 E2E、API（33/33）、单元（全过）三层回归均绿。

**回归汇总（2026-08-26）**：单元 6/6、API 33/33、E2E 5/5，全部通过；原始报告见 `docs/testing/reports/`（`uc13-unit-2026-08-26.txt`、`uc13-api-2026-08-26.txt`、`uc13-e2e/`）。
