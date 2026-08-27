# DanmakuStream 文档中心

本目录按“项目管理—架构—模型—测试—追溯—成员交付”组织。文件名统一使用英文 kebab-case；PlantUML 图继续使用稳定的业务编号，便于需求追溯和自动校验。

## 目录导航

| 目录 | 内容 | 入口 |
| --- | --- | --- |
| `project/` | 课程要求、开发计划、协作规范、完整用例清单 | [项目文档](project/) |
| `architecture/` | 单体与 Kubernetes 部署设计 | [部署设计](architecture/deployment-design.md) |
| `models/` | PlantUML 源文件、SVG/PNG 与检查记录 | [模型索引](models/README.md) |
| `testing/` | 测试用例和原始测试报告 | [测试用例](testing/test-cases/) |
| `traceability/` | 需求—设计—代码—测试追溯 | [UC13 追溯表](traceability/uc13.md) |
| `contributions/` | 按成员归档的领域交付 | [成员 D 交付](contributions/member-d/README.md) |

## 项目文档

- [课程任务书（2026 夏）](project/course-requirements-2026-summer.pdf)
- [软件开发计划](project/software-development-plan.md)
- [两周开发计划](project/two-week-development-plan.md)
- [Git 协作规范](project/git-workflow.md)
- [协作与验收规则](project/collaboration-and-acceptance.md)
- [业务场景用例清单](project/use-case-catalog.md)

## 当前交付缺口

以下状态按仓库现有文件盘点，不把计划中的材料计为已完成：

1. UC01、UC06、UC07、UC08、UC11 缺少系统级、组件级和对象级三层模型，共 15 张图；项目也缺少覆盖 UC01～UC13 的总用例图。
2. 除 UC13 及成员 D 的 UC05、UC09、UC10 外，其余用例缺少成套的独立说明、测试用例/结果与追溯材料；还缺 UC01～UC13 总追溯表。
3. 微服务阶段缺少服务划分图、服务接口清单、数据表归属矩阵、跨服务调用与失败处理说明，以及不少于三个业务微服务的部署与验证材料。
4. 云原生实验缺少 HPA 扩缩容原始数据、故障注入与降级/超时/熔断结果，以及单体和微服务在相同条件下至少三轮的性能对比原始数据与分析。
5. CI 当前只完成测试和镜像构建；缺镜像推送、Kubernetes 自动部署、部署后健康检查及失败阻断的实际工作流，也缺部署/回滚脚本和 HPA 清单。
6. 项目管理缺每日站会简报、看板/统计截图和汇总证据；最终答辩材料、技术总结、成员权重确认、AI 使用说明尚未归档。

详细提交范围以课程任务书为准。新增文档后请同步更新本索引和对应追溯表。
