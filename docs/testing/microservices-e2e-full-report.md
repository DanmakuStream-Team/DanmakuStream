# 微服务版 13/13 E2E 测试报告

> 生成方式：`MICRO_E2E_FULL_SUITE=1 bash scripts/run-microservices-e2e.sh`
> 报告位置：`artifacts/microservices-e2e/`（含 compose.log / compose-ps.txt / playwright-report-full / junit-microservices-full.xml）
> 失败保留现场：`MICRO_E2E_KEEP_STACK=1 MICRO_E2E_FULL_SUITE=1 bash scripts/run-microservices-e2e.sh`

---

## 0. 环境信息

| 项 | 值 | 实际运行填入 |
|---|---|---|
| 前端端口 | `MICRO_FRONTEND_PORT=18080` |  |
| 网关端口 | `MICRO_GATEWAY_PORT=18888` |  |
| RTMP / HLS 端口 | `19350 / 18081` |  |
| Compose 栈名 | `danmakustream-e2e` |  |
| Playwright 配置 | `frontend/playwright.microservices-full.config.ts` |  |
| 浏览器 | Chromium (Desktop Chrome) |  |
| 运行模式 | Serial (workers=1)，CI 重试 1 次 |  |
| 执行命令 | `MICRO_E2E_FULL_SUITE=1 bash scripts/run-microservices-e2e.sh` |  |
| 报告生成时间 | YYYY-MM-DD HH:MM |  |

---

## 1. 微服务连通性（基线 2/2）

| 用例 | 说明 | 预期 | 实际 |
|---|---|---|---|
| 网关可用、服务目录完整、内部接口不暴露 | `/gateway/health` + `/api/v1/platform/services` + 404 for `/internal/...` | 网关返回 api-gateway；三服务目录齐全；内部 404 |  |
| 前端经网关注册 + JWT 跨三服务访问 | 注册 → 存 JWT → GET auth/me → videos → users/me/videos → live → users/me/history | 三服务全部 200 / data 非空 |  |

**结果**：基线 2 / 2 通过 = ✅ / ❌

---

## 2. 13 UC 微服务版 E2E 结果

> 代码位置：`frontend/e2e-microservices/ucXX-*.spec.ts`
> 骨架 skip 标记：`test.skip(...)` 表示**因依赖未打通暂不执行**；领域负责人补全后可移除 `.skip`。

| 编号 | 用例名 | 负责人 | 当前状态 | 是否需要跨服务 | 实际结果 | 备注 / 阻塞原因 |
|---|---|---|---|---|---|---|
| UC01 | 注册、登录、资料维护 | 成员 A | ✅ 已实现（非 skip） | 单 user-service |  |  |
| UC02 | 搜索公开视频并进入播放页 | 成员 C | ✅ 已实现（非 skip） | 单 content-service |  |  |
| UC03 | 取消上传 / 重新投稿 / 转码失败 | 成员 C | ⏭️ skip：content-service 上传路径+媒体卷 | content-service + OBS |  | 上传路由 + media volume 需映射 |
| UC04 | 审核员发布 / 拒绝视频 | 成员 C | ✅ 已实现（非 skip） | content-service admin |  |  |
| UC05 | 弹幕 / 评论 / 点赞 / 收藏 | 成员 D | ⏭️ skip：跨服务视频 join | engagement + content |  | engagement 查 content 的 videoId 标题字段 |
| UC06 | 资料库 / 观看历史 / 稍后再看 | 成员 E | ✅ 已实现（非 skip） | 单 engagement-service |  |  |
| UC07 | 关注 / 分组 / 黑名单 / 动态 | 成员 B | ⏭️ skip：动态模块跨服务 | user + content (动态表) |  | 通知中心推送需打通 |
| UC08 | 会员套餐 / 订阅 / 幂等支付 | 成员 B | ⏭️ skip：支付 mock | user-service 订单 + 支付回调 |  | 需支付回调 mock 服务 |
| UC09 | 直播预约 / 提醒 | 成员 D | ⏭️ skip：跨服务作者信息 join | engagement + user-service |  | 预约卡片的作者昵称跨服务 |
| UC10 | 开播 / 点赞赠礼 / 结束直播 | 成员 D | ⏭️ skip：SRS + WS + webhook | engagement + SRS |  | SRS webhook 回调 + 串流推流 |
| UC11 | 私信 / WebSocket / 媒体分享 / 幂等 | 成员 B | ⏭️ skip：媒体上传路径 + WS 网关 | engagement + SRS/媒体卷 |  | /ws/chat 网关路由 + 文件上传映射 |
| UC12 | 创作者数据分析切换 | 成员 C | ⏭️ skip：跨服务 engagement 聚合 | content + engagement 聚合 |  | 观看 / 收藏增长跨表聚合 |
| UC13 | 管理员用户/视频管理 | 成员 E | ⏭️ skip：三服务 admin 路由 | user+content 各自 /admin |  | 网关路由 + 管理员权限边界 |

**已实现非 skip 用例数：5 / 13**（UC01, UC02, UC04, UC06, 基线 2）
**骨架待补全用例数：8 / 13**（UC03, UC05, UC07, UC08, UC09, UC10, UC11, UC12, UC13 = 8）

**结果**：微服务 E2E 全量用例覆盖率 = 实际通过 / 13 = **__ / 13**

---

## 3. 典型失败分析（若有）

### 失败用例 / 错误摘要

| 用例 | 首次失败版本 | 根因 | 修复 commit | 回归结果 |
|---|---|---|---|---|
| （例）UC05 | — | engagement video_id 表缺少 video_id_title 关联 | — | — |

### 未打通 / skip 的阻塞项（需后续迭代）

1. **文件上传与媒体卷共享**：content-service 上传接口与 Nginx 媒体目录需在 Compose 中配置一致的 bind mount（UC03 / UC11）。
2. **跨服务 JOIN 数据聚合**：engagement 查 content 视频标题、UC12 analytics 跨服务统计需补 internal API + 服务间 INTERNAL_API_TOKEN 调用（UC05 / UC12 / UC09）。
3. **SRS Webhook / 推流回传**：UC10 开播后串流 key 生成、直播间 1人观看事件、礼物入账需 SRS → engagement 回传。
4. **支付回调 mock**：UC08 演示支付完成后 demo-pay 端点内部接口需在微服务 user-service 暴露。
5. **管理员后台路由**：UC13 /admin/users 和 /admin/videos 需在 Nginx 网关拆分路由到 user-service / content-service 内部 admin。
6. **内容动态模块**：UC07 发布动态 → 粉丝通知 需 content-service 动态 publish + user-service follow notify 推送。

---

## 4. 修复与行动项

| # | 行动项 | 负责人 | 截止 | 状态 |
|---|---|---|---|---|
| 1 | content-service 上传媒体卷共享配置 + 校验 | 成员 C | | 未开始 |
| 2 | engagement ↔ content 视频信息 internal API（INTERNAL_API_TOKEN） | 成员 D + 成员 C | | 未开始 |
| 3 | SRS webhook 回调到 engagement + UC10 开播流程 | 成员 D | | 未开始 |
| 4 | UC08 demo-pay 内部端点打通 + 订单幂等 | 成员 B | | 未开始 |
| 5 | 网关追加三服务 /admin/* 路由映射 + 鉴权 | 成员 E（公共层） | | 未开始 |
| 6 | 通知中心推送（动态 + 特别关注 + 预约开播） | 成员 B | | 未开始 |
| 7 | UC12 analytics 聚合查询跨服务接口 | 成员 C | | 未开始 |
| 8 | 以上 7 项完成后，批量移除 8 个 ucXX.skip → 全量回归 13/13 | 全员 + 成员 E | | 未开始 |

---

## 5. 执行痕迹（填入 artifacts 路径）

- 容器状态快照：`artifacts/microservices-e2e/compose-ps.txt`
- 全部容器日志：`artifacts/microservices-e2e/compose.log`
- Playwright HTML 报告：`artifacts/microservices-e2e/playwright-report-full/index.html`
- JUnit XML：`artifacts/microservices-e2e/junit-microservices-full.xml`
- 截图 / 录像 / trace：`artifacts/microservices-e2e/test-results-full/`
