# HPA 扩缩容与依赖故障演练

> 对应清单：`deploy/k8s/microservices/hpa.yaml`；脚本：`scripts/k3s-hpa-chaos.sh`；手动流水线：`.github/workflows/microservice-resilience.yml`。

## 1. 验收目标

| 项目 | 通过条件 | 自动保存的证据 |
|---|---|---|
| HPA 扩容 | 固定压力下目标服务从 1 个 Pod 增加到至少 2 个 | 副本/CPU 时间线、负载请求数、平均、P95、错误率、Pod 与 Event |
| HPA 缩容 | 压力停止后 480 秒内回到 `minReplicas=1` | `hpa-timeline.csv`、可直接截图的 `hpa-timeline.svg` |
| 依赖故障 | 临时停止 content-service 后，engagement 在 2.5 秒内明确返回 503/504 | 请求响应、耗时和 `requestId` |
| 故障隔离 | user/engagement 健康接口继续成功，容器重启总数不增加 | 健康检查、restart 快照和三服务日志 |
| 恢复 | 演练退出时恢复 content-service 和规范 HPA | rollout 输出、最终 HPA/Pod 快照 |

三个 HPA 均使用 `autoscaling/v2`，CPU 目标 60%，范围 1～4 个副本。扩容无稳定窗口；缩容有 60 秒稳定窗口，避免压力短暂下降造成抖动。资源利用率依赖各 Deployment 已设置的 CPU requests。

## 2. 前置条件

```bash
NS=danmakustream-microservices
sudo k3s kubectl -n "$NS" get deploy,pod,hpa
sudo k3s kubectl -n "$NS" top pods
```

`kubectl top pods` 必须返回 CPU/内存；若显示 Metrics API unavailable，先修复 k3s 自带的 Metrics Server。不要通过降低 HPA 阈值伪造扩容结果。

集群还需要能拉取固定镜像 `curlimages/curl:8.10.1`，它只用于集群内有限时长压测和故障探针。

## 3. Actions 一键演练

进入 Actions → `microservice-resilience` → Run workflow：

### HPA

- `exercise=hpa`
- `service=content-service`（也可选择另外两个服务）
- `duration_seconds=180`
- `concurrency=80`
- `confirm_chaos=false`

如果 180 秒压力不足以超过 60% CPU，先查看 artifact 中的 CPU 时间线，再逐步提高 concurrency；不得直接删除失败证据。

### 故障

- `exercise=chaos`
- 勾选 `confirm_chaos`

演练会暂时删除 content-service HPA 并把 Deployment 缩到 0，随后使用合成的 10 分钟 JWT 调用 engagement 的发送弹幕接口，验证跨服务依赖失败映射。JWT 密钥只在服务器内存中使用，不写日志和 artifact。脚本通过 EXIT trap 恢复副本和 HPA；工作流的 `always()` 步骤会再次应用仓库中的规范 HPA 清单。

## 4. Artifact 与截图

每次运行生成 `microservice-resilience-<exercise>-<sha>`：

- `exercise.log`：完整演练过程和负载汇总；
- `hpa-timeline.csv`：每 10 秒一次的原始副本/CPU数据；
- `hpa-timeline.svg`：扩容和缩容曲线，可直接用于报告截图；
- `cluster-evidence.txt`：HPA、Pod、资源指标、Event 和四个工作负载日志。

报告必须记录 Action URL、commit SHA、并发、时长、初始/峰值/最终副本、请求数、P95、错误率和结论。未实际运行时只能写“脚本/清单已完成”，不能写“扩缩容通过”。

## 5. 边界与风险

- MySQL 当前仍为单副本，HPA 只验证业务服务横向扩展，不代表数据库高可用。
- user/content 共享 `ReadWriteOnce` 媒体卷；单节点 k3s 可多 Pod 挂载，同一卷跨节点调度可能失败。多节点生产部署需要 RWX 对象存储/NFS，或把媒体改为对象存储。
- WebSocket 长连接不会因新 Pod 出现而自动迁移；缩容时现有连接可能重连。多副本 WS 广播需要 Redis Pub/Sub，属于增强项。
- 故障演练只允许手动触发并使用 production Environment 审批，不能放进普通 push/PR 流水线。
