# 匹配 / 排位 SOP

匹配系统负责将玩家按技术水平、延迟、游戏模式等维度分组进入对局，并在对局结束后进行结算（积分/段位更新）。排查匹配和排位异常时，核心是还原对局完整的生命周期日志链。

---


## 概念介绍

### MMR 与段位积分

**MMR（Matchmaking Rating）** 是系统内部用于匹配的隐藏评分，反映玩家真实技术水平，不直接对玩家展示。**段位积分**是玩家可见的排名数值，两者由不同算法维护，更新时机也相互独立——积分更新后段位可能有最多 30 秒延迟。

胜负对积分的影响由 ELO/Glicko2 模型决定：以弱胜强赢分更多、扣分更少；以强负弱扣分更多、赢分更少。**连胜保护**、**保护期**、**新赛季校准期**等特殊规则会覆盖标准积分算法，此类情况下赢了也可能不加分或少加分，属正常逻辑。

---

### 挂机检测

系统通过采样玩家在对局中的行为数据（移动距离、操作频率、战斗贡献）计算 **`afk_score`**，达到阈值（通常 80~100 分）才触发惩罚。以下情况不会触发检测：
- 对局时长过短（未进入检测窗口）
- 玩家有移动行为但战斗贡献极低（需综合战斗贡献分判断）

惩罚按累计次数递增：首次警告 → 短期禁赛 → 长期禁赛。

---

### 机器人（Bot）填充机制

当匹配等待超时且人数不足时，部分模式允许用官方 AI（`fill_bot`）填充空位，系统会在对局开始界面显式告知玩家。**排位模式严禁出现 `fill_bot`**，若发现需立即上报技术团队。新手引导局和训练模式使用 `training_bot`，属正常设计，不构成投诉依据。

---

### 对局生命周期

一场完整的对局由以下阶段依次组成，每个阶段均会产生对应日志：

1. **匹配请求**：玩家点击匹配后，系统记录 MMR、游戏模式、所在大区等信息，开始寻找合适对手
2. **匹配成功/分组**：系统完成分组，生成 `match_id`，记录所有参与玩家列表及 MMR 分布
3. **对局开始**：分配对局服务器，生成全局唯一的 `battle_id`，后续所有阶段均以此为主键关联
4. **对局中行为采样**：以采样方式记录玩家移动、操作频率、延迟等行为数据，用于挂机检测和机器人识别
5. **对局结束**：记录胜负结果、对局时长、结束原因（正常/断线/超时），是判断是否触发特殊积分规则的关键节点
6. **结算**：根据 ELO/Glicko2 模型计算积分变化并更新段位，记录变化前后的积分值和变更原因
7. **结算确认（客户端 ACK）**：客户端收到并确认结算数据，若未收到则下次登录时重新同步

排查问题时，通常需要按上述顺序逐步追溯，找到链路断点所在的阶段。

---

## 排查步骤

### 场景一：赢了却掉段/掉星

* **Step 1 — 查询对局结果日志**：参考 [game-audit-logstore → 查询玩家近期对局结果](stores/game-audit-logstore.yaml)
* **Step 2 — 查询结算日志**：参考 [game-audit-logstore → 查询指定对局的结算记录](stores/game-audit-logstore.yaml)。`settlement_status` / `rank_change_reason` / `network_status_at_settle` 各值含义见 [game-audit-logstore → settlement_status](stores/game-audit-logstore.yaml)
* **Step 3 — 确认结算数据同步状态**：参考 [game-audit-logstore → 查询结算数据同步状态](stores/game-audit-logstore.yaml)
* **Step 4 — 特殊规则核查**：以下情况胜利局也可能扣分，属于正常逻辑：连胜保护失效后首次胜利仍可能因综合评分下降扣分；MMR 与段位积分是两套独立系统，MMR 赢了但积分可能仍在调整期；以弱胜强赢分少，以强负弱扣分多

---

### 场景二：匹配到机器人投诉

* **Step 1 — 查询对局玩家列表**：参考 [game-audit-logstore → 查询对局玩家列表（机器人核查）](stores/game-audit-logstore.yaml)
* **Step 2 — 查询可疑玩家的机器人标记**：参考 [game-audit-logstore → 查询玩家机器人标记](stores/game-audit-logstore.yaml)。`is_bot` / `ai_label` 各值含义见 [`game-audit-logstore` → is_bot / ai_label](stores/game-audit-logstore.yaml)
* **Step 3 — 查询 ELO 匹配计算过程**：参考 [game-audit-logstore → 查询匹配算法过程](stores/game-audit-logstore.yaml)
* **Step 4 — 核查可疑账号行为特征**：参考 [game-audit-logstore → 分析可疑账号行为特征（机器人检测）](stores/game-audit-logstore.yaml)


---

### 场景三：队友挂机但未被惩罚

* **Step 1 — 查询挂机行为数据**：参考 [game-audit-logstore → 分析对局内挂机行为数据](stores/game-audit-logstore.yaml)
* **Step 2 — 查询反挂机系统检测结果**：参考 [game-audit-logstore → 查询挂机检测结果](stores/game-audit-logstore.yaml)
* **Step 3 — 查询举报系统记录**：参考 [game-audit-logstore → 查询指定对局的挂机举报记录](stores/game-audit-logstore.yaml)
* **Step 4 — 核查该玩家历史挂机记录**：参考 [game-audit-logstore → 查询玩家 30 天内挂机累计次数](stores/game-audit-logstore.yaml)


---

## FAQ

**Q：结算积分计算逻辑是什么？**

- 使用 ELO / Glicko2 模型计算积分变化
- 胜利积分 = `base_win_points × (1 + performance_bonus_ratio) × expected_difficulty_factor`
- 失败扣分 = `base_lose_points × (1 + expected_difficulty_factor)`
- 对阵强于自己（MMR 差距大）的对手：胜利得分更多，失败扣分更少
- `performance_bonus` 基于 KDA、输出、贡献等维度综合评分

---

**Q：什么情况下胜利局不加积分？**

- 对局时长不足最低有效时长（一般为 5 分钟）
- 服务端检测到本局存在作弊行为，全局结算被取消
- 结算时玩家处于掉线状态超过阈值，触发掉线积分规则
- 保护期/新赛季校准期采用特殊积分规则

---

**Q：机器人玩家在哪些模式下允许出现？**

| 场景 | 说明 |
|------|------|
| 新手引导局 | 全程对战 AI |
| 训练模式 | 可选 AI 对手 |
| 排队等待超时填充（fill_bot） | 等待超过 5 分钟且人数不足时启用，须在对局开始界面显式告知玩家 |
| **排位模式** | **不允许出现 fill_bot，若发现请立即上报技术团队** |

---

## 注意事项

- 结算日志在对局服务器（battle server）上生成，网络波动可能导致日志传输延迟最长 5 分钟
- 积分系统和段位系统是异步解耦的，积分更新后段位更新可能有最多 30 秒延迟
- 掉线处理规则：掉线 < 30 秒重连不扣分，30~120 秒按时长比例处理，超过 120 秒视为放弃
- 历史对局回放数据保留 60 天，超过后无法查询 `battle_behavior_log` 行为采样日志
