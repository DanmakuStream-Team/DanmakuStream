# DanmakuStream 文档中心

本目录按“项目管理—架构—模型—测试—追溯—成员交付”组织。文件名统一使用英文 kebab-case；PlantUML 图继续使用稳定的业务编号，便于需求追溯和自动校验。

## 目录导航

| 目录 | 内容 | 入口 |
| --- | --- | --- |
| `project/` | 课程要求、开发计划、协作规范、完整用例清单 | [项目文档](project/) |
| `architecture/` | 单体、Kubernetes 与微服务架构规范 | [部署设计](architecture/deployment-design.md) · [微服务统一规范](architecture/microservices/service-standards.md) |
| `deploy/` | CI/CD、镜像与部署运维说明 | [CD 流水线](deploy/cd-pipeline.md) · [微服务 CI](deploy/microservice-ci.md) · [微服务 CD](deploy/microservice-cd.md) · [HPA 实验](deploy/hpa-experiment.md) · [HPA/故障演练](deploy/hpa-chaos.md) |
| `models/` | PlantUML 源文件、SVG/PNG 与检查记录 | [模型索引](models/README.md) |
| `testing/` | 测试用例和原始测试报告 | [测试用例](testing/test-cases/) |
| `traceability/` | 需求—设计—代码—测试追溯（总表 + 单用例） | [总追溯表 master.md](traceability/master.md) |
| `contributions/` | 按成员归档的领域交付（B/D/E） | [成员 D 交付](contributions/member-d/README.md) · [成员 B](contributions/member-b/README.md) |

## 项目文档

- [课程任务书（2026 夏）](project/course-requirements-2026-summer.pdf)
- [软件开发计划](project/software-development-plan.md)
- [两周开发计划](project/two-week-development-plan.md)
- [Git 协作规范](project/git-workflow.md)
- [协作与验收规则](project/collaboration-and-acceptance.md)
- [业务场景用例清单](project/use-case-catalog.md)

## 架构规范

- [单体与 Kubernetes 部署设计](architecture/deployment-design.md)
- [微服务统一规范](architecture/microservices/service-standards.md)
- [微服务独立 CI 与镜像规范](deploy/microservice-ci.md)
- [三微服务自动部署与排障手册](deploy/microservice-cd.md)
- [HPA 自动扩缩容实验](deploy/hpa-experiment.md)
- [HPA 扩缩容与依赖故障演练](deploy/hpa-chaos.md)
- [最终版本冻结与 A 组答辩手册](delivery/release-freeze-and-demo.md)

## 当前交付缺口

以下状态按仓库现有文件盘点，不把计划中的材料计为已完成：

1. UC01、UC06、UC07、UC08、UC11 缺少系统级、组件级和对象级三层模型，共 15 张图；项目也缺少覆盖 UC01～UC13 的总用例图。
2. 除 UC13 及成员 D 的 UC05、UC09、UC10 外，其余用例缺少成套的独立说明、测试用例/结果与追溯材料；还缺 UC01～UC13 总追溯表。
3. 三个业务微服务、网关、独立 CI、微服务 E2E 和 SHA 自动部署已建立；完整 13/13 微服务 E2E 与实际集群验收材料仍需在流水线运行后归档。
4. 四个无状态工作负载的 HPA 配置（autoscaling.yaml 统一管理）、可重复压测、扩缩容曲线和依赖故障脚本已建立；仍需在实际 k3s 运行并归档，熔断实现和单体/微服务至少三轮性能对比仍未完成。
5. 单体 CD 与微服务 SHA 镜像推送/自动部署已建立；微服务 CD 检查探针、版本、HPA 和 metrics-server，失败回滚归档日志；最终 tag 由 release-freeze 创建。
6. 项目管理缺每日站会简报、看板/统计截图和汇总证据；最终答辩材料、技术总结、成员权重确认、AI 使用说明尚未归档。

详细提交范围以课程任务书为准。新增文档后请同步更新本索引和对应追溯表。
