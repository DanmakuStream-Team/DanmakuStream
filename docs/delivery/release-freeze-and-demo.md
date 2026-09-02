# 最终版本冻结与 A 组答辩手册

## 1. 冻结条件

最终版本只能从 `main` 冻结，并同时满足：

- `dev → main` PR 已合并，工作区无未提交修改；
- main 精确 SHA 的总 CI 为 success；
- 三服务最终 `go test ./...` 通过；
- 微服务 E2E、自动部署、HPA 扩/缩容和故障演练均有 Action URL 与 artifact；
- 三个 `/api/v1/version` 返回相同目标 commit SHA；
- 部署/回滚、测试报告、追溯表和答辩材料中的版本号一致。

当前单体对比版本已经使用 `v1.0`，微服务最终标签建议使用 `microservices-v1.0.0`，避免覆盖或误解单体标签。最终名称由组内确认后输入流水线，不在代码中写死。

## 2. 自动冻结

在 Actions 中选择 `release-freeze`，必须从 main 运行并填写：

- `release_tag`：确认后的标签；
- `expected_sha`：要冻结的 main 完整 40 位 SHA；
- `confirm_freeze=FREEZE`。

流水线会核对 main HEAD、精确 SHA 的 CI、Compose/K8s 渲染和三服务测试；全部通过且 production Environment 批准后才创建 annotated tag。相同标签指向同一 SHA 时可安全重跑；若标签已指向其他提交则立即失败，绝不移动旧标签。

## 3. A 组现场演示顺序（约 8 分钟）

1. **版本与 CI/CD**：打开最终 tag 和 main SHA，展示三服务 CI、E2E、`microservice-cd-<sha>` artifact。
2. **Kubernetes 状态**：展示 Deployment、Pod、Service、HPA；查询三个版本接口，说明镜像标签和响应 commit 一致。
3. **HPA**：打开 `hpa-timeline.svg`，说明固定并发下 1→N→1 的时间线，同时展示请求数、P95、错误率。
4. **故障隔离**：展示 content-service 停止期间 engagement 2.5 秒内返回 503/504，user/engagement 健康且 restart 不增加。
5. **自动恢复/回滚**：展示 content-service 恢复、HPA 恢复；再说明发布失败使用 `rollout undo`，数据库/PVC 不自动回滚。
6. **边界说明**：如实说明 MySQL 单副本、RWO 媒体卷和多副本 WebSocket 的限制，不宣称尚未实现的高可用能力。

## 4. 现场命令

```bash
NS=danmakustream-microservices
sudo k3s kubectl -n "$NS" get deploy,pod,svc,hpa -o wide
sudo k3s kubectl -n "$NS" top pods
for service in user-service content-service engagement-service; do
  sudo k3s kubectl -n "$NS" exec "deploy/$service" -- \
    wget -qO- http://127.0.0.1:8080/api/v1/version
  echo
done
```

手动回滚指定服务：

```bash
sudo k3s kubectl -n danmakustream-microservices rollout undo deploy/content-service
sudo k3s kubectl -n danmakustream-microservices rollout status deploy/content-service --timeout=240s
```

## 5. 最终证据索引

冻结前把下表填入最终测试报告或技术报告；空项不能标记完成。

| 证据 | URL/文件 | SHA/结论 |
|---|---|---|
| main 总 CI | 待填写 | 待填写 |
| 三服务 CI | 待填写 | 待填写 |
| 13/13 微服务 E2E | 待填写 | 待填写 |
| k3s 自动部署 | 待填写 | 待填写 |
| HPA 扩/缩容 | 待填写 | 待填写 |
| content 依赖故障 | 待填写 | 待填写 |
| 单体/微服务性能对比 | 待填写 | 待填写 |
| 最终 annotated tag | 待填写 | 待填写 |

建议提前下载所有 artifact 并准备录屏。GitHub 或公网现场不可用时，用未修改的原始 artifact/录屏演示，不临时伪造输出。
