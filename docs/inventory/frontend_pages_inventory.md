# 前端页面盘点清单

| 盘点日期 | 盘点人 | 项目 |
|---------|--------|------|
| 2026-08-25 | [郭谦] | DanmakuStream |

> **说明**：基于路由配置文件自动化扫描生成，页面功能状态需进一步确认。

---

## 一、页面清单

### 1. 公开页面（无需登录）

| 序号 | 页面路径 | 页面名称 | 对应组件 | 状态 | 备注 |
|------|---------|---------|---------|------|------|
| 1 | `/` | 首页 | HomePage | 待确认 | 首页入口 |
| 2 | `/login` | 登录页 | LoginPage | 待确认 | 支持 redirect 参数 |
| 3 | `/register` | 注册页 | RegisterPage | 待确认 | - |
| 4 | `/live` | 直播列表 | LiveListPage | 待确认 | 支持 `?create=1` 参数 |
| 5 | `/live/:id` | 直播间 | LiveRoomPage | 待确认 | 动态路由 |
| 6 | `/video/:id` | 视频详情 | VideoDetailPage | 待确认 | 动态路由 |
| 7 | `/user/:id` | 用户主页 | UserProfilePage | 待确认 | 动态路由 |
| 8 | `/404` | 404页面 | NotFoundPage | 待确认 | 通配符路由 `/:pathMatch(.*)*` |

### 2. 认证页面（需登录）

| 序号 | 页面路径 | 页面名称 | 对应组件 | 状态 | 备注 |
|------|---------|---------|---------|------|------|
| 9 | `/creator` | 创作者中心 | CreatorDashboardPage | 待确认 | requiresAuth |
| 10 | `/creator/upload` | 视频上传 | VideoUploadPage | 待确认 | requiresAuth |
| 11 | `/subscriptions` | 订阅列表 | SubscriptionPage | 待确认 | requiresAuth |
| 12 | `/me/:kind` | 用户资料库 | UserLibraryPage | 待确认 | requiresAuth，支持 history/liked/collections/downloads |
| 13 | `/me/tags` | 标签偏好 | TagAffinityPage | 待确认 | requiresAuth |

### 3. 管理后台（需权限）

| 序号 | 页面路径 | 页面名称 | 对应组件 | 权限要求 | 状态 | 备注 |
|------|---------|---------|---------|---------|------|------|
| 14 | `/admin` | 管理仪表盘 | AdminDashboardPage | requiresStaff | 待确认 | - |
| 15 | `/admin/danmaku` | 弹幕管理 | AdminDanmakuPage | requiresStaff | 待确认 | 核心功能 |
| 16 | `/admin/videos` | 视频管理 | AdminVideosPage | requiresStaff | 待确认 | - |
| 17 | `/admin/users` | 用户管理 | AdminUsersPage | requiresAdmin | 待确认 | - |
| 18 | `/admin/operations` | 运维管理 | AdminOperationsPage | requiresAdmin | 待确认 | - |
| 19 | `/admin/infrastructure` | 基础设施 | AdminInfrastructurePage | requiresAdmin | 待确认 | - |

---

## 二、页面分类汇总

| 模块 | 页面数量 | 页面路径 |
|------|---------|---------|
| 公开页面 | 8 | `/`, `/login`, `/register`, `/live`, `/live/:id`, `/video/:id`, `/user/:id`, `/404` |
| 认证页面 | 5 | `/creator`, `/creator/upload`, `/subscriptions`, `/me/:kind`, `/me/tags` |
| 管理后台 | 6 | `/admin`, `/admin/danmaku`, `/admin/videos`, `/admin/users`, `/admin/operations`, `/admin/infrastructure` |
| **合计** | **19** | |

---

## 三、路由参数说明

| 页面 | 参数 | 用途 |
|------|------|------|
| `/login` | `?redirect=xxx` | 登录后跳转地址 |
| `/live` | `?create=1` | 快速创建直播 |
| `/me/:kind` | `history/liked/collections/downloads` | 用户资料分类 |
| `/:pathMatch(.*)*` | - | 404兜底路由 |

---

## 四、权限矩阵

| 权限标识 | 页面数量 | 页面列表 |
|---------|---------|---------|
| 无需登录 | 8 | `/`, `/login`, `/register`, `/live`, `/live/:id`, `/video/:id`, `/user/:id`, `/404` |
| requiresAuth | 5 | `/creator`, `/creator/upload`, `/subscriptions`, `/me/:kind`, `/me/tags` |
| requiresStaff | 3 | `/admin`, `/admin/danmaku`, `/admin/videos` |
| requiresAdmin | 3 | `/admin/users`, `/admin/operations`, `/admin/infrastructure` |

---

## 五、待确认事项

- [ ] 上述19个页面是否全部完整？是否有产品规划但未在路由中配置的页面？
- [ ] 各页面的功能状态（已完成/开发中/未开始/需重构）？
- [ ] `/` 首页的具体功能范围是什么？（视频推荐/直播推荐/混合？）
- [ ] 管理后台的 `requiresStaff` 和 `requiresAdmin` 权限划分是否正确？
- [ ] 是否有移动端适配要求？是否有独立的H5页面？
- [ ] 本期迭代必须交付的页面是哪些？

---

## 六、后续行动

| 序号 | 任务 | 负责人 | 截止日期 |
|------|------|--------|---------|
| 1 | 与产品确认页面完整性和状态 | [开发] | 本周内 |
| 2 | 标注每个页面的完成状态 | [开发] | 本周内 |
| 3 | 识别缺失页面，评估工作量 | [开发+产品] | 本周内 |