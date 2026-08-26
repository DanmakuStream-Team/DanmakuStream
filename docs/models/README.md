# PlantUML 模型索引

本目录是项目正式图的统一来源。每张正式图均保留 `.puml` 源文件，并在相同目录导出同名 `.svg` 和 `.png`。Markdown 文档优先嵌入 SVG，PNG 用于不支持 SVG 的汇报或平台。

| 分类 | 图 | 编号/文件 |
| --- | --- | --- |
| 用例 | 互动与直播域用例图 | `usecase/UC05-UC09-UC10-overview` |
| 类图 | 互动与直播域概念类图 | `object/DOMAIN-CLASS-D` |
| 系统级 | UC05、UC09、UC10 系统级顺序图 | `system/SYS-SEQ05`、`SYS-SEQ09`、`SYS-SEQ10` |
| 组件 | 当前实现组件图 | `component/COMPONENT-D` |
| 组件级 | UC05、UC09、UC10 组件级顺序图 | `component/COMP-SEQ05`、`COMP-SEQ09`、`COMP-SEQ10` |
| 对象 | 当前实现类图 | `object/IMPLEMENTATION-CLASS-D` |
| 对象级 | UC05、UC09、UC10 对象级顺序图 | `object/OBJ-SEQ05`、`OBJ-SEQ09`、`OBJ-SEQ10` |
| 部署 | 改造前单体部署图 | `deployment/DEPLOY-MONO` |
| 部署 | 改造后 Kubernetes 部署图 | `deployment/DEPLOY-K8S` |

重新导出时，在项目根目录执行：

```powershell
java -jar path/to/plantuml.jar -charset UTF-8 -tsvg "docs/models/**/*.puml"
java -jar path/to/plantuml.jar -charset UTF-8 -tpng "docs/models/**/*.puml"
```

`_theme.puml` 是公共样式，不是独立正式图。新增图必须引用该样式，并遵循 `SYS-SEQxx`、`COMP-SEQxx`、`OBJ-SEQxx` 等编号。

提交前运行 `python scripts/check_diagram_assets.py`。流水线会检查五个标准目录、部署图、顺序图编号、每张图的 SVG/PNG 导出物，并拒绝在正式 Markdown 文档中重新使用 Mermaid 图。

本轮逐图检查结果见 [正式图完整性检查记录](./图检查记录.md)。

## 用例三层模型编写样例

UC13 是 UC01～UC12 的统一参考样例。每个用例负责人需要交付：

1. `usercase/UCxx.md`：用例说明，包含参与者、前后置条件、主流程、异常流程、业务规则、可验证结果和代码对应。
2. `system/SYS-SEQxx.puml`：把系统当作黑盒，只画参与者与系统之间的业务消息。
3. `component/COMP-SEQxx.puml`：画页面、网关、中间件、Handler、数据库等组件协作。
4. `object/OBJ-SEQxx.puml`：选择该用例的代表性主流程，使用真实类/对象和方法名展开。
5. `../tests/UCxx-test-cases.md`：定义 UNIT、INT、E2E 测试及关键断言。
6. `../traceability/UCxx-traceability.md`：连接需求、模型、代码和测试结果。

## 复制要求

- 复制 UC13 文件后，统一替换 `13` 为目标用例编号。
- 必须重新核对真实参与者、接口、页面、Handler、Model 和异常状态码。
- 一张图可以选择顺序图、状态图或活动图；全组建议优先统一使用顺序图。
- 每张 `.puml` 必须同时导出 `.svg`，但不能只提交 SVG。
- 测试未执行时写“待执行”，不得预填“通过”。
