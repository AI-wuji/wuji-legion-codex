# 无极军团 2.0 / Wuji Legion 2.0

无极军团 2.0 是一个面向 ChatGPT Codex 的系统级 Skill 中枢，也是一个有界的多 Agent 协作与能力进化系统。它不是一组领域 Skill 的简单集合，也不是多个 Agent 各自决策的聊天框架，而是负责理解任务、选择并调用合适的 Skill 和工具，组织彼此独立的 Agent 并行执行，再由唯一主脑统一合并结果、验证真实产物并完成交付。

无极军团的另一项核心能力是对 Skill、插件和工作流进行蒸馏、融合与进化：保留有价值的脚本、模板、资产和入口，通过真实行为对照决定能力的接入、替换、升级或淘汰，让分散的外部能力逐步沉淀为可调用、可验证、可演化的系统能力。2.0 以冷挂载完整能力包、确定性 Go 路由、稀疏上下文选择和进化门禁支撑这条主链路。

Wuji Legion 2.0 is a system-level Skill hub for ChatGPT Codex: a bounded multi-agent coordination and capability-evolution system. It does not replace domain Skills; it understands the task, selects and invokes the right Skills and tools, coordinates independent agents in parallel, and lets one main brain merge the work, verify real artifacts, and complete delivery.

Its second core function is distilling, fusing, and evolving Skills, plugins, and workflows. Wuji retains valuable scripts, templates, assets, and entrypoints, then uses real behavioral comparisons to admit, upgrade, replace, or retire capabilities. Cold-mounted complete packages, deterministic Go routing, sparse context selection, and evolution gates keep this orchestration chain bounded and verifiable.

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

## 2.0 的核心原则 / Core Principles

- **单一主脑，单一写权限**：阿极负责需求归一化、依赖分析、并行分发、最终合并和完成判断。2.0 不再使用 Nuwa、第二路由、默认会审或常驻桥接。
- **能力包优先于摘要**：保留真实的 Skill、脚本、模板、资产、UI 和入口。规则摘要不能冒充已经融合的能力。
- **冷挂载与有界并行**：专家不是常驻人格；只接收任务契约、必要上下文句柄和完整冷能力包。独立分支可以并行，依赖关系保持顺序。
- **证据驱动的能力生命周期**：`known -> doctrine-only -> assets-retained -> callable -> behavior-verified -> primary`。只有通过真实行为验证的能力才会被称为已融合。
- **上下文与委派都有预算**：默认使用小型稳定前缀和任务级检索，不把完整历史、原始日志或大型图谱常驻在上下文中；跨模型不假设共享缓存，只有内容寻址工件和总重放成本同时通过门禁才委派。
- **上游只作内部原子**：旧项目和外部 Skill 需要经过清点、对照测试和裁决；合适的部分蒸馏融合，不合适的部分剔除，不原封不动回迁。

## 当前状态 / Current Status

2.0 已完成一次完整基线审计和发布前验收：

- 14 个能力包完成旧版迁移裁决；104 个旧版对象和 52 条工作树路径均有明确的更新、蒸馏、补落地或剔除结论。
- `fusion_audit`、`optimization_audit`、`context_bloat_audit` 均通过。
- Go 单元测试、静态检查、Skill 校验、依赖安全审计和并发输出压力测试通过。
- 生产仓库不保存 API Key、Token、会话内容、`node_modules`、`dist` 或临时缓存。

能力的具体可信度以 `capabilities/*/manifest.json` 为准。`callable` 表示宿主可以挂载并调用；`behavior-verified` 和 `primary` 才表示已经通过对应行为验证，项目不会把 smoke 探针结果包装成完整融合。

## 能力表面 / Capability Surfaces

2.0 将用户面能力收束为按场景挂载的统一入口：

| 场景 | 统一能力包 |
| --- | --- |
| 编码与审查 | `code`、`code-review` |
| 上下文与进化 | `context`、`evolution` |
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
./bin/wuji.exe route --query "修改登录页并验证真实路由"
./bin/wuji.exe context-select --workspace . --query "fix code workerPlan in internal/core/route.go" --max-bytes 2048
```

代码任务需要先用同一查询生成工件，路由器才会评估 Terra 委派是否实际划算：

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

## 模型与提供商边界 / Model Boundaries

- 阿极负责规划、架构、合并、唯一写权限和高风险判断，控制面固定使用 `gpt-5.6-sol`；这不代表所有分支都应消耗 Sol。
- `wuji route` 输出的每个 worker 都包含真实 `model` 与有序 `fallback_models`。通过成本门禁的独立实现分支可使用 `gpt-5.6-terra`，依赖实现的验证由阿极顺序完成；广域研究与机械提取分支可使用 `gpt-5.6-luna`，执行宿主必须按结果真正委派。演示与写作默认不分发，只有显式并行且交接自包含时才允许降级。
- `model_class` 只是分类标签，不能作为模型切换已经发生的证据；只有后台出现对应模型调用，或者执行记录确认了实际 fallback，才算模型路由生效。
- 每个 worker 必须返回路由声明的完整执行回执，包括模型尝试、实际模型、payload 哈希/字节、token/cache、计费基线、实际成本、节省额和阿极验收；只有回执通过 `validate-receipt`，才算该分支完成且确实省钱。
- `primary` 不是 manifest 自报标签：只有演化替换生成并通过校验的内容寻址 promotion receipt，连同归档基线，才可进入 `primary`。
- 图像和视频提供商按项目规则路由，并在失败时回退；凭据只从当前进程环境读取，绝不写入规则或仓库。
- Sol、Terra、Luna 之间不假设共享提示缓存。代码上下文覆盖率低于 60%、没有代码摘录、任务契约超过 2048 字节、共享上下文超过 4096 字节、总重放超过 8192 字节、工件不匹配/已过期或需要父任务上下文时，一律留在阿极。worker 按“稳定能力前缀、上下文 capsule、任务契约”的固定顺序接收确定性 prompt；Terra/Luna 只在模型不可用或生成前 provider 错误时直退 Sol，生成后的低质量结果由阿极拒绝，不再跨模型付费重试。

## 目录说明 / Repository Layout

- `SKILL.md`：系统级热路径规则。
- `capabilities/*/manifest.json`：能力生命周期、来源、入口和探针的事实源。
- `cmd/`、`internal/`：确定性 Go CLI 和运行时核心。
- `scripts/`：构建、审计、测试、安装和冷源更新脚本。
- `migration/`：旧版对象与工作树路径的逐项迁移裁决证据。
- `references/`：架构、能力契约和运行时边界文档。

旧版已从当前主线退出，但发布时会保留在 `legacy-v1-backup` 分支和 `legacy-v1-final` 标签中，便于回溯和对照；主分支只承载 2.0。

## 许可与凭据 / License and Credentials

本仓库不包含 API Key、Token 或 Codex 会话内容。使用外部提供商时，请通过当前进程环境或宿主的安全凭据机制提供密钥，并在发布前检查工作树和构建产物。
