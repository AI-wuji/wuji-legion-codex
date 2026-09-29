# 决策与上下文方法借鉴：2026-09-26

本轮检查 Jev、Laya 及其相近项目，目标是寻找能增强无极军团 3.0 的最小方法，而不是把外部运行时直接装进热路径。

## 结论

| 候选 | 保留价值 | 当前决定 |
|---|---|---|
| Laya | 低成本、结构化的 System 1 决策 | 候选 shadow provider，尚未接入本仓库；不直接删除历史、不决定完成 |
| fast-jev-compaction | 工具调用/结果配对、最近记录保护、压缩回退 | 只借鉴算法，不直接安装 Claude Code Hook 或远端 Jev |
| jev-compaction | superseded/stale 关系、外部记忆仓、可恢复 recall | 仅吸收同一 footprint 下重复结果和成功写入后旧读取的确定性关系；外部记忆仓与自动 recall 未实现 |
| jev-mcp | verify/screen/gate/classify/rerank 等窄工具 | 作为未来门禁工具的参考，不注册万能 MCP |
| jevals | allow/modify/block/escalate 门禁语义与 Laya backend | 作为决策结果模型参考，不把评分当事实 |
| Paperclip | 长期项目、预算、心跳、外层 Agent 控制面 | 暂不进入 3.0 核心，未来作为外部运行管理器 |
| Agent-Native | 一个 Action 供 Agent、UI、HTTP、MCP、CLI 共用 | 只吸收 Action 契约，不引入整套应用运行时 |
| Whittle | 可证明的无损/失败开放/代码直通压缩 | 只吸收安全不变量；不接入代理、后台服务或模型路由 |
| Context Ledger | 提交边界 + 可恢复指针 | 吸收“保留句柄、按需重建”原则；不增加第二套提交钩子/账本 |
| codex-agent-mem | 本地优先、范围/会话隔离、拉取式 continuity pack、`not_modified` | 与现有知识/需求/执行状态重叠；只作为设计参照，不接入第二 MCP 状态库 |
| Headroom / Distil | 缓存前缀冻结、可恢复扩展、真实账单测量 | 当前宿主没有可证明 cache/billing 回执；不宣称节省，不接代理 |
| RTK | 窄命令输出过滤、拥有权标记、失败开放 | 保持可选外部输出过滤，不改核心 shell 或全局安装 |
| jev-routing | 对 Codex/不同模型的工具引导负收益证据 | 作为拒绝通用 Jev 路由的实测反证；先做配对基准，不接 A/B |

## 已落地

`wuji context-reduce` 新增了 schema v2 的可恢复确定性上下文投影，`wuji context-recall` 提供受校验的冷归档召回。它们只是准备材料，尚未接入 Codex 宿主的对话压缩钩子：

- 用户、系统、助手消息默认保留；
- `protected` 与 `critical` 记录保留；
- 最近记录保留；
- 同一 `tool + footprint + 原文` 的成功旧读取可标记为 `identical-observation`；
- 成功写入后同一 `footprint` 的旧读取可标记为 `stale-after-write`；
- 工具调用与结果通过 `pair_id` 一起归档；
- 归档记录随结果返回，原始输入的 SHA-256 和顺序可以独立恢复校验；调用方须把 `archived` 冷存，不能仅保存 `visible`；
- 结果同时报告原始、可见、归档和可节省的记录字节数；这只是投影指标，不等于 provider 账单 token 或真实上下文节省；
- `context-recall` 先重新验证 schema、顺序、哈希和确定性策略，再从归档记录中做有限词法检索，只返回原始记录；命中带 `pair_id` 的记录时按原顺序带回完整工具调用/结果对，不生成摘要；
- 任何输入不满足关系条件时保持原样；
- 结果明确声明它只是可恢复投影，不构成事实核实或完成证据。
- 新增 `wuji context-recall`：输入完整 schema v2 投影，先通过恢复/哈希/确定性策略复核，再对 `archived` 做有限词法检索，只返回原始记录和匹配字段；不做语义判断、不注入用户记忆、不宣布完成。

命令从标准输入读取一个有界 JSON：

```powershell
@'
{
  "records": [
    {"id":"u1","role":"user","kind":"message","text":"保留用户约束"},
    {"id":"r1","role":"tool","kind":"tool_result","tool":"Read","operation":"view","footprint":"src/a.go","pair_id":"p1","succeeded":true},
    {"id":"w1","role":"tool","kind":"tool_result","tool":"Write","operation":"write","footprint":"src/a.go","pair_id":"p2","succeeded":true}
  ],
  "archive_stale": true
}
'@ | ./bin/wuji.exe context-reduce
```

保存输出到宿主受控的冷制品路径后，可用 `wuji context-reduce --restore` 从该投影重建完整原始记录。CLI 自身不持久化、压缩或修改任何 Codex 会话；把完整 `visible + archived` 结果送回主提示不会节省上下文。
需要找回旧证据时，将完整投影送给 `wuji context-recall --query <路径或关键词>`；召回结果仍须由 Aji/宿主重新核对，不能当作完成或事实证明。

## Laya 边界

本地 Laya 服务可以承担候选判断、是否需要工具、候选 Skill 排序和压缩建议，但它的输出必须满足：

1. 只作为建议或 shadow 记录；
2. 结构化输入不得包含密钥、完整历史或未压缩的大型工具输出；
3. 必须允许 `abstain` 或低置信度返回阿极；
4. 不得直接修改需求图、任务图、模型回退、完成状态或验收状态；
5. 任何自动收缩必须先通过确定性保护层，并保留可恢复原文。

## 未宣称

本记录不宣称 Laya、Jev、Paperclip 或任何相近项目已经融合为无极的 primary 能力。真正晋级仍需要当前仓库的 callable 入口、真实行为探针、独立哈希和对应场景的回归证据。
