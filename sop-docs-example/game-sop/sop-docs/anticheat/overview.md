# 外挂 / 作弊误判申诉 SOP

反作弊系统（Anti-Cheat Engine）通过**客户端行为采集 → 服务端规则引擎 → 机器学习模型**三层检测机制，识别并处置外挂/作弊行为。申诉排查的核心是还原触发处罚的完整证据链，判断是否存在误判。

---

## 检测架构

### 层一：客户端行为采集

在客户端记录操作行为序列，上报至服务端分析（存储在 `behavior_sequence_log` 中）：

- 移动速度（`move_speed`）
- 点击频率（`click_frequency`）
- 鼠标/手势轨迹（`input_trace`）
- 内存修改检测（`memory_scan_result`）
- 进程注入检测（`process_inject_detect`）
- 帧时间间隔（`frame_delta_ms`）

### 层二：服务端规则引擎

基于预定义阈值规则实时检测异常行为（触发结果存储在 `cheat_detection_log` 中），各规则 ID、检测逻辑与阈值见 [game-audit-logstore → rule_id / rule_name](stores/game-audit-logstore.yaml)。

### 层三：机器学习模型

综合行为序列和历史数据给出置信度评分（`ml_score` 字段），各区间含义及对应处置见 [game-audit-logstore → ml_score](stores/game-audit-logstore.yaml)。

---

## 处罚等级

处罚等级由 `action` 字段记录，取值及触发条件见 [game-audit-logstore → action](stores/game-audit-logstore.yaml)。

---

## 排查步骤

### 场景一：外挂误判申诉排查

> ⚠️ **高风险操作**：必须有完整证据链支持才可解封，不可仅凭玩家说辞操作。

* **Step 1 — 查询处罚记录详情**：参考 [game-audit-logstore → 查询封禁处罚记录（反作弊申诉核查）](stores/game-audit-logstore.yaml)
* **Step 2 — 分析触发规则的合理性**：根据 `rule_id` 选择对应核查思路：
  * `speed_hack_001`：确认 `detected_value` 与 `threshold` 差距（超出 < 10% 可疑误判）；核查是否有服务端延迟/网络抖动导致移速计算失真；对比同对局其他玩家移速，排除全局异常
  * `aim_bot_001`：确认 `sample_count` 是否足够（< 30 时统计不可靠）；核查玩家历史 `headshot_rate` 趋势（高技术玩家可能长期保持高爆头率）；核查是否使用高帧率设备/专业外设
  * `auto_click_001`：确认 `input_interval_stddev` 测量时段（手柄/宏按键 stddev 也会偏低但属合法设备）；核查玩家设备类型
* **Step 3 — 查询客户端行为序列**：参考 [game-audit-logstore → 查询行为序列（误判核查）](stores/game-audit-logstore.yaml)。`memory_scan_result` 各值含义见 [`game-audit-logstore` → memory_scan_result](stores/game-audit-logstore.yaml)
* **Step 4 — 查询同设备/同 IP 关联账号**：参考 [game-audit-logstore → 查询同设备封禁账号数量](stores/game-audit-logstore.yaml)
* **Step 5 — 查询历史处罚记录**：参考 [game-audit-logstore → 查询玩家历史处罚汇总](stores/game-audit-logstore.yaml)
* **Step 6 — 综合判断与处理**：

  | 条件 | 建议处理 |
  |------|---------|
  | `ml_score < 0.7` AND `memory_scan_result=clean` AND 无历史处罚 | 建议解封，补偿封禁期间损失（道具/积分） |
  | `ml_score 0.7~0.9` AND 单次触发 AND 网络延迟异常 | 缩短封禁时长，转为警告级别，持续监控 |
  | `ml_score > 0.9` OR `memory_scan_result=infected` OR 多规则同时触发 | 维持封禁，告知玩家触发原因（脱敏后），提供二次申诉通道 |
  | 同设备多账号封禁 AND `ml_score > 0.8` | 维持封禁，关联设备列入高风险监控 |

---

### 场景二：查询玩家申诉状态

参考 [game-audit-logstore → 查询申诉状态](stores/game-audit-logstore.yaml)

> `appeal_status` 各值含义及 SLA 见 [`game-audit-logstore` → appeal_status](stores/game-audit-logstore.yaml)

---

## FAQ

**Q：玩家说从来没有外挂，游戏时间也不长，怎么解释高爆头率？**

- 高游戏技术本身不构成作弊，需结合 `memory_scan_result` 和多规则触发情况判断
- 如果只有单一指标异常且其他检测均为 `clean`，建议重新审查规则阈值设置
- 可对比该玩家与同段位前 1% 玩家的数据分布，判断是否属于统计异常

---

**Q：玩家说别人举报他太多次所以被封了？**

举报次数**不直接触发**封禁，封禁由反作弊检测引擎独立判断。举报只会触发系统优先审查该玩家，最终处罚依据是技术检测数据。

---

**Q：玩家说使用了外设/宏但不是外挂怎么处理？**

- 硬件外设（鼠标宏/键盘连点）属于游戏协议的灰色地带，不同游戏政策不同
- 需查阅游戏《用户协议》中对于硬件辅助的明确规定
- 若游戏明确禁止，即使是硬件宏也视为违规；若无明文禁止，可酌情解封但记录

---

**Q：永久封禁后玩家的充值金额怎么处理？**

- 原则上已消费的虚拟货币不予退款（游戏规则内消费属于正常行为）
- 未消费的钻石是否退款需依照游戏运营政策，部分情况下可扣除作弊所得后按比例退款
- 具体退款走财务流程，需提交内部申请

---

## 注意事项

- 反作弊日志属于高敏感数据，具体规则阈值和算法细节**不可对外披露**（防止外挂开发者绕过）
- 封禁操作不可直接由客服执行，必须经过反作弊团队审核
- 解封操作需要高级客服或反作弊团队二次确认，并在 [`game-audit-logstore`](stores/game-audit-logstore.yaml) 的 `gm_op_log` 中记录操作员 ID 和原因
- 申诉处理 SLA：普通申诉 3 个工作日，高级申诉（永久封禁）5 个工作日
- 一个账号最多可提交 3 次申诉，超过次数后需走特殊通道
