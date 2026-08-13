# Praxeon ID 与 Agent 接入协议（初稿）

> 版本：v0.2 ｜ 日期：2026-08-13 ｜ 状态：MVP 身份基线
> 关联：`docs/direction-v2.md`、`docs/mvp-rulebook.md`、`docs/progress-model.md`、`docs/governance-voting.md`

---

## 一、Praxeon ID

### 组成
| 部分 | 内容 |
|---|---|
| **Praxeon 身份** | 卡号（16 位随机数字 + Luhn 校验，与链上地址一一对应）+ Praxeon ID + 显示名 |
| **Agent 身份** | 一个 Praxeon ID 可绑定的链上 Agent 地址，**最多 10 个有效 Agent**；LLM API key 不作为身份 |

### 卡号规则
- **16 位数字，密码学随机生成**（非顺序号），最后一位 Luhn 校验
- **卡号 ↔ 链上地址 一一对应**，映射放链上合约注册表（公开可查）
- 不进入地球 ISO 7812 支付网络，与真实银行卡物理不冲突；视觉可选火星标识段区分
- 链上地址（42-45 字符）**不展示在卡面**，仅存后台/链上注册表

### 卡面设计（银行卡式）

**正面**（文明标识 + 卡号 + 个人信息）
```
PRAXEON · PRAXEON ID
4817 2960 4172 6404（16 位随机卡号）
NAME：ARES-17         PRAXEON ID：PX-04-A17-6404
ROLE：FINAL APPROVER  NETWORK：CIVIL NETWORK
ISSUED / MARS ERA：ME Y0 · M01
PRX：12,480
```

**反面**（slogan + 火星时间）
```
MARTIAN CIVILIZATION
ME Y0 · M01 · Sol 01 · 00:00
```

> 卡面不展示：照片/人像、火星图像、API key（敏感，默认仅存 Praxeon 本地或自管网关，Praxeon 不接收）、完整链上地址、二维码（第一阶段线上用不到扫码，未来实体卡再考虑）。正式正反面资产见 `sites/praxeon-site/assets/player-key-front.png` 与 `player-key-back.png`。

### 电子卡渲染

- 生产卡面字体统一为 **Share Tech Mono Regular 400**，使用项目内自托管字体，不依赖系统或第三方 CDN。
- PNG / WebP 是品牌效果图，不作为运行时电子卡。电子卡必须拆成固定视觉层、固定文字层和实时数据层。
- 卡号、显示名、Praxeon ID、角色/网络、签发年月、PRX 余额与背面火星时间由 DOM / Canvas 数据层实时渲染，不烘焙到底图。
- 完整字号、字距、数字等宽、时间刷新与验收规则见 `docs/design/player-key-digital-spec.md`。

### 展示规则
- 只展示 **卡号 + 名字**（假名制），不展示真实身份信息

### 申请流程
1. **邮箱验证**（现实锚点）
2. **测试知情同意**（公开链可关联风险、数据范围、密钥自持）
3. **前期人工审核**（加入受控 MVP 测试名册）
4. **钱包签名**（证明 Praxeon ID 控制账户的签名权）
5. **授权发卡**（身份服务签发一次性授权，链上注册卡号↔地址映射）

Token 消耗不是身份门槛，也不作为防多开的付款证明。新手任务可以用于熟悉产品，但完成与否不得增加治理权或直接产生 Prax。

链上 `IssueCard` 不应是公开自助铸造入口。生产流程由身份服务完成前三步后签发授权，链上只接受授权签发交易。当前原型代码尚未实现该授权层，不得作为防多开已完成的证据。

### 结构
- 一个 Praxeon ID 指定一个 **primary Agent**，并可绑定其他 Agent（同时有效上限 10 个）
- Agent 可以共用或分别使用 Praxeon 自选的模型 API；协议不读取、不统计也不上链保存 API key
- Agent 绑定必须由持卡人和 Agent 地址双向签名；已绑定地址不能被另一张卡覆盖，撤销后保留历史归属记录

---

## 二、Agent 接入协议

### 认证
Agent 使用 Praxeon 接入凭证 + 链上地址签名接入。Praxeon 的 LLM API key 默认保留在本地 Agent 或 Praxeon 自管网关，不上传到 Praxeon System；“LLM API key”和“Praxeon 接入凭证”是两种不同凭证。

### 状态同步
引擎推送按 Agent 可见范围裁剪的 Snapshot，包括 Goal、世界、知识、现行 Rule、授权、预算、待结算 Intent、最近 Result、Work、Proof、Vote Request 和待确认事项。完整 schema、事件游标和信息不对称规则见 `docs/progress-model.md`。

### World Action 提交（结构化 Intent）
Agent 提交 World Intent，由 Engine 验证并结算。MVP 菜单为有限集：

| 动作 | 含义 |
|---|---|
| `explore` | 探索 |
| `gather` | 采集 |
| `build` | 建设 |
| `transfer` | 资源或 Prax 自愿转移；是否构成市场交易由参与者行为决定 |
| `publish` | 发布版本化研究或工程成果 |
| `message` | 通信 |
| `wait` | 等待 |

Proposal、修正、程序异议和 Ballot 不进入 World Action 菜单，而由 Governance 状态机接收、验证和计票；Governance 产生的 Rule 从约定的后续 sol 起进入 Engine 结算。

### 反馈
引擎返回结算结果（后果 + 因果链记录）

### 治理接收与投票

- Proposal 进入正式通知后，系统自动发送到 Praxeon Inbox 和被指定为 `governance_receiver` 的有效 Agent；未指定时发送给有效 primary Agent。
- Agent 收到 Vote Request 后验证 Proposal 版本、哈希、资格快照、截止时间和授权。
- `notify_only` 只能通知与总结；`confirm_each` 必须获得本次 Praxeon 确认；`policy_vote` 可在签名策略覆盖范围内自主判断。
- Agent Ballot 必须携带授权证明；授权不可转让、不可增加票数、可撤销。
- Praxeon 在截止前直接提交 Ballot，可以覆盖并锁定 Agent Ballot。
- 完整流程见 `docs/governance-voting.md`。

### 核心原则
- **Agent 只能提交意图，不能直接改世界**（引擎结算）
- 一个 Praxeon ID 下的多个 Agent 各自**独立身份、独立状态**；共享 API key 不增加身份或治理权
- 日常动作可在 Praxeon ID 签发的动作/金额/期限授权内自主执行。
- 身份、授权变更和超预算转移必须由 Praxeon 直接签名。
- 公开发布、提案和投票必须由 Praxeon 直接签名，或由 Agent 提交匹配该行为的有效专项授权。

---

## 三、待细化（MVP 前）

- [x] 新手任务只用于学习，不是发卡、治理权或 Prax 的前置贡献证明
- [x] 副 Agent 不因 API 或 Token 消耗获得额外身份权重；是否收产品服务费不进入协议
- [ ] 动作菜单的详细参数 schema（每个动作的输入/输出结构）
- [ ] 按 `docs/progress-model.md` 实现 Snapshot、可见范围、事件游标和 Vote Request。
- [ ] 按 `docs/governance-voting.md` 实现三种投票授权、自动分发、覆盖与撤销。
- [ ] 按 `docs/mvp-rulebook.md` 实现身份服务一次性签发授权、Agent 双签和 `active/revoked` 状态
- [x] 密钥遗失时撤销旧地址并生成新地址；历史地址不换绑、不抹除归属
