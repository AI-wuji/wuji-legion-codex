# 无极军团 3.0 / Wuji Legion 3.0

无极军团 3.0 是一个面向 ChatGPT Codex 的系统级 Skill 中枢。其目标与路由契约是让不懂 Skill、插件和 MCP 的用户用自然语言提出要求，由 Terra 阿极完成理解、白帽式判断、能力组合和最终沟通，并调配已验证的插件、MCP、Skill、工具和执行节点；需要执行时，确定性参谋部维护任务图并只调度真正需要的执行节点。阿极不会为了迎合而附和错误结论，会明确指出风险、证据不足和不可行之处。当前 Go 运行时提供该契约的有界路由和状态机制；原生 Codex 宿主的实际派发、回执接入和全局计费控制尚未完成，不能据此宣称已自动完成真实调度。

无极军团的另一项核心能力是对 Skill、插件和工作流进行蒸馏、融合与进化：保留有价值的脚本、模板、资产和入口，通过真实行为对照决定能力的接入、替换、升级或淘汰，让分散的外部能力逐步沉淀为可调用、可验证、可演化的系统能力。当前自动反馈只会在版本化的终态执行结果上记录有界的身份/结果候选事件；它不自动写入已验证知识、不触发路由复用，也不晋级能力。已有 `knowledge-record` / `knowledge-query` 仍分别代表显式记录和事件触发检索。能力晋级和替换仍必须经过真实对照、证据校验和进化门禁，不能把一次成功或模型自报当成自动升级依据。2.0 以冷挂载完整能力包、确定性 Go 路由、稀疏上下文选择和进化门禁支撑这条主链路。

Wuji Legion 3.0 is a system-level Skill hub for ChatGPT Codex. Its target and routing contract is for Terra Aji to handle natural-language understanding, universal PonyTail judgment, capability composition, communication, and final reporting, while deterministic General Staff state schedules only the execution nodes a task needs. The current Go runtime provides bounded routing and state mechanisms for that contract; native Codex host dispatch/receipt integration and global billing control are not yet complete, so this must not be represented as completed automatic real-world scheduling.

Its second core function is distilling, fusing, and evolving Skills, plugins, and workflows. Wuji retains valuable scripts, templates, assets, and entrypoints, then uses real behavioral comparisons to admit, upgrade, replace, or retire capabilities. Current automatic feedback records only bounded identity/outcome candidate events from versioned terminal execution results; it does not create verified knowledge, trigger route reuse, or promote capabilities. Cold-mounted complete packages, deterministic Go routing, sparse context selection, and evolution gates keep this orchestration chain bounded and verifiable.

## 架构文档

- [Codex 3.0 当前架构](references/architecture.md)
- [无极军团 3.0 完整方案](references/architecture/wuji-legion-3.0-blueprint.md)
- [无极军团 3.0 全图谱架构](references/architecture/wuji-legion-3.0-graph-architecture.md)
- [进化主帅与能力谱系融合](references/architecture/wuji-legion-3.0-evolution-commander.md)

## 项目演进 / Evolution

无极军团的版本演进不是把旧项目原封不动搬到新平台，而是持续做能力清点、蒸馏、验证和架构纠偏。下面的时间线只记录无极军团本身的公开主线；同一账号下的其他仓库属于相关产品或能力实验，不自动等同于本项目的版本。

| 阶段 | 主要变化 |
| --- | --- |
| `v6.0 - v6.1` · 2026-05-28 至 2026-05-30 | 建立轻量状态机和单主帅内核；将普通生图、真 PPTX 与 HTML 演示稿拆成清晰的能力路线。 |
| `v8.0 - v9.2` · 2026-05-30 至 2026-05-31 | 对全局规则和入口做整流，明确“无极军团是总框架、领域能力是内部模块”；开始系统吸收工程流程、测试、审查和发布门禁。 |
| `v9.3 - v10.8` · 2026-05-31 至 2026-06-03 | 建立蒸馏与进化闭环，压缩重复专家，落地执行底座、任务状态、证据等级、收口检查、工作流门禁和真实成品验证。 |
| `v11.3` · 2026-06-09 至 2026-06-20 | 发布旧版运行内核，继续完成来源池融合、上下文膨胀诊断和首轮路由优化，为 2.0 重构保留可验证的经验与资产。 |
| `2.0` · 2026-07-12 | 面向 Codex 重新设计：保留有价值的能力原子，剔除旧版弯路；收束为阿极单一主脑、单一写权限、冷挂载完整能力包、有界并行、确定性 Go 路由和基于真实行为的能力进化。旧版保留在 `legacy-v1-backup` 分支与 `legacy-v1-final` 标签中。 |
| `2.0.1` · 2026-07-12 | 修复模型降级的真实经济性：排除锁文件检索噪声，增加上下文覆盖率与代码证据门禁，向 worker 发送确定性上下文 capsule 和真实任务契约，按实际重放字节计费，并将失败回退限制为生成前不可用错误。 |
| `2.0.2` · 2026-07-12 | 落地搜索优先与任务内模型粘性：非平凡解决方案任务先做最多 3 来源、90 秒的 Luna 预检；机械只读任务使用 Luna，独立高推理判断交给 Sol；每个 worker 固定 session key，并以 schema v2 回执验证真实模型调用。 |
| `2.0.3` · 2026-07-12 | 落地有界关系图：工作区图先定位文件/符号，经验图只在失败、复用、能力缺失和验证事件触发；记录带显式 scope、根因和验证哈希，查询、索引词和候选展开均有硬上限。新增真实行为探针，并修复同大小同时间源码变化导致的陈旧索引风险。 |
| `2.0.4` · 2026-07-13 | 纠正模型执行面：阿极默认改为 Terra；Luna 用于机械/检索分支，Sol 限为显式高推理、只读、单次裁决子代理。路由 JSON 不再被视为模型切换证据，宿主必须创建对应模型的真实子代理。 |
| `2.0.5` · 2026-07-13 | 曾补入 `wuji dispatch` 作为 CLI 契约检查适配器；随后确认它不能证明原生模型执行。CLI 请求模型、已生成结果与提供商侧实际模型、token、缓存、计费遥测必须严格区分。 |
| `2.0.6` · 2026-07-13 | 收紧 GPT 5.6 执行面：Terra 固定为阿极主脑与唯一写入者；子代理只允许精确调用 Sol 或 Luna，均单次、无自动回退。执行器在启动前拒绝 `gpt-5.4`、`gpt-5.4-mini`、Terra 和其他 GPT worker 名称；用户显式选择 Grok 或其他非 GPT 供应商/模型时不受此策略限制。 |
| `2.0.7` · 2026-07-13 | 主动路由与真实宿主收口：除简单问答外任务均路由；新增 `orchestrate` 与高风险 `change-capsule`，预检后有界并发执行。Windows 宿主改为直接调用 npm 的 Node 入口，避免包装层截断多行契约；拒收泛化问句和“被沙箱阻断/结果不可用”的伪交付。CLI 模型请求与供应商后台实际模型、token、计费继续严格分开陈述。 |
| `2.0.8` · 2026-07-13 | 修正真实模型执行面：Go CLI 只负责路由、上下文和契约校验；当前 Codex 宿主必须以路由声明的精确 Sol/Luna 模型创建原生子代理。原生 agent 标识与请求模型才是执行证据，CLI 参数不再冒充后台实际调用。 |
| `2.1.0` · 2026-07-17 | 建立任务熔断、需求/决策投影、执行图谱、安全门禁、官员建议与能力谱系可达性验证等运行底座，并验证阿极、参谋部和执行节点的职责边界。 |
| `3.0.0` · 2026-09-05 | 阿极融合跨领域 PonyTail 最小正确原则与白帽判断；参谋部保留名称但收束为确定性任务状态和调度机制；默认模型为 Terra，Terra 不可用时在生成前回退到 Sol，Luna 不作为阿极默认模型。真实宿主回执与独立验证继续是完成的唯一依据。 |

## 无极生态 / Wuji Ecosystem

这些仓库是“无极”生态的不同产品线和探索阶段，不是平行的军团中枢。它们只有通过 2.0 的能力契约、真实行为验证和进化门禁后，才可能进入当前主链。

| 仓库 | 定位 |
| --- | --- |
| [ComfyUI-WujiToolbox](https://github.com/AI-wuji/ComfyUI-WujiToolbox) | 面向 ComfyUI 的官方 API 直连工具。 |
| [wuji-dev-skill](https://github.com/AI-wuji/wuji-dev-skill) | 面向 Trae IDE 的开发 Skill 系统。 |
| [wuji-AIMotionComic-PromptSet](https://github.com/AI-wuji/wuji-AIMotionComic-PromptSet) | 无极 AI 漫剧的标准化生产指令集。 |
| [wuji-comfyui-skill](https://github.com/AI-wuji/wuji-comfyui-skill) | 面向 ComfyUI 插件开发、搜索、分析和融合的领域 Skill。 |
| [wuji-lobster-legion](https://github.com/AI-wuji/wuji-lobster-legion) | 多 Agent 协作与指挥链的早期探索。 |
| [wuji-dev-claude](https://github.com/AI-wuji/wuji-dev-claude) | 面向 Claude Code 的开发 Skill 系统。 |
| [wuji-legion-codex](https://github.com/AI-wuji/wuji-legion-codex) | 面向 ChatGPT Codex 的系统级 Skill 中枢，本仓库当前主线。 |

## 3.0 的核心原则 / Core Principles

> 规范角色边界：Terra Aji 是唯一用户沟通、需求表格/图谱、PonyTail 判断、能力组合和最终汇报入口；Terra 不可用时阿极在生成前回退到 Sol，Luna 不作为阿极默认模型。参谋部保留为确定性任务状态与调度机制，不是常驻 Sol 子代理；它不执行任务、不写制品、不合并、不验收。任务级 worker 只按自身声明链在生成开始前回退，一旦生成即保持模型与 session 粘性。完成必须同时有真实宿主执行和独立验证证据。

- **阿极融合 PonyTail，参谋部机制化**：阿极对所有领域先判断回答、行动或无需行动，优先复用已有能力，再选最小正确路径；参谋部只维护任务图、调度和证据状态，不作为模型 worker。
- **能力包优先于摘要**：保留真实的 Skill、脚本、模板、资产、UI 和入口。规则摘要不能冒充已经融合的能力。
- **冷挂载与有界并行**：专家不是常驻人格；只接收任务契约、必要上下文句柄和完整冷能力包。独立分支可以并行，依赖关系保持顺序。
- **证据驱动的能力生命周期**：`known -> doctrine-only -> assets-retained -> callable -> behavior-verified -> primary`。只有通过真实行为验证的能力才会被称为已融合。
- **上下文与图谱都有预算**：默认使用小型稳定前缀和任务级检索，不把完整历史、原始日志或大型图谱常驻在上下文中；工作区图先缩小候选，经验图只在事件触发时查询；跨模型不假设共享缓存，只有内容寻址工件和总重放成本同时通过门禁才委派。
- **先查现成方案，但严格区分预检与调研**：非平凡故障、依赖、路由、缓存、架构、迁移和集成任务先按“官方、GitHub、社区”做一次受限预检，默认最多 3 个来源、90 秒；找到决定性证据立即停止。用户明确要求全网或全面调研时，转入既有 search 能力，以证据覆盖、信息饱和和时间预算停止，不设三来源总数上限；确定性小修改和离线任务不搜索。
- **上游只作内部原子**：旧项目和外部 Skill 需要经过清点、对照测试和裁决；合适的部分蒸馏融合，不合适的部分剔除，不原封不动回迁。

## 当前状态 / Current Status

3.0 将阿极判断、机制化参谋部、可用性回退、跨领域 PonyTail 和证据门禁收敛到唯一入口。本轮优化的实现与验证状态以 [优化记录](references/release/optimization-2026-09-06.md) 为准；不以历史文案或单个 smoke 结果代替当前验收证据。

- 14 个能力包完成旧版迁移裁决；104 个旧版对象和 52 条工作树路径均有明确的更新、蒸馏、补落地或剔除结论。
- `fusion_audit`、`optimization_audit`、`context_bloat_audit`、Go 测试、静态检查、安全审计和压力测试都必须以当前提交对应的可复现输出为准；本页不把历史或局部运行表述为全量通过。
- 当前自动反馈仅为版本化终态执行结果的身份/结果候选事件，最多保留 1024 条或 2 MiB；它不是自动验证知识、路由改进或能力晋级。有限重试/续跑、工作区图清理时限与探针取消仍须保留失败路径与独立验证证据后才可标记完成。
- 默认 `verify` 对缺失或不完整的非 primary 冷源不作失败处理，但已发现的来源仍须完整校验；完整审计要求 secondary 可用，optional 缺失只作信息提示。
- 生产仓库不保存 API Key、Token、会话内容、`node_modules`、`dist` 或临时缓存。

能力的具体可信度以 `capabilities/*/manifest.json` 为准。`callable` 表示宿主可以挂载并调用；`behavior-verified` 和 `primary` 才表示已经通过对应行为验证，项目不会把 smoke 探针结果包装成完整融合。尚无真实行为证据、独立 SHA-256 和必要 promotion/baseline 的能力必须标为待证，不能称为已完成或已融合。

## 能力表面 / Capability Surfaces

3.0 将用户面能力收束为按场景挂载的统一入口：

| 场景 | 统一能力包 |
| --- | --- |
| 编码与审查 | `code`、`code-review` |
| 上下文、关系图与进化 | `context`、`knowledge`、`evolution` |
| 研究与数据 | `search`、`data` |
| 文档与演示 | `documents`、`presentation` |
| 前端与视觉 | `frontend`、`visual` |
| 图像与视频 | `image`、`video` |
| 写作与安全 | `writing`、`security` |

演示能力对外只暴露 `wuji-web-deck` 和 `wuji-editable-deck` 两个统一 Skill；HTML-PPT、Slidev、PPT Master、Huashu 等只作为内部模板、组件或验证资产，不再作为用户需要选择的并行产品。

## 快速开始 / Quick Start

在 PowerShell 中：

```powershell
git clone https://github.com/AI-wuji/wuji-legion-codex.git
cd wuji-legion-codex

./scripts/build.ps1
./bin/wuji.exe route --query "修改登录页并验证真实路由" --workspace . > .wuji/route.json
./bin/wuji.exe context-select --workspace . --query "fix code workerPlan in internal/core/route.go" --max-bytes 2048
```

随后由当前 Codex 宿主读取 `.wuji/route.json`，对每个符合条件的 `preflight_workers` / `workers` 以其精确 `model`、`session_key` 创建原生只读子代理；预检完成前不得启动后续 worker。`wuji dispatch` 和 `codex exec` 仅可用于本地契约兼容检查，不能作为模型已切换或已执行的证据。

### Feishu 冷源挂载

干净克隆不包含官方 `feishu-lark` 的完整 `official/` 树：该树是本地已有官方 Skill 的外部 junction，刻意不随仓库打包。要启用 Feishu 能力，先在无凭据前提下挂载一份已获准的完整官方 Skill 到 manifest 支持的任一路径：仓库内 `capabilities/feishu/skills/feishu-lark`，或用户级 `$env:USERPROFILE/.codex/skills/feishu-lark`。该挂载必须至少包含 `SKILL.md`、`agents/openai.yaml`、`official/lark-shared/SKILL.md`、`official/lark-doc/SKILL.md`、`official/lark-wiki/SKILL.md` 和 `official/lark-task/SKILL.md`；可用 junction 指向该本地已有副本，避免复制凭据或将其提交到仓库。

挂载只提供冷源指令，并不认证账号或授予云端访问。认证与实际读取/修改仍须由用户在宿主中显式完成；缺少该官方 Skill 时，克隆仓库本身不能宣称具备 Feishu 调用能力。

代码任务需要先用同一查询生成工件，路由器才会评估 Sol 委派是否实际划算：

```powershell
$query = "fix code workerPlan in internal/core/route.go"
$context = ./bin/wuji.exe context-select --workspace . --query $query --max-bytes 2048 | ConvertFrom-Json
./bin/wuji.exe route --query $query --context-artifact $context.artifact_path
```

快速回归检查：

```powershell
./scripts/test.ps1
```

完整验收（包含较慢的演示能力探针）：

```powershell
./scripts/test.ps1 -Full
```

安装 CLI 到当前会话或用户 PATH：

```powershell
./scripts/install-wuji-path.ps1
./scripts/install-wuji-path.ps1 -User
```

## 受控进化 / Controlled Evolution

只检查冷源：

```powershell
./scripts/update-cold-sources.ps1
```

评估候选能力（默认不写入）：

```powershell
./bin/wuji.exe evolve --candidate ./candidate-manifest.json
```

只有在同一 fixture 下完成上游与现有路线的真实对照、候选不降级并明确确认后，才使用 `--apply`。应用替换前会归档旧 manifest。

## 关系图检索 / Bounded Graph Retrieval

项目检索默认先走可重建的工作区关系图，再读取候选源文件。经验解决方案不会在每个任务启动时自动查询；只有遇到失败、显式复用、能力缺失或验证追踪事件时，才通过 `knowledge-query` 按索引定位解决方案。对符合条件的失败反馈，可在显式提交已验证知识时使用 `knowledge-record --feedback-id ... --feedback-store ...` 复用候选身份；该入口会验证候选资格和证据，但不会自动查询、自动写入知识或改变后续路由。图谱只返回位置和紧凑关系，正文按需读取；普通任务不允许全量扫描历史。

```powershell
./bin/wuji.exe graph-sync --workspace .
./bin/wuji.exe knowledge-query --trigger explicit-reuse --kind failure --key "browser timeout" --scope global
```

金字塔层级只负责缩小检索展开范围，不能自动阻止事实层增长。因此 2.0 同时使用节点替换、显式作用域、陈旧重建、验证文件 SHA-256、索引引用上限、候选/结果预算和受限 fallback。真实白帽也遵循同一证据规则：路由里的 `internal_adversarial_pass` 不是 officer 已执行的证明，只有独立审查产物存在并通过校验才算完成。

## 模型与提供商边界 / Model Boundaries

- 默认或显式 GPT 选择采用同一单链：Terra 阿极只负责用户交流、需求表格图谱、跨领域 PonyTail 判断、能力组合与最终汇报；Terra 不可用时阿极在生成前回退到 Sol。每个需要执行的任务由确定性参谋部维护任务图、调度、回执失败和需求复核；它不是模型 worker，不能执行、写制品、合并或验收。执行节点的模型按任务选择，不能把参谋部状态当作执行证据。
- 除简单、自包含问答外，所有任务均积极路由。`wuji route` 输出的 `preflight_workers` 必须先于 `workers` 执行，两阶段禁止并行；预检改变方案时必须作废旧计划并重新路由。阿极默认可用性链为 `gpt-5.6-terra -> gpt-5.6-sol`，任务级 worker 可按其声明链在生成前且出现模型不可用或生成前提供商错误时回退。每个 GPT worker 都包含稳定 `session_key`，一旦生成开始即固定模型，不允许 A/B 验证、质量重试或生成后切换。显式非 GPT 选择保留既有 capability/provider mode，不生成 GPT worker。执行宿主必须按结果真正委派；路由 JSON 不等于执行。需要生成或修改代码、演示、写作及其他制品时，执行节点获得 `scoped-artifact-write`，参谋部和阿极保持制品只读。
- `model_class` 只是分类标签，不能作为模型切换已经发生的证据。只有当前 Codex 宿主按精确模型创建的原生 agent 标识与结果，才能证明该分支被真实委派；`wuji dispatch`、`codex exec` 的参数或输出仅能证明 CLI 契约。后台对应模型调用可进一步佐证提供商侧实际模型；当前 CLI 不提供 token、缓存和计费遥测，不能伪造节省。
- 每个执行节点必须保留宿主派发标识、session key、模型尝试、payload 哈希/字节、失败种类和独立行为验证；没有真实执行和验证证据不得算分支完成。没有供应商遥测时模型实际标识、token/cache、计费基线、实际成本和节省额必须明确为不可用，不能伪造。
- `primary` 不是 manifest 自报标签：只有演化替换生成并通过校验的内容寻址 promotion receipt，连同归档基线，才可进入 `primary`。
- 图像和视频提供商按项目规则路由，并在失败时回退；凭据只从当前进程环境读取，绝不写入规则或仓库。
- Sol、Terra、Luna 之间不假设共享提示缓存。代码上下文覆盖率低于 60%、没有代码摘录、任务契约超过 2048 字节、共享上下文超过 4096 字节、总重放超过 8192 字节、工件不匹配/已过期或需要父任务上下文时，执行节点不得接收不完整代码上下文，staff 必须缩小任务或重新规划。worker 按“稳定能力前缀、上下文 capsule、任务契约”的固定顺序接收确定性 prompt，并在任务开始选定模型后保持 session 粘性；GPT 5.6 worker 一次尝试、无模型切换。低质量结果由任务图中的验证节点判为未通过并返回 staff 继续调度，不再降级、切换模型或付费重试。

## 目录说明 / Repository Layout

- `SKILL.md`：系统级热路径规则。
- `capabilities/*/manifest.json`：能力生命周期、来源、入口和探针的事实源。
- `cmd/`、`internal/`：确定性 Go CLI 和运行时核心。
- `scripts/`：构建、审计、测试、安装和冷源更新脚本。
- `migration/`：旧版对象与工作树路径的逐项迁移裁决证据。
- `references/`：架构、能力契约和运行时边界文档。

旧版已从当前主线退出，但发布时会保留在 `legacy-v1-backup` 分支和 `legacy-v1-final` 标签中，便于回溯和对照；主分支承载 3.0。

## 许可与凭据 / License and Credentials

本仓库不包含 API Key、Token 或 Codex 会话内容。使用外部提供商时，请通过当前进程环境或宿主的安全凭据机制提供密钥，并在发布前检查工作树和构建产物。
