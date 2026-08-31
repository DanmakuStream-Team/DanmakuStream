# Issue、PR 与 Done 验收规则

## 1. 每个 Issue 必填内容

每个任务统一使用“业务用例任务”模板，并始终保留以下信息：

| 内容 | 创建 Issue 时 | 进入 Review 前 | 进入 Done 前 |
| --- | --- | --- | --- |
| 负责人 | 必须填写并设置 Assignee | 不得为空 | 与实际完成人一致 |
| 分支 | 必须填写计划分支 | 必须是实际开发分支 | 分支内容已通过 PR 合并 |
| 用例编号 | 必须填写 `UC01`—`UC13` | 与代码、设计和测试一致 | 已写入追溯记录 |
| 验收条件 | 必须写成可勾选、可验证的结果 | 已逐项验证 | 全部勾选 |
| PR | 可暂写“待补” | 必须填写编号或链接 | 已由另一名组员 Review 并合并 |
| 测试证据 | 可暂写“待补” | 必须给出测试文件、命令和结果 | 单元、集成/API、E2E 证据无未解释失败 |
| 文档证据 | 可暂写“待补” | 必须给出仓库路径或链接 | 需求、设计和追溯记录均已更新 |

验收条件不能只写“功能完成”或“测试通过”，应写成可观察结果。例如：

> 同一用户重复预约同一直播计划时，不产生重复记录；取消预约后预约人数减 1，刷新页面后状态保持一致。

## 2. 看板状态流转

| 状态 | 进入条件 | 操作 |
| --- | --- | --- |
| Backlog | 已记录任务，但范围或负责人可能尚未冻结 | 补负责人、用例编号和初步验收条件 |
| Ready | 范围、负责人、分支和验收条件均明确 | Assignee 认领并准备开发 |
| In Progress | 已建立分支并开始修改 | 持续更新 Issue，提交中引用 `#Issue编号` |
| Review | PR 已创建，测试和文档证据已写入 | 邀请另一名组员 Review，修复检查意见 |
| Blocked | 存在无法由负责人独立消除的依赖、权限或环境阻塞 | 在 Issue 写明阻塞原因、责任方和下次检查时间；解除后回到原状态 |
| Done | PR 已合并，全部验收项勾选，证据完整 | 关闭 Issue；关闭后由看板工作流移入 Done |

不要直接把未关闭的卡片拖到 Done。建议在 GitHub Project 的 **Workflows** 中启用：

1. `Item closed` → 将 `Status` 设为 `Done`。
2. `Pull request merged` → 不直接设为 Done，只保留或进入 `Review`，等待 Issue 证据核对。
3. `Item reopened` → 将 `Status` 设为 `In Progress`。

仓库的 `Issue Done Evidence Gate` 会在 Issue 被关闭时检查负责人、分支、用例、PR、测试/文档证据和全部复选框。证据不完整时会自动重新打开 Issue，因此需要在仓库 **Settings → Actions → General → Workflow permissions** 中允许 **Read and write permissions**。

## 3. 分支、提交和 PR

- 一个 Issue 原则上对应一个短期分支，例如 `feature/sxh-uc10-live-chat-20260826`。
- 提交信息建议为 `type: summary (#Issue编号)`，例如 `test: cover live chat failure paths (#39)`。
- PR 正文使用仓库模板，并通过 `Closes #Issue编号` 建立自动关联。
- PR 的 base 使用团队当前约定的集成分支；未经团队确认不要直接以 `main` 为 base。
- 重要代码必须由负责人之外的另一名组员批准后合并。

## 4. 测试与文档证据写法

测试证据至少写清：测试层级、文件或场景、执行命令、通过/失败数量、运行环境和结果链接。截图或录屏应放入团队约定的证据目录，或上传到 Issue/PR 后粘贴链接。

文档证据使用仓库内可点击路径，至少能追溯到：用例说明、系统级图、组件级图、对象级图、代码模块、单元测试、集成/API 测试和 E2E 测试。某一测试层级确实不适用时，必须写明原因，不能留空。

## 5. 现有 Issue 的调整方法

对已经创建的 Issue，无需删除重建：编辑正文，复制新模板的七个栏目并补齐；在右侧设置 **Assignees**；将分支名、PR 编号和证据路径更新为实际内容。核对完成前保持在 Backlog、Ready、In Progress、Review 或 Blocked，全部满足后再关闭。

项目看板统一状态为：`Backlog → Ready → In Progress → Review → Done`，并另设 `Blocked`。状态名称和大小写应保持一致。
