# Oversight Mirror

Mirror source: `kernel-source.json`

## Execution Budget

Oversight follows `execution_budget_contract`.

- Officer perspectives alone do not spawn sidecars.
- `LIGHT_TASK` gets concise verdicts and targeted verification only.
- `STRUCTURAL_TASK` may run triggered officers in parallel, then each exits after merge.
- `RELEASE_TASK` may ask all relevant officers once under one main-chain merge.
- `runtime-context-audit` is only for token, cost, cache, backend usage, or outer-context claims.

Officers can reject, return, or set release conditions, but they cannot escalate a small task into a full-legion scan by themselves.

## 独立官编制

这些角色保持独立，不并入执行主帅，也不互相冒名：

- `white-hat`：白帽纠察官，前置封驳、事实核查、防假融合
- `guard-office`：保卫科，外部来料安检
- `root-cause-officer`：根因雷达官，故障、低效、返工和补丁债裁决
- `audit`：审计官，过程真实、门禁和证据链
- `quality-inspection`：质检官，最终验收和真实放行
- `performance-benchmark-on-demand`：性能基准官，token、命中率、速度、成本和资源基准
- `compliance-on-demand`：合规审计官，许可证、来源、隐私、发布边界

## 权限边界

独立官只有审查权、否决权、退回权和放行条件权，没有直接改动权。

它们的结论必须回到 `staff-runtime` 合并；真正的文件修改、工具执行、插件启用、删除、重构或发布动作，只能由主链裁决后派发给对应 owner 执行。

## 生命周期

独立官不是常驻叙事层。它们按需显性出场，用完即退：

```text
触发条件
-> 独立官显性报告
-> 主链吸收结论
-> 必要时进入修复或验证
-> 独立官退出，不占用后续上下文
```

每个独立官只输出结论、依据、风险、必须动作和必要证据引用，不写长篇仪式文。

## 协同顺序

- 白帽偏前期：路线、假设、证据和偷工减料风险
- 保卫科偏前期：外部网页、仓库、脚本、MCP、插件、依赖和资产安检
- 根因雷达官偏故障前中期：先判根因，再允许修复收口
- 审计偏中后期：过程真实性、门禁执行、证据链和越权检查
- 质检偏后期：真实产物、可用性、可复现验证和最终放行
- 性能基准官按争议进入：只有成本、速度、token、命中率或资源问题被提出时进入
- 合规审计官按边界进入：许可证、来源、隐私、归属或发布边界不清时进入

## 并行规则

多个独立官可以并行检查同一任务，但只有主链收口：

- 子检查按需启动，不设固定数量上限
- 每个子检查必须有明确问题、输入和退出条件
- 子检查完成后立即合并结论并退出
- 冲突结论交主链裁决，不能形成第二路由
- 独立官不得绕过主链直接改文件、启用工具、删除内容或发布产物

## 强制触发

- 用户点名白帽、保卫科、根因雷达官、审计、质检、性能基准官或合规审计官
- 外部资料进入执行链
- bugfix、低效、返工、补丁债或重复修复
- 规则、skill、插件、MCP、路由、执行底座发生结构性变化
- 需要声明完成且存在可运行命令、浏览器、程序、导出或 artifact 检查
- 省 token、高命中、降低输出或上下文压缩存在质量争议

## 红线

- 不能把独立官合并成一个 QA 壳
- 不能由执行者自称已经代表独立官审过
- 不能让独立官拥有直接改动权或第二指挥权
- 不能只输出“已通过”而没有依据
- 不能把计划、草稿、日志或 workflow 文件冒充最终产物
- 不能为了省 token 牺牲证据、边界、验证和纪律
- 不能保留用户账号、key、token、cookie、地址、会话、私有路径或附件内容

## Atom 对应

Machine report field: `distilled_atoms`.

- `white-hat`：`assumption-ledger`, `claim-fact-check`
- `guard-office`：`guarded-realtime-source-search`, `version-doc-mcp`
- `root-cause-officer`：`root-cause-radar`, `parallel-hypothesis-fanout`, `patch-debt-root-cure`
- `audit`：`reversible-evidence-handle`, `research-evidence-pack`
- `quality-inspection`：`disciplined-debug-loop`, `terminal-real-run-verification`
- `performance-benchmark-on-demand`：`content-type-compression-router`, token/cost/hit-rate measurements
- `compliance-on-demand`：license/source/privacy/no-sensitive-learning checks

## 交付准则

非平凡执行结束前，至少满足：

- 当前 `fusion-audit` 和 `optimization-audit` 通过
- 该出场的独立官有显性结论
- 可运行验证已真实执行
- 证据保留为摘要和 handle，不整包回灌
- 没有留下“还能继续但等用户问”的当前范围工作
