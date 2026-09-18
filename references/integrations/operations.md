# 冷集成操作指南

本页只在专家交接、显式用户记忆或可选上下文/图适配器任务中读取。入口可调用不等于上游能力已融合。

## 专家交接

```powershell
./bin/wuji.exe expert-bridge select --root . --query "修复可复现的并发故障"
./bin/wuji.exe expert-bridge prepare --root . --workspace . --query "修复可复现的并发故障" --worker worker.json --task-instance task-1 --graph-version 1 --execution-node exec-1 --attempt attempt-1
./bin/wuji.exe expert-bridge dispatch --contract contract.json --dry-run
./bin/wuji.exe expert-bridge verify --contract contract.json --receipt receipt.json
```

`prepare` 绑定专家目录、callable capability manifest、工作区、worker/session、任务图版本和 attempt，并生成内容哈希。`verify` 只做一致性检查：重新计算结果与独立证据文件哈希，并拒绝越界路径、身份不一致和陈旧 attempt；它固定返回 `host_execution_verified: false`、`graph_mutation_allowed: false`，不会把任务标成成功。当前 Codex 宿主没有暴露本仓库所需的原生模型/session 绑定入口；因此 dry-run、CLI dispatch 或 consistency verify 都不证明专家模型真的执行，也不证明六类专业工作流完整。

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
