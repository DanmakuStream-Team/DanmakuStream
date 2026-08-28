# CI 红灯/绿灯阻断证据（2026-08-27）

> 对应课程任务书要求："保留两条真实流水线证据——一次测试失败且后续镜像/部署未执行；修复后全部通过。"
> 采集方式：PR [#70](https://github.com/DanmakuStream-Team/DanmakuStream/pull/70)（分支 `codex/day3-traceability-ci-proof`），依次推送绿灯提交 → 故意失败提交 → 回滚提交，全部为真实 GitHub Actions 运行，无人工编辑。

## 时间线（三次运行）

| # | commit | 内容 | 运行 | 结论 |
|---|---|---|---|---|
| 1 | `52281df` | 总追溯表（正常代码） | [33032495917](https://github.com/DanmakuStream-Team/DanmakuStream/actions/runs/33032495917) | ✅ success（基线绿灯） |
| 2a | `3a7e14b` | 加入故意失败的单元测试（未修复 CI 前） | [33032768023](https://github.com/DanmakuStream-Team/DanmakuStream/actions/runs/33032768023) | ⚠️ **success（暴露缺陷，见下）** |
| 2b | `a80130f` | CI pipefail 修复 + 仍保留失败测试 | [33033392243](https://github.com/DanmakuStream-Team/DanmakuStream/actions/runs/33033392243) | 🔴 **failure（有效红灯）** |
| 3 | `fec5e71` | 回滚失败测试 | [33033631085](https://github.com/DanmakuStream-Team/DanmakuStream/actions/runs/33033631085) | ✅ success（修复后绿灯） |

## 红灯运行详情（run 33033392243）

Job 级结论（API：`/actions/runs/33033392243/jobs`）：

| Job | 结论 |
|---|---|
| 前端依赖安装与构建 | success（并行独立，不受影响） |
| **后端编译与单元测试** | **failure** |
| API/集成测试（INT-TC13） | **skipped**（依赖 backend-test） |
| **构建前后端镜像** | **skipped**（依赖全部测试通过） |

失败日志摘录（job log，`##[error]` 前原始输出）：

```text
=== RUN   TestRedLightEvidenceDeliberateFailure
    red_light_evidence_test.go:9: 故意失败：验证 CI 在单元测试失败时停止后续 api-test 与 docker-build（RED-LIGHT EVIDENCE）
--- FAIL: TestRedLightEvidenceDeliberateFailure (0.00s)
FAIL	danmakustream/backend/internal/middleware	0.005s
FAIL
##[error]Process completed with exit code 1.
```

**结论：单元测试失败 → 镜像构建（docker-build）被跳过，未产出任何镜像。阻断规则生效。**

## 演练过程中发现并修复的真实缺陷（重要）

第一次红灯提交（`3a7e14b`，run 33032768023）暴露：`go test ./... -v | tee unit-test-output.txt` 中 **`tee` 吞掉了 go test 的非零退出码**——日志里明确出现 `FAIL danmakustream/backend/internal/middleware`，但该步骤与整个 job 仍显示 success，四个 job 全部放行（含镜像构建）。也就是说：**修复前，CI 并不能真正阻断失败测试，违反任务书红线。**

排查过程：对比 job 结论（success）与 job 日志（含 `--- FAIL`）→ 定位到管道退出码问题 → 修复：为 `Unit tests` 与 `Run API test suite` 两个步骤显式加 `set -o pipefail`（commit `a80130f`）。

修复验证即上表 2b：同样一个失败测试，修复后 backend-test 正确 failure，下游全部 skipped。**这也直接满足任务书"现场说明一次部署失败是怎么查出来的"的答辩素材要求。**

## 复现方式（教师/助教核对）

1. 打开 PR #70 的 Checks 或上表运行链接，核对 2b 运行的 job 矩阵与日志。
2. 或自行复演：`git checkout a80130f && git push origin HEAD:tmp-red-drill`（会触发同样红灯；勿合入）。

## 遗留说明

- `3a7e14b` 那次"失败但显示 success"的运行保留不删，作为缺陷存在的原始证据；修复提交 `a80130f` 的说明中已注明原因。
- 红灯测试文件已在 `fec5e71` 回滚，`dev`/`main` 不含任何故意失败代码。
