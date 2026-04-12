# 灵视 VisionLive

在线视频与直播平台 · 课程大作业

## 项目简介

灵视是一个面向内容创作者与普通用户的综合视频社区，支持视频点播、实时弹幕、轻量级直播、评论互动和后台管控。系统采用前后端分离的多语言微服务架构，并集成 AI 课代表（视频摘要）与智能标签功能作为加分项。

## 技术栈

| 层级 | 技术 |
|---|---|
| 前端 | Vue 3 · TypeScript · Arco Design · xgplayer · Pinia |
| 后端 | Go · Go-Zero · GORM · MySQL · Redis · MinIO |
| AI 服务 | Python · FastAPI · LangChain · DeepSeek |
| 部署 | Docker Compose · Nginx |

## 目录结构

```
DanmakuStream/
├── frontend/          # Vue 3 前端应用
├── backend/           # Go-Zero API 服务
├── ai-service/        # Python FastAPI AI 微服务
├── scripts/           # 数据库初始化脚本
├── docker-compose.yml
└── README.md
```

## 快速启动

### 前置要求

- Docker & Docker Compose
- Go 1.22+（本地开发）
- Node.js 20+（本地开发）
- Python 3.12+（本地开发）

### 一键启动（Docker）

```bash
# 1. 复制 AI 服务环境变量配置
cp ai-service/.env.example ai-service/.env
# 编辑 ai-service/.env，填入 LLM API Key

# 2. 启动所有服务
docker compose up -d

# 访问地址
# 前端：http://localhost
# 后端 API：http://localhost:8080
# AI 服务：http://localhost:8000
# MinIO 控制台：http://localhost:9001 (minioadmin / minioadmin)
```

### 本地开发启动

```bash
# 启动基础设施（MySQL / Redis / MinIO）
docker compose up mysql redis minio -d

# 后端
cd backend
go mod tidy
go run api/main.go -f etc/config.yaml

# 前端（新终端）
cd frontend
npm install
npm run dev
# 访问 http://localhost:5173

# AI 服务（新终端）
cd ai-service
cp .env.example .env   # 填写 API Key
pip install -r requirements.txt
uvicorn app.main:app --reload --port 8000
```

## 核心功能

### 用户端
- 注册 / 登录 / 个人主页 / 关注关系
- 视频首页浏览 · 关键词搜索 · 标签筛选
- 视频播放 · 拖拽进度 · 点赞 · 收藏
- 实时弹幕发送（视频 & 直播间）
- 评论区互动

### 创作者端
- 视频上传（分片）· 封面设置 · 元数据编辑
- 开启 / 关闭直播间

### 管理员端
- 视频内容审核（通过 / 拒绝）
- 违规弹幕屏蔽

### AI 加分项
- 视频课代表：自动生成三点式摘要
- 智能标签：根据标题 & 简介提取内容标签

## API 概览

| 方法 | 路径 | 说明 | 权限 |
|---|---|---|---|
| POST | `/api/v1/auth/register` | 注册 | 公开 |
| POST | `/api/v1/auth/login` | 登录，返回 JWT | 公开 |
| GET | `/api/v1/videos` | 视频列表 | 公开 |
| GET | `/api/v1/videos/:id` | 视频详情 | 公开 |
| POST | `/api/v1/videos/upload` | 上传视频 | 登录 |
| GET | `/api/v1/danmaku/:videoId` | 拉取弹幕 | 公开 |
| POST | `/api/v1/danmaku` | 发送弹幕 | 登录 |
| POST | `/api/v1/live` | 开启直播间 | 登录 |
| WS | `/ws/live/:id` | 直播间实时弹幕 | 登录 |
| PUT | `/api/v1/admin/videos/:id/status` | 审核视频 | 管理员 |
| PUT | `/api/v1/admin/danmaku/:id/block` | 屏蔽弹幕 | 管理员 |

## WebSocket 弹幕协议

连接地址：`ws://host/ws/live/:roomId?token=<JWT>`

**客户端发送：**
```json
{ "type": "danmaku", "content": "弹幕内容", "color": "#FFFFFF" }
```

**服务端推送（弹幕）：**
```json
{ "type": "danmaku", "payload": { "userId": 1, "content": "...", "color": "#FFFFFF", "time": 1718000000 } }
```

**服务端推送（在线人数）：**
```json
{ "type": "viewer_count", "payload": 42 }
```

## 数据库模型

| 表 | 说明 |
|---|---|
| users | 用户（含角色：user / creator / admin） |
| videos | 视频（含状态：pending / approved / rejected） |
| danmakus | 弹幕（视频时间点 + 样式） |
| comments | 评论（支持多级回复） |
| live_rooms | 直播间（状态机：idle / live / ended） |
| follows | 关注关系 |
| likes / collects | 点赞 / 收藏记录 |

## 开发规范

- 分支：`main`（稳定）· `dev`（集成）· `feature/*`（新功能）· `fix/*`（缺陷修复）
- 提交格式：`type: summary`，例如 `feat: 添加弹幕 WebSocket 网关`
- 接口变更必须同步更新设计文档，禁止只改代码不改文档
- 所有大模型生成的代码须经人工 Review，并记录在《大模型使用说明.md》

## 项目进度

| 阶段 | 截止时间 | 交付物 |
|---|---|---|
| 计划阶段 | 第7周（4/19） | 软件开发计划书 ✅ |
| 需求阶段 | 第8周（4/26） | 软件需求规格说明书 |
| 概要设计 | 第10周（5/10） | 软件概要设计说明书 |
| 详细设计 & 编码 | 第12周（5/24） | 软件详细设计说明书 + V0.5 |
| 系统集成 | 第14周（6/7） | Beta 版可运行系统 |
| 验收交付 | 第15周（6/14） | 测试报告 · 部署文档 · 用户手册 · 演示视频 |
