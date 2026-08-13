# Civilization vertical slice

本目录是 Praxeon 第一条可执行纵向切片，不是独立“游戏系统”。它把同一事实源提供给人的进展视图和 Agent 的机器视图：

`Praxeon ID → Goal / Snapshot → Proposal → Vote Request → Authorization / Confirmation → Ballot → Decision → Civilization Goal`

## 已实现

- `api/v1/`：由 `proto/mars/civilization/v1/` 生成的 Msg、Query、REST gateway 与数据类型。
- `domain/`：不依赖存储实现的确定性状态机和边界测试。
- Praxeon Goal 与 Agent Goal 由所属 Praxeon 创建；Civilization Goal 只能由通过的治理 Proposal 实体化。
- Proposal 需要不同 Praxeon second；通知时冻结 eligible snapshot，按 governance Agent、primary Agent 顺序路由 Vote Request。
- 无 Agent 路由时保留投票资格，并显式记录 delivery failure；Agent 需对精确 Proposal hash 回执。
- Agent 没有独立票。`notify_only` 不能提交；`confirm_each` 绑定一次性 Proposal/版本/选择；`policy_vote` 受类型、领域、Proposal ID、风险和期限约束。
- 同一 Praxeon 的多个 Agent 共享一张当前 Ballot；Praxeon 直接 Ballot 可覆盖并锁定后续 Agent 更新。
- 实质性修订生成新版本、重新 second/分发并使旧 Request/Ballot 失效；提案类型不可借修订切换。
- quorum 为 60%；abstain 只进入 quorum；普通事项简单多数，高影响事项使用精确整数 2/3。
- 公开事件与 Praxeon 私有事件共用单调 cursor，供 Agent 幂等续读；Snapshot 引用同一 cursor。

## 尚未实现

当前 `Engine` 使用进程内存，只用于冻结语义与验证不变量，尚未成为可上线的 Cosmos 模块。下一层必须完成：

1. Cosmos Collections keeper、genesis、索引、迁移和确定性事件持久化。
2. 现有 identity 模块适配，验证 Praxeon ID、Agent 绑定与真实交易 signer，而不是信任消息字段。
3. Msg Server、Query Server、模块注册、App Wiring 与链上端到端测试。
4. Rule、Charter、Revoke、procedure 的 Decision 实体化；当前只有 Civilization Goal 实体化。
5. Proposal Package、反对材料、程序异议和大内容的链下寻址存储。

协议变更后从 `app/chain/mars/` 运行：

```bash
go tool buf lint --path proto/mars/civilization/v1
proto_tmp=$(mktemp -d /tmp/praxeon-proto-gen.XXXXXX)
go tool buf generate --path proto/mars/civilization/v1 --template proto/buf.gen.gogo.yaml --output "$proto_tmp"
cp "$proto_tmp"/mars/x/civilization/api/v1/*.go x/civilization/api/v1/
go test ./x/civilization/...
```
