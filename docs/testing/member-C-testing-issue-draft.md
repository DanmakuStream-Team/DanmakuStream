# GitHub Issue 创建草稿

## 创建方法

1. 打开仓库 `Issues`，选择 `New issue` → `业务用例任务`。
2. 使用下面的标题和正文，负责人替换为自己的 GitHub 用户名。
3. 将 Issue 加入组织项目 `Danmaku project`，开始处理后把状态设为 `In Progress`。
4. 当前没有远程分支、PR 或 GitHub Actions run，因此相应内容保持“待补”，相关复选框不要勾选。
5. 后续推送、创建 PR 并获得另一名成员检查后，再补链接和勾选完全满足的项目。

## 建议标题

`[成员C][UC02/UC03/UC04/UC12] 内容域自动化测试、报告与 CI`

## 建议正文

```markdown
## 负责人

@你的GitHub用户名

## 分支

`test/member-c-content-tests`（当前仅本地，尚未创建远程分支）

## 用例编号

UC02、UC03、UC04、UC12；关联 #32、#34、#33、#42。

## 任务范围

- 要完成：成员 C 单元、API 集成、E2E 测试，测试报告及三层 CI 阻断。
- 不包含：修复测试发现的业务缺陷、创建远程分支或部署环境。

## 验收条件

- [x] UC02/03/04/12 核心单元测试执行通过。
- [x] API 阻断断言执行通过并生成报告。
- [x] Playwright 四条完整流程执行通过并生成 HTML 报告。
- [x] backend-test、api-test、e2e-member-c 均成为 docker-build 前置任务。
- [ ] 五项已知业务 GAP 已另行确认处理策略。
- [ ] 远程 GitHub Actions 已运行并提供链接。

## Pull Request

待补。

## 测试证据

- 测试设计：`docs/testing/test-cases/member-c-uc02-03-04-12.md`
- 单元报告：`docs/testing/reports/UC02-03-04-12-unit-test-report-20260827.txt`
- API 报告：`docs/testing/reports/UC02-03-04-12-api-test-report-20260827.txt`
- E2E 报告：`docs/testing/reports/UC02-03-04-12-e2e-test-report-20260827.txt`
- HTML：`docs/testing/reports/UC02-03-04-12-e2e-report/index.html`
- 总报告：`docs/testing/reports/UC02-03-04-12-test-report-20260827.md`

## 文档证据

- 三层模型：`docs/models/system/`、`docs/models/component/`、`docs/models/object/`
- 测试追溯：`docs/testing/test-cases/member-c-uc02-03-04-12.md`

## Done 门禁

- [ ] 负责人、远程分支和用例编号准确。
- [ ] 所有验收条件均已完成。
- [ ] PR 已关联并由另一名组员检查。
- [ ] PR 已合并到目标分支。
- [ ] GitHub Actions 三层测试全部通过。
- [ ] 已知 GAP 均已修复或形成独立 Issue。
```
