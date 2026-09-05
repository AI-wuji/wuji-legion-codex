# Wuji Legion Codex 2.1 实施与验收状态

本文件以 `wuji-legion-codex-2.1-direction.md` 为唯一实施标准，区分当前仓库已经实现的 CLI/持久化契约与当前 Codex 宿主尚未证明的自动执行能力。它不把 Pi 的宿主能力、模型调用回执或设计文字当作完成证据。

## 宿主边界

路由、稀疏上下文、需求/决策/执行/验收图、经验图和能力谱系均已有受测试的 Go CLI 与结构化存储实现；这些命令只生成确定性契约或写入受限图谱，不会自行创建 Codex 原生子代理。`orchestrate` 和 `dispatch` 明确返回待宿主派发状态。任务级常驻 `gpt-5.6-sol` staff、实例/session 复用和局部任务图失效仍需真实宿主回执与行为证据；没有这些证据必须记为“待证”，不得声称已完成。

只有同时具备原生 agent ID、精确请求模型、session key、结果句柄、输入工件和独立行为验证的回执，才能把某次执行节点标为已执行；staff 回执只证明调度，不证明任务执行。路由 JSON、CLI receipt、mount、smoke 和测试中的 `native-host-dispatch-required` 都不算该证据。

## 已实现的 2.1 切片

| 方向文档要求 | Codex 中的实现 | 统一入口 |
| --- | --- | --- |
| Luna 交流/需求图/最终汇报与任务级 staff | 规则要求每个非纯交流任务绑定常驻 `gpt-5.6-sol` staff；staff 只维护任务图、调度、回执失败和需求复核，真实宿主复用与局部失效状态仍须验证 | `route`、`orchestrate`、`dispatch`、`validate-receipt` |
| 需求、决策与稀疏上下文 | 版本化需求/决策图、内容寻址的投影和精确修订失效 | `requirement-record`、`decision-record`、`requirement-project` |
| 执行图与最终验收 | 执行节点绑定精确需求修订；验收只接受成功节点中完全一致的产物和验证句柄 | `execution-record`、`execution-result`、`execution-project`、`acceptance-reconcile` |
| 原始对话保持冷证据 | 只保存不透明消息句柄与修订关系，支持按需双向定位，不保存聊天正文 | `conversation-link`、`conversation-resolve` |
| 无穷失败熔断 | 对相同尝试、策略重复和无进展阈值执行确定性门禁 | `task-gate`、`task-record` |
| 事件触发的经验、保卫与审查 | 经验图由失败/复用事件触发；安全门禁先行，官员仅给出冷建议，不常驻调用模型 | `knowledge-record`、`knowledge-query`、`security-gate`、`officer-select` |
| 最小来源生命周期与影响分析 | 同一来源版本复用最小判定；新版本仅从能力谱系计算受影响节点 | `source-assess`、`source-impact` |
| 跨图受限溯源 | 以稳定句柄记录不可变边，并在每次读取时重新检查读者 ACL | `provenance-record`、`provenance-resolve` |
| 融合资产统一调用 | 稳定 `asset_id`、兼容条件、入口与资产 SHA-256 组成可信调用契约 | `lineage-sync`、`asset-select` |
| 图谱保留治理 | 对话证据到期后先归档再从热索引移除；其余图先执行 schema 和哈希校验 | `graph-govern` |

## 可验证边界

- 所有持久化状态默认写入项目根目录下的 `.wuji`，也可通过显式存储路径或环境变量定向；不在 C 盘保留唯一规范副本。
- 验收、对话证据、溯源、来源评估和保留治理均使用限定大小的句柄和结构化记录；拒绝密钥、密码、聊天正文和未限制的日志。
- `callable`、挂载或 smoke 不能声明为“已融合”；只有可调用入口、真实资产与行为验证满足既有生命周期规则时才可提升。
- 不引入第二路由、第二工作流或 Nuwa。所有新增能力均是现有 Go `wuji` CLI 的子命令。
- `task_staff` 是宿主必须实际创建的常驻 `gpt-5.6-sol` 实例契约；其回执只能证明任务图调度，不能证明任一执行节点或能力已完成。CLI 输出该契约，但不会把 `dispatch`、路由 JSON 或模拟收据标记为执行成功。
- 审计将完整的仓库源树（含可调用资产）限制为 `1.75 MiB`；热路径仍分别受根 `SKILL.md` 6 KB、嵌套场景 Skill 6 KB 和清单总量 64 KB 门禁约束。

## 不属于 Codex 2.1 的宿主能力

- 从同一个主聊天窗口物理删除或重置历史消息。
- Pi 的 `AgentSession`、桌面 RPC 控制、跨会话运行时隔离和聊天/推理的物理双通道。
- 未由 Codex 宿主回传原生 agent ID、模型、输入工件和结果句柄时，把 CLI 规划结果当作子代理实际执行。

这些边界在方向文档中被明确保留；它们不是本实现的缺失项，也不会被标记为已完成。

## 相关回归

- `internal/core/phase_one_governance_test.go` 覆盖验收拒绝陈旧或无验证执行、会话证据归档、ACL 拒绝、来源版本复用/影响，以及兼容资产调用。
- `cmd/wuji/main_test.go` 通过真实 CLI 入口覆盖上述治理命令与可信资产调用契约。
- 完整交付前必须执行 `gofmt`、定向测试、`go test ./...`、`git diff --check` 和 Git 状态审计；具体结果以本次交付记录为准。
