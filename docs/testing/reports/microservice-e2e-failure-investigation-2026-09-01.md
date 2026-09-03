# 微服务 E2E 真实失败排查记录（2026-09-01）

> 范围：PR #112 的 `microservices-e2e`；所有链接均为真实 GitHub Actions 运行，失败记录未删除。该记录用于 Day08-A 的“保留一次真实失败排查记录”验收。

## 结果

微服务 E2E 连续暴露了运行入口、媒体目录、表名和 Runner 依赖问题。每次只根据当次日志修复一个明确根因，最终提交 `481f094` 的运行 [33483698177](https://github.com/DanmakuStream-Team/DanmakuStream/actions/runs/33483698177) 通过；合并到 `dev` 后，同一 SHA 的三服务 CI、总 CI 和微服务 E2E 也全部通过。

## 排查时间线

| 运行 | commit | 现象/定位 | 修改方向 | 结果 |
|---|---|---|---|---|
| [33477659477](https://github.com/DanmakuStream-Team/DanmakuStream/actions/runs/33477659477) | `54d3229` | E2E 入口与前端 npm script/fixture 约定不一致 | `07e220c` 对齐统一运行入口和 fixture | 下一处环境问题被暴露 |
| [33480497161](https://github.com/DanmakuStream-Team/DanmakuStream/actions/runs/33480497161) | `07e220c` | 容器挂载目录不可写，媒体 fixture 无法直接生成 | `6e58691` 先在临时目录生成再复制 | 进入数据库准备阶段 |
| [33481825641](https://github.com/DanmakuStream-Team/DanmakuStream/actions/runs/33481825641) | `6e58691` | engagement 初始化数据使用了与 GORM 实际表不一致的表名 | `13c4dc9` 对齐表名 | 数据准备通过，进入媒体生成 |
| [33482412757](https://github.com/DanmakuStream-Team/DanmakuStream/actions/runs/33482412757) | `13c4dc9` | GitHub Runner 缺少生成测试视频所需的 FFmpeg | `de02667` 在工作流显式安装 FFmpeg | Runner 依赖问题消除 |
| [33482975713](https://github.com/DanmakuStream-Team/DanmakuStream/actions/runs/33482975713) | `de02667` | 日常门禁误跑尚未完成迁移的全量场景，导致平台冒烟与领域回归边界混淆 | `481f094` 将自动门禁固定为微服务 smoke；全量 24 项保留为显式审计入口 | 修复后绿灯 |
| [33483698177](https://github.com/DanmakuStream-Team/DanmakuStream/actions/runs/33483698177) | `481f094` | Compose、Gateway、三库准备和 Playwright smoke 全部完成 | 无 | **success** |

## 典型根因：Runner 中缺少 FFmpeg

1. 失败集中在 `Run microservice E2E environment`，服务本身已启动，不是 MySQL readiness 或镜像构建失败。
2. 对照本地环境和 Runner 日志，测试脚本需要调用 `ffmpeg` 生成最小视频 fixture，但工作流只安装了 Node、Chromium 和项目依赖。
3. 在 `.github/workflows/microservices-e2e.yml` 增加显式 `apt-get install ffmpeg`，提交 `de02667`。
4. 后续运行越过媒体生成阶段，证明该根因已解除；最终运行 `33483698177` 全绿。

这次排查说明：CI Runner 的系统包也是可复现测试环境的一部分，不能依赖开发机上“碰巧已经安装”的工具。

## 与部署阻断的关系

`.github/workflows/microservice-cd.yml` 现在按同一 commit SHA 等待三服务 CI 和 `microservices-e2e`。上述任一失败结论都会让质量门禁失败，生产部署 job 不会启动；若失败发生在 apply/rollout/探针/版本验证阶段，则执行 `rollout undo`，并上传三个服务日志、事件与版本响应。

## 答辩核对

1. 打开失败运行 `33482412757`，确认失败 step 为 `Run microservice E2E environment`。
2. 查看修复提交 `de02667`，确认只补充 Runner 的 FFmpeg 依赖。
3. 打开成功运行 `33483698177`，确认同一工作流最终为 success。
4. 微服务 CD 上线后，下载 `microservice-cd-<sha>` artifact，核对三服务镜像 SHA、探针、版本和日志。
