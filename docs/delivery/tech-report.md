# DanmakuStream 技术报告（Tech Report）

> **项目名称**：DanmakuStream 弹幕视频直播平台
> **报告版本**：v1.0
> **里程碑**：M1 需求 → M2 基建 → M3 功能 → M4 测试与部署改造
> **报告日期**：YYYY-MM-DD
> **报告人/团队**：_____________

---

## 1. 项目总结

### 1.1 背景与目标

DanmakuStream 是一个面向创作者与社区的弹幕视频 + 直播平台，覆盖视频投稿审核、弹幕墙互动、直播推流 + 实时互动、创作者会员、个人视频资料库、用户私信、创作者数据分析和平台后台治理，并完成从 Docker Compose 单体到 Kubernetes 编排的部署改造。**交付目标**是 PRD 定义的 13 个核心用例（UC01~UC13）100% 追溯覆盖，关键用例 E2E 通过。

### 1.2 交付清单（是/否交付）

| 类别 | 交付物 | 状态 | 位置 |
|---|---|---|---|
| 业务用例 | UC01~UC13 代码实现（共 13 UC） | | `frontend/src/pages/` + `backend/internal/handler/v1/` |
| 直播媒体服务 | SRS 4（RTMP→HLS）配置 | | [srs.conf](../deploy/srs.conf) |
| 推荐系统（离线） | ItemCF / 基线 / 评测脚本 | | [recommendation/](../recommendation/) |
| 图档（PlantUML） | 总用例图 | ✅ 交付 | [usecase-overview.puml](../models/usecase/usecase-overview.puml) |
| 图档（PlantUML） | UC06 系统/组件/对象级顺序图 | ✅ 交付 | [sys-seq06-library.puml](../models/system/sys-seq06-library.puml)、[comp-seq06-library.puml](../models/component/comp-seq06-library.puml)、[obj-seq06-progress.puml](../models/object/obj-seq06-progress.puml) |
| 图档（PlantUML） | 单体/K8s 部署图 | ✅ 交付 | [deployment-monolith.puml](../models/deployment/deployment-monolith.puml)、[deployment-k8s.puml](../models/deployment/deployment-k8s.puml) |
| 用例说明 | UC06 详细文档 | ✅ 交付 | [UC06-personal-library.md](../usecase/UC06-personal-library.md) |
| 需求追溯 | 13 用例追溯矩阵 | ✅ 交付 | [traceability-matrix.md](../traceability/traceability-matrix.md) |
| 测试 | E2E 测试计划 | ✅ 交付 | [e2e-test-plan.md](../testing/e2e-test-plan.md) |
| 测试 | E2E 测试报告 | ✅ 交付（骨架，执行时填实） | [e2e-test-report.md](../testing/e2e-test-report.md) |
| 压测（Day 9） | k6 脚本 × 3 核心接口 | ✅ 交付 | `benchmarks/k6/01-public-video-search.js`、`02-auth-login.js`、`03-library-history.js` |
| 压测（Day 9） | 一键对比脚本（单体 vs 微服务各 3 轮，自动数据播种+资源采样+聚合） | ✅ 交付 | [run-benchmark-comparison.sh](../../scripts/run-benchmark-comparison.sh) |
| 压测（Day 9） | 对比报告模板（吞吐/Avg/P95/错误率/CPU/内存 对比表+行动项） | ✅ 交付（骨架，执行时填实） | [benchmark-comparison-template.md](../testing/benchmark-comparison-template.md) |
| 部署 | Docker Compose 单体 → K8s 设计 | ✅ 交付 | §4 详述 + 部署图 |

### 1.3 关键技术栈（实际版本填实）

| 层 | 技术 | 版本 |
|---|---|---|
| 前端 | Vue 3 + TS + Vite + Element Plus + Pinia | |
| 测试 | Playwright（Chromium） | |
| 后端 | Go + Gin + GORM | |
| DB | MySQL 8.0 | |
| 实时通信 | Gorilla WebSocket（danmaku/chat 双 Hub） | |
| 媒体 | SRS 4（RTMP→HLS） | |
| 离线推荐 | Python（NumPy/Pandas 按需） | |
| 容器 | Docker + docker-compose.yml | |
| 编排（改造后） | Kubernetes（Deployment/StatefulSet/HPA/Ingress） | |

### 1.4 里程碑完成度

| 里程碑 | 交付物 | 完成度 | 偏差 |
|---|---|---|---|
| M1 需求对齐 | 13 UC 业务场景清单冻结、PRD 确认 | | |
| M2 基础设施 | 后端脚手架 + 前端路由 + DB 模型 + Docker Compose + Playwright 基建 | | |
| M3 功能实现 | 13 UC 代码 + UC06 三层图 + UC13 自动化 | | |
| M4 测试与部署改造 | K8s 部署图 + E2E 计划/报告 + 13 UC 追溯表 + 最终技术报告 | | |

### 1.5 遗留问题与后续计划

| 编号 | 问题 | 影响 | 计划 |
|---|---|---|---|
| R-01 | UC11 大附件上传暂不支持分片与断点续传 | 弱网下的大附件重传成本较高 | 下一迭代补分片上传、断点续传与失败重试 |
| R-02 | 推荐能力未接入线上 `/recommendations` 路由 | 当前仅有本地标签聚合 + 离线脚本 | 下一迭代接入线上 API + A/B 评估 |
| R-03 | WebSocket 多 Pod 水平扩容一致性（Hub 内存态） | K8s 多副本弹幕不互通 | 引入 Redis Pub/Sub 统一消息平面 |
| R-04 | SRS 在现有 CI 环境缺部署 | UC10-01 媒体链路常被标 [MEDIA-REQUIRED] 阻塞 | docker-compose srs service 纳入统一 CI Job |

---

## 2. AI 使用说明

### 2.1 使用场景与工具清单

> **原则**：AI 仅用于**加速生成骨架、模板、代码提示、测试步骤设计**，不替代业务评审、安全复核与最终代码审查。所有 AI 产出必须通过人工 review + 自动化验证。

| 场景 | 使用的 AI 工具 | 输入 | 输出 | 人工复核环节 |
|---|---|---|---|---|
| 图档与文档反向生成 | Trae（本会话） | 现有代码库 + 业务用例清单 | 总用例图、UC06 三层顺序图、部署图、用例说明、矩阵、计划/报告、技术报告 | 架构评审 + 逐字段核对代码路径与接口 |
| 代码补全/签名建议 | IDE 内置（如 Copilot） | 本地上下文 | Handler 模板、模型 JSON tag、前端 SFC 骨架 | Diff Review + go test / tsc --noEmit |
| 测试用例设计 | Trae + 人工细化 | PRD 主/异常流程 | 13 UC ≥ 39 条 E2E 步骤 + 通过标准 | Playwright 实际执行通过 |
| 缺陷初筛 | 通用 LLM（如使用） | 失败日志 + trace | 根因候选 | 开发复现+抓包+DB 校验后确认 |
| 部署脚本/图档草拟 | Trae + 人工 | docker-compose.yml、SRS conf | 单体/K8s 部署图 Puml + K8s manifest 草稿 | kubectl dry-run + 压测 |

### 2.2 使用边界（硬性红线）

1. **业务决策不交给 AI**：角色边界 user/creator/moderator/admin、视频状态机、订阅价格周期、到期/预约时间阈值等 PRD 硬约束，AI 不做最终判断。
2. **安全代码强制人工复审**：JWT 中间件、CORS、媒体上传白名单、SQL 拼接（必须参数化）、后台 staff/admin 守卫。
3. **密钥与测试数据分离**：AI 生成的计划统一使用 `Test1234!` 类通用密码；生产密钥仅人工注入 Secret/ConfigMap。
4. **编号一致**：PlantUML 严格按 `SYS/COMP/OBJ-SEQxx` 与 [models/README.md](../models/README.md) 的约定。

### 2.3 验证流水线

```
AI 生成骨架/代码/图档/文档
        ↓
人工 Diff Review
        ↓
静态检查：go vet / tsc --noEmit / eslint / plantuml 语法
        ↓
动态验证：go test ./... ；npx playwright test
        ↓
结果写入 [e2e-test-report.md](../testing/e2e-test-report.md) 与追溯矩阵
```

### 2.4 真实 AI 占比（交付时填实）

| 类别 | 总条目/行数 | 其中 AI 辅助 | 人工编写 | AI 辅助占比 |
|---|---|---|---|---|
| 后端 Go (*.go) | | | | % |
| 前端 Vue/TS (*.vue,*.ts) | | | | % |
| PlantUML (*.puml) | | | | % |
| 测试 (*_test.go / *.spec.ts / shell) | | | | % |
| 文档 (*.md) | | | | % |
| **加权合计（按工作量）** | — | — | — | **%** |

---

## 3. 个人权重

### 3.1 多人协作模板（复制行加至合计 100%）

| 成员 | 需求/PRD | 架构/设计图 | 后端开发 | 前端开发 | 测试/自动化 | 运维/部署改造 | 文档/追溯 | **行内小计** |
|---|---|---|---|---|---|---|---|---|
| 成员 A（例：后端主程） | 3% | 5% | 20% | 0% | 3% | 4% | 5% | **40%** |
| 成员 B（例：前端主程） | 2% | 3% | 0% | 20% | 5% | 2% | 3% | **35%** |
| 成员 C（例：测试/运维） | 1% | 2% | 0% | 0% | 10% | 6% | 6% | **25%** |
| **合计** | **6%** | **10%** | **20%** | **20%** | **18%** | **12%** | **14%** | **100%** |

### 3.2 单人项目（默认）

| 成员 | 需求/PRD | 架构/设计图 | 后端开发 | 前端开发 | 测试/自动化 | 运维/部署改造 | 文档/追溯 | **行内小计** |
|---|---|---|---|---|---|---|---|---|
| 本人（姓名） | 15% | 15% | 15% | 15% | 15% | 10% | 15% | **100%** |

### 3.3 证据链（支撑权重）

| 证据 | 位置 |
|---|---|
| Git 贡献者分布 | `git shortlog -sn --all` |
| 缺陷修复记录 | GitHub Issues/PRs |
| 追溯表签核 | [traceability-matrix.md](../traceability/traceability-matrix.md) 核对清单 |
| E2E 落地清单 | [e2e-test-plan.md](../testing/e2e-test-plan.md) §2 |

---

## 4. 部署改造：单体 → Kubernetes

配套部署图：
- 改造前：[deployment-monolith.puml](../models/deployment/deployment-monolith.puml)
- 改造后：[deployment-k8s.puml](../models/deployment/deployment-k8s.puml)

### 4.1 改造前（Docker Compose Monolith）

- **主机**：单台 Docker Host，共享故障域。
- **4 容器**：frontend（Nginx+Vue）、backend（Gin，单实例）、srs（单实例）、mysql（单实例）。
- **存储**：宿主机目录 bind mount，承载视频/封面/HLS 分片。
- **路由**：Nginx 反代 `/api`、`/ws`、`/live/*.m3u8`。
- **适用**：开发联调、演示、小规模上线。
- **问题**：单点故障、难以水平扩容、MySQL/SRS 无法独立升级。

### 4.2 改造后（Kubernetes）

| 组件 | 资源 | 副本 | 说明 |
|---|---|---|---|
| Frontend | Deployment + Service + HPA | ≥ 2 | Nginx+Vue；按 CPU/请求扩缩 |
| Backend API/WS | Deployment + Service + HPA | ≥ 2 | Gin + WS；按 CPU/连接数扩缩 |
| Worker（Schedule/Expiration） | Deployment | 1+（leader 或幂等） | 避免重复扫描/重复发券 |
| SRS | Deployment / StatefulSet | 1+ | RTMP/HLS + PVC 绑定 HLS 分片 |
| MySQL | StatefulSet 或云托管 | 主从/高可用 | 建议托管 RDS/Aurora/Cloud SQL |
| 配置/密钥 | ConfigMap + Secret | — | 分离挂载，不进镜像 |
| 入口 | Ingress + LoadBalancer | — | `/`→Web，`/api`/`/ws`→API，`/live`→SRS |
| 存储 | PersistentVolumeClaim | — | 视频/封面/HLS；支持快照扩容 |

### 4.3 改造验证清单

| 验证项 | 方法 | 结果（交付时填） |
|---|---|---|
| Frontend ≥ 2 Pod 流量均摊 | k6/ab 压测 + accesslog 分布 | |
| Backend WebSocket 多 Pod 稳定 | 并发 1k 观众，3 分钟在线人数不重复计 | |
| Worker 幂等 | 手动触发多次扫描；DB 仅变更 1 次 | |
| SRS RTMP→HLS | FFmpeg 推流→下载主清单+TS 分片 | |
| MySQL StatefulSet 备份恢复 | kill pod / 恢复快照 | |
| PV 扩容不丢文件 | 扩容后 `ls` 文件数一致 | |
| HPA 阈值触发 | 模拟高 CPU/连接；副本数升+降 | |
| **单体 vs 微服务吞吐差（BM01/BM02/BM03）** | Day9 脚本：`bash scripts/run-benchmark-comparison.sh` 各 3 轮（见 `artifacts/benchmarks/comparison.md`） | |
| **单体 vs 微服务 P95（BM01/02/03）** | 同上，k6 p(95) 指标对比，阈值见 §5.1 | |
| **单体 vs 微服务错误率** | 同上，`http_req_failed` 阈值 <2%/单接口 <5%/合计 | |
| **单体后端进程 vs 三服务总内存 RSS** | `docker stats` 20s 采样（`*-stats.csv`），见压测报告 §2 | |

---

## 5. 质量与风险

### 5.1 质量指标

| 指标 | 目标 | 实际 | 达成 |
|---|---|---|---|
| 13 UC 追溯率 | 100% | | |
| E2E（UC13 5 条先落地） | 100%（阻塞单独标注） | | |
| 后端单测覆盖率 | ≥ 60% | | |
| E2E 条目总数 | ≥ 39 条 | | |
| 关键 API P95（单体 3 轮平均） | BM01≤800ms / BM02≤1500ms / BM03≤1200ms | BM01: __ms / BM02: __ms / BM03: __ms | |
| 关键 API P95（微服务 3 轮平均） | 同单体阈值 | BM01: __ms / BM02: __ms / BM03: __ms | |
| 压测错误率（三轮平均） | < 5% | % | |
| WS 断连恢复 | 3s 重试、计数不重复 | | |
| P0/P1 严重缺陷 | 0 | | |

> 压测脚本：`bash scripts/run-benchmark-comparison.sh`；报告模板：`docs/testing/benchmark-comparison-template.md`；
> 原始产物：`artifacts/benchmarks/{monolith,microservices}-bm{01,02,03}-r{1,2,3}.json/.txt`、`comparison.csv/.md`。

### 5.2 风险清单

| 风险 | 概率 | 影响 | 缓解 | 状态 |
|---|---|---|---|---|
| UC11 大附件上传缺少断点续传 | | | 后续补分片上传、断点续传与弱网重试测试 | |
| SRS CI 不可达 → UC10 媒体阻塞 | | | compose srs service + [MEDIA-REQUIRED] 标记 | |
| 推荐能力线上接入延期 | | | 本地标签聚合先满足主流程 | |
| WS 多 Pod 消息不互通 | | | Redis Pub/Sub（下一迭代） | |
| K8s 多 Pod 本地 localStorage liked/collections 不一致 | | | 明确文档说明「liked/collections/downloads 本机」，符合 UC06 设计 | |

---

## 6. 附录：交付物索引

| 交付物 | 路径 |
|---|---|
| 总用例图 | [usecase-overview.puml](../models/usecase/usecase-overview.puml) |
| UC06 系统级顺序图 | [sys-seq06-library.puml](../models/system/sys-seq06-library.puml) |
| UC06 组件级顺序图 | [comp-seq06-library.puml](../models/component/comp-seq06-library.puml) |
| UC06 对象级顺序图（进度） | [obj-seq06-progress.puml](../models/object/obj-seq06-progress.puml) |
| 单体部署图 | [deployment-monolith.puml](../models/deployment/deployment-monolith.puml) |
| K8s 部署图 | [deployment-k8s.puml](../models/deployment/deployment-k8s.puml) |
| UC06 用例说明 | [UC06-personal-library.md](../usecase/UC06-personal-library.md) |
| 13 UC 追溯矩阵 | [traceability-matrix.md](../traceability/traceability-matrix.md) |
| E2E 测试计划 | [e2e-test-plan.md](../testing/e2e-test-plan.md) |
| E2E 测试报告 | [e2e-test-report.md](../testing/e2e-test-report.md) |

---

**报告人签名**：_____________  日期：YYYY-MM-DD
**审阅人签名**：_____________  日期：YYYY-MM-DD
