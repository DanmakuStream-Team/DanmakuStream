# DanmakuStream 项目技术报告（骨架）

> **项目名称**：DanmakuStream 弹幕视频直播平台
> **报告版本**：v1.0（骨架，交付时逐项填实）
> **报告周期 / 里程碑**：M1 需求对齐 → M2 基础设施 → M3 功能开发 → M4 测试与部署改造
> **报告日期**：YYYY-MM-DD
> **报告人 / 团队**：_____________

---

## 1. 项目总结

### 1.1 项目背景与目标

> 简述：面向二次元/创作者社区的弹幕视频 + 直播平台，提供视频投稿审核、弹幕墙互动、直播推流与实时互动、创作者会员、个人资料库、创作者数据分析与平台后台治理等能力。
>
> **目标**：交付符合 PRD 定义的 13 个核心用例（UC01~UC10、UC12~UC14），实现单体到 Kubernetes 的部署改造，追溯率 100%，关键用例 E2E 通过。

### 1.2 范围与交付清单

| 类别 | 交付物 | 是否已交付（是/否） | 备注 |
|---|---|---|---|
| 业务用例 | UC01~UC10、UC12~UC14 共 13 UC 实现（UC11 空缺待确认） | | |
| 前端 | Vue 3 + Vite + Element Plus + Pinia 页面与组件（见 frontend/src/pages） | | 路由守卫与鉴权已落地 |
| 后端 | Go + Gin + GORM + MySQL API 与 WebSocket 服务（见 backend/internal） | | ScheduleWorker / ExpirationWorker 已启动 |
| 直播服务 | SRS 4（RTMP 推流 + HLS 分发），见 deploy/srs.conf | | 媒体链路独立验证 |
| 推荐系统（离线） | Python ItemCF / 基线 / 评测脚本，见 recommendation/ | | 线上 API 按 UC14 接入进度 |
| 设计图档 | PlantUML：总用例图、UC06 三层顺序图、部署图等 | | 导出 SVG/PNG |
| 测试 | 单元、API（UC13 脚本样例）、E2E（UC13 已落地 + 其余 12 UC 计划与报告骨架） | | Playwright 框架 |
| 部署改造 | Docker Compose（单体）→ Kubernetes（Deployment/StatefulSet/HPA/Ingress） | | 详见 §4 |
| 文档 | UC06 用例说明、13 用例追溯表、E2E 计划/报告、最终追溯表、本技术报告 | | |

### 1.3 关键技术栈

| 层 | 技术 | 版本（实际） |
|---|---|---|
| 前端框架 | Vue 3 + TypeScript + Vite | |
| UI 库 | Element Plus | |
| 状态管理 | Pinia | |
| E2E | Playwright (Chromium) | |
| 后端语言 | Go | |
| Web 框架 | Gin | |
| ORM | GORM | |
| 数据库 | MySQL | |
| 实时通信 | Gorilla WebSocket（自研 Hub：danmaku/chat） | |
| 直播媒体 | SRS | |
| 离线推荐 | Python（NumPy/Pandas 按需） | |
| 容器 | Docker + docker-compose.yml | |
| 编排（改造后） | Kubernetes | |

### 1.4 项目成果摘要（按里程碑）

| 里程碑 | 核心交付 | 实际完成度 | 偏差说明 |
|---|---|---|---|
| M1 需求对齐 | 13 UC 业务场景清单冻结、PRD 确认 | | |
| M2 基础设施 | 后端脚手架 + 前端路由 + DB 模型 + Docker Compose + Playwright 基建 | | |
| M3 功能实现 | 13 UC 代码实现 + UC06/UC13 三层图 + UC13 自动化 | | |
| M4 测试与部署改造 | K8s 部署图 + E2E 计划/报告骨架 + 13 UC 追溯表 | | |

### 1.5 遗留问题与后续计划

> 列出未完成项：例如 UC14 推荐线上 API、UC11 签核、SRS CI 环境、剩余 8 UC 三层图等。
>
| 编号 | 问题 | 影响范围 | 建议计划 |
|---|---|---|---|
| R-01 | | | |
| R-02 | | | |

---

## 2. AI 使用说明

### 2.1 使用场景与工具清单

> **原则**：AI 仅用于**加速生成**骨架、模板、代码提示与测试用例设计，**不替代架构决策、业务评审与最终代码审查**。所有 AI 产出均经过代码审查与人工验证。

| 场景 | 使用的 AI 工具/服务 | 输入物 | 输出物 | 人工复核环节 |
|---|---|---|---|---|
| 需求文档与图档反向生成 | Trae（本会话） | 现有代码库 + 业务用例清单 | 总用例图、UC06 三层顺序图、部署图、用例说明、追溯表、E2E 计划/报告、技术报告骨架 | 架构评审 + 逐字段核对代码路径与接口 |
| 代码补全与函数签名建议 | IDE 内置 AI 补全（如 Copilot） | 本地上下文 | Handler 模板、模型 JSON tag、前端 SFC 骨架 | Diff Review + 单元/集成测试通过 |
| 单元/API 测试用例设计 | Trae + 人工细化 | PRD 主/异常流程 | UNIT/INT 编号、断言要点 | 执行结果与覆盖率门限 |
| E2E 步骤与断言设计 | Trae（本会话 E2E 计划 §2） | 13 UC 主/异常流程 | 39+ 条 E2E 步骤与通过标准 | Playwright 实际录制/回放通过 |
| 缺陷初筛与根因提示 | 通用 LLM（如使用） | 失败日志 + trace | 根因候选列表 | 开发复现、抓包与 DB 校验后确认 |
| 部署图与运维脚本草拟 | Trae + 人工 | docker-compose.yml、SRS 配置 | DEPLOY-MONO / DEPLOY-K8S Puml 与 K8s manifest 草稿 | kubectl dry-run 与压测验证 |

### 2.2 使用边界与风险控制

1. **业务决策不交给 AI**：角色权限边界（staff/admin）、状态机（pending→approved→rejected）、价格与订阅周期等 PRD 硬约束，AI 不做最终判断。
2. **安全相关代码必须人工审查**：JWT 中间件、CORS、SQL 参数化、媒体上传路径白名单等，AI 建议必须经过至少一次人工安全 Review。
3. **测试数据与密钥规则**：AI 生成的计划一律使用 `Test1234!` 等通用测试密码，绝不允许出现在真实环境；生产密钥由人工注入 Secret / ConfigMap，不写入仓库。
4. **图档命名与编号一致**：PlantUML 编号规则（SYS/COMP/OBJ-SEQxx）严格按照 [README.md](file:///d:/DanmakuStream/docs/models/README.md)，AI 生成后人工核对编号避免错位。

### 2.3 生成内容的验证流程

```
AI 生成骨架/代码/脚本
        ↓
人工 Diff Review（IDE 或 PR Reviewer）
        ↓
静态检查：go vet / tsc --noEmit / eslint
        ↓
动态验证：go test ./... / npx playwright test
        ↓
结果写入《E2E 测试报告》与《最终追溯表》
```

### 2.4 真实 AI 使用占比（交付时填实）

| 类别 | 行数 / 条目总数 | 其中 AI 辅助生成（估算） | 人工编写 | AI 辅助占比 |
|---|---|---|---|---|
| 后端 Go 代码（*.go） | | | | % |
| 前端 Vue/TS 代码（*.vue,*.ts） | | | | % |
| PlantUML 图档（*.puml） | | | | % |
| 测试脚本（*_test.go / *.spec.ts / shell） | | | | % |
| 文档（*.md） | | | | % |
| **加权合计（按工作量）** | — | — | — | **%** |

---

## 3. 个人权重（成员贡献度，多人协作时填写）

> **说明**：
> - 若为单人项目：个人权重 = 100%。
> - 若为团队项目：每人一行，按维度拆分（需求/架构/后端/前端/测试/运维/文档），四列相加等于 100% 行内小计，所有人小计之和 = 100%。
> - 权重由团队互评与交付物签核确认，不作为唯一考核依据，与代码提交数、缺陷数、追溯表签字页共同构成综合证据。

### 3.1 多人协作模板（可复制行）

| 成员 | 需求与 PRD | 架构与设计图 | 后端开发 | 前端开发 | 测试与自动化 | 运维与部署改造 | 文档与追溯 | **行内小计** |
|---|---|---|---|---|---|---|---|---|
| 成员 A（例：后端主程） | 3% | 5% | 20% | 0% | 3% | 4% | 5% | **40%** |
| 成员 B（例：前端主程） | 2% | 3% | 0% | 20% | 5% | 2% | 3% | **35%** |
| 成员 C（例：测试/运维） | 1% | 2% | 0% | 0% | 10% | 6% | 6% | **25%** |
| **合计** | **6%** | **10%** | **20%** | **20%** | **18%** | **12%** | **14%** | **100%** |

### 3.2 单人项目（默认）

| 成员 | 需求与 PRD | 架构与设计图 | 后端开发 | 前端开发 | 测试与自动化 | 运维与部署改造 | 文档与追溯 | **行内小计** |
|---|---|---|---|---|---|---|---|---|
| 本人（姓名） | 15% | 15% | 15% | 15% | 15% | 10% | 15% | **100%** |

### 3.3 证据链（支撑权重）

| 证据 | 路径 |
|---|---|
| Git 提交统计（贡献者分布） | git shortlog -sn --all |
| 缺陷修复记录 | GitHub Issues / PRs（实际填真实链接） |
| UC 负责人签名 | [final-traceability-skeleton.md](file:///d:/DanmakuStream/docs/traceability/final-traceability-skeleton.md) §6 审批表 |
| 自动化测试实现清单 | [e2e-test-plan.md](file:///d:/DanmakuStream/docs/tests/e2e-test-plan.md) |

---

## 4. 部署改造说明（单体 → K8s）

> 配套图：
> - 改造前：[DEPLOY-MONO.puml](file:///d:/DanmakuStream/docs/models/deployment/DEPLOY-MONO.puml)
> - 改造后：[DEPLOY-K8S.puml](file:///d:/DanmakuStream/docs/models/deployment/DEPLOY-K8S.puml)

### 4.1 改造前（Docker Compose 单体）

- **主机**：单台 Docker Host，共享故障域。
- **容器**：frontend（Nginx + 静态资源）、backend（单实例 Gin）、srs（单实例）、mysql（单实例）。
- **存储**：宿主机目录 bind mount，存放视频、封面、HLS 分片。
- **路由**：Nginx 反向代理 `/api`、`/ws`、`/live/*.m3u8`。
- **适用场景**：开发联调、演示、小规模上线。
- **已知问题**：单点故障、难以水平扩容、MySQL 与 SRS 不可独立升级。

### 4.2 改造后（Kubernetes）

- **Frontend Deployment**：Nginx + Vue，replicas ≥ 2，CPU/流量 HPA。
- **Backend Deployment**：Gin API + WebSocket，replicas ≥ 2，连接数 HPA。
- **Worker Deployment**：Schedule / Expiration Worker，leader 选举或幂等扫描。
- **SRS Deployment / StatefulSet**：RTMP/HLS 与 PV 绑定（若多副本需共享存储方案）。
- **MySQL**：建议 StatefulSet 或云托管数据库（RDS/Aurora/Cloud SQL）。
- **PersistentVolume**：视频/封面/HLS 分片走 PVC，支持快照与扩容。
- **ConfigMap / Secret**：配置与密钥解耦，分别挂载到各 Deployment。
- **Ingress / LoadBalancer**：统一入口，按路径路由 `/`、`/api`、`/ws`、`/live`、`/rtmp`。
- **收益**：滚动升级、水平扩缩容、故障自愈、蓝绿发布、资源隔离。

### 4.3 改造验证清单

| 验证项 | 方法 | 结果（交付时填） |
|---|---|---|
| Frontend ≥ 2 Pod 流量均摊 | ab / k6 压测 + 查看 accesslog | |
| Backend WebSocket 多 Pod 连接稳定 | 并发 1k 观众，3 分钟不重复计数 | |
| Worker 幂等（UC08 到期/UC09 提醒） | 手动触发多次扫描，DB 只产生 1 次变更 | |
| SRS RTMP/HLS 回放 | FFmpeg 推流 → HLS 下载主清单与 TS | |
| MySQL StatefulSet 主从/备份 | 故障切换演练 + 备份恢复 | |
| PV 扩容不丢数据 | 容量扩容后 ls 文件数一致 | |
| HPA 触发阈值 | 模拟高 CPU/连接数，观察 replicas 上升/回落 | |

---

## 5. 质量与风险回顾

### 5.1 质量指标（交付时填实）

| 指标 | 目标值 | 实际值 | 是否达成 |
|---|---|---|---|
| 13 UC 追溯率 | 100% | | |
| E2E 通过率（UC13 5 条先落地） | 100%（阻塞允许标注） | | |
| 后端单元测试覆盖率 | ≥ 60% | | |
| 前端 E2E 覆盖率（按用例） | 13 UC × ≥ 3 条 | | |
| 关键 API P95 响应 | ≤ 1s | | |
| WebSocket 断连自动恢复 | 3s 重试 + 不重复计数 | | |
| 严重/高危缺陷（P0/P1） | 0 | | |

### 5.2 风险清单（交付时填实）

| 风险 | 概率 | 影响 | 缓解措施 | 状态 |
|---|---|---|---|---|
| UC11 未签核导致范围蔓延 | | | 冻结 UC11 为后续增量交付 | |
| SRS 在 CI 环境不可达 | | | docker-compose srs service + [MEDIA-REQUIRED] 单独标记 | |
| 推荐系统线上 API 延期 | | | 本地标签聚合先满足 UC14 主流程 | |
| WebSocket 水平扩容一致性 | | | Hub 外部化（Redis Pub/Sub，后续迭代） | |

---

## 6. 附录：关键交付物索引

| 交付物 | 路径 |
|---|---|
| 总用例图 | [UC-OVERVIEW.puml](file:///d:/DanmakuStream/docs/models/usecase/UC-OVERVIEW.puml) |
| UC06 用例说明 | [UC06.md](file:///d:/DanmakuStream/docs/models/usercase/UC06.md) |
| UC06 三层顺序图 | SYS/COMP/OBJ-SEQ06.puml（docs/models/system|component|object/） |
| 改造前部署图 | [DEPLOY-MONO.puml](file:///d:/DanmakuStream/docs/models/deployment/DEPLOY-MONO.puml) |
| 改造后部署图 | [DEPLOY-K8S.puml](file:///d:/DanmakuStream/docs/models/deployment/DEPLOY-K8S.puml) |
| 13 用例追溯表（详细） | [all-uc-traceability.md](file:///d:/DanmakuStream/docs/traceability/all-uc-traceability.md) |
| E2E 测试计划 | [e2e-test-plan.md](file:///d:/DanmakuStream/docs/tests/e2e-test-plan.md) |
| E2E 测试报告骨架（填实版） | [e2e-test-report-skeleton.md](file:///d:/DanmakuStream/docs/tests/reports/e2e-test-report-skeleton.md) |
| 最终追溯表骨架（填实版） | [final-traceability-skeleton.md](file:///d:/DanmakuStream/docs/traceability/final-traceability-skeleton.md) |

---

**报告人签名**：____________________    **日期**：YYYY-MM-DD
**审阅人签名**：____________________    **日期**：YYYY-MM-DD
