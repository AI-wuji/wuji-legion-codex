# Native runtime bounded-claim evidence (2026-09-19)

本轮真实 native review 节点 `/root/native_review` 发现无进展计数仍按 strategy 分裂，以及 lease 文档与实现不一致。主任务复核另发现 policy ID 重命名和 legacy 接口降级问题。修复节点 `/root/native_runtime`（请求模型 `gpt-5.6-sol`，本地关联 `native-runtime-20260919/runtime`）将 guarded state 改为 task-only 寻址、完整 policy 固定、跨策略无进展停止，并让每个 accepted record 在同一原子状态写入中释放 lease；后续 attempt 必须 fresh claim。主任务最后补齐 legacy record 漏传预算参数的拒绝检查和回归测试。旧 `task+policy` 状态仅在唯一且完整 policy 匹配时保守读取，否则拒绝。

验证命令：

```powershell
$go = & scripts/resolve-locked-go.ps1 -Root (Get-Location).Path
& $go test ./...
& $go vet ./...
./scripts/build.ps1
```

结果：全套 Go 测试、go vet、Skill 校验和本地构建均通过。覆盖并发 claim、跨策略 no-progress、progress reset、transient attempt 耗尽、重复 attempt、policy 改名/弱化、legacy gate/record 降级、过期 deadline/lease 和 stale finish。另使用重建后的真实 CLI 和隔离状态目录验证：active lease、错误 lease、legacy record 漏参数、policy 改名、同策略与跨策略无进展停止，六项均通过。修复前也曾以 task-claim 保护真实 native review 调用，确认占用期间第二次申请被拒绝。

本地 Skill junction 指向当前仓库；重建 bin/wuji.exe 更新本地入口。热指令压缩只代表文本字节减少，未测量实际 token 或计费节省。

边界：这是仓库内的持久化调用门禁，不是宿主全局计费或自动中断硬限制。宿主仍未提供 effective model、token、cache 或 billing 证明；strict expert-bridge 的宿主执行证明缺口仍在。Graphify 保持关闭；完整 audit 的既有源码体积门禁仍未转绿。未覆盖或未声称这些边界已解决。
