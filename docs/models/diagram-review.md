# 正式图完整性检查记录

检查日期：2026-08-26  
检查范围：`docs/models/` 中 46 张正式图；下表为已完成人工复核的 30 张核心图。
检查项：PlantUML 语法、源文件与 SVG/PNG 对应、编号、内容完整性、层级边界、异常路径、空类/空组件、连线和文字可读性。

| 图 | 检查重点 | 结果 |
| --- | --- | --- |
| `USECASE-B` | 成员 B 范围、系统边界及参与者关联 | 通过；明确仅覆盖 UC01/07/08/11 |
| `CLASS-B` | 成员 B 领域实体、属性、多重性与实现依赖 | 通过 |
| `COMPONENT-B` | 成员 B 当前 Vue/Gin/GORM/Chat Hub 单体组件边界 | 通过 |
| `SYS-SEQ01/07/08/11` | 系统黑盒边界、主流程和失败分支 | 通过 |
| `COMP-SEQ01/07/08/11` | 页面、网关、组件及返回调用链 | 通过 |
| `OBJ-SEQ01/07/08/11` | 真实 Handler、Logic、`*gorm.DB`、model 与方法名 | 通过；可追溯当前代码 |
| `UC05-UC09-UC10-overview` | 系统边界、三个参与者与三个用例关联 | 通过；关联改为标准无方向 UML 关联线 |
| `DOMAIN-CLASS-D` | 业务实体、核心属性、多重性与关系语义 | 通过；已补齐视频点赞、视频收藏、直播预约、直播点赞、通知属性 |
| `SYS-SEQ05` | 用户—系统边界、互动主流程与非法输入路径 | 通过 |
| `SYS-SEQ09` | 主播/用户/计划任务、创建冲突与不可预约路径 | 通过 |
| `SYS-SEQ10` | 主播/观众/SRS、媒体与互动、校验/持久化失败路径 | 通过 |
| `COMPONENT-D` | Vue、Gin、MySQL、SRS 组件边界和依赖方向 | 通过 |
| `COMP-SEQ05` | Page、API、Router、Handler、DB 的调用层级 | 通过；已补齐逐层返回链路 |
| `COMP-SEQ09` | Page、API、Router、Handler、Worker、DB 的调用层级 | 通过；已补齐逐层返回链路 |
| `COMP-SEQ10` | HTTP、WebSocket、Hub、MySQL、SRS 的职责划分 | 通过；已补齐创建直播返回链路 |
| `IMPLEMENTATION-CLASS-D` | 页面、Handler、Hub、Client、数据模型的属性、方法和依赖 | 通过；所有空类已补齐，Video 及关联实体已接入关系图 |
| `OBJ-SEQ05` | 页面对象、API、鉴权、Handler、关系对象与事务 | 通过；已修正跨层直接返回 |
| `OBJ-SEQ09` | 页面对象、计划/预约实体、事务和 Worker | 通过；已修正跨层直接返回 |
| `OBJ-SEQ10` | WebSocket、Client、Hub、事务、房间和礼物对象 | 通过；包含权限/慢速模式/持久化失败分支 |
| `DEPLOY-MONO` | 单机容器、协议、数据库、共享卷和故障域 | 通过 |
| `DEPLOY-K8S` | 集群接入、Deployment/StatefulSet、PV、配置、密钥和 HPA | 通过；已删除规划/目标免责声明标记 |

## 本轮修正

1. 补齐概念类图中五个空实体的关键业务属性。
2. 补齐实现类图中页面、WebSocket Handler 和全部数据模型的代表性属性或方法。
3. 增加 Video 与弹幕、评论、点赞、收藏及相关 Handler 的依赖/关联。
4. 修正组件级和对象级顺序图中数据库直接返回页面的跨层画法，恢复逐层响应。
5. 将用例图参与者与用例之间的箭头改为标准无方向关联线。
6. 删除 Kubernetes 图中的附加规划性标记和免责声明。
7. 将成员 B 范围图改名为 `USECASE-B` / `CLASS-B`，对象图改为真实 Handler、Logic、GORM DB 与 model 调用。
8. Kubernetes 图收敛为 user/content/engagement 三服务及 Schema 所有权，跨域访问改为内部 API。
9. 新增成员 B 当前单体组件图 `COMPONENT-B`，并为成员 B 全部正式图统一引用 `_theme.puml`。
10. UC01 补齐 `UpdateMeHandler`/`UploadAvatarHandler`；UC07 补齐分组和屏蔽；UC08 补齐支付幂等事务；UC11 改为真实 `chat.Hub.CreateAndBroadcast` 调用链。

自动检查命令：`python3 scripts/check_diagram_assets.py`。该检查会拒绝缺失 SVG/PNG、编号不规范、正式文档使用 Mermaid，以及类图存在空类的提交。

## 尚待人工复核

UC02、UC03、UC04、UC12、UC13 的 15 张三层顺序图，以及 `DEPLOY-MONO-DETAILED` 已具备源文件和导出物，但尚未在本记录中逐图登记布局、调用层级和异常路径的人工复核结论。
