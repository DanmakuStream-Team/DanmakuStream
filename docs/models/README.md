# PlantUML 模型索引

本目录是项目正式图的统一来源。正式图统一使用 PlantUML，保留 `.puml` 源文件，并在相同目录导出同名 `.svg` 和 `.png`。Markdown 文档优先嵌入 SVG，PNG 用于不支持 SVG 的汇报或平台。

## 目录与图类

- `usecase/`：用例说明、需求追溯和总用例图。
- `class/`：四个用例共享的领域类图。
- `system/`：需求说明书层面的系统级顺序图。
- `component/`：概要设计层面的组件图和组件级顺序图。
- `object/`：详细设计层面的实现类图和对象级顺序图。
- `deployment/`：改造前单体和改造后 Kubernetes 部署图。

当前成员 B 用例图：`UC01`、`UC07`、`UC08`、`UC11`。
当前成员 D 用例图：`UC05`、`UC09`、`UC10`。
平台管理样例：`UC13`。

## 编号

统一采用：`REQxx / UCxx / SYS-SEQxx / COMP-SEQxx / OBJ-SEQxx / UNIT-TCxx / INT-TCxx / E2E-TCxx`。

## 公共主题与导出

新增图可以引用 `_theme.puml` 公共样式。重新导出时，在项目根目录执行：

```powershell
java -jar path/to/plantuml.jar -charset UTF-8 -tsvg "docs/models/**/*.puml"
java -jar path/to/plantuml.jar -charset UTF-8 -tpng "docs/models/**/*.puml"
```

若 PowerShell 未展开递归通配符，可使用：

```powershell
Get-ChildItem docs/models -Recurse -Filter *.puml | ForEach-Object { plantuml -tsvg $_.FullName }
Get-ChildItem docs/models -Recurse -Filter *.puml | ForEach-Object { plantuml -tpng $_.FullName }
```

提交前运行 `python scripts/check_diagram_assets.py`。检查内容包括标准目录、部署图、顺序图编号、每张图的 SVG/PNG 导出物，以及正式 Markdown 中不得重新使用 Mermaid 图。

## 用例交付约定

每个用例负责人需要交付：

1. `usecase/UCxx.md`：参与者、前后置条件、主流程、异常流程、业务规则、可验证结果和代码对应。
2. `system/SYS-SEQxx.puml`：把系统当作黑盒，只画参与者与系统之间的业务消息。
3. `component/COMP-SEQxx.puml`：画页面、网关、中间件、Handler、数据库等组件协作。
4. `object/OBJ-SEQxx.puml`：使用真实类/对象和方法名展开代表性主流程。
5. `../tests/UCxx-test-cases.md`：定义 UNIT、INT、E2E 测试及关键断言。
6. `../traceability/UCxx-traceability.md`：连接需求、模型、代码和测试结果。

测试未执行时写“待执行”，不得预填“通过”。每张 `.puml` 必须同时提交 `.svg` 和 `.png`。

## GitHub Projects

Issue 状态统一为 `Backlog → Ready → In Progress → Review → Done`，另设 `Blocked`。每个 Issue 必须包含负责人、分支、用例编号、验收条件、PR 和测试/文档证据；证据不齐全不得进入 `Done`。
