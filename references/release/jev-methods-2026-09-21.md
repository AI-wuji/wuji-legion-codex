# Jev 方法借鉴：有界选择与证据收缩

本轮只实现两个本地确定性辅助入口，不安装 Jev、Laya、Reticle 或新模型服务，不变更能力晋级状态。

## 来源与采用范围

- [hermes-jev-skills](https://github.com/kerpopule/hermes-jev-skills)：核查 `jevkit/skillpick.py`、`tests/test_skillpick_gate.py` 与 MIT 许可证（2026 Steve Darlow）。借鉴有限候选、允许不选、失败返回原流程；不复制其模型调用、批量并发或上下文注入。其第二阶段使用技能描述，不应描述为完整技能正文。
- [jev-search](https://github.com/superagents-lab/jev-search)：核查 `src/lib/pipeline.ts`、`test/pipeline.test.ts` 与 MIT 许可证（2026 Search1API）。借鉴结果收缩、来源与失败状态保留；本地实现不包含其搜索后端或 Jev 排序。上游模拟测试不作为本地效果证据。

这是方法借鉴和本地实现，不是安装上游软件或完整融合其模型能力；无实测 token 节省比例、延迟收益或专业质量提升承诺。

## 行为变化

`expert-bridge select` 从“并列时取目录第一项”改为明确返回 `selected`、`ambiguous`、`none`。后两者返回阿极原有路由，`prepare` 拒绝交接；最多列出 5 个并列候选，命中数不作置信度。词面匹配的固有限制仍在。

`search-select` 对已检索结果做离线有界去重，保留输入顺序、证据 URL 与来源错误。它不读取网页、不做语义排名、不核实事实。既有 task lease、deadline、总尝试预算和跨策略无进展停止逻辑保持不变。

入口用法见[集成操作指南](../integrations/operations.md)。专家选择写入根 Skill 的交接规则；搜索收缩按需写入 research Skill，无新增全局 hook 或常驻专家。

## 验证结果与限制

- 锁定 Go 的 `go test ./...`、`go vet ./...` 通过；统一测试中的 Go 格式与根/嵌套 Skill 校验通过。
- 独立运行新二进制：代码修复任务唯一匹配；“请调研并制作PPT”从旧版静默选 research 改为返回 research/document-deck 并列；无匹配与反触发返回原路由。搜索 fixture 的 6 条输入保留 4 条不同证据，删除 1 条重复与 1 条无效协议；保留 fragment、业务版本参数和来源超时信息。它只验证这些边界，不证明总体任务质量提升。
- 既有任务门禁真实 smoke：第二个并发 claim 被 `active-lease` 拒绝，跨策略两次无进展后的新 claim 被 `no-progress-limit` 拒绝。
- 本地 CLI 证据：`.wuji/jev-methods-cli-evidence.json`，从当前已安装目录重跑后的验证器独立 SHA-256 为 `4613A43F19A92A50B502B2BA5EA385F555E0839C68F67F25872C9C4207497256`。此文件是忽略的本地制品，不随仓库分发。二进制 SHA-256 为 `28E8E141FA825DE8054B9218297B22AC5646B282C889E90B4EB1B57119029FA6`；当前 Codex junction 下二进制哈希一致。
- 统一 fast audit 未通过：15 个 capability 检查后，OfficeCLI Word sentinel 写入因缺少 `System.Private.Xml, Version=10.0.0.0` 失败。未调整或跳过此门禁，不宣称整体审计通过。另按 audit 的统计口径检查，源码总体约 3.11 MB，仍超过 1,835,008 字节上限；未删改无关文件或放宽门禁。
- 未伪造宿主模型/计费证明，strict expert-bridge 的宿主证明缺口保持不变；没有能力自动晋级。发布目标为仓库 `main`，本轮验证不等于整个军团验收通过。

## 当前 Codex 启用核查

当前用户的 `.agents/skills/wuji-legion-codex-3-0` 与兼容入口 `.codex/skills/wuji-legion-codex-2-1` 均链接到本工作区，PATH 中的 `wuji` 也解析到本次构建。没有发现禁用该入口的 `skills.config` 配置。

在本次会话中，已通过 `.agents/skills/wuji-legion-codex-3-0/bin/wuji.exe` 实际调用 `expert-bridge select` 与 `search-select`，并从安装目录执行四类专家选择和搜索边界验证。介绍文案请求没有词面匹配时，确实返回阿极原流程继续处理；搜索启用样例保留一条官方文档来源并去掉重复记录。

根 Skill 和 research Skill 的调用说明已实际读取。按照 [Codex 官方技能文档](https://developers.openai.com/codex/skills/)，链接目录受支持、技能变更自动检测；这只涉及技能发现，不能推导为默认阿极身份或白帽对话行为已经生效。按照 [AGENTS.md 官方说明](https://developers.openai.com/codex/guides/agents-md/)，启动指令在会话开始时构建；修改后应在新会话验证。`default_prompt` 是调用技能时的提示元数据，不是全局常驻指令。此轮 CLI 验证证明当前入口可执行与本会话已调用，不证明白帽行为、后台常驻或未来任务自动调用。

后续用户反馈指出实际对话仍有迎合。本机既有 `developer_instructions` 已包含阿极和白帽约束，因此不能断言原因是未配置；当前桌面会话是否加载该字段仍缺少直接证据。本次将项目规则和本机既有指令中的白帽要求明确为普通聊天也生效、纠正错误前提、不因催促改口、按新证据修正且不刻意唱反调。配置落盘不作行为通过证明；后续验收须观察新会话对错误前提、要求无证据确认成功、施压赞同和合理请求的实际回答，不能仅检查身份自报或规则文本。

本次配置 TOML 解析通过，项目 AGENTS 为 4,796 字节，未超过既有 5,000 字节上限。一次 60 秒上限、只读、ephemeral 的 CLI 对话探针收到 `turn.completed`；但输出采集未保留回答正文，无法审核行为，因此不计通过，也未重复调用。桌面新会话的实际加载和持续不迎合仍待验收。
