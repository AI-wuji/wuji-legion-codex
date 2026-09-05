<div align="center">

# 无极军团 3.0

**让不懂 Skill、插件和 MCP 的人，也能用自然语言调动 AI，把事情办成。**

*Your goals, in plain language. An AI legion for Codex.*

[![Platform](https://img.shields.io/badge/Platform-Codex-111111?style=flat-square)](https://openai.com/codex/)
![Version](https://img.shields.io/badge/Version-3.0-0b6e4f?style=flat-square)
![Development](https://img.shields.io/badge/Development-Public-2563eb?style=flat-square)

[开始使用](#开始使用) · [它解决什么](#它解决什么) · [当前状态](#当前状态) · [文档](#验证与文档) · [受控进化](#受控进化)

</div>

无极军团不是又一个要求用户学习指令、插件名和模型名的工作流集合。你只需说明目标；阿极负责理解任务、提出必要质疑，并在已验证的 Skill、插件、MCP、工具与执行节点之间选择最小正确路径。

它不把“自动化”当成盲从。阿极采用白帽判断：证据不足、要求不合理或方案有风险时，会明确说出来，而不是为了迎合给出看似顺从的答案。

## 它解决什么

| 你面对的情况 | 无极军团的处理方式 |
| --- | --- |
| 不知道该用哪个 Skill、MCP 或插件 | 从已验证能力中按任务选择，不要求先记住工具名。 |
| 任务跨越代码、研究、文档、设计或数据 | 按场景挂载最小能力包，避免把所有规则和上下文塞进一次请求。 |
| 自动化容易越跑越多、越改越乱 | 路由、重试、图谱、上下文和反馈都有显式上限及证据门槛。 |

## 三个坚持

1. **自然语言优先**：用户描述想要的结果；系统负责选择合适的已验证能力。
2. **先判断，再行动**：阿极会复用已有路径、指出错误和风险，拒绝无依据的复杂化。
3. **越用越好用，但不自我吹捧**：失败、复用和验证结果可以沉淀为有界候选；只有真实对照与独立证据才能升级能力。

## 如何工作

```text
用户的自然语言目标
        |
        v
阿极：理解、PonyTail 最小正确判断、白帽提醒、最终沟通
        |
        v
确定性参谋部：任务图、依赖、预算、状态与失败回执
        |
        v
按需执行节点：挂载经验证能力，执行并接受独立验证
```

阿极是唯一面向用户的入口。参谋部是确定性状态和调度机制，不是常驻模型代理，也不写制品或宣布完成。真正的执行必须有宿主产生的执行回执和独立验证证据。

按当前可用能力与宿主授予的权限选择，不把未安装、未验证或未授权的能力说成已经可运行。

| 你会怎么说 | 可能按需挂载的能力包 |
| --- | --- |
| “帮我找出这个登录问题并修好。” | `code`、`code-review` |
| “把这些资料做成一份能讲清楚的汇报。” | `documents`、`presentation`、`writing` |
| “分析这份表格，告诉我异常和结论。” | `data` |
| “做一份有来源、有证据的行业调研。” | `search`、`writing` |
| “给这个产品页做前端和视觉调整。” | `frontend`、`visual`、`image`、`video` |

## 当前状态

本仓库提供 Go CLI 与 3.0 的有界路由、任务状态、上下文/图谱和能力门禁机制。它不是对所有宿主能力的完成宣言。

| 已有机制 | 仍未完成或不可宣称 |
| --- | --- |
| 路由、图规模、重试、上下文、探针和候选反馈均有配置的边界 | 原生 Codex 宿主的真实自动派发与回执接入尚未完成。 |
| 候选反馈仅记录版本化终态的身份/结果，最多 1024 条或 2 MiB | 候选不会自动成为已验证知识、自动改变路由或自动晋级能力。 |
| `knowledge-record` 可在显式提交时对合格候选做证据门控复用 | 已验证经验到路由复用的自动学习闭环尚未完成。 |
| 能力需经 callable、行为验证、对照与晋级门禁 | 没有宿主计费遥测时，不存在全局 token 或费用硬上限。 |

最近验收的真实范围、fast audit 的未通过项和后续限制见[优化记录](references/release/optimization-2026-09-06.md)。该记录明确：修复后的全包 Go 测试、vet 和 build 已通过；fast audit 仍被三项 size gate 拦截，因此不能称为全绿发布。

## 开始使用

**前提**：Windows PowerShell、Git，以及 Go 1.25+。构建脚本不会自动下载 Go。克隆后构建 CLI：

```powershell
git clone https://github.com/AI-wuji/wuji-legion-codex.git
cd wuji-legion-codex
$env:WUJI_GO = (Get-Command go).Source
./scripts/build.ps1
./bin/wuji.exe route --query "修改登录页并验证真实路由" --workspace .
```

`route` 输出的是任务契约，不会自行执行。宿主必须按其中的阶段、精确模型与 `session_key` 创建原生执行节点，并按任务契约授予所需的受限读写权限；JSON 或 CLI 参数本身不构成模型已调用、制品已写入或任务已完成的证据。

可选择安装到当前会话或用户 PATH：

```powershell
./scripts/install-wuji-path.ps1
./scripts/install-wuji-path.ps1 -User
```

后者需要重启 Codex 或终端以取得新的 PATH。

### Feishu 冷源挂载

干净克隆刻意不包含官方 `feishu-lark` 的完整 `official/` 树；它是本地已有官方 Skill 的外部 junction，不能被当作仓库自带能力。启用前，在不复制凭据的前提下，将获准的完整官方 Skill 挂载到以下任一路径：

```text
capabilities/feishu/skills/feishu-lark
$env:USERPROFILE/.codex/skills/feishu-lark
```

挂载至少需包含 `SKILL.md`、`agents/openai.yaml`、`official/lark-shared/SKILL.md`、`official/lark-doc/SKILL.md`、`official/lark-wiki/SKILL.md` 与 `official/lark-task/SKILL.md`。挂载只提供冷源指令，不认证账号，也不授权云端访问。

## 受控进化

系统可以保留真实失败、来源评估、复用结果和验证线索，让后续选择有可追溯的经验；但“自动记录”不等于“自动接纳”。候选必须通过真实对照、独立行为验证、内容哈希和晋级门禁，才可能成为可用能力或 `primary`。

```powershell
# 检查冷源，不写入
./scripts/update-cold-sources.ps1

# 评估候选；默认不修改任何能力
./bin/wuji.exe evolve --candidate ./candidate-manifest.json
```

对已验证的失败候选进行显式知识记录时，可在完整的 `knowledge-record` 命令上额外传入 `--feedback-id` 与 `--feedback-store`；字段要求以 `./bin/wuji.exe knowledge-record --help` 为准。

## 验证与文档

```powershell
# 快速回归
./scripts/test.ps1

# 完整验收（包含较慢的演示能力探针）
./scripts/test.ps1 -Full
```

| 需要了解 | 从这里开始 |
| --- | --- |
| 当前架构与运行时边界 | [Codex 3.0 当前架构](references/architecture.md) |
| 3.0 目标设计 | [完整方案](references/architecture/wuji-legion-3.0-blueprint.md) · [全图谱架构](references/architecture/wuji-legion-3.0-graph-architecture.md) |
| 能力的证据等级与验证契约 | [能力契约](references/capability-contract.md) · `capabilities/*/manifest.json` |
| 本轮优化的实际验收状态 | [优化记录](references/release/optimization-2026-09-06.md) |
| 项目规则与入口 | [SKILL.md](SKILL.md) · [AGENTS.md](AGENTS.md) |
| 同一初心的其他宿主实现 | [dsh-wuji-legion-global](https://github.com/AI-wuji/dsh-wuji-legion-global) · [dsh-wuji-legion-mode](https://github.com/AI-wuji/dsh-wuji-legion-mode) |

<details>
<summary>开发说明与演进</summary>

无极军团 3.0 延续的是“保留可验证资产、淘汰不可验证复杂度”的主线，而不是把旧项目原样搬到 Codex。2.0 建立了冷挂载能力包、确定性 Go 路由、稀疏上下文选择和演化门禁；3.0 将阿极的跨领域 PonyTail 最小正确判断与白帽立场收束到唯一用户入口，并把参谋部明确为非模型的任务状态机制。

能力生命周期为 `known -> doctrine-only -> assets-retained -> callable -> behavior-verified -> primary`。只有 `callable`、`behavior-verified` 与 `primary` 可以激活；smoke 只能证明可调用，不能证明已融合。旧版可从 `legacy-v1-backup` 分支与 `legacy-v1-final` 标签回溯。

### 项目演进

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

</details>

## 凭据与贡献

仓库不保存 API Key、Token、Codex 会话内容或本地构建缓存。外部服务凭据只应通过宿主或当前进程环境提供；提交前请检查工作树与构建产物。
