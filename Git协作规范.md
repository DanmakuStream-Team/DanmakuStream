# Git 协作规范

本文档用于 DanmakuStream 团队开发协作。所有成员开发前请先阅读，避免分支混乱、覆盖他人代码或把未完成内容合入稳定分支。

## 分支说明

项目采用以下分支模型：

```text
main        稳定版本分支，用于阶段性交付、演示和验收
dev         集成开发分支，所有功能完成自测后先合入这里
feature/*   功能开发分支，每个任务单独创建
fix/*       缺陷修复分支
docs/*      文档修改分支
```

不要直接在 `main` 上开发。一般也不要直接在 `dev` 上写功能代码。

## 开发流程

开始一个新任务时，从最新的 `dev` 创建自己的功能分支：

```bash
git switch dev
git pull origin dev
git switch -c feature/s1-be-video-upload
```

开发完成后提交并推送：

```bash
git status
git add .
git commit -m "feat: implement video upload api"
git push -u origin feature/s1-be-video-upload
```

然后在 GitHub 上创建 Pull Request：

```text
base: dev
compare: feature/s1-be-video-upload
```

PR 通过 review 后再合入 `dev`。`dev` 稳定后，再由负责人发起 PR 合入 `main`。

## 分支命名

分支名使用小写英文、数字和短横线，表达清楚任务内容。

推荐格式：

```text
feature/任务编号-简短说明
fix/问题编号-简短说明
docs/文档说明
```

示例：

```text
feature/s1-be-video-list
feature/s1-be-video-detail
feature/s1-be-video-upload
feature/s1-fe-video-page
fix/login-token-expire
docs/api-design
```

## 提交信息规范

提交信息使用：

```text
type: summary
```

常用 type：

```text
feat      新功能
fix       修复问题
docs      文档修改
style     代码格式调整，不影响逻辑
refactor  重构，不新增功能也不修 bug
test      测试相关
chore     构建、脚本、依赖等杂项
```

示例：

```bash
git commit -m "feat: implement video list api"
git commit -m "fix: correct video detail view count"
git commit -m "docs: add git workflow guide"
git commit -m "test: add core api smoke script"
```

## 合并规则

`feature/*` 合入 `dev` 前，需要满足：

```text
1. 功能已完成，不是半成品
2. 本地能正常编译或启动
3. 相关接口已手动测试或脚本测试
4. 没有提交本地临时文件、密钥、测试视频等无关文件
5. PR 描述里写清楚改了什么、怎么测试
```

`dev` 合入 `main` 前，需要满足：

```text
1. dev 当前版本可以完整启动
2. 核心功能通过冒烟测试
3. 没有明显阻塞级 bug
4. 团队确认可以作为阶段性交付版本
```

## Pull Request 规范

PR 标题建议和 commit 类似：

```text
feat: implement video detail api
fix: handle invalid video id
docs: update development plan
```

PR 描述建议包含：

```text
改动内容：
- 实现 GET /api/v1/videos/:id
- 增加 view_count 自增
- 增加接口测试脚本

测试方式：
- GOCACHE=/tmp/go-build go test ./...
- ./scripts/test_core_apis.sh

注意事项：
- 该接口只返回 status=approved 的视频
```

## 本地测试建议

后端开发前先启动基础设施：

```bash
docker compose up mysql redis minio -d
```

启动后端：

```bash
./scripts/run_backend_dev.sh
```

运行核心接口测试：

```bash
./scripts/test_core_apis.sh
```

也可以只做 Go 编译检查：

```bash
cd backend
GOCACHE=/tmp/go-build go test ./...
```

## 不要提交的内容

以下内容不要提交到仓库：

```text
.env
node_modules/
dist/
本地测试视频和封面
临时日志
IDE 私人配置
密钥、Token、数据库密码
```

测试视频建议放在本地目录，例如：

```text
local-test-assets/
```

并加入 `.gitignore`。

## 常用命令

查看当前状态：

```bash
git status --short --branch
```

同步 `dev`：

```bash
git switch dev
git pull origin dev
```

创建功能分支：

```bash
git switch -c feature/your-task-name
```

推送当前分支：

```bash
git push -u origin HEAD
```

将远端 feature 分支合入本地 `dev`：

```bash
git switch dev
git pull origin dev
git merge origin/feature/branch-name
git push origin dev
```

## 冲突处理原则

遇到冲突时，不要直接删除看不懂的代码。先确认冲突内容是谁改的、为什么改。

处理流程：

```bash
git status
```

打开冲突文件，找到：

```text
<<<<<<< HEAD
当前分支内容
=======
被合并分支内容
>>>>>>> branch-name
```

手动保留正确内容后：

```bash
git add 冲突文件
git commit
```

如果不确定如何处理，先在群里说明冲突文件和冲突位置，再一起决定。

## 团队约定

```text
1. 不直接 push 到 main
2. 功能开发从 dev 拉 feature 分支
3. 每个 PR 尽量只做一个任务
4. 合并前先自测
5. 不提交无关格式化和大范围重构
6. 不覆盖他人代码，不确定先沟通
```

