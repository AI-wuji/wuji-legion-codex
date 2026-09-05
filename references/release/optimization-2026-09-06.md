# 2026-09-06 优化记录

## 目标

保持无极军团 3.0 的初心：用户只需自然语言描述目标，由阿极进行白帽式判断并选择已验证的插件、MCP、Skill、工具和执行节点；系统只能在有限的时间、重试、上下文、并发和成本边界内工作。

## 本轮清单

| 项目 | 状态 | 完成判据 |
| --- | --- | --- |
| 调度器的重派、续跑和图规模上限 | Go 回归通过；宿主未接入 | 单图最多 1024 节点；每节点最多 3 次尝试、每条谱系最多 2 次重派。预算耗尽时须阻断下游。 |
| 工作区图重建、写入和清理的时限 | Go 回归通过 | 重建共享 5 秒 deadline；构建最多 32 MiB、查询最多 16 MiB；generation 容量上限为 8，旧代有界回收；清理每棵树最多 512 项，fallback 受 512 源文件、目录项、时间和字节预算限制。 |
| 探针及其子进程取消 | Go 回归通过 | 终止等待本身有上限，超时不会无限阻塞，且应覆盖真实子进程场景。 |
| 执行结果到经验记录的反馈 | Go 回归通过；待扩展验证 | 仅在版本化终态执行结果上写入有界身份/结果候选事件，最多 1024 条或 2 MiB。`knowledge-record --feedback-id/--feedback-store` 可在显式提交时门控复用符合条件的失败候选；不自动成为已验证经验、不自动触发路由复用、不晋级能力。 |
| capability 审计的默认来源范围 | Go 回归通过；快速审计受阻 | 默认 `verify` 忽略缺失/不完整的非 primary 冷源，但校验已发现来源的完整性；完整审计要求 secondary 可用，optional 缺失仅作信息提示。 |

## 边界

- 调度器目前是进程内 Go 库；除非宿主派发和回执的真实证据已归档，不能称为已接入实际 Codex 宿主。
- 宿主未提供 token、缓存或计费遥测时，系统不能执行真实成本核算或声称节省金额；配置上的字节/次数预算不等于供应商账单控制。
- 候选资料、已验证经验和 `primary` 能力不是同一等级。候选需对照；经验需独立验证文件及 SHA-256；`primary` 还需内容寻址 promotion receipt 和归档基线。
- Windows 上的 `taskkill` 仍是 best-effort 进程树终止，不是 Job Object 级的全局强制隔离；日志磁盘增长和证据目录清理也尚不能称为全局完全有界。
- 合法的大型工作区图达到 generation 容量时，单次同步可能暂时失败；后续调用会先做有界增量回收后恢复，内部不得以无限重试掩盖该状态。
- 以下验收只证明列明的本地测试范围，不代表宿主级完成或全能力融合。每项只有在当前代码、失败路径和独立验证证据均归档后，才可更新为“已验证”。

## 尚未完成

- 当前运行时尚未接入原生宿主调度；全局 token/计费硬停止也不能在缺少宿主遥测时实现。
- 已验证经验到后续路由复用的自动闭环尚未实现。候选事件仅供受控诊断和后续显式、证据门控的 `knowledge-record --feedback-id/--feedback-store` 处理。
- 已安装的 3.0 `.agents` 接点与旧目录名 `.codex` 接点均指向本仓库，无需为本轮优化重新安装。

## 冻结验收（2026-09-06）

- 首次全包测试曾因图谱清理测试 fixture 的 5 秒同步超时失败；fixture 纠偏后，`go test -timeout 120s ./...`、`go vet ./...`、`scripts/build.ps1` 和 `git diff --check` 均退出码为 0。对应哈希见 `.wuji/validation-20260906/final-artifact-hashes.txt` 与测试日志；最终二进制 SHA-256 为 `97D046F9D1BF3A520972491503CAA24DDEB92E883BD9402153F536968B5C657A`。
- 根 Skill 与 15 个嵌套 Skill 的结构验证通过。Feishu 验证曾因 GBK 编码失败；定向设置 `PYTHONUTF8=1` 后通过，该环境依赖应保留在后续跨环境验证范围内。
- 历史快速审计失败记录仍保留：Windows PowerShell watchdog 约 30 秒后得到真实退出码 1；`audit-fast-final3.stderr.log` 记录 OfficeCLI 行为探针缺少 `System.Private.Xml, Version=10.0.0.0`，随后 DOCX sentinel 写入失败。当前本机 OfficeCLI `v1.0.147` 行为探针已通过，旧 XML 故障当前不复现。
- 最后一轮 fast audit 自然退出码为 1，已验证 15 个 Skill 后被三项 size gate 拦截：Skill `8101 > 6000`、agents `6450 > 5000`、source `2906116 > 1835008`。门禁未取消、阈值未放宽；应进一步压缩（若后续证实为统计误计，再以独立证据修正）。因此本轮不是 overall green，也不能将 fast audit 或 full audit 称为通过。
- alias 补丁后的定向验证均退出码为 0：`internal/core` 的 `Test.*Feedback`、CLI 的 `TestKnowledgeRecordFailureFeedbackBridge`、`go vet ./...` 与 `scripts/build.ps1`。full audit、race 检查和安全扫描本轮未运行。没有原生宿主级调度回执，也没有已验证经验到路由复用的自动学习闭环。
- 验收结束时未发现活动的 `wuji` 或 `go` 验证进程；失败日志保留为回归证据，而不是通过记录。
