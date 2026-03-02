# 核心概念

## 玩家账号（Player Account）

> 别名：用户账号、user

游戏中玩家的唯一身份标识，通过 `user_id` 全局唯一定位。

- 一个玩家账号可以绑定多个登录方式（手机号、第三方 OAuth、游客 ID）
- 账号状态：`active`（正常）/ `banned`（封禁）/ `frozen`（冻结）/ `deleted`（注销）
- 玩家在不同区服（`server_id`）有独立的角色（role）数据，但账号级数据（充值记录、绑定信息）全局共享

**关键字段**

| 字段 | 说明 |
|------|------|
| `user_id` | 全局唯一玩家 ID |
| `role_id` | 区服内角色 ID，格式：`{server_id}_{user_id}` |
| `account_status` | 账号状态 |
| `platform` | 登录平台（ios / android / pc） |
| `channel` | 渠道来源（appstore / google_play / huawei / oppo 等） |

---

## 区服（Game Server）

> 别名：server、大区

游戏服务器的逻辑分区，每个区服有独立的玩家社区和经济系统。

- `server_id` 唯一标识一个区服，格式一般为 `region_code + server_seq`，例如 `cn_s1`
- 跨区服操作（如转服）会产生额外的迁移日志
- 合服（server_merge）会将多个区服数据合并，需特别注意 `role_id` 的映射关系

**关键字段**

| 字段 | 说明 |
|------|------|
| `server_id` | 区服唯一 ID |
| `server_status` | 区服状态（running / maintenance / merged） |
| `region` | 物理部署地域 |

---

## 订单（Order）

> 别名：充值订单、支付订单

玩家充值行为产生的唯一订单，贯穿支付全流程。

- 状态流转：`created → paying → paid → delivering → delivered / failed / refunded`
- `order_id` 由游戏服务器生成，`cp_order_id` 是第三方支付渠道的订单号
- 幂等校验基于 `order_id` 实现，防止重复发货
- 退款状态：`refund_pending → refunded / refund_failed`

**关键字段**

| 字段 | 说明 |
|------|------|
| `order_id` | 游戏侧订单 ID（全局唯一） |
| `cp_order_id` | 第三方支付渠道订单号 |
| `user_id` | 下单玩家 |
| `server_id` | 下单区服 |
| `amount` | 充值金额（分为单位） |
| `currency` | 货币类型（CNY / USD 等） |
| `channel` | 支付渠道（alipay / wechat / appstore / google_play 等） |
| `product_id` | 充值商品 ID |
| `status` | 订单状态 |
| `retry_count` | 回调重试次数 |
| `idempotent_key` | 幂等键 |

---

## 物品（Item）

> 别名：道具、装备、货币

游戏内所有可持有资产的统称，包括装备、消耗品、货币（钻石/金币）等。

- 每件物品有唯一的 `item_uuid`（实例级），同时有 `item_id`（模板级，表示物品类型）
- 生命周期：**获得（acquire）→ 持有（hold）→ 使用/消耗（consume/use）→ 丢弃/分解/交易（drop/dismantle/trade）**
- 背包（bag）是玩家持有物品的容器，有容量上限

**关键字段**

| 字段 | 说明 |
|------|------|
| `item_uuid` | 物品实例 UUID（全局唯一，一件具体装备对应一个 uuid） |
| `item_id` | 物品模板 ID（同种道具共享） |
| `item_name` | 物品名称 |
| `item_type` | 物品类型（weapon / armor / consumable / currency / material 等） |
| `rarity` | 稀有度（common / rare / epic / legendary） |
| `owner_id` | 当前持有者 user_id |
| `acquire_type` | 获得方式（drop / purchase / craft / mail / activity / trade 等） |
| `acquire_time` | 获得时间戳（Unix ms） |

---

## 背包流水（Bag Transaction）

> 别名：物品变更流水、道具流水

记录玩家背包中每一次物品增减变化的**不可变**流水日志，是排查道具消失问题的核心数据源。

**关键字段**

| 字段 | 说明 |
|------|------|
| `txn_id` | 流水 ID（全局唯一） |
| `user_id` | 玩家 ID |
| `role_id` | 角色 ID |
| `item_uuid` | 物品实例 ID |
| `item_id` | 物品模板 ID |
| `change_type` | 变更类型（add / remove / transfer） |
| `change_reason` | 变更原因（purchase / consume / expire / admin_op / craft / trade / drop 等） |
| `qty_before` | 变更前数量 |
| `qty_after` | 变更后数量 |
| `delta` | 变化量（正数增加，负数减少） |
| `source_system` | 来源系统（pay / craft / battle / mail / shop / gm_tool 等） |
| `timestamp` | 操作时间戳（Unix ms） |
| `operator` | 操作者（user_id 或 system / gm_uid） |

---

## 匹配（Matchmaking）

> 别名：组队匹配、天梯匹配

将玩家按照技术水平（ELO/MMR）和偏好（模式/地区）分组进入对局的系统。

- 匹配请求产生 `match_request_id`，完成后生成 `match_id`，对局结束后生成 `battle_id`
- 排位系统在对局结束后根据 `battle_result` 更新段位积分（`rank_points`）

**关键字段**

| 字段 | 说明 |
|------|------|
| `match_request_id` | 匹配请求 ID |
| `match_id` | 匹配成功 ID |
| `battle_id` | 对局 ID |
| `user_id` | 玩家 ID |
| `mode` | 游戏模式（ranked / casual / custom） |
| `mmr_before` / `mmr_after` | 匹配前/后 MMR 值 |
| `rank_points_before` / `rank_points_after` | 排位积分（赛前/赛后） |
| `battle_result` | 对局结果（win / lose / draw / abandoned） |
| `settlement_status` | 结算状态（success / failed / rollback） |
| `is_bot` | 是否机器人玩家（true / false） |

---

## 反作弊系统（Anti-Cheat System）

> 别名：反外挂、ACE

检测和处置游戏内作弊行为的系统，包括速度外挂、透视、自动瞄准等。

- 触发处罚时产生 `cheat_event_id`，记录触发规则、阈值和客户端行为数据
- 处罚等级：**警告（warn）→ 临时封禁（temp_ban）→ 永久封禁（perm_ban）**
- 申诉流程：`appeal_created → under_review → approved`（解封）/ `rejected`（维持处罚）

**关键字段**

| 字段 | 说明 |
|------|------|
| `cheat_event_id` | 作弊事件 ID |
| `user_id` | 被检测玩家 |
| `rule_id` | 触发规则 ID |
| `rule_name` | 规则名称（speed_hack / aim_bot / wall_hack / packet_manipulation 等） |
| `detected_value` | 检测到的异常数值 |
| `threshold` | 规则阈值 |
| `confidence` | 置信度（0~1） |
| `action` | 执行的处罚动作 |
| `ban_duration` | 封禁时长（秒，-1 表示永久） |
| `appeal_id` | 申诉 ID |
| `device_id` | 设备 ID |
| `ip_address` | 客户端 IP |

---

## 社交系统（Social System）

> 别名：公会、好友、聊天、举报

包含公会（guild）、好友（friend）、聊天（chat）、举报（report）等社交功能，每次社交操作产生对应的流水日志。

**关键字段**

| 字段 | 说明 |
|------|------|
| `guild_id` | 公会 ID |
| `event_id` | 事件 ID |
| `sender_id` | 发送者 user_id |
| `receiver_id` | 接收者 user_id |
| `gift_id` | 礼物模板 ID |
| `gift_uuid` | 礼物实例 ID |
| `message_id` | 消息 ID |
| `report_id` | 举报 ID |
| `channel_id` | 聊天频道 ID（world / guild / private） |

---

> 日志分层架构、各 Logstore 的字段说明及常用查询语句，详见 [stores/README.md](./stores/README.md)。
