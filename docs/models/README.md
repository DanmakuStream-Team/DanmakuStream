# PlantUML 模型索引

本目录是项目正式图的统一来源。每张正式图均保留 `.puml` 源文件，并在相同目录导出同名 `.svg` 和 `.png`。Markdown 文档优先嵌入 SVG，PNG 用于不支持 SVG 的汇报或平台。

| 分类 | 覆盖范围 | 编号/文件 |
| --- | --- | --- |
| 用例 | 互动与直播域总览 | `usecase/UC05-UC09-UC10-overview` |
| 用例说明 | UC13 | `usecase/UC13.md` |
| 系统级 | UC02、03、04、05、09、10、12、13 | `system/SYS-SEQxx` |
| 组件级 | UC02、03、04、05、09、10、12、13 | `component/COMP-SEQxx` |
| 对象级 | UC02、03、04、05、09、10、12、13 | `object/OBJ-SEQxx` |
| 领域设计 | 互动与直播域 | `object/DOMAIN-CLASS-D`、`component/COMPONENT-D`、`object/IMPLEMENTATION-CLASS-D` |
| 部署 | 单体、单体详图、Kubernetes | `deployment/DEPLOY-MONO`、`DEPLOY-MONO-DETAILED`、`DEPLOY-K8S` |

目前共有 31 个 PlantUML 正式源文件（不含 `_theme.puml`），全部应保留同名 SVG 和 PNG。UC01、UC06、UC07、UC08、UC11 的三层模型尚未提交。

重新导出时，在项目根目录执行：

```powershell
java -jar path/to/plantuml.jar -charset UTF-8 -tsvg "docs/models/**/*.puml"
java -jar path/to/plantuml.jar -charset UTF-8 -tpng "docs/models/**/*.puml"
```

`_theme.puml` 是公共样式，不是独立正式图。新增图必须引用该样式，并遵循 `SYS-SEQxx`、`COMP-SEQxx`、`OBJ-SEQxx` 等编号。

提交前运行 `python3 scripts/check_diagram_assets.py`。流水线会检查五个标准目录、部署图、顺序图编号、每张图的 SVG/PNG 导出物，并拒绝在正式 Markdown 文档中重新使用 Mermaid 图。

本轮逐图检查结果见 [正式图完整性检查记录](./diagram-review.md)。

## 用例三层模型编写样例

UC13 是 UC01～UC12 的统一参考样例。每个用例负责人需要交付：

1. `usecase/UCxx.md`：用例说明，包含参与者、前后置条件、主流程、异常流程、业务规则、可验证结果和代码对应。
2. `system/SYS-SEQxx.puml`：把系统当作黑盒，只画参与者与系统之间的业务消息。
3. `component/COMP-SEQxx.puml`：画页面、网关、中间件、Handler、数据库等组件协作。
4. `object/OBJ-SEQxx.puml`：选择该用例的代表性主流程，使用真实类/对象和方法名展开。
5. `../testing/test-cases/ucxx.md`：定义 UNIT、INT、E2E 测试及关键断言。
6. `../traceability/ucxx.md`：连接需求、模型、代码和测试结果。

## 复制要求

- 复制 UC13 文件后，统一替换 `13` 为目标用例编号。
- 必须重新核对真实参与者、接口、页面、Handler、Model 和异常状态码。
- 一张图可以选择顺序图、状态图或活动图；全组建议优先统一使用顺序图。
- 每张 `.puml` 必须同时导出 `.svg`，但不能只提交 SVG。
- 测试未执行时写“待执行”，不得预填“通过”。
