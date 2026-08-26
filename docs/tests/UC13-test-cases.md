# UC13 测试用例设计

> 状态说明：本文只定义测试范围和断言。未执行的用例统一标记为“待实现/待执行”，不能提前填写通过。

## 1. 测试范围

- 视频审核与弹幕屏蔽。
- 用户角色查询和修改。
- 横幅、公告的查询与增删改。
- 存储、流量、CPU、在线和直播指标查询。
- 普通用户、审核员、管理员之间的权限边界。

## 2. 单元测试（UNIT-TC13）

| 编号 | 测试目标 | 输入/条件 | 关键断言 | 当前状态 |
|---|---|---|---|---|
| UNIT-TC13-01 | AdminMiddleware 权限 | role=user/moderator/admin | 仅 admin 放行，其余返回 403 | 待实现 |
| UNIT-TC13-02 | StaffMiddleware 权限 | role=user/moderator/admin | moderator/admin 放行，user 返回 403 | 待实现 |
| UNIT-TC13-03 | 用户角色参数校验 | user/moderator/admin/非法值 | 三种合法值通过，非法值返回 400 | 待实现 |
| UNIT-TC13-04 | 横幅参数校验 | 空标题/合法标题 | 空标题返回 400 | 待实现 |
| UNIT-TC13-05 | 公告参数校验 | 空内容、非法时间、合法时间 | 对应返回 400 或构造成功 | 待实现 |
| UNIT-TC13-06 | 审核状态校验 | 合法/非法视频状态 | 非法状态返回 400 | 待实现 |

## 3. API/集成测试（INT-TC13）

| 编号 | 接口/场景 | 前置条件 | 关键断言 | 当前状态 |
|---|---|---|---|---|
| INT-TC13-01 | GET /admin/videos | moderator token | 200，返回分页待审视频 | 待执行 |
| INT-TC13-02 | PUT /admin/videos/:id/status | 待审且媒体就绪 | 审核状态持久化；公开查询结果同步变化 | 待执行 |
| INT-TC13-03 | PUT /admin/videos/:id/status | 视频未转码/不存在 | 返回 403/404，状态不被错误修改 | 待执行 |
| INT-TC13-04 | PUT /admin/danmaku/:id/block | moderator token | 200，blocked=true | 待执行 |
| INT-TC13-05 | PUT /admin/users/:id/role | admin token | 200，数据库角色与响应一致 | 待执行 |
| INT-TC13-06 | PUT /admin/users/:id/role | moderator/user token | 403，数据库不变 | 待执行 |
| INT-TC13-07 | Banner CRUD | admin token | 创建、更新、查询、删除结果一致 | 待执行 |
| INT-TC13-08 | Announcement CRUD | admin token | 内容和时间持久化正确 | 待执行 |
| INT-TC13-09 | GET /admin/infrastructure | admin token | 返回 storage/traffic/cpu/online 及来源 | 待执行 |
| INT-TC13-10 | 任意后台接口 | 无 token/无效 token | 返回 401，无数据修改 | 待执行 |

## 4. 端到端测试（E2E-TC13）

| 编号 | 端到端场景 | 操作步骤摘要 | 通过标准 | 当前状态 |
|---|---|---|---|---|
| E2E-TC13-01 | 审核员处理内容 | 登录审核员→审核视频→屏蔽弹幕→刷新页面 | 状态保存且普通页面结果同步 | 待执行 |
| E2E-TC13-02 | 管理员修改角色 | 登录管理员→搜索用户→修改角色→刷新 | 新角色保持，权限随之生效 | 待执行 |
| E2E-TC13-03 | 管理员维护运营内容 | 新建/修改/删除横幅和公告 | 后台与展示端结果一致 | 待执行 |
| E2E-TC13-04 | 管理员查看基础设施 | 进入监控页→刷新指标 | 指标、单位、阈值状态和来源可见 | 待执行 |
| E2E-TC13-05 | 普通用户越权 | 普通用户访问后台页面/API | 页面拦截或接口返回 403，无数据修改 | 待执行 |

## 5. 测试证据要求

- 保存测试命令、环境版本、测试数据初始化方式和原始输出。
- 每个失败结果关联 GitHub Issue，修复后保留重新执行结果。
- CI 中的失败测试必须阻断镜像和部署步骤。
- 测试执行后再把“当前状态”改为通过或失败，并在追溯表填写报告链接。
