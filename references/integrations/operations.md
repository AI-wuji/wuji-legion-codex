# 冷集成操作指南

本页只在专家交接、显式用户记忆或可选上下文/图适配器任务中读取。入口可调用不等于上游能力已融合。

## 专家交接

```powershell
./bin/wuji.exe expert-bridge select --root . --query "修复可复现的并发故障"
./bin/wuji.exe expert-bridge prepare --root . --workspace . --query "修复可复现的并发故障" --worker worker.json --task-instance task-1 --graph-version 1 --execution-node exec-1 --attempt attempt-1
./bin/wuji.exe expert-bridge dispatch --contract contract.json --dry-run
./bin/wuji.exe expert-bridge verify --contract contract.json --receipt receipt.json
```

`prepare` 绑定专家目录、callable capability manifest、工作区、任务图版本和 attempt，并生成内容哈希。实际 native collaboration spawn/followup 前，运行时 worker 必须检查 `--lease`、lease duration、deadline、task-wide 无进展停止和跨策略共享的原子任务总尝试预算；`task-claim` 获取 lease，每个被接受的 `task-record` 都携带并释放本次 lease，后续 attempt 必须 fresh claim。只有 success 或停止条件使任务终止，progress 重置无进展计数。完整 policy 固定到同一任务。collaboration spawn 可接收请求模型并返回原生 agent ID，但调用者不能提供宿主证明的 `session_key`；本地 sticky `session_key` 仅用于关联。

```powershell
$claim = ./bin/wuji.exe task-claim --store .wuji/task-circuits --task task-1 --strategy expert-bridge --policy bounded-native-v1 --max-no-progress 2 --max-attempts 2 --deadline-seconds 120 --lease-seconds 60 --attempt attempt-1 | ConvertFrom-Json
# Spawn/follow up only when $claim.allowed; record the observed outcome with $claim.lease_id.
./bin/wuji.exe task-record --store .wuji/task-circuits --task task-1 --strategy expert-bridge --policy bounded-native-v1 --max-no-progress 2 --max-attempts 2 --deadline-seconds 120 --lease-seconds 60 --attempt attempt-1 --outcome success --lease $claim.lease_id
```

`verify` 是严格的 legacy consistency gate：重算结果与独立证据文件哈希，并拒绝越界路径、身份不一致和陈旧 attempt。它要求 receipt 中的 effective-model 以及 billing baseline/savings；当前宿主未独立提供这些字段时，不得伪造以通过 verify。此时 `host_execution_verified: false`、`graph_mutation_allowed: false`，不得把任务标为成功。原生调用记录加独立制品测试仍可作为有用的执行/行为证据，但不能替代 strict expert-bridge verify、不能声称模型已获宿主证明，也不能完成任务图。dry-run、CLI dispatch 或 consistency verify 同样不证明专家模型真的执行或六类专业工作流完整。

## 用户记忆

只有用户明确要求保存时才调用 `remember`，并把该确认写入 `--provenance`。默认工作区隔离；跨项目共享必须显式选择命名的 `--shared-scope`。`recall` 结果是用户确认记录，不是外部事实证据。

```powershell
./bin/wuji.exe user-memory remember --workspace . --key "review-style" --value "先列阻断问题" --provenance "用户明确要求保存" --ttl 720h
./bin/wuji.exe user-memory recall --workspace . --query "review"
./bin/wuji.exe user-memory revoke --workspace . --key "review-style" --expected-version 1
```

不要保存 API Key、Token、密码、会话原文或未经用户确认的推断。修改已有记录时必须传当前 `--expected-version`，撤销时也建议传入；TTL 为零表示不自动过期。

## Context-mode

只允许按任务显式使用无状态 execute 适配来处理大输出，并强制超时、工作区和输出预算。不要启用自动 prompt 捕获、session 恢复注入、indexed memory reuse 或全局 hook。生成参数经真实 MCP 执行并通过独立复算：10,000 行、775,450 字节中文日志正确统计出 104 条失败。另测得上游 1,000 ms 超时有效；同一索引 fixture 却可在当前与无关显式项目路径下返回，因此相关索引功能保持禁用。详见[兼容契约](context-mode-compatibility.md)与[验收证据](../release/integration-2026-09-19.md)。

```powershell
./bin/wuji.exe context-mode-prepare --workspace . --input ./bounded.jsonl --operation jsonl-count > call.json
# 宿主显式执行 call.json 中的 ctx_execute 后：
./bin/wuji.exe context-mode-validate --contract ./call.json --result ./result.json
```

`prepare` 只生成 `prepared_only: true` 的契约，不证明工具已运行。

## Graphify

Graphify 仅作为无 hook 的可选 code-only pilot：

```powershell
./scripts/graphify-pilot.ps1 -Workspace . -File @('cmd/wuji/main.go')
./scripts/test-graphify-pilot.ps1
```

它不替换主工作区图，不安装 watcher、MCP 或模型提供商。适配器测试只证明边界逻辑；只有固定上游版本在隔离环境中的真实运行才可能证明提取行为。详见[Graphify pilot](graphify-pilot.md)。
