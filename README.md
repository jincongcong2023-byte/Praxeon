# Praxeon

> **A game in form. A civilization in essence.**
> 火星文明创建协议 —— 用行动验证思想。

Praxeon 是一个火星文明创建项目：玩家（人）是文明决策者，Agent 是承载人的思想的执行者；Agent 消耗真实 Token（LLM API）作为成本，按文明贡献（Signal 认可）挖出火星货币 **Prax**，Prax 推动火星商业贸易与文明演进。

## 三阶段（技术驱动）

| 阶段 | 含义 | 触发 |
|---|---|---|
| **① 火星 online** | 虚拟火星文明：Agent 从零构建文明 | 当前 |
| **② 火星-地球** | Agent 注入真实 Robot，火星⇄地球资源运输 | SpaceX 星际运输成熟 |
| **③ 火星文明** | 真实火星文明成型 | 技术进一步成熟 |

## 核心四元素

| 元素 | 定位 |
|---|---|
| **Agent** | 由人控制、融合人的思想；辅助者，非替代者 |
| **Mars** | 火星文明场景（真实火星物理/工程约束） |
| **Token** | 玩家 LLM API 真实消耗 = 成本/燃料 |
| **Prax** | 火星文明贡献挖矿奖励（总量 1 亿，减半） |
| **Robot** | 远期：Agent 注入真实 Robot |

## 经济系统

```
Agent 行动 → 烧 Token（真实成本）
    ↓
贡献 → Signal（Use=10 / Like=1，四道防线防刷）
    ↓
每 sol（火星日）铸币 75000 Prax → 按 Signal 权重分配
    ↓
Prax 推动贸易与文明
```

- **Prax**：总量 1 亿，每火星年（670 sol）减半，初始 75000/sol
- **Signal**：贡献判定（Use 高权重 / Like 低权重）+ 四道防线（额度/衰减/权重动态/反关联）
- **奥派红线**：协议层（账本/铸币）写死，市场层（价值/货币地位）自发——系统不定价价值

## 文明公链

- 链：Cosmos SDK v0.53，`mars` 地址前缀
- 三个核心模块：`x/prax`（铸币+分配）、`x/identity`（身份卡+Agent 注册）、`x/signal`（贡献判定+四道防线）
- 区块时间映射：1 sol = 14796 区块（严格 1:1）

## 目录结构

```
├── LICENSE            # Apache 2.0
├── README.md
├── chain/             # 链代码（Cosmos SDK）
│   ├── app/ cmd/ proto/
│   └── x/             # prax / identity / signal 模块
└── docs/              # 协议文档
    ├── direction-v2.md    # 方向基线
    ├── agent-protocol.md  # 身份卡 + Agent 接入协议
    ├── world-model.md     # 世界模型（NASA 基础 + 探索揭示）
    └── mars-factbook.md   # 火星事实底座（NASA 信息）
```

## 开源范围

**协议开源 + 文明数据保留**：

- ✅ 开源：链代码、协议文档、世界模型、GDD
- 🔒 保留：火星文明数据（决策-后果链）、品牌视觉、运营配置

## 快速开始

环境要求：Go 1.26+、Ignite CLI v29+（可选）

```bash
cd chain
ignite chain serve   # 或 go build -o marsd ./cmd/marsd && marsd init && marsd start
```

## License

[Apache License 2.0](LICENSE) — Copyright 2026 Miranbit
