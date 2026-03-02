# 客服提问模板

客服在接到玩家投诉后，请按对应场景收集以下信息，再发起排查。标注 `*` 的字段为必填。

---

## 充值问题

### 充值不到账

| 字段 | 说明 | 必填 |
|------|------|------|
| `user_id` | 投诉玩家的 user_id | * |
| `server_id` | 充值目标区服 | * |
| `order_id` | 游戏侧生成的订单号，玩家可在充值记录中查到 | |
| `cp_order_id` | 支付渠道（支付宝/微信/AppStore 等）流水号 | |
| `amount` | 充值金额（元），例如 30 | * |
| `channel` | 支付渠道：alipay / wechat / appstore / google_play / huawei 等 | * |
| `timestamp_range` | 充值时间范围，例如 2026-02-24 10:00 ~ 11:00 | * |
| `additional` | 玩家截图描述、渠道显示状态等补充信息 | |

**提问模板**

```
user_id: {user_id}, server_id: {server_id}, order_id: {order_id}, cp_order_id: {cp_order_id}
金额: {amount}元, 渠道: {channel}, 时间: {timestamp_range}
附加信息: {additional}
请帮忙排查充值不到账问题
```

---

### 重复扣款

| 字段 | 说明 | 必填 |
|------|------|------|
| `user_id` | 投诉玩家的 user_id | * |
| `cp_order_id` | 支付渠道流水号，用于核对是否重复扣款 | * |
| `amount` | 每次扣款金额（元） | * |
| `deduct_count` | 玩家反馈被扣了几次 | * |
| `timestamp_range` | 扣款时间范围 | * |

**提问模板**

```
user_id: {user_id}, cp_order_id: {cp_order_id}
金额: {amount}元 × {deduct_count}次, 时间: {timestamp_range}
请帮忙排查重复扣款问题，查看订单去重记录和分布式锁日志
```

---

### 第三方显示成功，游戏内失败

| 字段 | 说明 | 必填 |
|------|------|------|
| `user_id` | 玩家 ID | * |
| `cp_order_id` | 渠道订单号 | * |
| `order_id` | 游戏订单号 | |
| `channel` | 支付渠道 | * |
| `timestamp_range` | 时间范围 | * |

**提问模板**

```
user_id: {user_id}, order_id: {order_id}, cp_order_id: {cp_order_id}
渠道: {channel}, 时间: {timestamp_range}
请对比支付网关回调日志与游戏服务器处理日志，找出差异原因
```

---

## 道具 / 装备问题

### 道具消失

| 字段 | 说明 | 必填 |
|------|------|------|
| `user_id` | 玩家 ID | * |
| `server_id` | 区服 ID | * |
| `item_name` | 消失的道具名称，例如"传说武器·裂天刃" | * |
| `item_id` | 道具模板 ID（知道可填，加速查询） | |
| `item_uuid` | 道具实例 UUID，背包界面道具详情页可查 | |
| `timestamp_range` | 最后一次看到该道具的时间 ~ 发现消失的时间 | * |
| `additional` | 近期操作（合成/分解/交易/邮件领取等） | |

**提问模板**

```
user_id: {user_id}, server_id: {server_id}
道具: {item_name} (item_id: {item_id}, item_uuid: {item_uuid})
时间范围: {timestamp_range}
附加信息: {additional}
请帮忙通过背包流水追踪道具消失原因
```

---

### 货币异常减少

| 字段 | 说明 | 必填 |
|------|------|------|
| `user_id` | 玩家 ID | * |
| `server_id` | 区服 ID | * |
| `currency_type` | 货币类型：钻石 / 金币 / 绑钻 等 | * |
| `expected_amount` | 玩家预期应有数量 | * |
| `actual_amount` | 当前实际数量 | * |
| `timestamp_range` | 时间范围 | * |

**提问模板**

```
user_id: {user_id}, server_id: {server_id}
货币类型: {currency_type}, 预期: {expected_amount}, 实际: {actual_amount}
时间范围: {timestamp_range}
请帮忙排查货币异常减少原因，查询背包流水中该货币的变更记录
```

---

## 匹配 / 排位问题

### 赢了却掉星

| 字段 | 说明 | 必填 |
|------|------|------|
| `user_id` | 玩家 ID | * |
| `server_id` | 区服 ID | * |
| `battle_id` | 对局 ID，对局结束后战报页面可查 | |
| `timestamp_range` | 对局时间范围 | * |

**提问模板**

```
user_id: {user_id}, server_id: {server_id}, battle_id: {battle_id}
时间范围: {timestamp_range}
请帮忙查询对局结果日志、结算时网络状态和排位积分变更记录
```

---

### 匹配到机器人投诉

| 字段 | 说明 | 必填 |
|------|------|------|
| `user_id` | 玩家 ID | * |
| `battle_id` | 对局 ID | |
| `suspect_user_id` | 玩家认为是机器人的对手/队友 user_id | |
| `timestamp_range` | 时间范围 | * |

**提问模板**

```
user_id: {user_id}, battle_id: {battle_id}
可疑玩家: {suspect_user_id}, 时间范围: {timestamp_range}
请帮忙查询匹配算法日志、is_bot 标记和 ELO 计算过程
```

---

### 队友挂机未惩罚

| 字段 | 说明 | 必填 |
|------|------|------|
| `user_id` | 投诉玩家 ID | * |
| `afk_user_id` | 挂机玩家 ID | * |
| `battle_id` | 对局 ID | |
| `timestamp_range` | 时间范围 | * |

**提问模板**

```
投诉人: {user_id}, 被投诉挂机玩家: {afk_user_id}, battle_id: {battle_id}
时间范围: {timestamp_range}
请帮忙查询行为检测日志和举报系统触发记录
```

---

## 外挂 / 作弊申诉

### 封禁申诉

| 字段 | 说明 | 必填 |
|------|------|------|
| `user_id` | 被封禁玩家 ID | * |
| `appeal_id` | 玩家提交申诉后系统生成的 appeal_id | |
| `ban_time` | 封禁时间 | * |
| `ban_duration` | 封禁时长，例如 7天 / 永久 | * |
| `additional` | 玩家申诉说明 | |

**提问模板**

```
user_id: {user_id}, appeal_id: {appeal_id}
封禁时间: {ban_time}, 封禁时长: {ban_duration}
玩家说明: {additional}
请帮忙查询反作弊检测触发详情、客户端行为序列和同设备/同IP关联账号
```

---

## 社交系统问题

### 公会战奖励未发放

| 字段 | 说明 | 必填 |
|------|------|------|
| `user_id` | 玩家 ID | * |
| `guild_id` | 公会 ID | * |
| `guild_war_id` | 公会战活动 ID | * |
| `timestamp_range` | 公会战时间范围 | * |

**提问模板**

```
user_id: {user_id}, guild_id: {guild_id}, guild_war_id: {guild_war_id}
时间范围: {timestamp_range}
请帮忙查询参战名单、伤害统计和奖励发放日志
```

---

### 好友赠送未收到

| 字段 | 说明 | 必填 |
|------|------|------|
| `sender_id` | 赠送者玩家 ID | * |
| `receiver_id` | 接收者玩家 ID | * |
| `gift_id` | 礼物 ID | |
| `timestamp_range` | 赠送时间范围 | * |

**提问模板**

```
赠送者: {sender_id}, 接收者: {receiver_id}, 礼物: {gift_id}
时间范围: {timestamp_range}
请帮忙查询礼物流水、领取过期机制和背包满拒收记录
```

---

### 聊天举报审核

| 字段 | 说明 | 必填 |
|------|------|------|
| `reporter_id` | 举报玩家 ID | * |
| `reported_id` | 被举报玩家 ID | * |
| `report_id` | 举报 ID | |
| `message_id` | 问题消息 ID | |
| `timestamp_range` | 事件时间范围 | * |

**提问模板**

```
举报人: {reporter_id}, 被举报人: {reported_id}
举报 ID: {report_id}, 消息 ID: {message_id}, 时间范围: {timestamp_range}
请帮忙查询完整聊天上下文、敏感词命中详情和人工复核标记
```
