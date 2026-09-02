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

一键启动微服务栈（三服务 + 网关，`docker-compose.microservices.yml`）并运行跨网关烟测门禁：

```bash
cd frontend && npm run test:e2e:micro
```

默认门禁验证前端、网关、三个服务、独立 Schema、JWT 和内部接口隔离，不回退到单体服务。单体模式行为不变（`npm run test:e2e`）。

需要审计 24 条单体业务用例迁移进度时，可显式运行全量模式：

```bash
cd frontend && MICRO_E2E_FULL_SUITE=1 npm run test:e2e:micro
```

全量模式会启用 `E2E_MICROSERVICES=1`，按 v1.1 归属准备三库数据并把媒体夹具复制到 content-service。该命令用于暴露尚未完成的网关路由、服务接口和响应契约，不作为微服务环境 PR 的绿色门禁；用例逐项兼容后再迁入 `frontend/e2e-microservices/`。

真实失败、定位和修复时间线见 [微服务 E2E 失败排查记录](reports/microservice-e2e-failure-investigation-2026-09-01.md)。自动部署后的三服务日志、探针和版本响应由 `microservice-cd-<sha>` Actions artifact 保存。

成员 B 第 8 天 user-service 公开/内部 API 回归范围、命令与真实执行状态见 [user-service 回归报告（2026-09-02）](reports/user-service-regression-2026-09-02.md)。

成员 C 第 8 天 content-service 公开/内部 API 路由清单、真实 MySQL 回归范围与执行状态见 [content-service 回归报告（2026-09-02）](reports/content-service-regression-2026-09-02.md)。
