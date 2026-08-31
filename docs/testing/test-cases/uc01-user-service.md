# UC01 测试用例与执行结果

执行证据：[UC01/UC06 测试报告（2026-08-29）](../reports/uc01-uc06-test-report-20260829.md)。

## 单元测试（UNIT-TC01）

实现文件：`backend/internal/logic/auth/auth_logic_test.go`。

| 编号 | 测试目标 | 关键断言 | 当前状态 |
|---|---|---|---|
| UNIT-TC01-01 | 注册输入校验 | 空昵称、空密码在访问数据库前被拒绝 | 通过 |
| UNIT-TC01-02 | 登录输入校验 | 缺少用户标识时统一返回认证失败 | 通过 |
| UNIT-TC01-03 | JWT 签发与解析 | claims 含 userId、username、role、iat、exp；错误密钥不能验签 | 通过 |

密码哈希、数据库唯一性及资料持久化由现有 MySQL 集成测试
`backend/integration/member_b_integration_test.go` 覆盖，不重复伪造数据库单元测试。

## API/集成测试（INT-TC01）

| 编号 | 接口/场景 | 关键断言 | 当前状态 |
|---|---|---|---|
| INT-TC01-01 | POST `/auth/register` | 注册成功、密码不以明文保存、返回 token | 已由成员 B 集成测试覆盖 |
| INT-TC01-02 | 重复注册 | 返回业务错误且不产生重复用户 | 已由成员 B 集成测试覆盖 |
| INT-TC01-03 | POST `/auth/login` | 正确凭据返回 JWT，错误密码被拒绝 | 已由成员 B 集成测试覆盖 |
| INT-TC01-04 | GET `/auth/me` | 有效 token 返回当前用户，无效 token 返回 401 | 已由成员 B 集成测试覆盖 |
| INT-TC01-05 | PUT `/users/me` | 简介更新持久化且只修改当前用户 | 已由成员 B 集成测试覆盖 |

## E2E 测试（E2E-TC01）

实现文件：`frontend/e2e/uc01-user.spec.ts`。

| 编号 | 场景 | 通过标准 | 当前状态 |
|---|---|---|---|
| E2E-TC01-01 | 注册成功 | 跳转首页，token/userInfo 写入本地存储，`/auth/me` 可用 | 通过 |
| E2E-TC01-02 | 重复昵称 | 页面显示错误，停留注册页且不建立会话 | 通过 |
| E2E-TC01-03 | 登录与会话持久化 | 登录后刷新 token 保持且 `/auth/me` 返回当前用户 | 通过 |
| E2E-TC01-04 | 错误密码 | 页面显示错误，停留登录页且不建立会话 | 通过 |
| E2E-TC01-05 | 编辑个人简介 | 页面更新成功且接口查询结果持久化 | 通过 |

CI 入口：统一 `e2e` Job 执行 `npm run test:e2e`（全部用例），失败阻断镜像构建。
