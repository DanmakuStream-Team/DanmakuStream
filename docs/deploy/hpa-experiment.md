# HPA 自动扩缩容实验

## 1. 覆盖范围

| 工作负载 | CPU request | CPU limit | HPA 范围 | 触发目标 |
| --- | ---: | ---: | ---: | ---: |
| `user-service` | 100m | 500m | 1～5 | CPU 60% |
| `content-service` | 100m | 500m | 1～5 | CPU 60% |
| `engagement-service` | 100m | 500m | 1～5 | CPU 60% |
| `gateway` | 50m | 200m | 1～5 | CPU 60% |

当前微服务 k3s/CD 清单管理上表四个无状态工作负载和 MySQL，因此四个可安全水平复制的 Deployment 已全部覆盖。MySQL 使用单副本和 `ReadWriteOnce` PVC，不属于可直接水平复制的无状态工作负载；直接给它配置 HPA 会产生多主写入和数据损坏风险。前端与 SRS 目前不在这套微服务 k3s 清单中，后续迁入时应先补资源请求，再单独增加 HPA。Compose 也不提供 HPA；自动扩缩容只在 Kubernetes 环境生效。

## 2. 部署前检查

```bash
NS=danmakustream-microservices
sudo k3s kubectl wait --for=condition=Ready node --all --timeout=60s
sudo k3s kubectl top node
sudo k3s kubectl -n "$NS" top pod --containers
sudo k3s kubectl -n "$NS" get hpa
```

四个 HPA 的 `TARGETS` 必须显示数值而不是 `<unknown>`。如果指标不可用，检查 metrics-server：

```bash
sudo k3s kubectl -n kube-system get deploy,pod | grep metrics-server
sudo k3s kubectl -n kube-system logs deploy/metrics-server --tail=100
```

## 3. 施加 HTTP 压力

下面的临时 Job 从集群内部持续访问网关、用户、内容和互动路径。`parallelism` 可根据单机资源从 10 开始逐步增加，避免把节点本身压垮。

```bash
NS=danmakustream-microservices
cat <<'YAML' | sudo k3s kubectl -n "$NS" apply -f -
apiVersion: batch/v1
kind: Job
metadata:
  name: hpa-http-load
spec:
  parallelism: 20
  completions: 20
  template:
    spec:
      restartPolicy: Never
      containers:
        - name: load
          image: busybox:1.36
          command: ["sh", "-c"]
          args:
            - |
              end=$(( $(date +%s) + 240 ))
              while [ "$(date +%s)" -lt "$end" ]; do
                wget -qO- http://gateway/gateway/health >/dev/null &
                wget -qO- http://gateway/api/v1/livez >/dev/null &
                wget -qO- http://gateway/api/v1/videos >/dev/null &
                wget -qO- http://gateway/api/v1/live-schedules >/dev/null &
                wait
              done
YAML
```

在另一个终端每 10 秒保存一次 HPA、副本数和资源使用率：

```bash
NS=danmakustream-microservices
mkdir -p artifacts/hpa
while true; do
  date -u '+%Y-%m-%dT%H:%M:%SZ'
  sudo k3s kubectl -n "$NS" get hpa
  sudo k3s kubectl -n "$NS" get deploy user-service content-service engagement-service gateway
  sudo k3s kubectl -n "$NS" top pod --containers
  sleep 10
done | tee "artifacts/hpa/timeline-$(date -u +%Y%m%dT%H%M%SZ).log"
```

## 4. 观察缩容并清理

Job 结束后继续观察至少 3 分钟。HPA 的缩容稳定窗口为 60 秒，副本会逐步回到 1：

```bash
NS=danmakustream-microservices
sudo k3s kubectl -n "$NS" get hpa --watch
sudo k3s kubectl -n "$NS" delete job hpa-http-load --ignore-not-found
```

验收记录必须包含：固定并发和持续时间、Pod 数时间线、吞吐量、平均/P95、错误率、CPU/内存、测试机配置和原始输出。上述 Job 用于触发扩缩容；吞吐和延迟请使用项目选定的同一份 `hey`、k6 或其他压测脚本采集，不能用截图代替原始结果。

本地隔离集群的策略运行验证见 [HPA 本地运行验证（2026-09-02）](../testing/reports/hpa-local-runtime-2026-09-02.md)。该记录验证 HPA 控制链路；正式性能结论仍以真实业务镜像和实际 k3s 的测试为准。
