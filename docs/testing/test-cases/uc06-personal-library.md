# UC06 测试用例与执行结果

用例说明：[UC06 个人视频资料库](../../contributions/member-e/uc06-personal-library.md)。
执行证据：[UC01/UC06 测试报告（2026-08-29）](../reports/uc01-uc06-test-report-20260829.md)。

## 1. 测试范围

- 服务端观看历史、播放进度与稍后再看的持久化。
- 鉴权、非法参数、未审核视频、用户数据隔离、单项删除和整库清空。
- 前端历史/稍后再看页面的展示、删除、清空和未登录重定向。

## 2. 单元测试（UNIT-TC06）

实现文件：`backend/internal/handler/v1/user/library_handler_test.go`、
`backend/internal/handler/v1/collection/collection_handler_test.go`。

| 编号 | 测试目标 | 关键断言 | 当前状态 |
|---|---|---|---|
| UNIT-TC06-01 | 分页参数归一化 | 默认值、负数、非法数字、超限 pageSize 均得到安全值 | 通过 |
| UNIT-TC06-02 | 视频 ID 校验 | 正整数通过，0/负数/非数字返回 400 | 通过 |
| UNIT-TC06-03 | 历史响应映射 | 视频/作者/保存时间正确映射 | 通过 |
| UNIT-TC06-04 | 进度计算 | 正常比例取整，超时长钳制到 100%，零时长为 0 | 通过 |
| UNIT-TC06-05 | 合集 ID 与响应映射 | 非法合集 ID 返回 400，合集所有者字段正确 | 通过 |

## 3. API/集成测试（INT-TC06）

实现文件：`tests/api/uc06-library-test.sh`，真实启动后端与 MySQL 执行 30 项断言。

| 编号 | 场景 | 关键断言 | 当前状态 |
|---|---|---|---|
| INT-TC06-01～02 | 鉴权与参数异常 | 无 token 401；非法 ID、负位置 400 | 通过 |
| INT-TC06-03～05 | 保存与读取观看历史 | DB 持久化、进度计算、超时长钳制、未审核视频拒绝 | 通过 |
| INT-TC06-06～07 | 历史隔离与删除 | 用户间隔离；单项删除和清空后无记录 | 通过 |
| INT-TC06-08～10 | 稍后再看状态与列表 | 状态切换、列表包含目标视频、用户间隔离 | 通过 |
| INT-TC06-11～12 | 稍后再看异常与删除 | 未审核视频拒绝；单项删除和清空成功 | 通过 |

## 4. E2E 测试（E2E-TC06）

实现文件：`frontend/e2e/uc06-library.spec.ts`。

| 编号 | 端到端场景 | 通过标准 | 当前状态 |
|---|---|---|---|
| E2E-TC06-01 | 历史展示与单项删除 | 保存的进度在历史页可见，可继续观看并从 UI 删除 | 通过 |
| E2E-TC06-02 | 清空观看历史 | 二次确认后页面与服务端列表均为空 | 通过 |
| E2E-TC06-03 | 稍后再看闭环 | 添加后跨页面可见，状态为 saved，UI 清空后服务端为空 | 通过 |
| E2E-TC06-04 | 未登录访问 | 自动跳转登录页并携带 redirect 参数 | 通过 |

## 5. CI 与证据

- 单元测试：`backend-test` 和 `test-uc06` Job。
- API：`api-test` Job 执行 `tests/api/uc06-library-test.sh`。
- E2E：`e2e-uc01-uc06` Job 执行 `npm run test:e2e:uc01-uc06`。
- 三层任一失败，`docker-build` 不执行；CI 上传 API 输出、Playwright HTML、trace、截图和视频。
