# 成员 B：四个用例顺序图索引

本文件只作为索引。正式图统一使用 PlantUML，源文件位于 `docs/models/`，不再在 Markdown 中嵌入 Mermaid 图。

- [四个用例总览图](docs/models/usecase/USECASE-ALL.puml)
- [四个用例领域类图](docs/models/class/CLASS-ALL.puml)

## UC01 用户注册、登录与资料维护

追溯编号：`REQ01 / UC01 / SYS-SEQ01 / COMP-SEQ01 / OBJ-SEQ01`

- [用例说明](docs/models/usecase/UC01.md)
- [系统级顺序图](docs/models/system/SYS-SEQ01.puml)
- [组件级顺序图](docs/models/component/COMP-SEQ01.puml)
- [对象级顺序图](docs/models/object/OBJ-SEQ01.puml)

## UC07 关注关系、分组、屏蔽与内容通知

追溯编号：`REQ07 / UC07 / SYS-SEQ07 / COMP-SEQ07 / OBJ-SEQ07`

- [用例说明](docs/models/usecase/UC07.md)
- [系统级顺序图](docs/models/system/SYS-SEQ07.puml)
- [组件级顺序图](docs/models/component/COMP-SEQ07.puml)
- [对象级顺序图](docs/models/object/OBJ-SEQ07.puml)

## UC08 创作者会员订阅

追溯编号：`REQ08 / UC08 / SYS-SEQ08 / COMP-SEQ08 / OBJ-SEQ08`

- [用例说明](docs/models/usecase/UC08.md)
- [系统级顺序图](docs/models/system/SYS-SEQ08.puml)
- [组件级顺序图](docs/models/component/COMP-SEQ08.puml)
- [对象级顺序图](docs/models/object/OBJ-SEQ08.puml)

## UC11 用户私信与媒体分享

追溯编号：`REQ11 / UC11 / SYS-SEQ11 / COMP-SEQ11 / OBJ-SEQ11`

- [用例说明](docs/models/usecase/UC11.md)
- [系统级顺序图](docs/models/system/SYS-SEQ11.puml)
- [组件级顺序图](docs/models/component/COMP-SEQ11.puml)
- [对象级顺序图](docs/models/object/OBJ-SEQ11.puml)

## 图纸规范

- 所有正式图使用 PlantUML，并导出同名 SVG/PNG。
- 编号统一为 `REQxx / UCxx / SYS-SEQxx / COMP-SEQxx / OBJ-SEQxx / UNIT-TCxx / INT-TCxx / E2E-TCxx`。
- GitHub Projects 状态统一为 `Backlog → Ready → In Progress → Review → Done`，另设 `Blocked`。
- Issue 必须包含负责人、分支、用例编号、验收条件、PR 和测试/文档证据；证据齐全后才可进入 `Done`。
