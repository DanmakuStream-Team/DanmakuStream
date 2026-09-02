# Day 9 性能压测对比报告（单体 vs 微服务）

> 执行命令：`bash scripts/run-benchmark-comparison.sh`
> 产物目录：`artifacts/benchmarks/`（compose 日志、k6 json+txt、容器资源采样 CSV、comparison.csv/.md）
> 调参覆盖：
> - 并发 VUS：BM01=50 / BM02=30 / BM03=40
> - 单轮持续时间：60s
> - 架构各跑 3 轮取平均（默认，可通过 `BENCHMARK_ROUNDS` 调整）

---

## 0. 环境信息

| 项 | 值 | 实际运行填入 |
|---|---|---|
| 执行日期 | — | 2026-__-__ |
| 宿主机 OS / CPU / 内存 | Docker Desktop on Windows/macOS/Linux |  |
| Docker 版本 / Compose 版本 | Docker Engine 25+ / Compose v2 |  |
| MySQL 版本 / 配置 | mysql:8.0（共享镜像；单体单库 / 微服务 3 库×3 账号） |  |
| Go 版本（单体 backend） | 由 `backend/Dockerfile` 定义 |  |
| Go 版本（3 个微服务） | 各自 service Dockerfile |  |
| Nginx 网关版本 | nginx:1.27-alpine |  |
| k6 版本 | grafana/k6:0.52.0（容器运行，宿主无需安装） |  |
| 压测方式 | host network 下的 k6 容器 → 宿主机端口 → Compose 栈 |  |
| 单体网关端口 | `28888`（BENCHMARK_MONO_PORT 可调） |  |
| 微服务网关端口 | `38888`（BENCHMARK_MICRO_PORT 可调） |  |
| 数据预置 | 单体：seed-test-data + 压测脚本补 BM-公开视频；微服务：seed-microservices-e2e-data.sh |  |

---

## 1. 核心指标对比（单体 vs 微服务，各 N=3 轮平均）

> 运行脚本 `run-benchmark-comparison.sh` 会自动在 `artifacts/benchmarks/comparison.md` 生成此表。
> 此处为**待填实**的人工模板，填入一次原始测量结果即可提交答辩。

### 1.1 BM01 公开视频搜索 `GET /api/v1/videos?keyword=`

| 轮次 | 架构 | 吞吐 (req/s) | 平均 (ms) | P95 (ms) | 错误率 (%) | 总请求数 | 备注 |
|---|---|---:|---:|---:|---:|---:|---|
| 1 | 单体 |  |  |  |  |  | |
| 2 | 单体 |  |  |  |  |  | |
| 3 | 单体 |  |  |  |  |  | |
| 平均 | 单体 |  |  |  |  |  | **3轮平均** |
| 1 | 微服务 |  |  |  |  |  | |
| 2 | 微服务 |  |  |  |  |  | |
| 3 | 微服务 |  |  |  |  |  | |
| 平均 | 微服务 |  |  |  |  |  | **3轮平均** |
| 差异 | 微服务−单体（%） |  |  |  | — | — | **正=更好/更高** |

### 1.2 BM02 用户登录 `POST /api/v1/auth/login`（含每轮先 register 再 login）

| 轮次 | 架构 | 吞吐 (req/s) | 平均 (ms) | P95 (ms) | 错误率 (%) | 总请求数 | 备注 |
|---|---|---:|---:|---:|---:|---:|---|
| 1 | 单体 |  |  |  |  |  | bcrypt/JWT |
| 2 | 单体 |  |  |  |  |  | |
| 3 | 单体 |  |  |  |  |  | |
| 平均 | 单体 |  |  |  |  |  | **3轮平均** |
| 1 | 微服务 |  |  |  |  |  | |
| 2 | 微服务 |  |  |  |  |  | |
| 3 | 微服务 |  |  |  |  |  | |
| 平均 | 微服务 |  |  |  |  |  | **3轮平均** |
| 差异 | 微服务−单体（%） |  |  |  | — | — | **正=更好/更高** |

### 1.3 BM03 个人资料库历史 `GET /api/v1/users/me/history`（setup 预埋 300 条历史）

| 轮次 | 架构 | 吞吐 (req/s) | 平均 (ms) | P95 (ms) | 错误率 (%) | 总请求数 | 备注 |
|---|---|---:|---:|---:|---:|---:|---|
| 1 | 单体 |  |  |  |  |  | 已鉴权+分页+JOIN |
| 2 | 单体 |  |  |  |  |  | |
| 3 | 单体 |  |  |  |  |  | |
| 平均 | 单体 |  |  |  |  |  | **3轮平均** |
| 1 | 微服务 |  |  |  |  |  | |
| 2 | 微服务 |  |  |  |  |  | |
| 3 | 微服务 |  |  |  |  |  | |
| 平均 | 微服务 |  |  |  |  |  | **3轮平均** |
| 差异 | 微服务−单体（%） |  |  |  | — | — | **正=更好/更高** |

---

## 2. 资源占用对比（压测 BM03 期间 20s 采样）

| 架构 | 容器名 | CPU% 平均 | 内存 RSS 平均 | 内存 Limit | 备注 |
|---|---|---:|---:|---:|---|
| 单体 | mysql |  |  |  | 单库 danmakustream |
| 单体 | srs |  |  |  | （空闲，仅供依赖） |
| 单体 | backend（Go 单体） |  |  |  | **业务承载核心** |
| 单体 | nginx-gateway |  |  |  | 网关轻量 by-path 转发到 backend |
| 单体 | frontend |  |  |  | （空闲静态资源） |
| 微服务 | mysql |  |  |  | 3 库×3 账号 |
| 微服务 | srs |  |  |  | （空闲） |
| 微服务 | user-service |  |  |  | 账号/关注/订阅 |
| 微服务 | content-service |  |  |  | 视频/审核/分析 |
| 微服务 | engagement-service |  |  |  | 弹幕/评论/直播/资料库 |
| 微服务 | gateway（nginx） |  |  |  | 网关 by-path 分发到三服务 |
| 微服务 | frontend |  |  |  | （空闲静态资源） |

> 自动采集 CSV：`artifacts/benchmarks/monolith-stats.csv` 与 `microservices-stats.csv`
> 若需更精准的 CPU 曲线：接入 cAdvisor + Prometheus 并在 Grafana 截图粘贴到 §6。

---

## 3. 结论摘要（答辩稿要点）

1. **吞吐差异**：BM01（公开搜索，读多）单体 vs 微服务 = ___:___；BM03（资料库，单服务）___ 更高。
2. **延迟差异**：BM02（登录，CPU 密集密码哈希）__架构 Avg ___ms，___ 更高 P95；原因分析：_____。
3. **资源占用差异**：单体单 backend 进程 vs 微服务 3 个 Go 进程总内存 = ___MB vs ___MB（多出 ___MB，约 +___%）。
4. **可扩展性启示**：微服务可单独扩容 user-service（登录热点）、engagement-service（资料库/直播热点），在 Nginx upstream 调整权重即可线性扩展；单体需整体扩容副本+数据库主从。
5. **稳定性**：所有场景错误率均 <___%（低于 2% 阈值），符合验收门槛。

---

## 4. 压测阈值（验收门槛）

| 指标 | 阈值 | 实际是否达成 |
|---|---|---|
| 全部 3 接口 `http_req_failed`（k6 built-in） | < 2% | __ |
| BM01 p95 | < 800ms | __ |
| BM02 p95 | < 1500ms | __ |
| BM03 p95 | < 1200ms | __ |
| 3 接口合计错误率（3 轮平均） | < 5% | __ |

---

## 5. 原始产物索引

```
artifacts/benchmarks/
├─ run.log                                   # 脚本主日志（含 compose wait_url、k6 返回码、采样开始结束）
├─ monolith-bm01-r{1..3}.json / .txt         # BM01 单体 k6 summary-export + stdout
├─ monolith-bm02-r{1..3}.json / .txt         # BM02 单体
├─ monolith-bm03-r{1..3}.json / .txt         # BM03 单体
├─ microservices-bm01-r{1..3}.json / .txt    # BM01 微服务
├─ microservices-bm02-r{1..3}.json / .txt    # BM02 微服务
├─ microservices-bm03-r{1..3}.json / .txt    # BM03 微服务
├─ monolith-stats.csv                        # 单体容器 CPU/内存 20s 采样聚合
├─ microservices-stats.csv                   # 微服务容器 CPU/内存 20s 采样聚合
├─ comparison.csv                            # 脚本自动生成的扁平行（可粘贴进 Excel 画对比图）
└─ comparison.md                             # 脚本自动生成的 Markdown 报告（同目录下 comparison.md 模板）
```

---

## 6. 截图 / 附件占位

- [ ] 单体 3 轮 k6 终端输出截图（bm01/02/03 各一张）
- [ ] 微服务 3 轮 k6 终端输出截图（bm01/02/03 各一张）
- [ ] 单体 docker stats 热图（或 top/htop 截图）
- [ ] 微服务 docker stats 热图（三服务+gateway 各 CPU/内存柱）
- [ ] 可选：Prometheus/Grafana 的 RPS、P95、CPU 时间序列曲线截图
