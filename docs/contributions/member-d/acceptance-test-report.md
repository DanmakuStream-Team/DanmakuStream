# 成员 D 验收测试记录

## 1. 执行信息

- 最近执行日期：2026-08-27。
- 覆盖范围：UC05、UC09、UC10（开发计划旧编号 UC04、UC07、UC06 分别对应当前三个业务场景）。
- 代码分支：`codex/d-engagement-tests-20260826`；已同步 `origin/dev` 的文档重组基线 `fdc44ad`。
- 运行环境：Windows 11 `10.0.26100`、Go `1.26.2 windows/amd64`、Node.js `24.13.0`、npm `11.6.2`、Gin Handler、GORM、本机 MySQL（TCP 3306）。
- 数据隔离：测试运行时创建名称为 `danmakustream_d_sxh_<时间戳>` 的临时数据库，完成后自动删除，不使用开发业务库。

## 2. 本次测试汇总

| 层级 | 执行数 | 通过 | 失败 | 未执行 | 结论 |
| --- | ---: | ---: | ---: | ---: | --- |
| 单元测试场景 | 11 | 11 | 0 | 0 | 通过 |
| 集成/API 场景组 | 10 | 10 | 0 | 0 | 通过 |
| 前端类型检查与生产构建 | 1 | 1 | 0 | 0 | 通过，有既有构建警告 |
| FFmpeg—RTMP—SRS—HLS 媒体 E2E | 1 | 1 | 0 | 0 | 通过 |
| Playwright 浏览器 E2E | 3 | 3 | 0 | 0 | Chromium 通过 |
| **已执行合计** | **26** | **26** | **0** | **0** | **有效测试全部通过** |

单元测试按实际执行叶子场景统计：Super Chat 五个价值边界，加上热度、时间解析、计划状态、观众去重、Hub 事件和持久化失败六个场景。集成/API 按三个业务用例、三个数据库故障组，以及计划启动、计划通知、聊天权限和真实 WebSocket 往返四组统计。断网重连仍保留为现场增强检查，不计入上述自动化场景。

## 3. 自动化执行结果

| 证据 | 覆盖内容 | 结果 |
| --- | --- | --- |
| `go test -count=1 -v ./...` | 后端全量单元测试和普通包编译 | 通过 |
| `TestSuperChatDisplaySeconds` | 49/50/200/500/1000 五个价值边界 | 5/5 通过 |
| `TestRoomHeat` | 在线人数、点赞和礼物价值的热度计算 | 通过 |
| `TestParseScheduleTime` / `TestIsValidScheduleStatus` | 合法/非法时间及计划状态规则 | 2/2 通过 |
| `TestCountRoomViewersDeduplicatesAuthenticatedUsers` | 登录用户多连接去重、匿名连接计数、监控连接排除 | 通过 |
| `TestBroadcastEventAndSendEvent` | Hub 事件序列化、发送及满通道非阻塞 | 通过 |
| `TestSaveDanmakuPropagatesPersistenceFailure` | 实时弹幕数据库错误向上返回，广播前置门控 | 通过 |
| `TestEngagementUseCasesWithMySQL/UC05_*` | 弹幕发送/列表、评论树/排序/回复/删除、评论点赞切换、视频点赞/收藏切换、未审核/未登录/非法 ID/不存在资源 | 通过 |
| `TestEngagementUseCasesWithMySQL/UC09_*` | 列表与预约状态、关注者通知、时间/标题校验、同主播同时间冲突、预约切换、本人预约、普通用户/管理员取消及不存在资源 | 通过 |
| `TestEngagementUseCasesWithMySQL/UC10_*` | 列表/详情/主播管理、创建/复用/结束后重启、聊天设置边界与越权、互动汇总/监控/热度榜、点赞/礼物/SC、越权下播及异常资源 | 通过 |
| `TestProcessDueLiveSchedulesIsIdempotent` | 到期计划重复扫描只创建一个直播间、一条通知和一份主播统计；取消计划不启动 | 通过 |
| `TestScheduleNotificationEmptyAndSelfBranches` | 无预约、本人预约和本人关注时不产生错误通知 | 通过 |
| `TestCanSendChatModesAndSlowMode` | everyone/followers/members 身份校验、会员有效期与慢速模式，主播绕过限速 | 通过 |
| `TestHubWebSocketRoundTripAndPersistenceFailure` | 真实 WebSocket 注册、在线人数、弹幕往返、默认样式和数据库错误事件 | 通过 |
| `TestEngagementDatabaseFailureResponses` | UC05/UC09/UC10 依赖表不可用时返回 500，且不伪报成功 | 3/3 通过 |
| `npm run test:e2e:d` | UC05 弹幕、评论、点赞、收藏及刷新持久化；UC09 页面创建预约并由另一用户预约/取消；UC10 创建直播、点赞、赠礼与结束 | Chromium 3/3 通过 |
| `npm run build` | Vue TypeScript 检查与生产构建 | 通过；存在既有包体积提示，无编译错误 |
| FFmpeg → RTMP → SRS → HLS | 30 秒测试音视频推流；读取主清单、媒体清单和首个 TS 分片 | 通过；三级 HTTP 均为 200，分片 162620 字节 |

集成测试日志中的 `record not found` 是首次点赞、收藏、预约和直播点赞前查询“关系尚不存在”的预期分支，测试最终退出码为 0，不属于失败。

## 4. 覆盖率

| 统计口径 | 覆盖语句 | 总语句 | 覆盖率 |
| --- | ---: | ---: | ---: |
| 用户 D 用例函数口径 | 924 | 1023 | **90.3%** |
| 用户 D 相关文件保守口径 | 940 | 1618 | **58.1%** |

用例函数口径只统计 UC05、UC09、UC10 实际涉及的评论、普通视频弹幕、视频点赞/收藏、计划预约、直播管理/互动和 WebSocket Hub 函数。保守口径按这些函数所在的完整文件统计，因此还包含视频上传/转码/后台管理、高级弹幕/审核和已从验收范围删除的回放查询等非成员 D 用例代码。覆盖率是补充证据，不能代替业务场景、异常流程和 E2E 验收。

## 5. 执行过程中的环境问题

| 阶段 | 现象 | 判定与处理 |
| --- | --- | --- |
| 首次 Go 测试启动 | 沙箱不能访问用户级 Go 构建缓存 | 尚未进入测试代码，不计业务失败；允许缓存访问后原命令重跑通过 |
| 首次 MySQL 连接 | `root/password` 返回 MySQL 1045 | 使用中的本机数据库与 Compose 示例密码不同；改用项目运行配置后，在隔离测试库重跑通过 |
| 首次媒体轮询 | 已成功推流，但脚本未发现 `.ts` | SRS 返回两级 HLS 清单；修正为“主清单→媒体清单→分片”后复验通过，不计产品失败 |

## 6. 可重复执行方法

普通测试：

```powershell
cd backend
go test -count=1 -v ./...
```

MySQL 集成测试需要提供有权创建临时数据库的测试账号：

```powershell
$env:DANMAKU_TEST_ADMIN_DSN='<测试管理员DSN>'
go test -count=1 -tags=integration ./internal/handler/v1/live ./internal/logic/danmaku ./integration -v
Remove-Item Env:DANMAKU_TEST_ADMIN_DSN
```

## 7. E2E 状态与现场检查项

自动化集成测试已经从 HTTP Handler 入口覆盖三个用例的业务写入、权限、异常分支和数据库结果；真实 RTMP 推流与 SRS/HLS 输出也已通过。Playwright 已自动完成 UC05 的弹幕、评论、点赞、收藏与刷新持久化，UC09 的创建、预约、取消与刷新持久化，以及 UC10 的直播创建、点赞、赠礼和结束流程。HTML 报告位于 `docs/testing/reports/engagement-e2e/index.html`。

最终现场仍建议补充两项视觉证据：

1. UC05：自动化流程已通过；最终汇报时保存一张代表性的成功页面截图。
2. UC10：打开同一账号多个标签页确认在线人数不重复；断网恢复后确认自动重连。数据库写入失败门控、真实 WebSocket 往返、聊天权限和媒体链路已有自动化证据。
