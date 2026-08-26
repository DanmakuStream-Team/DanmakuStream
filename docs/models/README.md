# 用例三层模型编写样例

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
