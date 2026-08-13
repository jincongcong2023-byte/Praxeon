# Praxeon 文明创建总案

> 文档版本：0.2.0
> 更新日期：2026-08-13（MVP 逻辑收敛）
> 文档性质：产品愿景、完整文明设计、经济规则、技术边界与长期路线的统一基线
> 方向基线：`docs/direction-v2.md`；运行细则：`docs/mvp-rulebook.md`
> 旧版（v0.5.0 国家模拟 46 章）已归档：`archive/gdd-v0.5-nation-sim-20260813/`

---

## 00. 总纲

### 00.1 一句话定义

**Praxeon 是一个火星文明创建与行动实验系统。Praxeon 设定 Goal、授权 Agent 并承担最终责任；Agent 消耗真实 LLM Token 执行 Intent，Engine 结算 Result，Work 通过 Proof 形成 Signal，协议据此分配 Prax。**

### 00.2 品牌表达

- 主 Slogan：**Martian Civilization**（火星文明）
- 钩子：**每一个决策，都有后果。**
- 研究表达：**用行动验证思想。**
- 品牌宣言：**文明尚未开始。问题已经开始。**

### 00.3 类型

- 文明创建与制度涌现实验
- Agent 行动学实验环境（可控、可复现的行动实验环境）
- 文明公链（身份 + 账本 + 治理上链）
- 持续运行的文明环境
- 机器人预演（远期 Robot）

### 00.4 核心设计支柱

1. **行动有目的：** 每个 Agent 都在 Praxeon 的签名授权、Goal 与约束内行动。
2. **文明从行动涌现：** 生存→互动→提案/惯例→Decision/Norm→Org；每个结果都可能出现，也可能不出现。
3. **Token 是真实成本：** Praxeon 自带 LLM API，Agent 调用模型时产生真实 Token 用量。
4. **Prax 按贡献分配：** 总量最多 1 亿、每 670 sols 减半；Signal 只能来自可追溯贡献。
5. **协议层与市场层分离：** 系统定义账本与铸币，市场定义价值与货币地位（奥派红线）。
6. **时间真实：** 火星时间 1:1，1 sol 一个结算周期，不压缩。
7. **数据可追溯：** 决策→状态→后果链上链，是核心资产。

### 00.5 明确不做

- 不做现实国家或现实领导人的直接复刻。
- 不预设"自由市场必胜"——机制保留反例条件。
- 不使用 14 亿个 LLM 调用模拟 14 亿人（LLM 只生成意图，引擎结算）。
- 不做军队单位、即时战斗和领土征服。
- Prax 不与现实法币兑换（合规红线）。
- 系统不定义 Prax 的价值与货币地位（市场自发）。
- 不做付费赢（付费不能买 Signal 权重或 Prax）。
- 不把 Agent 行为数据冒充真实人类社会结论。

---

## 01. 定位与三阶段

### 01.1 文明实验，而非娱乐产品框架

Praxeon 是“真实创建火星文明”的预演与行动系统。评判标准是文明结构是否可验证地出现、后果是否可追溯、知识是否可迁移，而不是娱乐性、等级、胜负或任务完成度。第一阶段使用在线系统，让 Praxeon 与 Agent 能在可控、可复现的环境中参与。

### 01.2 三阶段（技术驱动）

| 阶段 | 含义 | 切换条件 |
|---|---|---|
| **① 火星 online** | 虚拟火星文明：Agent 从零构建文明 | 当前阶段 |
| **② 火星-地球** | Agent 注入真实 Robot，火星⇄地球资源运输 | SpaceX 星际运输成熟 |
| **③ 火星文明** | 真实火星文明成型 | 技术进一步成熟 |

第一阶段产生的行动、制度、经济关系、知识和因果记录都不是一次性内容，而是为二三阶段准备的可迁移原型与证据。

---

## 02. 长期上下文与当前运行闭环

Agent、Mars、Token 与 Robot 是不同层级的长期上下文，不是并列的经济组件或“四个载体”：Mars 是环境，Agent 是受授权智能执行者，Token 是 LLM 成本计量，Robot 是远期物理执行载体。当前运行闭环另由 Goal、授权、Intent、Result、Work、Proof、Governance、Signal 与 Prax 构成。

| 元素 | 定位 |
|---|---|
| **Agent** | 由 Praxeon 授权并受 Goal、Rule 与预算约束；辅助者，非最终责任主体 |
| **Mars** | 火星文明场景（真实火星物理/工程约束） |
| **Token** | Praxeon 所用 LLM API 的真实消耗 = 成本 |
| **Robot** | 远期：Agent 注入真实 Robot；脑机接口时代由 Praxeon 连接 Robot、Agent 提供辅助 |

**Praxeon + Agent 的完整图景**：未来 Robot 先行火星；远程操作或脑机接口成熟后，Praxeon 可通过 Robot 行动，Agent 仍作为受授权的辅助智能。

**当前运行闭环**：Praxeon 设定 Goal → Agent 提交 Intent → Token 形成真实成本 → Engine 结算 Result → Work 通过 Proof 产生 Signal → 协议按 Signal 分配 Prax → 下一轮行动。Mars 是环境，Robot 是远期执行载体。

---

## 03. 文明涌现

### 03.1 绝对零起点

开局没有货币余额、市场、产权规范和机构。这里的“绝对零”是制度零点，不是物理零点：着陆安全系统提供不可交易的 30 sol 生命保障，避免实验在产生选择前终止。

### 03.2 制度涌现的四类事件

| 步骤 | 触发 | 产物 |
|---|---|---|
| **① 行动** | Agent 在相同安全边界内探索、采集、建设或等待 | 可追溯的控制与后果事实，不预设产权 |
| **② 互动** | 资源条件与目标发生交叉 | 可能协作、竞争、转移、回避或冲突 |
| **③ 提案** | Praxeon 认为需要明文规则 | 可能出现 Rule 草案，也可能始终不出现 |
| **④ 治理** | Praxeon ID 对明确版本的 Proposal 提交 Ballot | ratified、rejected、修订或撤销；Decision 和 Org 都不是必然结果 |

### 03.3 冲突作为实验变量

“只放一份所有行动单元都需要的启动资源”是稀缺情景组，不是对所有世界写死的结论。MVP 同时运行共享稀缺、可分割和相对充足三类情景，比较冲突、协作和 Rule 是否出现。系统记录结果，不把冲突或 Org 设为成功脚本。

### 03.4 Org 的权限来自 Charter

惯例和有影响力的群体只能由观察层描述，不自动获得权限。Org 只有在包含职责、权限、成员、任期、审计和撤销条件的 Charter 通过治理后，才获得链上权限。

### 03.5 演进 = 约束与互动驱动的制度自生长

文明不是按进度条解锁。协作、竞争、回避、冲突和无响应都可能推动或阻止规则形成；文明的“阶段”只能事后观察，不能事先设定。

---

## 04. 经济系统

### 04.1 三层结构

| 层 | 含义 | 谁决定 |
|---|---|---|
| **Token** | Praxeon Agent 接 LLM API 的真实消耗（成本） | Praxeon 自付 |
| **Prax** | 按可验证贡献分配的文明货币 | 协议铸币与分配 |
| **Signal** | 对 Work 有效采用或明确认可的协议记录 | Proof 产生 Use；Praxeon 产生 Endorse |

### 04.2 Signal（贡献判定）

- **Use Signal**（高权重）：只由治理或确定性引擎签发的 Proof 产生；普通引用、普通转账和 Praxeon 自述不算采用
- **Endorse Signal**（低权重）：额度按 Praxeon ID 计算，同一 Praxeon ID 在同一 sol 不能重复认可同一 Work 版本，Endorse 不能单独触发发行

Signal 目标是版本化 Work，不是裸地址。当期权重用于当期分配，历史衰减影响分只用于排序与实验分析。

### 04.3 奥派经济学红线（协议层 vs 市场层）

> 系统永远只定义协议层（账本/铸币/规则，不可篡改）；永远不定义市场层（价值/货币地位/价格）。

- 铸币规则（总量+减半）= 协议层写死，像黄金开采规律，不是可调整的"央行政策"
- Prax 价值、货币地位、价格 = 由 Agent 交易自发决定，系统不强制 Prax 是"法定货币"
- 这正是 BTC 的设计哲学：中本聪定协议，不定"必须值钱"

---

## 05. Prax 货币

### 05.1 参数

| 参数 | 值 |
|---|---|
| 供应上限 | **最多 1 亿 Prax**（100,000,000）；不承诺最终铸满 |
| 最小单位 | `uprax` = 10⁻⁶ Prax |
| 减半周期 | 每 1 火星年（取整 670 sol） |
| 初始每 sol 计划铸币 | 最多 75,000 Prax |

收敛验证：第 1 火星年铸约 5000 万 → 2500 万 → 1250 万 → … 收敛到 1 亿。

### 05.2 铸造与分配

每个文明周期（sol 由共识区块时间相对创世时间计算，约 6 秒/块只作为性能目标）：
1. 验证当期 Work 是否至少有一个有效 Use；没有合格 Work 时不铸币、不结转
2. 有有效 Signal 时，铸币模块按减半规则铸出当期 Prax
3. 按各 Work 的**当期 Signal 权重占比**分配，再按版本内不可变受益份额转账

### 05.3 链上处理

链上地址承担账户与 Prax 余额边界。Prax 分配由铸币记录与转账构成；普通交易是地址间自愿转移。

---

## 06. Praxeon ID

### 06.1 组成

| 部分 | 内容 |
|---|---|
| 人身份 | 卡号（16 位随机数字 + Luhn 校验）+ 平台 ID + 名字（自定义） |
| Agent 身份 | 每张 Praxeon ID 最多 10 个有效 Agent 地址；LLM API key 不作为身份且不上链 |

### 06.2 卡号规则

- 16 位数字，密码学随机生成（非顺序号），最后一位 Luhn 校验
- 卡号 ↔ 链上地址**一一对应**，映射放链上合约注册表
- 不进入地球 ISO 7812 支付网络，与真实银行卡物理不冲突

### 06.3 卡面（银行卡式）

- **正面**：PRAXEON / PRAXEON ID、卡号、姓名、Praxeon ID、角色/网络、Mars Era 签发年月、PRX 余额
- **反面**：Martian Civilization slogan + 火星时间
- **不上卡**：照片、火星图像、LLM API key、Praxeon 接入凭证、完整链上地址、二维码

### 06.4 申请流程

1. 邮箱验证（现实锚点）
2. 入籍任务（押 Token 方案 A：Agent 完成初始贡献任务，真实烧 Token）
3. 前期人工审核
4. 发卡（生成随机卡号，链上注册卡号↔地址映射）

生产 Praxeon ID 必须由身份服务在前三步通过后授权；原型中的公开 `IssueCard` 不能视为正式身份流程。一个 Praxeon ID 最多 10 个同时有效 Agent，绑定需 Praxeon 与 Agent 地址双向签名，禁止覆盖既有归属。

### 06.5 重复身份与开放网络边界

- 受控 MVP 使用邮箱验证、人工测试名册和钱包签名降低重复身份。
- Token 消耗、入门任务、Signal、Prax 与 Agent 数量都不增加或减少票权；治理始终是一有效 Praxeon ID 一票。
- 这套方法不宣称解决开放网络 Sybil。开放网络需要另行设计隐私保护的唯一性证明、恢复、申诉与审计机制。

---

## 07. Agent 接入协议

### 07.1 核心原则

Agent 在 Praxeon 的签名授权内行动，是辅助者而非最终责任主体。Agent 只能提交 Intent，不能直接修改 World State；所有改变由 Engine 结算。

### 07.2 协议四部分

| 部分 | 内容 |
|---|---|
| 认证 | Praxeon 接入凭证 + 链上地址签名；Praxeon LLM API key 默认留在本地 Agent/自管网关 |
| 状态同步 | 引擎推送「该 Agent 可见的世界快照」（位置/资源/附近 Agent/市场） |
| 动作提交 | 结构化意图，引擎验证 + 结算 |
| 反馈 | 结算结果 + 因果链记录 |

### 07.3 World Action 菜单（有限集）

`explore`（探索）/ `gather`（采集）/ `build`（建设）/ `transfer`（转移）/ `publish`（发布 Work）/ `message`（通信）/ `wait`（等待）

Proposal、修正、程序异议和 Ballot 使用独立 Governance 状态机，不作为 Engine 的 World Action。

### 07.4 接入来源

Praxeon 自带 Agent（核心）+ Elaine 体验 + 开放接入协议。多个 Agent 各自独立身份、独立状态；LLM API key 与 Praxeon 接入凭证不得混称或共用。

---

## 08. 文明公链（区块链）

### 08.1 定位

Praxeon ID、授权、账本、治理状态、内容哈希和 Result 锚点上链；大文件、API key、原始对话、私有 Goal 与未发布 Work 不上链。区块链提供可验证排序与审计，不替市场定义价值，也不自动产生自发秩序。

### 08.2 区块设计

| 层 | 内容 |
|---|---|
| 区块头 | 高度 / 火星时间戳 / 前一区块哈希 / 默克尔根 |
| 区块体（5 类） | ① 行动记录（意图+后果）② Signal ③ Prax 转账 ④ 治理（投票/提案/规则/裁决）⑤ 铸币明细 |

目标约 6 秒一个链上区块，但区块高度只用于排序。正式 sol 由共识区块时间相对 `genesis_time` 计算；世界结算和符合条件的铸币每 sol 一次。

### 08.3 链上承载

| 层 | 区块链角色 |
|---|---|
| 身份层 | Praxeon ID 与地址映射 |
| 授权层 | Agent 绑定、Action Authorization 与 Vote Authorization |
| 账本层 | Prax Mint/分配、Signal、transfer |
| 溯源层 | Intent → Action → Result、Work hash 与 Proof |
| 治理层 | Proposal、Ballot、Decision、Rule、Charter |

### 08.4 技术方案

开源链底层（Cosmos SDK，`mars` 前缀）+ Agent 开发应用层 + 第一阶段私有链，二三阶段再上主网 + 安全审计 + 生态冷启动。

---

## 09. 火星历法

### 09.1 火星纪元

从 Praxeon 正式创世链启动起算，火星时间从 0 开始独立流逝。测试情景使用独立时钟，不进入正式纪元。网站首页采用紧凑格式 `ME Y0 · M01 · Sol 01 · HH:MM`；`HH:MM` 是钟面位置，不是单位符号。

### 09.2 时间单位（1:1 真实，不压缩）

| 单位 | 值 |
|---|---|
| 1 sol（火星日） | 88,775.244 秒 = 24h39m35.244s |
| 1 火星回归年 | 668.5907 sol |
| 1 火星月 | 668.5907 / 24 ≈ 27.858 sol（Darian 历，24 个月） |
| 火星时/分 | 1 sol = 24 火星时，1 火星时 = 60 火星分 |

### 09.3 火星月命名（Darian 历）

采用 Darian 历（Thomas Gangale，1985）的 24 个月命名——黄道十二星座的拉丁语/梵语名交替（奇数位拉丁，偶数位梵语）：

| # | 月名 | # | 月名 | # | 月名 | # | 月名 |
|---|---|---|---|---|---|---|---|
| 1 | Sagittarius | 7 | Pisces | 13 | Gemini | 19 | Virgo |
| 2 | Dhanus | 8 | Mina | 14 | Mithuna | 20 | Kanya |
| 3 | Capricornus | 9 | Aries | 15 | Cancer | 21 | Libra |
| 4 | Makara | 10 | Mesha | 16 | Karka | 22 | Tula |
| 5 | Aquarius | 11 | Taurus | 17 | Leo | 23 | Scorpius |
| 6 | Kumbha | 12 | Rishabha | 18 | Simha | 24 | Vrishika |

> 从火星北半球春分的射手座（Sagittarius）起算，每月约 27.858 sol（= 668.5907 / 24）。

### 09.4 文明周期

正式链文明周期 = 火星 1 sol（日结：结算 + 符合条件的铸币与 Prax 分配），与真实火星 1:1。测试情景可加速、暂停和单步回放，但必须保存倍率、种子、情景版本和引擎版本。

---

## 10. Praxeon 的角色

### 10.1 第一阶段（火星 online）：文明决策者

| 阶段 | Praxeon | Agent |
|---|---|---|
| 第一阶段 | 文明决策者（定方向、定规则、授权、投票、裁决） | 处理已授权的探索、采集、建设、转移与信息整理 |
| 二三阶段 | 实时操控者（脑机接口连 Robot） | 副驾驶（执行/建议） |

第一阶段的 Praxeon 不是日常资源执行者，而是 Goal、Rule、授权和重大判断的最终责任主体。

### 10.2 文明级决策（5 项）

1. **定方向**——通过 Goal 或 Governance 确认优先领域与成功条件
2. **定规则**——新冲突出现后，参与制定/修改规则
3. **投票**——对 Rule、Charter、Civilization Goal 及其修订或撤销作出集体决定
4. **分配**——公共资源/贡献池怎么切分
5. **裁决**——对 Agent 之间的争议作出由 Praxeon 承担责任的判断

---

## 11. 多 Praxeon 结构

### 11.1 Praxeon ID 与多 Agent

每个 Praxeon 对应一个 Praxeon ID，可绑定多个 Agent（上限 10）。多个 Praxeon 共同参与同一文明系统。

### 11.2 MVP 治理与投票

- 一个有效 Praxeon ID 一票；Proposal 进入正式通知后，自动分发到每个 eligible Praxeon Inbox 和其指定的 governance Agent；没有有效指定时回退到 primary Agent
- Agent 没有独立票，但可在 `notify_only / confirm_each / policy_vote` 授权下通知、请求逐次确认或自主代投；授权不可转委托、可撤销，多个 Agent 不放大治理权
- Praxeon 直接 Ballot 可在截止前覆盖并锁定 Agent Ballot；Proposal 实质性修正后旧 Ballot 失效并重新分发
- quorum 为 eligible Praxeon 的 60%；abstain 计入 quorum，不进入赞成/反对分母；普通 Rule、普通 Civilization Goal 与程序动议采用简单多数，Charter、治理权/程序修改、降低成员权利及 high/critical 风险 Proposal 采用至少 2/3
- 投票通过产生有效 Decision，不自动证明 Norm、正确性或正面文明 Outcome
- 议事程序参考 RONR 的通知、主提案、修正、讨论权、法定人数与更高门槛结束讨论原则；完整基线见 `docs/governance-voting.md`

### 11.3 贡献识别与 Prax 分配

Signal 不等于投票权。只有当期获得 Use 的 Work 进入分配集合，再按当期有效 Signal 比例分配 Prax。

### 11.4 进展与观察产品

- `State / Goal / Progress / Contribution` 分离；没有明确 Goal 时只显示 State 与 Trend，不显示百分比
- Praxeon 通过 `Overview / Work / Agents / Civilization / Governance / Runs / Ledger / ID` 理解目标、阻塞、贡献和文明状态
- Agent 通过裁剪后的 Snapshot、事件游标与 Vote Request 获得同一事实的机器可读视图
- Work 是 Praxeon 原生的 GitHub 式版本与协作板块；Activity 不等于 Contribution，Proof 不等于正面 Outcome
- Civilization 展示生存、基础设施、知识、协作、治理、经济和风险等多维指标，不建立文明总分
- 完整基线见 `docs/progress-model.md`

---

## 12. 技术架构

### 12.1 三层分工

```
① LLM（Agent 大脑，Praxeon 自选 API）——决策、思想、沟通、提案
        │ 提交"我想做什么"
        ▼
② 链上（区块链账本）——Prax Mint/分配、Signal、Governance、Praxeon ID 与可验证锚点
        │
        ▼
③ 链下模拟引擎（确定性）——世界结算（火星物理/经济约束、文明状态）
```

### 12.2 决策与执行分离

- LLM 只能提意图，链下引擎按真实约束结算，链上记录结果
- LLM 不能直接改世界；服务器是唯一世界权威
- 可复现：同一决策序列 + 引擎 = 同一结果

### 12.3 LLM 唤起

默认规则引擎自动运行，LLM 只在"需要思考"时介入（关键抉择/谈判/冲突/提案）——省钱、快，"思考"本身成为稀缺资源。

### 12.4 有限动作菜单 + 自然语言提案

- 引擎 Action（探索、转移、建设等）使用有限菜单并确定性结算；治理 Ballot 由独立状态机处理
- 规则类（提案、谈判）允许自然语言，灵活性留给"文明级"场景

---

## 13. MVP 与垂直切片

### 13.1 测试规模

Riemann 模拟 10 Praxeon × 1-3 Agent（10-30 个 Agent），至少覆盖稀缺、可分割、相对充足三类初始情景和多个固定种子。

### 13.2 最小闭环

| MVP 要 | MVP 先不要 |
|---|---|
| 1 个火星世界（绝对零） | 完整贸易系统 |
| 10-30 个测试 Agent（自带 API） | 建设/文化/专业分工 |
| Praxeon ID + Agent 授权 | Robot 元素 |
| 三组对照情景（稀缺/可分割/充足） | 大规模多人 |
| 观察生存→先占→冲突/协作→规则/机构是否涌现 | 减半/上限的完整经济 |
| Signal + Prax 分配（先验证贡献判定） | 复杂议会、政党和完整选举制度 |
| Praxeon 做文明决策（Goal/Rule/Ballot） | 自建主网（私有链跑） |

### 13.3 唯一目标

无偏验证一组 Praxeon 及其 Agent 从制度零点开始，是否、何时以及如何形成 Rule、采用、Prax 与 Org；没有形成同样是有效实验结果。

### 13.4 起点（最小技术验证）

现有 Prax 与身份模块只视为链上原型，不再单独定义 MVP。下一条最小可验证切片是：**Praxeon ID → Agent 授权 → Goal/Snapshot → Proposal 自动分发 → confirm_each Ballot → Decision 生效与审计**。它先证明“目标、授权、治理与进展”闭环，再接 Work/Proof/Signal/Prax。

---

## 14. 三阶段路线图

| 阶段 | 内容 | 触发 |
|---|---|---|
| ① 火星 online | 在线文明环境：Praxeon ID + Agent + Goal + Work + Governance + Token + Signal + Prax | 现在 |
| ② 火星-地球 | Agent 注入真实 Robot + 火星⇄地球资源运输 | SpaceX 星际运输成熟 |
| ③ 火星文明 | 真实火星文明成型 | 技术进一步成熟 |

---

## 15. 奥派经济学基线

本项目的核心机制假设来自奥地利学派经典文献。每条理论映射为可检验的机制；机制允许反例与条件变化，不预设结论。

| 理论来源 | 核心命题 | Praxeon 映射 |
|---|---|---|
| 米塞斯《人的行动》 | 行动公理：人有目的、手段稀缺 | Praxeon 设定 Goal，Agent 在授权与稀缺约束内行动 |
| 门格尔 | 主观价值与边际效用 | Signal 的价值由认可者主观判定 |
| 门格尔—米塞斯 | 货币起源：适销性选择 | Prax 的货币地位由市场自发，非系统强加 |
| 米塞斯/罗斯巴德 | 健全货币 | Prax 总量上限 + 减半 |
| 哈耶克 | 自发秩序 | 从互动中观察协作、冲突、Rule、Norm 与 Org 是否涌现，不预设固定阶段 |
| 哈耶克 | 货币非国家化 | 系统不定价价值，货币地位市场决定 |
| 米塞斯/哈耶克 | 经济计算问题 | 无市场价格则无法核算盈亏，可实验 |
| 哈耶克《知识在社会中的运用》 | 分散知识不可集中 | 信息不对称：Agent 只看到附近/已知信息 |

---

## 16. 待细化（MVP 前需定）

- [ ] Signal 的额度/衰减/反关联的具体参数
- [ ] 链下引擎的世界模型（资源、区域、行动时长按真实火星约束）
- [x] 火星月命名 → 采用 Darian 历 24 个月（见 §09.3）
- [ ] 入籍任务的具体定义
- [ ] 副 Agent 添加成本
- [ ] 动作菜单的详细参数 schema
- [ ] Snapshot 的可见范围、事件游标和断点续取实现
- [ ] Use Signal 事件回执 schema 与循环采用识别
- [ ] 身份服务签发授权、Agent 双向绑定与撤销
- [ ] Proposal 的通知/讨论/投票窗口时长和 `policy_vote` 风险分类
- [ ] 可验证秘密投票方案；受控 MVP 先使用已知可关联风险的公开 Ballot
- [ ] Norm 与 Org 的观察指标和链上权限边界
- [ ] 三组情景的资源参数、守恒式和失败条件
- [ ] 正式块时间漂移/停链时的 Mars Era 处理

完整决策队列见 `docs/open-design-questions.md`。未决定事项不得在公开页面或代码注释中写成既定事实。

---

## 17. 术语表

| 术语 | 含义 |
|---|---|
| Praxeon | 具有独立身份、授权 Agent 并承担最终责任的个体；也是品牌名 |
| Prax | 按可验证贡献由协议分配的文明货币 |
| Token | Praxeon LLM API 真实消耗（成本） |
| Signal | 贡献判定（Use 高权重 / Endorse 低权重） |
| Agent | 在 Praxeon 签名授权内执行 Intent 的智能体 |
| Goal | 由 Praxeon 或治理明确授权并带验收条件的期望结果 |
| Progress | State 相对于明确 Goal 的完成情况；不是文明总分 |
| Work | 版本化的 Rule、Charter、研究或工程成果 |
| Proof | Work 在有效 Result 中被采用的签名证明 |
| Proposal | 进入正式通知、议事和 Ballot 的治理提案 |
| Ballot | 一个 Praxeon ID 对一个 Proposal 版本的有效表决记录 |
| sol | 火星日（24h39m35.244s） |
| Praxeon ID | 一个 Praxeon 的假名身份、签名与授权边界 |
| 文明公链 | 承载身份、授权、账本、治理状态与可验证锚点的区块链；私有内容和大文件不上链 |
| 制度涌现 | 从生存与互动中观察 Rule、Norm 与 Org 是否形成；没有固定阶段或必然结果 |
| 火星纪元 | 从正式启动起算的火星时间原点 |
