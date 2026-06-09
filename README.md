# DanmakuStream

在线视频、弹幕与直播平台 · 课程大作业

## 项目简介

DanmakuStream 是一个前后端分离的视频社区系统，支持视频点播、实时弹幕、评论互动、收藏点赞、关注订阅、轻量级直播和后台审核管控。项目当前以 Docker Compose 为主要部署方式，前端容器内置 Nginx 用于静态资源、API、媒体文件和 WebSocket 反向代理。

## 技术栈

| 层级 | 技术 |
|---|---|
| 前端 | Vue 3 · TypeScript · Element Plus · Pinia · xgplayer |
| 后端 | Go · Gin · GORM · MySQL |
| 直播 | SRS · RTMP 推流 · HLS 播放 |
| 视频处理 | FFmpeg · HLS 分片 · 首帧封面 |
| AI 服务 | Python · FastAPI · LangChain · DeepSeek |
| 部署 | Docker Compose · 前端容器内置 Nginx |

## 目录结构

```text
DanmakuStream/
├── frontend/          # Vue 3 前端应用和 Nginx 配置
├── backend/           # Gin API 服务
├── ai-service/        # Python FastAPI AI 微服务
├── scripts/           # 数据库初始化和测试脚本
├── docker-compose.yml
└── README.md
```

## 快速启动

### 前置要求

- Docker 和 Docker Compose
- Go 1.22+，仅本地开发需要
- Node.js 20+，仅本地开发需要
- Python 3.12+，仅 AI 服务本地开发需要
- FFmpeg，本地运行后端转码时需要；Docker 后端镜像内会安装

### Docker Compose 启动

```bash
cp ai-service/.env.example ai-service/.env
# 按需编辑 ai-service/.env，填写 LLM API Key

docker compose up -d --build
```

默认访问地址：

| 服务 | 地址 |
|---|---|
| 前端 | `http://localhost` |
| 后端 API | 前端 Nginx 代理到 `/api/v1/*` |
| SRS RTMP | `rtmp://localhost:1935/live/<streamKey>` |
| SRS HLS | `http://localhost:8081/live/<streamKey>.m3u8` |
| AI 服务 | `http://localhost:8000` |

公网部署时需要把 [backend/etc/config.yaml](backend/etc/config.yaml) 里的直播地址改成公网可访问地址：

```yaml
Live:
  RTMPHost: "your-domain-or-ip:1935"
  HTTPHost: "your-domain-or-ip:8081"
```

如果这里仍然是 `localhost`，前端拿到的推流或播放地址会指向用户自己的电脑，直播会无法开启或无法观看。

### 本地开发启动

```bash
docker compose up mysql srs -d

cd backend
go mod tidy
go run api/main.go -f etc/config.yaml

cd ../frontend
npm install
npm run dev
```

本地前端默认访问 `http://localhost:5173`。

## 核心功能

### 用户端

- 注册、登录、退出登录
- 首页视频浏览，支持分页自适应填满最后一行
- 搜索视频和创作者，搜索栏支持下拉历史记录
- 视频播放、弹幕、评论、点赞、收藏
- 个人主页、关注关系、订阅页
- 个人主页支持头像裁剪上传、简介编辑、右键昵称改名
- 赞过的视频使用大拇指图标，收藏内容使用星星图标

### 创作者端

- 上传视频并异步转码为 HLS
- 上传封面；未上传封面时后端尝试截取首帧
- 上传过程中可终止上传；取消后不会生成待审核残留记录
- 查看自己视频的审核状态
- 开始直播和结束自己的直播

### 直播

- 直播页展示所有正在播出的直播间
- 直播间显示封面、主播、观看人数
- 用户通过页面创建直播间，系统返回 RTMP 推流地址和 HLS 播放地址
- 观看人数从本次开播开始统计，下播后清零，下次开播从 0 开始

### 后台与权限

系统使用 `role` 字段区分权限：

| 角色 | 权限 |
|---|---|
| `user` 普通用户 | 普通观看、互动、投稿、直播和个人资料功能 |
| `moderator` 内容审核员 | 只能进入审核后台，处理视频审核和弹幕屏蔽 |
| `admin` 超级管理员 | 拥有全部后台能力，包括审核、用户权限、服务器监控、运营工具 |

管理员和审核员登录后，顶部导航只显示“审核”，不显示首页、视频、直播、投稿等普通入口。

后台页面：

| 页面 | 说明 | 权限 |
|---|---|---|
| `/admin` | 后台入口 | `moderator` / `admin` |
| `/admin/videos` | 视频审核 | `moderator` / `admin` |
| `/admin/danmaku` | 弹幕治理 | `moderator` / `admin` |
| `/admin/users` | 用户与权限、角色分配 | `admin` |
| `/admin/infrastructure` | 服务器监控 | `admin` |
| `/admin/operations` | 首页横幅和系统公告 | `admin` |

服务器监控包含：

- 存储容量和容量预警
- 今日 / 本月下行流量
- CPU 使用率
- 当前在线 = 直播观看人数 + 视频文件连接数
- 当前直播间数和直播最高并发

## 管理员用户

初始化 SQL 只保留建库逻辑。部署后可以先注册一个普通账号，再通过数据库把该账号提升为超级管理员：

```bash
docker compose exec mysql mysql -uroot -ppassword danmakustream \
  -e "UPDATE users SET role='admin' WHERE nickname='你的昵称';"
```

之后用该账号登录，即可进入 `/admin/users` 给其他用户分配 `user`、`moderator` 或 `admin`。

## API 概览

| 方法 | 路径 | 说明 | 权限 |
|---|---|---|---|
| POST | `/api/v1/auth/register` | 注册 | 公开 |
| POST | `/api/v1/auth/login` | 登录，返回 JWT | 公开 |
| GET | `/api/v1/videos` | 视频列表 | 公开 |
| GET | `/api/v1/videos/:id` | 视频详情 | 公开 |
| POST | `/api/v1/videos/upload` | 上传视频 | 登录 |
| POST | `/api/v1/videos/:id/like` | 点赞 / 取消点赞 | 登录 |
| POST | `/api/v1/videos/:id/collect` | 收藏 / 取消收藏 | 登录 |
| GET | `/api/v1/users/following` | 当前用户关注列表 | 登录 |
| POST | `/api/v1/users/:id/follow` | 关注 / 取消关注 | 登录 |
| GET | `/api/v1/danmaku/:videoId` | 拉取弹幕 | 公开 |
| POST | `/api/v1/danmaku` | 发送弹幕 | 登录 |
| POST | `/api/v1/live` | 开启直播间 | 登录 |
| PUT | `/api/v1/live/:id/end` | 结束直播 | 房主 / 管理员 |
| WS | `/ws/live/:id` | 直播间实时弹幕和观看人数 | 登录 |
| GET | `/api/v1/admin/videos` | 审核视频列表 | 审核员 / 管理员 |
| PUT | `/api/v1/admin/videos/:id/status` | 更新视频审核状态 | 审核员 / 管理员 |
| GET | `/api/v1/admin/danmaku` | 弹幕治理列表 | 审核员 / 管理员 |
| PUT | `/api/v1/admin/danmaku/:id/block` | 屏蔽弹幕 | 审核员 / 管理员 |
| GET | `/api/v1/admin/users` | 用户与权限 | 管理员 |
| GET | `/api/v1/admin/infrastructure` | 服务器监控 | 管理员 |
| GET | `/api/v1/admin/banners` | 首页横幅 | 管理员 |
| GET | `/api/v1/admin/announcements` | 系统公告 | 管理员 |

## WebSocket 弹幕协议

连接地址：

```text
ws://host/ws/live/:roomId?token=<JWT>
```

客户端发送：

```json
{ "type": "danmaku", "content": "弹幕内容", "color": "#FFFFFF", "time": 0 }
```

服务端推送弹幕：

```json
{ "type": "danmaku", "payload": { "userId": 1, "content": "...", "color": "#FFFFFF", "time": 0 } }
```

服务端推送观看人数：

```json
{ "type": "viewer_count", "payload": 42 }
```

## 数据库模型

| 表 | 说明 |
|---|---|
| users | 用户，含角色：`user` / `creator` / `moderator` / `admin` |
| videos | 视频，含审核状态：`pending` / `approved` / `rejected` |
| danmakus | 弹幕，含时间点、颜色、字号、位置和屏蔽状态 |
| comments | 评论，支持回复和点赞 |
| live_rooms | 直播间，状态：`idle` / `live` / `ended` |
| live_schedules / live_reservations | 直播预约和预约关系 |
| notifications | 站内通知 |
| site_banners / site_announcements | 首页横幅和系统公告 |
| traffic_stats | 应用层流量统计 |
| follows | 关注关系 |
| likes / collects / comment_likes | 视频点赞、视频收藏、评论点赞 |

## 常用命令

```bash
# 构建并启动
docker compose up -d --build

# 查看服务状态
docker compose ps

# 查看后端日志
docker compose logs -f backend

# 重新构建前端和后端
docker compose build frontend backend
docker compose up -d frontend backend

# 后端测试
cd backend
GOCACHE=/tmp/go-build go test ./...

# 前端构建
cd frontend
npm run build
```

## 开发规范

- 分支：`main` 稳定分支，功能开发可使用 `feature/*`，缺陷修复可使用 `fix/*`
- 提交格式：`type: summary`，例如 `feat: add admin dashboard`
- 接口、权限或部署方式变化后，需要同步更新 README 或相关文档
- 重要后端变更至少运行 `GOCACHE=/tmp/go-build go test ./...`
- 重要前端变更至少运行 `npm run build`

## 项目进度

| 阶段 | 截止时间 | 交付物 |
|---|---|---|
| 计划阶段 | 第7周（4/19） | 软件开发计划书 |
| 需求阶段 | 第8周（4/26） | 软件需求规格说明书 |
| 概要设计 | 第10周（5/10） | 软件概要设计说明书 |
| 详细设计 & 编码 | 第12周（5/24） | 软件详细设计说明书 + V0.5 |
| 系统集成 | 第14周（6/7） | Beta 版可运行系统 |
| 验收交付 | 第15周（6/14） | 测试报告 · 部署文档 · 用户手册 · 演示视频 |
