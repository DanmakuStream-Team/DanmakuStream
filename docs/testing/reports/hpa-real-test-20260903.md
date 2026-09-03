# HPA 自动扩缩容实测报告（2026-09-03）

> 测试环境：公网 k3s 集群（deploy@47.76.86.151），命名空间 `danmakustream-microservices`
> 镜像版本：SHA `a14c294`（最新 CI 全绿提交）
> 测试工具：`scripts/k3s-hpa-chaos.sh`（k3s 内部 Job 施压 + HPA 轮询采样）
> 原始日志：`artifacts/hpa-real/{content,user,engagement}-service.log`

## 一、测试结论

四个无状态工作负载（三业务服务 + 网关）的 HPA 自动扩缩容**全部通过**：压力升高后 Pod 从 1 扩到 2，压力回落后自动缩回 1，全程零错误。

| 服务 | 副本变化 | 请求数 | 错误率 | 平均响应 | P95 响应 | 扩容触发 | 缩容完成 |
|---|---|---|---|---|---|---|---|
| content-service | 1→2→1 | 24,165 | 0.00% | 47.6ms | 108.2ms | CPU 94% 时 | ~102s |
| user-service | 1→2→1 | 24,476 | 0.00% | 47.6ms | 106.2ms | CPU >60% 时 | ~80s |
| engagement-service | 1→2→1 | 25,316 | 0.00% | 48.2ms | 110.4ms | CPU >60% 时 | ~85s |
| gateway | 1→2→1 | 23,357 | 0.00% | 49.5ms | 113.0ms | CPU 76% 时 | ~58s |

## 二、HPA 配置

```yaml
# deploy/k8s/microservices/autoscaling.yaml
spec:
  minReplicas: 1
  maxReplicas: 5
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 60   # CPU 达到 request 的 60% 触发扩容
  behavior:
    scaleUp:
      stabilizationWindowSeconds: 0    # 立即扩容
      selectPolicy: Max
      policies:
        - type: Percent, value: 100, periodSeconds: 30
        - type: Pods, value: 2, periodSeconds: 30
    scaleDown:
      stabilizationWindowSeconds: 60   # 缩容前等待 60 秒确认
      selectPolicy: Max
      policies:
        - type: Percent, value: 25, periodSeconds: 60  # 每分钟最多缩 25%
```

## 三、详细过程

### content-service（120 秒 · 40 并发）

```
时间线（节选自原始日志）：
03:26:20  current=1 desired=1 cpu=3%    ← 施压开始
03:26:46  current=1 desired=1 cpu=34%   ← CPU 开始攀升
03:27:11  current=1 desired=2 cpu=94%   ← HPA 决策扩容
03:27:23  current=2 desired=2 cpu=94%   ← 新 Pod 就绪
03:27:47  current=2 desired=2 cpu=54%   ← 负载分流，CPU 下降
03:28:50  （施压结束）
03:29:56  current=2 desired=1 cpu=1%    ← HPA 决策缩容（60s 稳定窗后）
03:30:07  current=1 desired=1 cpu=1%    ← 缩容完成
```

### user-service（120 秒 · 40 并发）

- 扩容触发：CPU 超过 60% 目标值
- 缩容完成：施压结束后约 69 秒期望回落，约 80 秒实际恢复到 1 副本
- 全程 24,476 请求，0 错误

### engagement-service（120 秒 · 40 并发）

- 扩容触发：CPU 超过 60% 目标值
- 缩容完成：施压结束后约 85 秒恢复到 1 副本
- 全程 25,316 请求，0 错误

### Gateway（120 秒 · 40 并发）

- 扩容触发：CPU 达到 76%（超过 60% 目标值）
- 缩容完成：施压结束后约 46 秒决策缩容（60s 稳定窗内），约 58 秒恢复到 1 副本
- 全程 23,357 请求，0 错误
- 时间线关键点：
  - 06:29:43 current=1 cpu=2%（施压开始）
  - 06:30:21 current=1→2 cpu=76%（HPA 决策扩容）
  - 06:30:33 current=2 cpu=74%（新 Pod 就绪，分流后 CPU 降至 36%）
  - 06:31:59 current=2 cpu=22%（施压结束）
  - 06:32:46 current=2→1 cpu=2%（HPA 决策缩容）
  - 06:32:57 current=1 cpu=2%（缩容完成）

## 四、验证的验收条件

- [x] **高压时 Pod 数量增加**：三个服务均从 1 副本扩到 2
- [x] **压力下降后数量回落**：三个服务均自动缩回 1 副本
- [x] **过程可复现**：脚本 `k3s-hpa-chaos.sh` 可重复执行，原始日志已归档
- [x] **HPA 目标值/副本数/容器资源使用率**：每 12 秒采样一次，完整时间线记录
- [x] **请求质量**：扩容期间零错误，P95 < 111ms

## 五、测试方法

```bash
# 在本机执行（通过 SSH 在 k3s 服务器上运行）
ssh -i ~/.ssh/danmakustream_cd -o IdentitiesOnly=yes deploy@47.76.86.151 \
  bash -s -- hpa <service-name> 120 40 < scripts/k3s-hpa-chaos.sh
```

脚本内部：
1. 预检：确认 HPA 存在、CPU 指标可用、服务健康
2. 创建 k3s Job 施压（40 并发 × 120 秒，targeting 服务 ClusterIP）
3. 每 12 秒轮询 `kubectl get hpa` 记录 current/desired/CPU/target
4. 施压结束后持续观察缩容过程
5. 输出 `LOAD_RESULT`（requests/errors/average/p95）和 `HPA_RESULT`（min/maxObserved/final）

## 六、限制与说明

2. **单节点限制**：k3s 单节点 maxReplicas=5，但实际只扩到 2 就满足了负载（CPU 已降至 54%），更高的副本数需要多节点集群才能验证。
3. **MySQL 无 HPA**：MySQL 使用 ReadWriteOnce PVC，直接扩副本有多主写入风险——数据库扩展不在本次无状态 Pod 扩缩容范围内。
4. **测试环境与生产差异**：40 并发在单节点上已能触发扩容，但生产负载分布可能不同。

## 七、证据索引

| 文件 | 内容 |
|---|---|
| `artifacts/hpa-real/content-service.log` | 51 行，完整扩缩容时间线 + 请求统计 |
| `artifacts/hpa-real/user-service.log` | 47 行，同上 |
| `artifacts/hpa-real/engagement-service.log` | 49 行，同上 |
| `artifacts/hpa-real/gateway.log` | 42 行，完整扩缩容时间线 + 请求统计 |
