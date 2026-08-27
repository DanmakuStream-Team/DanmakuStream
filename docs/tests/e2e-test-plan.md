# DanmakuStream E2E 测试计划

> **版本**：v1.0（骨架）
> **覆盖目标**：13 个用例（UC01~UC10、UC12~UC14；UC11 空缺待确认，不纳入本次 100% 追溯率计算）
> **状态约定**：本文件只定义计划、框架选型与用例清单，**实际通过/失败必须在执行后填入《E2E 测试报告》**，不得在此预填通过。

---

## 1. 框架选型与环境

### 1.1 推荐框架：Playwright（Chromium Headless）

| 维度 | 选型依据 |
|---|---|
| 项目既有实现 | `frontend/e2e/` 已落地 Playwright，样例见 [uc13-admin.spec.ts](file:///d:/DanmakuStream/frontend/e2e/uc13-admin.spec.ts)，可直接复用 fixture 与 global-setup |
| 浏览器能力 | 原生支持多浏览器（Chromium/Firefox/WebKit），可录制 trace、失败自动截图、录屏 |
| 登录机制 | 支持 `request` 上下文直接调后端 API 换 token，注入 localStorage，避免每次 UI 登录 |
| 工程化 | 与 Vite dev server 可无缝集成（见 [playwright.config.ts](file:///d:/DanmakuStream/frontend/playwright.config.ts)），CI 友好 |
| 报告 | 内置 HTML 报告 + 截图/trace 工件，符合交付物要求 |

> **备选**：Cypress / Selenium。不作为首推，因为项目已有 Playwright 落地实现，切换收益不足。

### 1.2 运行环境

```
操作系统  ：Windows / macOS / Linux（CI 推荐 Linux）
Node      ：>= v18
后端      ：go run backend/api/main.go -f backend/etc/config.yaml  监听 8080
MySQL     ：8.0+（docker run 或用户态实例，库名 danmakustream）
SRS       ：4.0+（UC10 媒体链路使用，可走 docker-compose）
前端      ：Vite dev server（由 playwright.config.ts 的 webServer 自动拉起，代理 /api → :8080）
```

### 1.3 启动脚本与目录约定

```
frontend/
  e2e/
    fixtures/auth.ts           ← loginViaApi / openAs 复用
    test-data.ts               ← 测试账号（e2euser1/e2euser2/... 密码 Test1234!）
    global-setup.ts            ← 运行前注册账号、准备视频/直播预约/弹幕等种子数据
    uc01-auth.spec.ts
    uc02-video.spec.ts
    ...
    uc14-tags.spec.ts
  playwright.config.ts         ← workers=1, retries=0, trace=retain-on-failure, screenshot=on
docs/tests/reports/<UCxx>-e2e-report/   ← html 报告输出目录，按用例号分子目录
```

运行命令：

```bash
cd frontend
# 一次性安装浏览器
npx playwright install chromium
# 运行全部 E2E（约 13 个 spec，每 spec 3 条用例，总约 39 条）
npx playwright test --reporter=html
# 单跑某个 UC，比如 UC06
npx playwright test e2e/uc06-library.spec.ts --reporter=html
# 打开报告
npx playwright show-report docs/tests/reports/e2e-report-html
```

---

## 2. 13 个用例的 E2E 清单（每个 UC ≥ 3 条）

### 2.1 UC01：用户注册、登录与资料维护

| 编号 | 场景 | 关键步骤摘要 | 通过标准 |
|---|---|---|---|
| E2E-UC01-01 | 注册并登录成功 | 访问 /register → 填唯一昵称 + 密码 `Test1234!` → 提交 → 跳登录 → 登录 → 进首页已登录态 | 路由不包含 /login；`authStore.isLoggedIn=true`；`/auth/me` 返回新用户 id |
| E2E-UC01-02 | 错误密码无法登录 | 登录页填正确昵称 + 错误密码 → 提交 | 停留在 /login；接口返回 401；首页不展示个人入口 |
| E2E-UC01-03 | 修改昵称与头像并持久 | 登录 → 进个人主页 → 修改昵称、上传头像图片 → 保存 → 刷新页面 | 新昵称和头像 URL 显示；再登录保持一致 |

### 2.2 UC02：视频发现、搜索与播放

| 编号 | 场景 | 关键步骤摘要 | 通过标准 |
|---|---|---|---|
| E2E-UC02-01 | 首页→详情跳转 | 首页列表点击任一已审核视频卡片 | 跳 `/video/:id`；详情页标题与列表标题一致；播放器 video 标签存在 |
| E2E-UC02-02 | 关键词搜索 | 首页搜索框输入 DB 中存在的视频标题关键词 → 回车 | 结果列表仅包含匹配条目；不匹配关键词返回 0 条 |
| E2E-UC02-03 | 未审核视频不可访问 | 构造 pending 视频 → 直连 `/video/:id` | 页面提示「视频不可播放」；播放器无 source；接口返回非 approved 状态过滤 |

### 2.3 UC03：创作者投稿与状态跟踪

| 编号 | 场景 | 关键步骤摘要 | 通过标准 |
|---|---|---|---|
| E2E-UC03-01 | 上传视频创建 pending 投稿 | 创作者登录 → 进 `/creator/upload` → 选 mp4 → 填标题/简介 → 提交 | 成功后跳转我的视频；DB 存在 status=pending 记录；媒体文件落盘 |
| E2E-UC03-02 | 取消上传不产生残留 | 进入上传页 → 选文件 → 未提交时离开（或取消按钮） | 我的视频列表无新条目；DB video 表行数不变 |
| E2E-UC03-03 | 创作中心状态与 DB 一致 | 查 DB 已知 3 条视频（pending/approved/rejected）→ 进创作中心 | 列表显示 3 条；每条状态标签与 DB `status` 字段一致 |

### 2.4 UC04：视频审核与发布

| 编号 | 场景 | 关键步骤摘要 | 通过标准 |
|---|---|---|---|
| E2E-UC04-01 | 审核员通过→首页可见 | moderator 登录 → 后台待审视频下拉「通过」→ 打开首页匿名访问 | 首页出现该视频；详情页可播放 |
| E2E-UC04-02 | 拒绝→创作者端可见 | moderator 选择「拒绝」→ 创作者登录进我的视频 | 对应条目状态为「已拒绝」；首页不展示该视频 |
| E2E-UC04-03 | 普通用户越权 | 普通 user token 直接调 `PUT /admin/videos/:id/status`，或访问 `/admin/videos` 路由 | 路由层重定向回首页；接口 403；DB status 不变 |

### 2.5 UC05：视频观看互动

| 编号 | 场景 | 关键步骤摘要 | 通过标准 |
|---|---|---|---|
| E2E-UC05-01 | 弹幕后刷新按时点回放 | 已审核视频详情页 → 拖到 30 秒 → 发送弹幕「E2E UC05」→ 刷新页面 → 播放到 30 秒 | 弹幕层在 30±1s 时显示该条；内容一致 |
| E2E-UC05-02 | 发表评论→列表出现 | 视频详情页填评论并提交 | 评论区出现新评论，作者为当前用户 |
| E2E-UC05-03 | 点赞/收藏切换且计数一致 | 点赞→再点赞；收藏→再收藏 | 按钮状态切换；前端计数 +1/-1 与刷新后 GET /videos/:id 响应中 likeCount/collectCount 一致 |

### 2.6 UC06：个人视频资料库管理

| 编号 | 场景 | 关键步骤摘要 | 通过标准 |
|---|---|---|---|
| E2E-UC06-01 | 保存进度→继续观看恢复 | 播放已审核视频到 120s（等待节流 save）→ 关页 → 进 `/me/history` 点「继续观看」 | `video.currentTime` 在 117~123s；进度条 progress≈20（600s 时长） |
| E2E-UC06-02 | 稍后再看切换跨端一致 | 视频页点稍后再看 → `/me/watchlater` 可见 → 另一浏览器登录同账号 → 打开 `/me/watchlater` | 两端列表均包含该 videoId；再点取消 → 两端同时移除 |
| E2E-UC06-03 | 删除单条与清空生效 | 历史≥3 条 → 删单条 → 列表少 1 → 点「清空」确认 → 空态页 | GET history list 数量逐条递减；最终 total=0；ElMessageBox 取消则不清空 |

### 2.7 UC07：关注关系与内容通知

| 编号 | 场景 | 关键步骤摘要 | 通过标准 |
|---|---|---|---|
| E2E-UC07-01 | A 关注 B→订阅页可见 | A 登录 → 访问 B 的个人主页 → 点关注 → 进 `/subscriptions` | B 出现在订阅列表；B 的粉丝数 +1 与 `/users/:id` 响应一致 |
| E2E-UC07-02 | 取消关注+分组 | 关注后分组加入「特别关注」→ 取消关注 | 分组内不再有 B；再次关注状态为未特别关注（默认） |
| E2E-UC07-03 | 屏蔽限制交互 | A 屏蔽 B → B 访问 A 主页尝试关注/发私信 | 关注按钮被禁用或返回 403；私信发送失败且不入库 |

### 2.8 UC08：创作者会员订阅

| 编号 | 场景 | 关键步骤摘要 | 通过标准 |
|---|---|---|---|
| E2E-UC08-01 | 配置方案→用户端可见 | 创作者登录 → 进入创作中心 → 配置会员价格/权益 → 启用 | 非会员观众访问创作者主页看到「订阅会员」入口与价格 |
| E2E-UC08-02 | 下单→demo 支付→订阅生效 | 观众进入订阅页 → 创建订单 → demo 支付 | 订单状态 paid；订阅表 status=active，expiresAt 为当前 +1 月 |
| E2E-UC08-03 | 自动续订开关+到期处理 | 订阅开启自动续订 → 把 expiresAt 调到过去（DB 直改）→ 等 Worker 扫描 | Worker 更新状态/续订；开关关闭则到期后自动变成 expired |

### 2.9 UC09：直播预约与用户预约

| 编号 | 场景 | 关键步骤摘要 | 通过标准 |
|---|---|---|---|
| E2E-UC09-01 | 主播创建→列表展示 | 主播登录 → 直播页填未来时间 + 标题 → 提交预约 | 直播列表出现该条；状态 pending；时间格式合法 |
| E2E-UC09-02 | 观众预约+取消 | 另一观众登录 → 点预约 → 预约人数 +1 → 再点取消 | reminder_count 从 0→1→0；接口不产生重复 LiveReservation 行 |
| E2E-UC09-03 | 非法时间/冲突拒绝 | 主播用过去时间建预约；同主播相同时间两个预约 | 第一请求 400；第二请求 409；DB 仅最多 1 条 |

### 2.10 UC10：直播发布、观看与实时互动

| 编号 | 场景 | 关键步骤摘要 | 通过标准 |
|---|---|---|---|
| E2E-UC10-01 | 主播开播→观众可见 | 主播进入 `/live/studio/:id` → 点开始直播（SRS 可选推流模拟）→ 观众端 `/live/:id` | 房间状态=live；播放 HLS src 可访问或进入 waiting 态 |
| E2E-UC10-02 | 观众发弹幕/SC→主播监看 | 观众端发送弹幕、赠礼、SC；主播端监看浮窗 | 监看浮窗滚动出现；互动计数与接口一致 |
| E2E-UC10-03 | 非房主越权+下播状态 | 非房主 token 调 `PUT /live/:id/end`；房主主动下播 | 前者 403；后者房间状态=ended，列表不再显示「直播中」 |

### 2.11 UC12：创作者数据分析

| 编号 | 场景 | 关键步骤摘要 | 通过标准 |
|---|---|---|---|
| E2E-UC12-01 | 核心指标展示 | 创作者数据中心 → 进入页面 | 展示播放、点赞、收藏、粉丝四类指标；数值非负，来源与 CreatorDailyStat 聚合一致 |
| E2E-UC12-02 | 切换统计范围 | 切换 7/30/90 天 | 折线图 x 轴日期范围变化；y 轴数值与后端 analytics 响应数组长度匹配 |
| E2E-UC12-03 | 越权不可访问 | 普通用户直连 `/creator?uid=other` 或调他人 analytics 接口 | 返回 403 或重定向；不泄露他人数据 |

### 2.12 UC13：平台审核、权限、运营与基础设施管理

> **已落地**，样例位于 [uc13-admin.spec.ts](file:///d:/DanmakuStream/frontend/e2e/uc13-admin.spec.ts)，本清单沿用其 5 条编号。

| 编号 | 场景 | 关键步骤摘要 | 通过标准 |
|---|---|---|---|
| E2E-TC13-01 | 审核员审核视频、屏蔽弹幕 | 登录审核员 → 改视频状态、屏蔽弹幕 → 刷新页面 | 标签持久化；普通端同步 |
| E2E-TC13-02 | 管理员改用户角色 | 搜索用户 → 改角色 moderator → 刷新并重登 | 返回 role=moderator；权限生效 |
| E2E-TC13-03 | 横幅/公告 CRUD | 新增→编辑→删除横幅与公告 | 后台列表+接口一致；最终无 E2E 残留 |
| E2E-TC13-04 | 基础设施指标可见 | 进入 `/admin/infrastructure` | 存储/流量/CPU/在线四区块可见；单位与说明正确 |
| E2E-TC13-05 | 普通用户越权 | 普通 user 访问后台页面/接口 | 路由守卫重定向；接口 403，无数据修改 |

### 2.13 UC14：用户标签偏好与个性化推荐

| 编号 | 场景 | 关键步骤摘要 | 通过标准 |
|---|---|---|---|
| E2E-UC14-01 | 标签亲和聚合 | 播放 3 个不同标签视频 → 产生历史 → 进入 `/me/tags` | 页面展示 top N 标签及次数；数值与历史视频 tags 字段求和一致 |
| E2E-UC14-02 | 推荐结果稳定 | 离线 recommendation 跑一次 ItemCF → 首页推荐区拉取 | 推荐视频 id 与脚本输出一致；刷新不产生随机排序（或 seed 固定） |
| E2E-UC14-03 | 标签过滤切换 | 在标签页选择某标签 → 过滤历史 | 列表仅包含该标签视频；再次点取消过滤恢复全量 |

---

## 3. 数据准备（Seed Data）

### 3.1 用户账号（global-setup 中注册或幂等重置）

| 昵称 | 密码 | 期望角色 | 用途 |
|---|---|---|---|
| e2e-plain | Test1234! | user | UC01 注册、UC02 普通观众、UC05 互动、UC06 资料库、UC07 A、UC08 观众 |
| e2e-creator | Test1234! | creator | UC03 投稿、UC08 方案、UC09 主播、UC10 主播、UC12 数据分析 |
| e2e-mod | Test1234! | moderator | UC04 审核员、UC13 审核 |
| e2e-admin | Test1234! | admin | UC13 管理员、基础设施操作 |
| e2e-target | Test1234! | user | UC07 B（被关注），UC08 主播等；DB 直改 role 按需切换 |

### 3.2 视频与媒体

1. 固定 2 条 approved 视频（时长 600s）：V1、V2（tags 分别为「动画,战斗」「教程,Go」，用于 UC14）。
2. 固定 1 条 pending：V3，用于 UC03/UC04 状态流转。
3. 固定 1 条 rejected：V4，用于 UC02 不可访问场景。
4. 所有种子数据加前缀 `E2E-UCxx-`，global-setup 结束时按前缀清理（幂等）。

### 3.3 直播与预约

- 1 个 LiveRoom（owner=e2e-creator，初始 idle），用于 UC10。
- 1 个未来 24 小时的 LiveSchedule，用于 UC09。
- SRS 若未就绪，UC10 仅校验 UI 流转与接口状态，媒体链路单独打 `[MEDIA-REQUIRED]` 标记。

---

## 4. 执行与阻断策略

1. **串行执行**：workers=1，避免并发下账号/DB 冲突。
2. **失败工件**：失败自动截图 + 录制 trace → 工件存 `docs/tests/reports/<uc>/artifacts/`。
3. **CI 阻断**：若任一 E2E 失败，不得合并 PR 与发布镜像（可由 GitHub Actions 统一执行）。
4. **回归**：修改 handler 或页面后，至少重跑受影响 UC 的 3 条用例；全量回归在 milestone 节点统一执行。

---

## 5. 交付物关联

| 交付物 | 路径 |
|---|---|
| 13 用例追溯表 | [all-uc-traceability.md](file:///d:/DanmakuStream/docs/traceability/all-uc-traceability.md) |
| E2E 报告骨架 | [e2e-test-report-skeleton.md](file:///d:/DanmakuStream/docs/tests/reports/e2e-test-report-skeleton.md) |
| 最终追溯表骨架 | [final-traceability-skeleton.md](file:///d:/DanmakuStream/docs/traceability/final-traceability-skeleton.md) |
| UC06 用例说明 | [UC06.md](file:///d:/DanmakuStream/docs/models/usercase/UC06.md) |
