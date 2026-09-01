# 自动化测试目录

测试组织参考 `origin/feature/uc13-tests`：每个用例分别维护单元、API 集成和 E2E 测试设计，使用统一追溯编号。

| 用例 | 测试设计 |
|---|---|
| UC01 | [uc01-user-service.md](./test-cases/uc01-user-service.md) |
| UC06 | [uc06-personal-library.md](./test-cases/uc06-personal-library.md) |
| UC07 | [uc07-relationships.md](./test-cases/uc07-relationships.md) |
| UC08 | [uc08-membership.md](./test-cases/uc08-membership.md) |
| UC11 | [uc11-messaging.md](./test-cases/uc11-messaging.md) |

执行约束：

1. 测试数据必须带用例前缀和时间戳，执行前后可清理。
2. API 测试失败以非零状态退出，并保存原始输出。
3. E2E 测试使用 Playwright，优先 API 换 token，再注入浏览器会话。
4. 未执行不能写“通过”；缺陷需要关联 Issue，修复后保留回归报告。

## 微服务 E2E（D06+）

一键在微服务栈（三服务 + 网关，`docker-compose.microservices.yml`）上运行全部 E2E：

```bash
cd frontend && npm run test:e2e:micro
```

内部行为：自动起栈/等网关健康 → `E2E_MICROSERVICES=1` 数据准备按 v1.1 归属拆到 user_db/content_db/engagement_db（root 直连）→ 媒体夹具经 `docker compose cp` 注入 content-service 媒体卷 → 走 compose 前端（:80）与网关（:8888）执行全部 spec。单体模式行为不变（`npm run test:e2e`）。
