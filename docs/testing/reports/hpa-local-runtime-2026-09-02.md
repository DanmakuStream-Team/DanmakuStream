# HPA 本地运行验证（2026-09-02）

## 结论

当前 `autoscaling.yaml` 的四个 HPA 均能正常读取 CPU 指标、触发扩容、遵守 `maxReplicas: 5`，并在停止压力后缩回 `minReplicas: 1`。本次结果判定为通过。

## 环境与范围

- Docker Desktop 4.66.1，8 CPU，约 11.4 GiB 可用内存。
- 隔离的单节点 k3s `v1.31.6+k3s1`，metrics-server 正常。
- 直接使用仓库的 `deploy/k8s/microservices/autoscaling.yaml`。
- 测试 Deployment 使用真实工作负载名称和相同 CPU request，但容器替换为本地已有的 `nginx:alpine`，避免连接数据库或污染项目数据。
- 本测试验证 Kubernetes/HPA 控制链路和策略，不代表真实业务接口的吞吐量、P95 或错误率。

## 结果

| 工作负载 | 初始副本 | 压测峰值副本 | 配置上限 | 停止压力后 |
| --- | ---: | ---: | ---: | ---: |
| `user-service` | 1 | 2 | 5 | 1 |
| `content-service` | 1 | 4 | 5 | 1 |
| `engagement-service` | 1 | 3 | 5 | 1 |
| `gateway` | 1 | 5 | 5 | 1 |

关键观测：

- 指标初次采样约需 60～90 秒，随后四个 HPA 的 `TARGETS` 均由 `<unknown>` 变为数值。
- 4 个压力 Pod 时，Gateway 先从 1 扩到 3，user/content 扩到 2。
- 8 个压力 Pod 时，engagement 也从 1 扩到 2；最终峰值见上表。
- Gateway 到达 5 后没有超过配置上限。
- 停止压力后 CPU 降到 0%，四个工作负载经过稳定窗口和指标采样延迟后全部回到 1。
- Kubernetes Event 明确记录 `SuccessfulRescale`，缩容原因为 `All metrics below target`。

## 本次发现的环境问题

1. 普通沙箱进程无法访问 Windows Docker 命名管道，会显示 `permission denied`；这不等于 Docker 引擎崩溃。应使用宿主权限再次只读确认。
2. Docker 拉取命令的前台输出提前结束，但镜像拉取和容器创建实际在后台完成。再次执行前必须先用 `docker ps -a` 和 `docker image inspect` 核对，避免重复启动。
3. k3s 容器生成的 kubeconfig 指向容器内部 `127.0.0.1:6443`，宿主实际映射为 `127.0.0.1:16443`；测试时使用 `kubectl --server` 覆盖，未修改全局 kubeconfig。
4. HPA 启动早期的 `<unknown>` 是 metrics-server 尚未完成 Pod 首轮采样，不应立即判定策略失败。
5. 单节点环境的 Pod 扩容只能改善并发调度和实例隔离，不能增加宿主机的物理 CPU/内存总量。

## 后续正式验收

部署到实际 k3s 后，应针对真实业务镜像重复压力测试，并保存固定并发、吞吐量、平均/P95、错误率、Pod 时间线和节点资源使用率。正式验收步骤见 `docs/deploy/hpa-experiment.md`。
