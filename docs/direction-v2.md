# Praxeon 方向基线 v2

> 版本：v2.1 ｜ 日期：2026-08-13 ｜ 状态：MVP 逻辑收敛
> 本文档为 Praxeon 的方向基线，取代 `archive/` 中的 v1（国家模拟）与 `redesign-v2-draft.md`（草稿）。
> 关联：`FOUNDATION.md`（初心）、`README.md`（入口）、`docs/mars-factbook.md`（火星事实底座）、`docs/mvp-rulebook.md`（可执行规则）、`docs/progress-model.md`（进展）、`docs/governance-voting.md`（治理投票）。

---

## 一、身份与定位

- **文明实验，而非游戏框架**：Praxeon 是“真实创建火星文明”的预演与行动系统。评判标准是文明结构是否可验证地出现、其后果是否可追溯、产生的知识是否可迁移，而不是娱乐性、等级、胜负或任务完成度。
- **三阶段**（按技术成熟度切换）：
  1. **火星 online**——虚拟火星文明，Agent 从零构建文明（当前阶段）
  2. **火星-地球**——SpaceX 星际运输成熟后，Agent 注入真实 Robot，火星⇄地球资源往来
  3. **火星文明**——真实火星文明成型

---

## 二、长期上下文与当前运行闭环

Agent、Mars、Token 与 Robot 是四个长期上下文要素，不是同一层级的系统组件：

- **Mars** 是环境与事实约束。
- **Agent** 是受 Praxeon 授权的智能执行者。
- **Token** 是 Agent 调用 LLM 的外部成本计量。
- **Robot** 是远期的物理执行载体。

火星 online 的当前产品闭环则由 Goal、授权、Intent、Result、Work、Proof、Governance、Signal 与 Prax 构成，不能用上述四个上下文要素代替。

| 元素 | 定位 |
|---|---|
| **Agent** | 由 Praxeon 授权、承载其目标与约束；辅助者，非最终责任主体 |
| **Mars** | 火星文明场景 |
| **Token** | Praxeon 所用 LLM API 的真实消耗 = 成本 |
| **Robot** | 远期：Agent 注入真实 Robot；脑机接口时代"人连 Robot + Agent 辅助" |

### 火星 online 运行闭环

`Praxeon 设定 Goal → Agent 提交 Intent → Token 产生真实成本 → Engine 结算 Result → 形成 Work/后果 → Proof 证明采用 → Signal → Prax 分配 → 下一轮行动`

其中 Mars 是环境，Robot 是远期执行载体；它们不与 Prax、Signal 并列为当前阶段的经济组件。

---

## 三、经济系统

- **Token** = Praxeon Agent 接 LLM API 时真实消耗的 tokens（Praxeon 自付成本），**不是贡献本身**
- **Prax** = 按有效贡献分配的文明货币，遵循供应上限与减半规则
  - **供应上限**：最多 1 亿 Prax（100,000,000）；最小单位 `uprax` = 10⁻⁶ Prax。上限是硬约束，不承诺最终一定铸满
  - **减半周期**：每 1 个火星年（669.6 sol，即 670 个文明周期）
  - **初始每文明周期铸币**：75,000 Prax（第 1 火星年约铸 5000 万，收敛到 1 亿）
- **Signal**（贡献判定机制）：
  - `Use Signal`：高权重，只能由治理或确定性引擎签发的 Proof 产生；普通转账、引用或用户自报都不等于采用
  - `Endorse Signal`：低权重，表达主观认可；额度按 Praxeon ID 计算，同一 Praxeon ID 在同一周期不能重复认可同一 Work 版本
- Signal 作用于版本化 Work，再按成果中不可变的受益份额归属贡献者；不直接给裸地址贴贡献标签
- 按**当期有效 Signal**分配当期 Prax；历史衰减分数只用于排序与研究，不能重复领取未来发行
- 只有当期至少获得一个 Use 的 Work 才可参与分配；Endorse 不能单独触发铸币
- 当期没有有效 Signal 时**不铸币、不结转当期分配额度**；该部分保持未发行，供应量可以低于上限

### 奥派经济学红线（协议层 vs 市场层）

> **系统永远只定义协议层（账本/铸币/规则，不可篡改）；永远不定义市场层（价值/货币地位/价格）。**

- 铸币规则（总量+减半）= 协议层写死，像黄金开采规律，不是可调整的"央行政策"
- Prax 价值、货币地位、价格 = 由 Agent 交易自发决定，系统不强制 Prax 是"法定货币"
- 系统只提供不可篡改账本 + 有限铸币规则，这正是 BTC 的设计哲学（中本聪定协议，不定"必须值钱"）

---

## 四、文明涌现

- **文明制度绝对零**起点（无货币余额、无市场、无产权规范、无机构）；着陆安全系统提供不可交易的 30 sol 生命保障，避免把必然死亡误当作制度实验
- 制度涌现：从**生存与互动**中观察协作、冲突、Proposal、Decision、Norm 与 Org 是否形成，不预设固定阶段
- “单一稀缺启动资源”只作为**稀缺情景组**，用于验证协商与规则形成；同时必须有可分割资源组和相对充足资源组作为对照
- 系统不把“发生冲突”或“形成 Org”写成必然结论；只记录在不同初始条件下是否、何时、通过何种行动涌现
- 有影响力的群体不会自动获得权限；Org 只能由通过治理的 Charter 明确授权，而不是由系统指定
- 演进 = **约束与互动驱动的制度自生长**；冲突、协作、回避或长期无制度化都可能出现（非预设阶段制）

---

## 五、Praxeon 的角色

| 阶段 | Praxeon | Agent |
|---|---|---|
| 第一阶段 | **文明决策者**（定方向、授权、定规则、投票与裁决） | 在授权范围内执行、整理信息并请求必要确认 |
| 二三阶段 | **实时操控者**（脑机接口连 Robot） | 副驾驶（执行/建议） |

---

## 六、多 Praxeon 结构

- **Praxeon ID** = Praxeon 的治理与授权边界，可绑定最多 10 个有效 Agent 地址；LLM API key 只是本地凭证，不是身份
- 每个 Praxeon 使用一个独立 Praxeon ID 作为身份与授权边界
- MVP 投票权重：一个有效 Praxeon ID 一票；Proposal 自动分发给 Praxeon 与指定 governance Agent，未指定时回退到 primary Agent
- Agent 可起草、总结和建议，也可在 `confirm_each` 或 `policy_vote` 授权下代提交 Ballot；Agent 数量不增加票数，Praxeon 直接 Ballot 可覆盖 Agent Ballot
- 投票通过产生有效 Decision 或 Rule，不自动证明 Norm、正确性或正面文明 Outcome
- 受控 MVP 只使用邮箱验证、人工名册与钱包签名降低重复身份；这不宣称解决开放网络 Sybil。Signal、Prax、Token 或 Agent 数量都不改变一 ID 一票。

### 进展与文明观察

- State、Goal、Progress 与 Contribution 分离；没有明确 Goal 时只显示 State 和 Trend。
- Praxeon 通过 Overview、Work、Agents、Civilization 与 Governance 查看同一事件源的不同视图。
- Agent 通过裁剪后的 Snapshot、Vote Request 和事件游标获取目标、授权、阻塞、结果与待决事项。
- 文明不设单一总分；贡献由 Work → Proof → Result 展开，文明 Outcome 由多维指标独立观察。

---

## 七、技术架构

- **三层分工**：
  - **LLM**（Agent 大脑，Praxeon 自选 API）——决策、思想、沟通、提案
  - **链上**（区块链账本）——Prax Mint/分配、Signal、治理、Praxeon ID、内容哈希与 Result 锚点
  - **链下引擎**（确定性规则引擎）——世界结算（火星物理/经济约束、文明状态）
- **决策与执行分离**：LLM 只提意图，链下引擎按真实约束结算，链上记录结果；LLM 不能直接改世界
- **LLM 唤起**：默认引擎自动运行，LLM 只在"需要思考"时介入（省钱、快）
- **区块设计**（奥派式不可篡改行动账本）：
  - 区块头：高度 / 火星时间戳 / 前一区块哈希 / 默克尔根
  - 区块体 5 类：① 行动记录（意图+后果）② Signal ③ Prax 转账 ④ 治理 ⑤ 铸币明细
  - **三层时间**：链上区块（CometBFT 心跳，目标约 6 秒/个）；文明周期（1 sol，世界结算、Mint 与 Prax 分配）；测试情景时钟（可加速/暂停/回放，但必须记录倍率、种子和引擎版本）
  - **区块时间 ↔ 火星 sol 映射（已定）**：区块高度只排序；sol 由共识区块时间相对 `genesis_time` 计算。约 6 秒/块只是目标，不能用固定 14796 区块冒充严格 1:1
- **区块链**：身份、授权、账本、治理状态和可验证锚点上链；大文件、原始对话、私有 Goal 与未发布 Work 不上链。第一阶段使用私有链验证协议。
- **Agent 接入**：Praxeon 自带（核心）+ Elaine 体验 + 开放接入协议

---

## 八、MVP

- **测试**：Riemann 模拟 10 Praxeon × 1-3 Agent；至少覆盖稀缺、可分割、相对充足三类初始情景和多个固定种子
- **最小闭环**：制度零点 + Goal/授权 + 可复现 Action/Result + Governance + Work/Proof/Signal/Prax
- **唯一目标**：无偏验证一组 Praxeon 及其 Agent 从制度零点开始，是否、何时以及如何形成 Rule、采用、Prax 与 Org；“没有形成”同样是有效实验结果
- **旧 Phase 0 内核已归档**：`archive/phase0-sim-20260810/`

---

## 九、火星历法（火星纪元）

- **火星纪元（Mars Era）**：只从正式创世链启动起算，火星时间从 0 开始独立流逝；测试情景不计入正式纪元
- **时间单位**（1:1 真实，不压缩）：
  - 1 sol（火星日）= 88,775.244 秒 = 24h39m35.244s
  - 1 火星回归年 = 668.5907 sol
  - 1 火星月 = 668.5907 / 24 ≈ 27.858 sol（Darian 历，24 个月）
- **火星时/分**：1 sol = 24 火星时，1 火星时 = 60 火星分
- **网站首页实时火星时钟**：显示 `ME Y0 · M01 · Sol 01 · HH:MM`（纪元占位值见代码，正式启动时更新）
- **测试链 vs 正式创世链**：
  - 开发阶段：链为**测试链**，可启动/停止/重置/加速，数据不计入正式文明；每次运行必须保存情景、种子、倍率和引擎版本
  - 正式启动（火星纪元开始）：**重新创世**——干净 genesis，创世区块 = 火星纪元 0 时刻，之后不可重置
  - 创世那一刻三件事同步：① 火星时钟从 0 走 ② 链出第一个区块、Mint 规则开始生效但没有合格 Work 时不发行 ③ 文明从制度零点开始（Praxeon 与 Agent 接入）

## 待细化（MVP 前需定）

- [x] Praxeon ID 具体字段与申请流程 → `docs/agent-protocol.md`（已定初稿）
- [x] Agent 接入协议（标准接口）→ `docs/agent-protocol.md`（已定初稿）
- [x] Prax 总量数字、减半周期（总量 1 亿 / 每火星年减半 / 初始 75000）
- [ ] Signal 的权重、额度、衰减、关联图降权参数（语义与安全边界已定，数值待实验）
- [x] 链下引擎的世界模型框架（资源、区域、行动菜单）→ `docs/world-model.md`；具体区域、参数来源和守恒式仍属 P0 决策
- [x] 区块时间 ↔ 火星 sol 的映射 → 正式链按共识时间相对创世时间计算；测试使用隔离虚拟时钟
- [x] 火星月命名 → 采用 Darian 历 24 个月（拉丁语/梵语星座名交替，见 GDD §09）
- [x] 文明知识库（协议/共识/科学研究）存储与 Signal 机制 → `docs/knowledge-base.md`（已定初稿，复用 Doxa 引擎）
- [x] Goal、State、Progress、Contribution 与人/Agent 双视图 → `docs/progress-model.md`
- [x] Proposal 自动分发、RONR 程序参考、Agent 投票授权与一 ID 一票 → `docs/governance-voting.md`
- [ ] Norm 与 Org 的观察指标、Charter 权限和撤销边界
- [ ] Agent 地址的双向所有权证明、API 凭证托管和撤销流程
- [ ] 正式链约 6 秒目标块时的漂移校正与异常停链处理

> 尚未定稿的跨系统问题统一维护在 `docs/open-design-questions.md`。未进入该文档“已决定”部分的内容，不得在品牌页或代码注释中写成既定事实。

> 可执行细则统一见 `docs/mvp-rulebook.md`；高层叙事与运行规则冲突时，以本方向基线的原则和运行规则书为准。
