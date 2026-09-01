# 微服务独立 CI 与镜像规范

## 流水线

三个服务各有独立入口，公共步骤由可复用工作流维护：

| 服务 | Workflow | 必须通过的检查 |
| --- | --- | --- |
| user-service | `.github/workflows/user-service-ci.yml` | `user-service-ci / test → image (user-service)` |
| content-service | `.github/workflows/content-service-ci.yml` | `content-service-ci / test → image (content-service)` |
| engagement-service | `.github/workflows/engagement-service-ci.yml` | `engagement-service-ci / test → image (engagement-service)` |

三条流水线在所有指向 `dev`、`main` 的 PR 上运行，也在 `dev` push 后运行。每条流水线严格按以下顺序执行：

```text
go build → go test → Docker build →（仅 dev push）推送 GHCR
```

测试或编译失败时，后续镜像步骤不会执行。建议在 `dev`、`main` 的 Branch protection 中把上表三个检查都设为 Required status checks。

## 不可变镜像

镜像只使用完整的 `github.sha` 作为发布标签：

```text
ghcr.io/danmakustream-team/user-service:<40位commit-sha>
ghcr.io/danmakustream-team/content-service:<40位commit-sha>
ghcr.io/danmakustream-team/engagement-service:<40位commit-sha>
```

构建同时注入 `SERVICE_VERSION`、`COMMIT_SHA`、`BUILD_TIME`，服务的 `/api/v1/version` 必须返回对应值。不得用 `latest`、`dev` 等浮动标签部署。

## 平台级检查

`.github/workflows/ci.yml` 的 `platform-config` 负责验证 Compose 展开结果和 Nginx 网关语法。它不重复执行三个服务的测试；单体系统测试和三个微服务流水线作为独立检查共同构成合并门禁。

## 数据库账号边界

三个应用账号只能访问各自 Schema：

| 账号 | Schema |
| --- | --- |
| `user_app` | `user_db.*` |
| `content_app` | `content_db.*` |
| `engagement_app` | `engagement_db.*` |

授权只包含运行期 DML 和当前启动期 AutoMigrate 所需的有限 DDL，不授予全局权限、跨 Schema 权限或 `GRANT OPTION`。生产口令和 DSN 只能通过 Kubernetes Secret 注入。
