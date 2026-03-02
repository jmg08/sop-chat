# 社交系统纠纷 SOP

社交系统涵盖公会（Guild）、好友（Friend）、聊天（Chat）、举报（Report）、礼物（Gift）五个子系统，每个子系统均有独立的事件日志，纠纷排查需跨日志关联分析。

---

## 子系统概览

### 公会系统（Guild System）

玩家组成公会共同参与公会战，战后系统按伤害与贡献值评定奖励资格，不达标的成员不计入发放名单。成员变更（加入/离开/踢出）和仓库物品操作均保留完整流水，可用于仲裁纠纷。

---

### 好友与礼物系统（Friend & Gift System）

好友间可互赠道具礼物，礼物有效期通常为 7 天。接收方背包满时礼物不直接写入背包，系统自动转存为邮件，清理背包后可手动领取；过期后附件无法找回。

---

### 聊天系统（Chat System）

聊天内容以哈希形式脱敏存储，原文查询需具备审核权限。违规消息由敏感词过滤系统自动拦截，敏感词命中列表（`sensitive_words_hit`）可对外展示，原文（`content`）不可对外披露。

---

### 举报系统（Report System）

玩家可对挂机、言语骚扰、作弊等行为发起举报，系统先执行自动审核，无法判定时进入人工审核队列（通常 24 小时内处理）。举报人信息对被举报方严格保密。

---

## 排查步骤

### 场景一：公会战奖励争议

玩家反馈参与了公会战但未收到奖励，或奖励数量不对。

* **Step 1 — 确认玩家是否在参战名单中**：参考 [game-audit-logstore → 查询公会战参战资格](stores/game-audit-logstore.yaml)
* **Step 2 — 查询奖励发放记录**：参考 [game-biz-logstore → 查询公会战奖励发放记录](stores/game-biz-logstore.yaml)
* **Step 3 — 查询玩家奖励领取记录**：参考 [game-biz-logstore → 查询玩家公会战奖励领取记录](stores/game-biz-logstore.yaml)
* **Step 4 — 核对伤害统计和贡献分**：参考 [game-audit-logstore → 查询公会战伤害与贡献统计](stores/game-audit-logstore.yaml)

---

### 场景二：好友礼物未收到

玩家反馈收到好友赠送通知，但背包中未出现礼物。

* **Step 1 — 查询礼物发送记录**：参考 [game-biz-logstore](stores/game-biz-logstore.yaml) 中的 `gift_send_log`（过滤 `sender_id` + `receiver_id`），确认礼物是否成功发送及过期时间（字段说明见 [game-biz-logstore → gift_send_log](stores/game-biz-logstore.yaml)）
* **Step 2 — 查询礼物领取状态**：参考 [game-biz-logstore](stores/game-biz-logstore.yaml) 中的 `gift_event_log`（过滤 `gift_uuid`，按时间升序）。`event_type` 各值含义见 [`game-biz-logstore` → event_type](stores/game-biz-logstore.yaml)
* **Step 3 — 确认背包流水**：参考 [game-biz-logstore → 查询礼物领取写入背包的流水](stores/game-biz-logstore.yaml)
* **Step 4 — 礼物系统限制核查**：每天每人最多发送 N 件礼物（N 查活动配置）；礼物有效期通常为 7 天，过期自动消除；接收者背包满时礼物不会自动写入，需清理背包后手动领取；跨区服好友无法互赠礼物

---

### 场景三：聊天举报与违规处置

玩家对他人发送骚扰/辱骂内容进行举报，或被举报玩家申诉处理不当。

* **Step 1 — 查询举报记录**：参考 [game-audit-logstore → 查询举报处理记录](stores/game-audit-logstore.yaml)
* **Step 2 — 查询被举报消息的完整上下文（需审核权限）**：参考 [game-audit-logstore → 查询聊天消息上下文（需审核权限）](stores/game-audit-logstore.yaml)
* **Step 3 — 查询被举报玩家的历史违规记录**：参考 [game-audit-logstore → 查询被举报玩家的历史违规记录](stores/game-audit-logstore.yaml)
* **Step 4 — 处理聊天禁言申诉**：查询禁言触发的 `message_id` 和 `sensitive_words_hit`；确认是否为误触发（正常词汇被误识别为敏感词）；如确认误判，解除禁言并将该词加入白名单（提交内容审核团队）；在 [`game-audit-logstore`](stores/game-audit-logstore.yaml) 的 `gm_op_log` 中记录操作员 ID 和解封原因

---

### 场景四：公会成员纠纷（踢人/退会争议）

* **Step 1 — 查询公会成员变更记录**：参考 [game-audit-logstore → 查询公会成员变更记录](stores/game-audit-logstore.yaml)
* **Step 2 — 核查操作权限**：`leader` 可踢所有非 leader 成员；`officer` 只能踢普通 `member`，不能踢其他 `officer` 或 `leader`。如发现权限越级操作，属于系统 Bug，需上报技术团队
* **Step 3 — 核查公会仓库物品归属**：参考 [game-audit-logstore → 查询公会仓库变更记录](stores/game-audit-logstore.yaml)


---

## FAQ

**Q：公会战开始后中途加入的成员能否获得奖励？**

一般情况下，公会战报名截止时间前加入公会的成员才有资格获得奖励。具体截止规则查活动配置（`guild_war_config`），通常是战前 24~48 小时。

---

**Q：礼物过期了能否补发？**

- 原则上礼物过期后无法找回（玩家应在有效期内主动领取）
- 如果是系统 Bug 导致礼物提前过期（`expire_time` 记录异常），可申请补偿
- 补偿操作需提供 `gift_uuid` 和系统异常证据，走 GM 补偿流程

---

**Q：举报后多久会有处理结果？**

| 类型 | 处理时间 |
|------|---------|
| 自动审核（明显违规） | 通常 1 小时内 |
| 人工审核 | 通常 24 小时内 |
| 复杂案例 | 最长 72 小时 |

举报结果会通过系统邮件通知举报人。

---

**Q：被举报的玩家能看到是谁举报他吗？**

举报人信息对被举报玩家**严格保密**，任何情况下不可泄露。举报记录中的 `reporter_id` 只有内部审核人员有权查看。

---

**Q：公会解散后公会仓库的物品去哪了？**

公会解散时，公会仓库物品会通过系统邮件按贡献值比例分配给成员。分配记录在 `guild_storage_log` 中的 `event_type=guild_disband_distribute` 记录中。解散时已离开公会的成员无法获得分配。

---

## 注意事项

- 聊天内容原文属于用户隐私数据，查询需要专门的审核权限，客服不得随意查阅
- 举报处置结果（具体违规原因）对被处置玩家只通知结果，不透露完整细节（防止规避）
- 公会仓库物品价值较高时的分配争议，需升级至运营主管介入，不可由客服单独处理
- 礼物系统跨服版本差异较大，排查前请确认玩家所在区服的版本号（`server_version`）
