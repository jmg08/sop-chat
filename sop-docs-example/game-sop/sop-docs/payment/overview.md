# 充值支付 SOP

充值支付模块负责处理玩家从发起充值到虚拟货币/道具到账的全流程，涉及游戏服务器、第三方支付网关、发货系统三个核心环节。

---

## 支付全链路流程

```
客户端发起充值
    ↓  pay_request_log          [user_id, server_id, product_id, amount, channel]
游戏服务器创建订单
    ↓  order_create_log         [order_id, user_id, amount, channel, status=created]
跳转第三方支付渠道
    ↓  pay_gateway_redirect_log [order_id, cp_order_id, channel, redirect_url]
第三方支付回调通知  ← 最关键节点
    ↓  pay_callback_log         [order_id, cp_order_id, callback_status, sign_verify_result]
幂等校验（防重复）
    ↓  idempotent_check_log     [order_id, idempotent_key, is_duplicate, lock_acquired]
发货（道具写入背包）
    ↓  deliver_log              [order_id, user_id, product_id, deliver_status, retry_count]
订单状态 → delivered
    ↓  order_update_log         [order_id, status=delivered, delivered_at]
```


---

## 排查步骤

### 场景一：充值不到账（用户付了钱未收到钻石）

* **Step 1 — 确认用户基本信息**：获取 `user_id`、`server_id`、充值金额、支付渠道、大概充值时间。如果玩家有订单截图，优先让玩家提供渠道订单号 `cp_order_id`
* **Step 2 — 查询游戏侧订单记录**：参考 [game-biz-logstore → 查询玩家最近充值订单](stores/game-biz-logstore.yaml)
* **Step 3 — 查询支付回调日志**：参考 [game-biz-logstore → 查询指定订单的支付回调](stores/game-biz-logstore.yaml)
* **Step 4 — 查询幂等校验记录**：参考 [game-biz-logstore → 查询幂等校验记录（by order_id）](stores/game-biz-logstore.yaml)
* **Step 5 — 查询发货流水**：参考 [game-biz-logstore → 查询发货记录](stores/game-biz-logstore.yaml)
* **Step 6 — 对比背包流水确认到账**：参考 [game-biz-logstore → 查询充值发货写入背包的流水](stores/game-biz-logstore.yaml)


---

### 场景二：重复扣款投诉

* **Step 1 — 通过 cp_order_id 查询所有相关回调**：参考 [game-biz-logstore → 通过渠道订单号查询回调（重复扣款排查）](stores/game-biz-logstore.yaml)
* **Step 2 — 检查幂等去重记录**：参考 [game-biz-logstore → 查询幂等校验记录（by cp_order_id）](stores/game-biz-logstore.yaml)
* **Step 3 — 核查发货次数**：参考 [game-biz-logstore → 统计某玩家在时间段内的发货次数](stores/game-biz-logstore.yaml)
* **Step 4 — 核查背包流水增量次数**：参考 [game-biz-logstore → 查询货币变更流水](stores/game-biz-logstore.yaml)（过滤 `change_reason="purchase"`）


---

### 场景三：第三方显示成功，游戏内失败

* **Step 1 — 用 cp_order_id 查回调日志**：参考 [game-biz-logstore → 查询指定订单的支付回调](stores/game-biz-logstore.yaml)（以 `cp_order_id` 作为过滤条件）。若回调未到达，可能原因：① 回调 URL 配置错误；② 游戏服务器回调接口异常（查 `game-system-logstore` 错误日志）；③ 网络防火墙拦截。处理方案：联系技术团队检查回调 URL 配置，或在后台手动触发补单
* **Step 2 — 如回调到达但签名验证失败**：参考 [game-biz-logstore → 查询签名校验失败的回调](stores/game-biz-logstore.yaml)
* **Step 3 — 如回调正常但未发货，查系统错误日志**：参考 [game-system-logstore → 查询指定订单的服务错误日志（支付失败根因排查）](stores/game-system-logstore.yaml)


---

## FAQ

**Q：玩家的充值记录在哪里查？**

在游戏 GM 工具中输入 `user_id` 可查看充值历史；也可在 [`game-biz-logstore`](stores/game-biz-logstore.yaml) 中执行 **[查询玩家最近充值订单](stores/game-biz-logstore.yaml)**。

---

**Q：如何判断是玩家侧问题还是游戏侧问题？**

- `pay_callback_log` 中 `callback_status=success` → 游戏服务器已收到回调，后续问题在发货环节
- `pay_callback_log` 中无记录 → 回调未到达，联系支付渠道或检查回调地址配置
- 背包流水中有 `change_reason=purchase` 记录 → 钻石已发放

---

**Q：AppStore / Google Play 充值回调和国内支付有什么不同？**

- AppStore 使用 `receipt` 验证机制，需向苹果服务器发起二次验证
- Google Play 使用 `purchase token` 验证，需通过 Google API 校验
- 这两个渠道的 `cp_order_id` 格式与国内不同，需用渠道特有的 `transaction_id` 或 `purchase_token` 核查
- 国际渠道回调延迟可能较高（最长 30 分钟），需告知玩家等待

---

**Q：玩家要求退款怎么处理？**

1. 先确认发货记录是否存在，若已发货需先回收道具/货币再处理退款
2. AppStore / Google Play 退款由平台处理，游戏侧无法直接操作
3. 国内支付渠道退款需联系财务走退款流程
4. 退款处理完成后需在 [`game-audit-logstore`](stores/game-audit-logstore.yaml) 的 `gm_op_log` 中记录操作日志（`operator_id + reason`）

---

## 注意事项

- 第三方支付回调可能存在延迟（通常 5 秒内，极端情况可达 30 分钟），不要过早判断到账失败
- 网络抖动可能导致回调被重复发送，幂等校验保证只发货一次
- AppStore 沙盒环境订单不会产生真实扣款，排查测试环境问题时注意区分
- 已发货订单不可直接冲正，必须走物品回收流程
