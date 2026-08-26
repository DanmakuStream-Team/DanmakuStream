# 模型图目录与编号

正式模型图统一使用 PlantUML，源文件扩展名为 `.puml`。导出文件与源文件同名，使用 `.svg` 和 `.png` 扩展名。

## 目录

- `usecase/`：用例说明和需求追溯 Markdown。
- `usecase/USECASE-ALL.puml`：四个用例的总用例图。
- `class/CLASS-ALL.puml`：四个用例共享的领域类图。
- `system/`：需求说明书层面的系统级顺序图。
- `component/`：概要设计层面的组件级顺序图。
- `object/`：详细设计层面的对象级顺序图。
- `deployment/`：部署图，包括改造前单体和改造后 Kubernetes。

## 编号

统一采用：`REQxx / UCxx / SYS-SEQxx / COMP-SEQxx / OBJ-SEQxx / UNIT-TCxx / INT-TCxx / E2E-TCxx`。

## 导出

在仓库根目录执行：

```powershell
plantuml -tsvg docs/models/**/*.puml
plantuml -tpng docs/models/**/*.puml
```

若 PowerShell 未展开递归通配符，可使用：

```powershell
Get-ChildItem docs/models -Recurse -Filter *.puml | ForEach-Object { plantuml -tsvg $_.FullName }
Get-ChildItem docs/models -Recurse -Filter *.puml | ForEach-Object { plantuml -tpng $_.FullName }
```

## GitHub Projects

Issue 状态统一为 `Backlog → Ready → In Progress → Review → Done`，另设 `Blocked`。每个 Issue 必须包含负责人、分支、用例编号、验收条件、PR 和测试/文档证据；证据不齐全不得进入 `Done`。
