# 日志存储说明（SLS Stores）

本目录描述游戏日志平台的 Project、Logstore 结构，包含每个 Logstore 的用途、字段说明及常用查询模板。各业务模块排查步骤中的查询语句均以本目录为字段参考依据。

---

## SLS Project 信息

| 属性 | 值 |
|------|-----|
| **Project** | `sop-demo` |
| **地域** | 按实际部署配置 |
| **访问方式** | SLS 控制台 / SDK / CLI |

---

## Logstore 总览

| Logstore | 用途 | 主要查询角色 | 保留期 |
|----------|------|------------|--------|
| [`game-biz-logstore`](./game-biz-logstore.yaml) | 业务流水日志：充值支付、物品背包、社交礼物/邮件 | 客服直接查询 | 365 天 |
| [`game-audit-logstore`](./game-audit-logstore.yaml) | 行为审计日志：对局结算、经济交易、反作弊、GM 操作 | 合规、仲裁、反作弊团队 | 730 天 |
| [`game-system-logstore`](./game-system-logstore.yaml) | 系统运行日志：服务异常、错误堆栈、第三方回调原始报文 | 技术团队排查 | 90 天 |

---

## 通用字段规范

所有业务流水日志**必须**包含以下四个基础字段：

| 字段 | 类型 | 说明 |
|------|------|------|
| `user_id` | string | 全局唯一玩家 ID |
| `server_id` | string | 区服 ID（格式：`region_code + server_seq`，如 `cn_s1`） |
| `timestamp` | long | 操作时间戳（Unix 毫秒） |
| `trace_id` | string | 跨系统链路追踪 ID，同一次用户操作产生的所有日志共享同一 `trace_id` |

> 通过 `trace_id` 可以在任意 Logstore 中检索同一次操作的完整上下文。

---

## 日志设计最佳实践

- 物品变更必须同时记录 `qty_before` 和 `qty_after` 快照，不能仅记录 `delta`
- 支付相关日志必须同时保留游戏侧 `order_id` 和渠道侧 `cp_order_id`，方便双边核对
- 敏感操作（GM 补偿、账号封禁）必须记录操作人 `operator_id` 和操作原因 `reason`
- 反作弊日志的规则阈值和算法细节属于高敏感数据，**不可对外披露**
