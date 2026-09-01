# 微服务 E2E 环境

该环境只启动 `user-service`、`content-service`、`engagement-service`、API Gateway、前端、MySQL 和 SRS，不会启动或回退到 `backend/` 单体服务。Playwright 从浏览器访问前端，再由前端 Nginx 和 API Gateway 进入三个业务服务。

## 一键运行

前置条件：Docker Engine、Docker Compose v2、Node.js 20，以及已安装的前端依赖和 Playwright Chromium。

```bash
cd frontend
npm ci
npx playwright install chromium
cd ..
bash scripts/run-microservices-e2e.sh
```

默认端口为前端 `18080`、网关 `18888`、RTMP `19350`、HLS `18081`，避免占用单体环境的端口。脚本使用独立 Compose project 和独立数据卷；无论成功或失败，都会将容器状态、日志和 Playwright 报告写入 `artifacts/microservices-e2e/`，随后删除容器与测试数据卷。

如需失败后保留现场：

```bash
MICRO_E2E_KEEP_STACK=1 bash scripts/run-microservices-e2e.sh
```

可用 `MICRO_FRONTEND_PORT`、`MICRO_GATEWAY_PORT`、`MICRO_RTMP_PORT`、`MICRO_HLS_PORT` 覆盖宿主机端口。保留现场后可执行：

```bash
docker compose --project-name danmakustream-e2e -f docker-compose.microservices.yml down --volumes
```

## 当前验收范围

- 网关健康检查和三服务目录可访问；
- `/internal/*` 不从公网网关暴露；
- 浏览器通过前端完成注册并保存登录态；
- 同一 JWT 能分别访问 user、content、engagement 三个服务；
- 三个 Schema 由各自最小权限账号初始化，测试结束后整体销毁；
- PR 和 `dev` push 会运行独立的 `microservices-e2e` 工作流并归档失败证据。

这是微服务运行环境和跨网关冒烟基线。现有 `frontend/e2e/` 全量业务用例仍保留给单体回归；待各服务内部 API 与测试数据工厂齐全后，再逐项迁移到 `frontend/e2e-microservices/`，不得用单体数据库脚本伪造微服务通过结果。
