# 道具 / 装备 SOP

物品系统管理玩家背包中所有虚拟资产（装备、消耗品、货币、材料等）的生命周期，每次变更均产生**不可变**的背包流水（`bag_txn_log`）。

---

## 物品生命周期

```
获得（acquire）  →  充值发货 / 掉落 / 任务奖励 / 邮件领取 / 合成 / 交易购入
持有（hold）     →  存在于背包中，可通过背包快照查询
使用/消耗        →  使用消耗品效果 / 强化装备 / 合成材料
流转（transfer） →  交易行卖出 / 赠送好友 / 公会仓库存取
消亡（destroy）  →  过期销毁 / 分解 / GM 回收
```

---

## 排查步骤

### 场景一：道具 / 装备消失（通用流程）

> **注意**：优先获取 `item_uuid`（物品实例 ID），可最精确地追踪单件装备。如果玩家无法提供，则用 `item_id + user_id + 时间范围` 查询。

* **Step 1 — 确认玩家最后登录状态（排查盗号）**：参考 [game-biz-logstore → 查询玩家登录记录（排查盗号）](stores/game-biz-logstore.yaml)
* **Step 2 — 通过背包快照确认物品最后存在时刻**：参考 [game-biz-logstore → 查询背包快照（追踪物品最后存在时刻）](stores/game-biz-logstore.yaml)
* **Step 3 — 通过背包流水追踪消失节点（核心步骤）**：参考 [game-biz-logstore → 查询物品背包流水（道具消失核心查询）](stores/game-biz-logstore.yaml)。最后一条 `change_type=remove` 即为消失原因，`change_reason` 各值含义见 [game-biz-logstore → change_reason](stores/game-biz-logstore.yaml)。根据值决定后续：`craft`/`dismantle` → Step 4；`trade_sell` → Step 5；`expire` → Step 6；`admin_op` → Step 7；空/不合理值 → 可能是系统 Bug，上报技术团队
* **Step 4 — 查询合成/分解记录（`change_reason=craft` 或 `dismantle` 时）**：参考 [game-biz-logstore → 查询合成/分解记录](stores/game-biz-logstore.yaml)
* **Step 5 — 查询交易记录（`change_reason=trade_sell` 时）**：参考 [game-audit-logstore → 查询物品交易记录](stores/game-audit-logstore.yaml)
* **Step 6 — 查询邮件领取记录（`change_reason=expire` 时）**：参考 [game-biz-logstore → 查询邮件附件领取情况](stores/game-biz-logstore.yaml)，确认道具是否以邮件附件形式存在但过期未领取
* **Step 7 — 查询 GM 操作记录（`change_reason=admin_op` 时）**：参考 [game-audit-logstore → 查询 GM 操作记录](stores/game-audit-logstore.yaml)


---

### 场景二：货币（钻石/金币）异常减少

* **Step 1 — 查询货币变更流水**：参考 [game-biz-logstore → 查询货币变更流水](stores/game-biz-logstore.yaml)。常见货币 `item_id` 见 [game-biz-logstore → item_id](stores/game-biz-logstore.yaml)
* **Step 2 — 按 change_reason 分类分析**：`change_reason` 各值含义见 [game-biz-logstore → change_reason](stores/game-biz-logstore.yaml)
* **Step 3 — 如确认是系统 Bug 导致的异常扣除**：上报技术团队，提供 `txn_id` 列表和具体 `delta` 值申请 GM 补偿。补偿操作必须在 [`game-audit-logstore`](stores/game-audit-logstore.yaml) 的 `gm_op_log` 中记录操作员 ID 和补偿原因

---

### 场景三：邮件附件道具未收到

* **Step 1 — 查询邮件发送记录**：参考 [game-biz-logstore → 查询邮件发送记录](stores/game-biz-logstore.yaml)
* **Step 2 — 查询邮件领取记录**：参考 [game-biz-logstore → 查询邮件附件领取情况](stores/game-biz-logstore.yaml)。`bag_full` / `recv_result` 各值含义见 [game-biz-logstore → recv_result](stores/game-biz-logstore.yaml)
* **Step 3 — 查询背包流水确认附件是否到账**：参考 [game-biz-logstore → 查询邮件附件写入背包的流水](stores/game-biz-logstore.yaml)

---

## FAQ

**Q：如何快速找到玩家的 item_uuid？**

- 方法 1：让玩家提供道具详情页的物品 ID（部分客户端会展示）
- 方法 2：通过 **[查询背包快照](stores/game-biz-logstore.yaml)** 用 `item_id` 查询，找到该玩家持有的 `item_uuid`
- 方法 3：通过物品获得时的背包流水（`change_type=add`）逆向找到 `item_uuid`

---

**Q：玩家说道具是被盗号偷走的怎么处理？**

1. 执行 **[查询玩家登录记录（排查盗号）](stores/game-biz-logstore.yaml)**，确认是否有异地/异设备登录
2. 执行 **[查询物品背包流水](stores/game-biz-logstore.yaml)**，过滤 `change_reason=trade_sell` 或 `admin_op`，确认道具是否被转移
3. 如确认盗号，协助玩家找回账号（修改密码/解绑设备）
4. 物品找回需走特殊申诉流程，附带完整证据链（登录 IP 异常记录 + 物品转移流水）

---

**Q：道具有效期是怎么定义的？**

- 每个物品模板（`item_id`）有 `expire_type` 字段：永久 / 固定截止时间 / 持有 N 天后过期
- 玩家持有的物品实例（`item_uuid`）有 `expire_time` 字段记录具体过期时间戳
- 过期删除会在 `bag_txn_log` 中产生 `change_reason=expire` 的记录

---

## 注意事项

- 背包流水是不可篡改的审计日志，任何物品变更都必须有对应流水；若没有流水则判定为系统 Bug
- 物品一旦过期销毁或玩家主动分解，原则上不予找回（需严格依照游戏运营政策）
- 交易行出售的物品若已被他人购买，无法直接回滚，需走特殊处理流程
- GM 工具批量操作（如发放补偿）必须先在测试环境验证，且需要至少两人审核
