# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

DanmakuStream（灵视 VisionLive）是一个视频社区平台，支持视频点播、实时弹幕、直播间、评论互动和后台管控。前后端分离的多语言微服务架构，课程大作业。

## 常用命令

### 本地开发

```bash
# 启动基础设施（仅 MySQL）
docker compose up mysql -d

# 后端（Go）
cd backend
go mod tidy
go run api/main.go -f etc/config.yaml       # 默认监听 :8080

# 前端（Vue 3）
cd frontend
npm install
npm run dev                                   # 默认监听 :5173，自动代理 /api、/media、/ws 到后端

# AI 服务（Python）
cd ai-service
cp .env.example .env
pip install -r requirements.txt
uvicorn app.main:app --reload --port 8000
```

### 构建与检查

```bash
# 后端编译检查
cd backend && go build ./...

# 后端测试
cd backend && GOCACHE=/tmp/go-build go test ./...

# 前端类型检查 + 构建
cd frontend && npm run build                  # vue-tsc --noEmit + vite build

# 前端 ESLint
cd frontend && npm run lint
```

### Docker 一键启动

```bash
cp ai-service/.env.example ai-service/.env     # 填写 LLM API Key
docker compose up -d                           # 前端 :80, 后端 :8080, AI :8000
```

### API 冒烟测试

```bash
./scripts/test_core_apis.sh
```

## 代码架构

### 后端（Go / Gin / GORM）

路由注册入口：`backend/api/main.go`。三层 Router Group：

- **公开路由** `v1.Group("")` — 无需登录
- **需认证路由** `v1.Group("").Use(authMW)` — Bearer JWT
- **管理员路由** `v1.Group("").Use(authMW, middleware.AdminMiddleware)` — admin 角色

Handler 按领域分包在 `internal/handler/v1/<domain>/`，每个 Handler 以 `svcCtx *svc.ServiceContext` 作为闭包参数注入。

**ServiceContext**（`internal/svc/service_context.go`）：全局单例，持有 `*gorm.DB`、`config.Config` 和 `VideoDir`。启动时 AutoMigrate 所有模型。

**中间件**：
- `middleware.AuthMiddleware(secret)` — JWT 验证，解析出 `userId`、`username`、`role` 写入 Gin Context
- `middleware.AdminMiddleware` — 检查 role == "admin"

**API 响应格式**：所有接口通过 `response.Ok(c, data)` / `response.Fail(c, code, msg)` 返回 `{ code: 0, message: "ok", data: ... }`。

**WebSocket 弹幕**：
- HTTP 升级点：`GET /ws/live/:id?token=<JWT>`
- Hub 模式（`internal/logic/danmaku/hub.go`）：全局单例 Hub，维护 `map[roomID]map[*Client]bool`，通过 channel 做 Register/Unregister/Broadcast
- Client.ReadPump 收到消息后直接写入 MySQL 再广播到同房间所有客户端
- 连接鉴权：在 `ws_handler.go` 的 HTTP 升级前从 query string 取 token 验证 JWT

**数据库**：MySQL 8.0，模型定义在 `internal/model/mysql/models.go`。核心表：`users`、`videos`、`danmakus`、`comments`、`live_rooms`、`follows`、`likes`、`collects`、`comment_likes`。

### 前端（Vue 3 / TypeScript / Pinia / Element Plus / Tailwind CSS v4）

- **Vite 代理**：开发时 `/api` → `:8080`，`/media` → `:8080`，`/ws` → `ws://:8080`
- **路由**（`src/router/index.ts`）：根布局 `DefaultLayout.vue` 包裹主要页面；`beforeEach` 守卫检查 `meta.requiresAuth`，未登录重定向到 `/login`
- **Store**（Pinia）：`auth.ts`（token + userInfo，持久化到 localStorage）、`video.ts`、`comment.ts`
- **HTTP 请求**（`src/utils/request.ts`）：axios 实例，baseURL `/api/v1`，自动附加 Bearer token，统一拦截 `code !== 0` 弹出错误提示，401 时自动 logout
- **自动导入**：`unplugin-auto-import`（vue/vue-router/pinia API 无需手动 import），`unplugin-vue-components`（Element Plus 组件按需导入）
- **弹幕渲染**：`DanmakuLayer.vue` 是 canvas-less 的 CSS animation 实现，absolute 定位覆盖在视频播放器上，通过 text-shadow 描边实现 Bilibili 风格可读性，支持 scroll/top/bottom 三种模式和动态行分配
- **弹幕颜色**：预设 15 色调色板；当存储颜色为 `#FFFFFF` 时按 danmaku ID 确定性随机分配颜色；用户可通过预设色块选择自定义颜色

### AI 服务（Python / FastAPI / LangChain）

独立微服务，端口 8000。路由在 `app/api/`，核心逻辑在 `app/services/`。提供视频摘要（summary）和智能标签（tags）两个端点。依赖 DeepSeek API。

## 分支与提交规范

- 分支：`main`（稳定）← `dev`（集成）← `feature/*` / `fix/*`
- 提交格式：`type: summary`，常用 type：`feat` / `fix` / `docs` / `refactor` / `chore`
- 不在 `main` 上直接开发，功能分支从 `dev` 拉出，PR 合回 `dev`

## 配置文件

- 后端配置：`backend/etc/config.yaml`（数据库连接、JWT secret、视频存储路径）
- AI 服务配置：`ai-service/.env`（LLM API Key，不提交）
- TypeScript strict 模式开启，`noUnusedLocals` / `noUnusedParameters` 均为 true

## 弹幕数据模型（Danmaku）

字段：`videoId`、`userId`、`content`（≤200 字符）、`time`（视频秒数偏移）、`color`（hex，默认 `#FFFFFF`）、`fontSize`（small/medium/large）、`type`（scroll/top/bottom）、`blocked`（管理员屏蔽标记）。

REST 拉取时自动过滤 `blocked = true` 的弹幕；WebSocket 广播时包含 `fontSize` 和 `danmakuType` 字段（JSON key 为 `danmakuType`，前端接收后映射为 `type`）。
