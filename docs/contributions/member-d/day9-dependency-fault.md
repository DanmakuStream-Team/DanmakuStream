# 9 月 3 日 engagement-service 依赖故障验收

## 1. 计划要求与当前状态

开发计划要求成员 D 验证：依赖服务故障时，`engagement-service` 在 2 秒内按设计超时或降级，且自身与其他 Pod 不发生连锁崩溃。

| 层级 | 验证内容 | 状态/证据 |
| --- | --- | --- |
| 客户端单元测试 | 下游 404、503、畸形响应、业务超时和传输超时的类型化错误 | 已自动通过，见 `internal/client/client_test.go` |
| API/MySQL 集成 | content 不可用、超时、畸形响应、媒体不可播放分别返回 503、504、502、409；故障后 `/livez`、`/health` 仍为 200 | 已自动通过，见 `tests/integration_test.go` |
| 覆盖率门槛 | `internal` 业务包低于 60% 时阻断 CI | 已通过：[Actions 运行 33581849649](https://github.com/DanmakuStream-Team/DanmakuStream/actions/runs/33581849649) |
| 真实三服务联调 | Compose、真实 user/content/engagement、网关和浏览器回归 | 已通过：[Actions 运行 33582138713](https://github.com/DanmakuStream-Team/DanmakuStream/actions/runs/33582138713) |
| K3s Pod 故障 | 缩容 content、检查 503/耗时/Pod UID/重启次数/其他工作负载、恢复后复测 | 脚本已准备；必须在授权的测试环境或维护窗口执行 |

## 2. K3s 验收脚本

脚本：`scripts/k3s-engagement-dependency-chaos.sh`

安全约束：

1. 只支持停止 `content-service`，不会操作数据库、PVC、Secret 或 `main` 分支；
2. 必须显式设置 `CONFIRM_CHAOS=engagement-dependency-outage`，避免误执行；
3. 必须使用专用测试账号 JWT 和可播放测试视频；脚本不会打印 JWT；
4. 记录原副本数，并通过退出陷阱在成功、失败或中断时恢复；
5. 两次点赞切换用于验证测试数据，第二次会恢复原始点赞状态；恢复依赖后再次执行两次；
6. 若 HTTP 不是 503、耗时超过 2 秒、engagement Pod UID 改变、重启次数增加或其他工作负载不 Ready，测试失败。

在 k3s 节点仓库根目录执行：

```bash
export CONFIRM_CHAOS=engagement-dependency-outage
export CHAOS_JWT='<专用测试账号 JWT>'
export CHAOS_VIDEO_ID='<可播放测试视频 ID>'
export CHAOS_GATEWAY_URL='http://127.0.0.1:30888'
bash scripts/k3s-engagement-dependency-chaos.sh
unset CHAOS_JWT CONFIRM_CHAOS
```

证据默认保存至 `artifacts/engagement-chaos/<UTC 时间>/`，包括：

- 故障前、故障中、恢复后的 Deployment、Pod、Endpoint 和事件；
- engagement-service 日志；
- 每次请求的 HTTP 状态与耗时；
- Pod UID、故障前后重启次数和最终 PASS 汇总。

## 3. 验收判定

通过标准：

- 依赖正常时互动请求为 200；
- content-service 无 Pod 时，同一请求在 2 秒内返回 503；
- engagement-service 的 `/api/v1/livez` 与 `/api/v1/health` 保持成功；
- engagement Pod UID 不变、容器重启次数不增加；MySQL、user、engagement、gateway 均保持 Ready；
- content-service 恢复到原副本数后，请求重新返回 200；
- 原始证据目录随测试报告归档，但不得提交 JWT、Secret 或真实账号信息。

当前没有在公网正式环境直接执行缩容操作；执行前需由 A 确认测试 Namespace/维护窗口和专用账号，避免影响在线演示。
