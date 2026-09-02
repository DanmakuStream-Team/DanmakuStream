# user-service 第 8 天回归报告（2026-09-02）

## 范围

- 负责人：成员 B
- 分支：`fix/p2-user-regression`
- 基线：2026-09-02 最新 `origin/dev`
- 服务：`services/user-service`
- 用例：UC01、UC07、UC08、UC11（计划原文将私信写作 UC12，仓库现行追溯编号为 UC11）

## 本次补强

1. 将生产路由注册提取为唯一的 `internal/server.Router`，生产启动与测试使用同一套路由，避免测试路由和真实服务漂移。
2. 增加带 `integration` build tag 的真实 MySQL API 回归：自动发现并触达 40 个公开 HTTP API 和 6 个内部 HTTP API。
3. 所有路由统一断言不返回 5xx，并断言入口 `X-Request-ID` 原样返回。
4. 对内部用户存在性、关注、双向屏蔽和有效会员状态增加真实数据成功断言。
5. 继续保留第 7 天 HTTP/WebSocket 视频分享、内部 Token、Request ID 与 content-service 客户端契约回归。

## 本次发现并修复的缺陷

- **现象**：同一进程依次创建两个隔离的 user-service 上下文时，WebSocket/私信 Hub 仍持有第一次上下文中已经回滚的数据库事务，后续视频分享报 `sql: transaction has already been committed or rolled back`。
- **原因**：Hub 使用进程级 `sync.Once`，无法识别测试、嵌入式服务或上下文重建。
- **修复**：以互斥锁保护 Hub 获取逻辑；当 `ServiceContext` 发生变化时创建与新上下文绑定的 Hub。生产单实例行为不变。
- **回归**：真实 MySQL 套件同时运行全路由测试和 HTTP/WebSocket 视频分享测试，确保上下文切换不再复用失效事务。

## 执行记录

| 检查 | 命令 | 本地结果 |
|---|---|---|
| 普通单元/API 回归 | `go test -p 1 -count=1 ./...` | 全部通过；本机 Windows 使用串行参数避免临时 `.test.exe` 文件锁干扰 |
| 集成套件编译 | `go test -run '^$' -tags=integration ./integration` | 通过 |
| 构建 | `go build ./...` | 通过 |
| 静态检查 | `go vet ./...` | 通过 |
| 真实 MySQL API 回归 | `go test -count=1 -tags=integration ./integration -v` | 通过：46/46 已注册公开/内部 HTTP 路由子测试通过，内部成功断言通过，HTTP/WebSocket 视频分享通过 |

## CI 阻断

`.github/workflows/user-service-ci.yml` 已启用独立 MySQL。真实集成命令位于 commit-SHA 镜像构建之前；普通测试或真实 MySQL API 回归任一失败，均不会构建或推送 user-service 镜像。

## 验收判定

本地隔离 MySQL 回归已经通过。最终合并仍须以本分支 GitHub Actions 中 user-service 独立流水线绿灯为准；CI 未完成前不提前标记 Done。
