# DanmakuStream E2E 测试计划（E2E Test Plan）

> **版本**：v1.0
> **覆盖范围**：13 用例（UC01~UC10、UC12~UC14）；UC11 空缺待确认，暂不纳入 100% 追溯率计算。
> **状态约定**：本文件只定义计划、框架选型与清单；实际通过/失败必须写入 [e2e-test-report.md](../testing/e2e-test-report.md)，不得在此预填通过。
> **关联追溯矩阵**：[traceability-matrix.md](../traceability/traceability-matrix.md)

---

## 1. 框架选型与环境

### 1.1 推荐框架：Playwright（Chromium Headless）

| 维度 | 选型依据 |
|---|---|
| 项目既有实现 | `frontend/e2e/` 已落地 Playwright，样例见 [uc13-admin.spec.ts](../frontend/e2e/uc13-admin.spec.ts)，可直接复用 fixture 与 global-setup |
| 浏览器能力 | 原生支持多浏览器（Chromium/Firefox/WebKit），可录制 trace、失败自动截图、录屏 |
| 登录机制 | 支持 `request` 上下文直接调后端 API 换 token，注入 localStorage，避免每次 UI 登录 |
| 工程化 | 与 Vite dev server 无缝集成（见 [playwright.config.ts](../frontend/playwright.config.ts)），CI 友好 |
| 报告 | 内置 HTML 报告 + 截图/trace 工件，符合交付物要求 |

> **备选**：Cypress / Selenium。不作为首推，因为项目已有 Playwright 落地实现，切换收益不足。

### 1.2 运行环境

```
操作系统  ：Windows / macOS / Linux（CI 推荐 Linux）
Node      ：>= v18
后端      ：go run backend/api/main.go -f backend/etc/config.yaml  监听 8080
MySQL     ：8.0+（docker 或用户态实例，库名 danmakustream）
SRS       ：4.0+（UC10 媒体链路使用，可走 docker-compose）
前端      ：Vite dev server（由 playwright.config.ts 的 webServer 自动拉起，代理 /api → :8080）
```

### 1.3 目录与命名约定

```
frontend/
  e2e/
    fixtures/auth.ts           ← loginViaApi / openAs 复用
    test-data.ts               ← 统一测试账号密码（Test1234!）与 E2E- 前缀数据
    global-setup.ts            ← 幂等注册账号 + 按 E2E-UCxx- 前缀清理并重建种子
    uc01-auth.spec.ts
    uc02-video.spec.ts
    uc03-upload.spec.ts
    uc04-review.spec.ts
    uc05-engagement.spec.ts
    uc06-library.spec.ts
    uc07-relationship.spec.ts
    uc08-membership.spec.ts
    uc09-schedule.spec.ts
    uc10-live.spec.ts
    uc12-analytics.spec.ts
    uc13-admin.spec.ts          ← 已落地
    uc14-tags.spec.ts
  playwright.config.ts         ← workers=1, retries=0, screenshot=on, trace=retain-on-failure
docs/testing/
  e2e-test-plan.md             ← 本文件
  e2e-test-report.md           ← 执行后真实报告（骨架与填空模板）
  artifacts/
    screenshots/                ← 每个用例 SC-UCxx-TCxx-(PASS/FAIL).png
    traces/                     ← 失败用例 zip trace
    html-report/                ← playwright 原生 HTML 报告
```

运行命令：

```bash
cd frontend
npx playwright install chromium           # 首次安装浏览器
npx playwright test --reporter=html        # 全量（13 个 spec，≥ 39 条）
npx playwright test e2e/uc06-library.spec.ts --reporter=html   # 单 UC
npx playwright show-report docs/testing/artifacts/html-report   # 查看结果
```

---

## 2. 13 用例清单（每 UC ≥ 3 条）

| 用例 | 编号 | 场景 | 关键步骤摘要 | 通过标准 |
|---|---|---|---|---|
| UC01 | E2E-UC01-01 | 注册并登录成功 | 访问 /register → 填唯一昵称 + `Test1234!` → 提交 → 自动跳登录 → 登录 → 首页已登录态 | 路由不含 /login；`authStore.isLoggedIn=true`；`/auth/me` 返回新用户 id |
| UC01 | E2E-UC01-02 | 错误密码无法登录 | /login 填正确昵称 + 错误密码 → 提交 | 停留在 /login；接口返回 401；顶部不显示个人入口 |
| UC01 | E2E-UC01-03 | 修改昵称与头像持久化 | 登录 → 个人主页 → 改昵称 + 上传头像图片 → 保存 → 刷新页面 | 新昵称和 avatar URL 刷新后保持；再登录仍一致 |
| UC02 | E2E-UC02-01 | 首页→详情跳转 | 首页点击任一已审核视频卡片 | 跳 `/video/:id`；详情标题与列表一致；播放器 video 标签存在 |
| UC02 | E2E-UC02-02 | 关键词搜索过滤 | 搜索框输 DB 中存在标题关键词 → 回车 | 结果仅匹配条目；不匹配关键词返回 0 条 |
| UC02 | E2E-UC02-03 | 未审核视频不可播放 | 直连 pending 视频 `/video/:id` | 页面提示不可播放；播放器无 src；接口按非 approved 过滤 |
| UC03 | E2E-UC03-01 | 创建 pending 投稿 | 创作者登录 → `/creator/upload` → 选 mp4 + 填标题 → 提交 | 成功后跳转我的视频；DB `status='pending'`；媒体文件落盘 |
| UC03 | E2E-UC03-02 | 取消上传无残留 | 进入上传页 → 选文件 → 未提交直接点离开 | 我的视频列表无新条目；DB video 表行数不变 |
| UC03 | E2E-UC03-03 | 创作中心状态一致 | DB 已知 3 条 pending/approved/rejected → 进创作中心 | 列表 3 条；状态标签与 DB `status` 一致 |
| UC04 | E2E-UC04-01 | 通过→首页可见 | moderator 登录 → 后台下拉「通过」→ 匿名访问首页 | 首页出现该视频；详情页可播放 |
| UC04 | E2E-UC04-02 | 拒绝→创作者端可见 | moderator 选择「拒绝」→ 创作者登录进我的视频 | 条目状态为已拒绝；首页不展示 |
| UC04 | E2E-UC04-03 | 普通用户越权 | 普通 user 访问 `/admin/videos` 或直接调 `PUT /admin/videos/:id/status` | 路由层重定向首页；接口 403；DB status 不变 |
| UC05 | E2E-UC05-01 | 弹幕按时点回放 | 已审核视频拖到 30s → 发弹幕「E2E UC05」→ 刷新 → 播放到 30s | 弹幕层在 30±1s 显示；内容一致 |
| UC05 | E2E-UC05-02 | 评论提交→列表出现 | 视频页填评论并提交 | 评论区新评论作者匹配当前用户 |
| UC05 | E2E-UC05-03 | 点赞/收藏切换计数一致 | 点赞→再点赞；收藏→再收藏 | 按钮状态切换；前端计数 +1/-1 与刷新后 `GET /videos/:id` likeCount/collectCount 一致 |
| UC06 | E2E-UC06-01 | 保存进度→继续观看恢复 | 播放到 120s → 等待节流 save → 关页 → `/me/history` 点「继续观看」 | `video.currentTime` 在 117~123s；进度 progress≈20（600s 时长） |
| UC06 | E2E-UC06-02 | 稍后再看跨端一致 | A 端点稍后再看 → B 端同账号进入 `/me/watchlater` | 两端列表包含同一 videoId；取消后两端同步移除 |
| UC06 | E2E-UC06-03 | 删除单条与清空生效 | 历史≥3 条 → 删单条 → 点清空（确认）| list 条数逐条递减；最终 total=0；取消二次确认则不清空 |
| UC07 | E2E-UC07-01 | A 关注 B→订阅页可见 | A 登录 → 访问 B 个人主页 → 点关注 → `/subscriptions` | B 出现在订阅列表；B 粉丝数 +1 与 `/users/:id` 响应一致 |
| UC07 | E2E-UC07-02 | 取消关注与分组 | 关注→分组加入特别关注→取消关注 | 分组内不再有 B；再次关注默认未特别关注 |
| UC07 | E2E-UC07-03 | 屏蔽后交互受限 | A 屏蔽 B → B 访问 A 主页尝试关注/发私信 | 关注按钮被禁用或 403；私信发送失败且不入库 |
| UC08 | E2E-UC08-01 | 方案配置→用户端可见 | 创作者进创作中心 → 配置会员价格/权益 → 启用 | 非会员观众访问创作者主页看到「订阅会员」入口与价格 |
| UC08 | E2E-UC08-02 | 下单→demo 支付→订阅生效 | 观众创建订单 → demo 支付 | 订单 status=paid；订阅表 status=active，expiresAt=当前+1 月 |
| UC08 | E2E-UC08-03 | 自动续订+到期 Worker | 开启自动续订 → DB 直改 expiresAt 过去 → 等 Worker 扫描 | Worker 更新状态或续订；关闭开关则到期变 expired |
| UC09 | E2E-UC09-01 | 主播创建→列表展示 | 主播填未来时间 + 标题 → 提交预约 | 列表出现该条；状态 pending；时间合法 |
| UC09 | E2E-UC09-02 | 观众预约+取消 | 另一观众点预约→ reminder_count +1 → 再点取消 | reminder_count 0→1→0；不产生重复 LiveReservation 行 |
| UC09 | E2E-UC09-03 | 非法时间/冲突拒绝 | 主播用过去时间；同主播同一时间两个预约 | 第一请求 400；第二请求 409；DB 最多 1 条 |
| UC10 | E2E-UC10-01 | 主播开播→观众可见 | 主播 `/live/studio/:id` 开始直播（SRS 可选推流模拟）| 房间状态=live；HLS src 可访问或进入 waiting |
| UC10 | E2E-UC10-02 | 观众互动→主播监看 | 观众端发弹幕/赠礼/SC；主播端监看浮窗 | 监看浮窗滚动出现；计数与接口一致 |
| UC10 | E2E-UC10-03 | 越权+下播状态 | 非房主调 `PUT /live/:id/end`；房主主动下播 | 前者 403；后者房间=ended，列表不再「直播中」 |
| UC12 | E2E-UC12-01 | 核心指标展示 | 创作者进入数据中心页面 | 播放/点赞/收藏/粉丝四类指标非负；与 CreatorDailyStat 聚合一致 |
| UC12 | E2E-UC12-02 | 切换时间范围 | 切换 7/30/90 天 | 折线图 x 轴变化；y 轴点数与 analytics 响应数组长度匹配 |
| UC12 | E2E-UC12-03 | 越权不可访问 | 普通用户直连 `/creator?uid=other` 或调他人接口 | 403 或重定向；不泄露他人数据 |
| UC13 | E2E-TC13-01 | 审核员处理内容 | 登录审核员 → 改视频状态、屏蔽弹幕 → 刷新页面 | 标签持久化；普通端同步 |
| UC13 | E2E-TC13-02 | 管理员修改角色 | 搜索用户 → 改角色 moderator → 刷新并重登 | role=moderator；权限生效 |
| UC13 | E2E-TC13-03 | 横幅/公告 CRUD | 新建→编辑→删除横幅与公告 | 后台列表+接口一致；最终无 E2E 残留 |
| UC13 | E2E-TC13-04 | 基础设施指标可见 | 进入 `/admin/infrastructure` | 存储/流量/CPU/在线四区块可见；单位与说明正确 |
| UC13 | E2E-TC13-05 | 普通用户越权 | 普通 user 访问后台页面/接口 | 路由守卫重定向；接口 403，无数据修改 |
| UC14 | E2E-UC14-01 | 标签亲和聚合展示 | 播放 3 个不同标签视频 → 产生历史 → `/me/tags` | top N 标签及次数与 video.tags 求和一致 |
| UC14 | E2E-UC14-02 | 推荐结果稳定 | 离线 recommendation 跑一次 ItemCF → 首页推荐区 | 推荐视频 id 与脚本输出一致；seed 固定则排序稳定 |
| UC14 | E2E-UC14-03 | 标签过滤切换 | 选中某标签 → 过滤历史 | 列表仅含该标签视频；再次取消过滤恢复全量 |

---

## 3. 数据准备（Seed Data）

### 3.1 测试账号（global-setup 幂等重置）

| 昵称 | 密码 | 期望角色 | 用途 |
|---|---|---|---|
| e2e-plain | Test1234! | user | UC01 注册、UC02 普通观众、UC05 互动、UC06 资料库、UC07 A、UC08 观众 |
| e2e-creator | Test1234! | creator | UC03 投稿、UC08 方案、UC09 主播、UC10 主播、UC12 数据分析 |
| e2e-mod | Test1234! | moderator | UC04 审核员、UC13 审核 |
| e2e-admin | Test1234! | admin | UC13 管理员、基础设施 |
| e2e-target | Test1234! | user | UC07 B（被关注）、UC08 主播；按需要 DB 直改 role |

### 3.2 视频与媒体（全部加前缀 `E2E-UCxx-`）

| 编号 | 标题 | 状态 | 时长 | tags | 用途 |
|---|---|---|---|---|---|
| V1 | E2E-UC02-已审核-1 | approved | 600 | 动画,战斗 | UC02 可播放、UC05 互动、UC06 历史进度 |
| V2 | E2E-UC14-已审核-2 | approved | 600 | 教程,Go | UC06 跨端 watchlater、UC14 标签聚合 |
| V3 | E2E-UC03-待审核 | pending | 300 | — | UC03 投稿流转、UC04 审核 |
| V4 | E2E-UC02-已拒绝 | rejected | 300 | — | UC02 不可访问、UC06 JOIN 自动过滤 |

### 3.3 直播与预约

- 1 个 LiveRoom：owner=e2e-creator，初始 idle。
- 1 个未来 24h 的 LiveSchedule（用于 UC09）。
- SRS 若不可用，UC10 仅校验 UI+接口状态，单条加 `[MEDIA-REQUIRED]` 标签单独阻塞管理。

---

## 4. 执行与阻断策略

1. **串行执行**：workers=1，避免并发账号/DB 冲突。
2. **失败工件**：失败自动截图 + trace zip → 存入 `docs/testing/artifacts/screenshots` 与 `traces/`。
3. **CI 阻断**：任一 E2E 失败，不得合并 PR 与发布镜像。
4. **回归**：改 handler 或页面至少重跑受影响 UC 的 3 条；milestone 节点全量回归。
