# 游戏运营 SOP - 知识库文件说明

本目录包含游戏运营客服场景的结构化知识库文件，用于 AI 问答与问题排查系统。
核心覆盖五大高频投诉/排查场景，每个场景均有完整的日志链路、关键字段定义及排查步骤。

## 📚 文件列表
| 文件名 | 说明 |
|--------|------|
| [core.md](./core.md) | 核心游戏概念：玩家模型、区服、订单、物品、匹配、反作弊、社交系统的业务定义 |
| [payment/overview.md](./payment/overview.md) | 充值不到账/重复扣款/第三方支付异常排查 SOP |
| [inventory/overview.md](./inventory/overview.md) | 道具/装备/货币消失排查 SOP，完整物品生命周期流水追踪 |
| [matchmaking/overview.md](./matchmaking/overview.md) | 匹配/排位异常投诉排查 SOP，含机器人检测、结算校验等 |
| [anticheat/overview.md](./anticheat/overview.md) | 外挂/作弊误判申诉排查 SOP，反作弊规则解析与证据链说明 |
| [social/overview.md](./social/overview.md) | 社交系统纠纷排查 SOP，含公会战、礼物赠送、聊天举报等场景 |

---

## 重要说明

1. 上面的文件可以调用 `SopRead` 工具读取，每个 overview.md 包含概念说明、排查步骤、FAQ、注意事项四个部分
2. **字段说明和查询语句**统一在 `stores/` 目录中维护，业务模块 SOP 中直接引用
3. 排查问题时必须先确认用户提供的 `user_id`、`server_id`、`order_id`（或对应关键 ID），再逐步查询日志链路
4. 日志分层（SLS Project 统一为 `sop-demo`）详见 [stores/README.md](./stores/README.md)
5. 核心概念优先查 `core.md` 了解；字段含义不明时查对应 Logstore 的字段说明，不要臆测
6. 所有排查链路均基于 `who/when/what/how` 四维原则：操作者、时间戳、操作对象、操作方式
7. 涉及账号封禁/资产补偿等高风险操作，必须在日志证据完整的情况下才可操作，且需要二次审核
