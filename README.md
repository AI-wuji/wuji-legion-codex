# 无极军团 Codex 版 / Wuji Legion for Codex

**让 Codex 从“会回答”升级成“真交付”。**  
无极军团不是再堆一堆 skill，也不是花哨的多 agent 表演。它是一套面向真实交付的执行系统：自动收敛角色、压低 token、提高命中率，把调研、代码、内容、PPT、HTML、配图、插件和规则升级都收口到可验证的结果。

**一句话卖点**  
更省 token，更少废话，更高命中，更稳交付。

## 为什么会让人想用

- 你不用自己学复杂工具链，阿极就是统一入口。
- 你不用反复提醒“别跑偏、别半成品、别假完成”，底层门禁会先拦。
- 你不用在简单任务上浪费高成本模型，它会先走低成本，再按难度升级。
- 你不用把历史偏好一次次重讲，它会把可用偏好蒸馏成离线优化闭环。
- 你不用担心一堆角色吵架，它默认单主帅负责到底，只有真缺能力才补位。

## 它卖的不是概念，卖的是结果

- 调研：查得更准，结论更短，更少信息噪音。
- 代码：Go 执行底座负责硬门禁，功能链路可构建、可测试、可审计。
- 内容：写作、脚本、方案、教程不再像 AI 拼接稿。
- PPT / HTML / UI：先走真正的生产引擎，再做预览、校验和修复，减少返工。
- 安全：白帽、质检、安全、合规独立，不让执行者自己给自己放行。
- 进化：只蒸馏有用能力，不做无穷叠加。

## 最适合谁

- 想把 Codex 当长期生产力系统的人
- 不会编程、但要高质量成品的人
- 讨厌半成品、试验版、待开发的人
- 在乎成本、稳定、命中率和最终质量的人

## 当前收口能力

- Go 执行底座：`wuji-cli` 已覆盖 `reference-guard`、`workflow-guard`、`claim-guard`、`time-guard`、`task`、`sync`、`audit`、`bench`、`bench-report`、`canon-report`、`route-task`、`context-pack`、`preview`、`asset-map`、`pptx-audit`、`pptx-preflight`、`pptx-batch-gate`、`mcp-guard`、`mcp-distill`、`feedback-log`、`feedback-dataset`、`prompt-candidate-audit`、`prompt-eval`、`prompt-distill`。
- Go 执行底座：`wuji-cli` 已覆盖 `reference-guard`、`workflow-guard`、`claim-guard`、`time-guard`、`task`、`sync`、`audit`、`bench`、`bench-report`、`code-map`、`closeout-check`、`canon-report`、`route-task`、`context-pack`、`preview`、`asset-map`、`pptx-audit`、`pptx-preflight`、`pptx-batch-gate`、`mcp-guard`、`mcp-distill`、`feedback-log`、`feedback-dataset`、`prompt-candidate-audit`、`prompt-eval`、`prompt-distill`。
- 模型分档：默认 `gpt-5.4-mini`，中档 `gpt-5.4`，高档 `gpt-5.5`，按任务难度推荐。
- 静态骨架：固定模型三档、核心路由骨架、内置插件归口和 MCP 默认裁决已沉入 Go 底座，`config.json` 只保留环境信息和覆盖项。
- 图像直出：普通 `生图/插图/海报/封面/配图` 已从 `ComfyUI` 重链剥离，`route-task` 直接落到 `imagegen` 低档直出链。
- PPT 主链：模板续写走 `Presentations template-following exact clone/edit`；从零高颜值走 `HTML-first -> editable PPTX`；Windows PowerPoint COM/MCP 只做最后一公里精修；Go 负责锁三张表、`style-lock`、`page-role-policy`、pilot 放行和收口 QA，默认主线不再先跑 `pptx-preflight`。
- 当前 `HTML-first` 的真实边界已明确：它优先走 `Playwright + dom-to-pptx` 的浏览器计算样式导出，能高保真保留静态 HTML/CSS 视觉，但仍不把 HTML/CSS 的动画、过渡和动态组件自动转成 PowerPoint 动画。
- 自进化闭环：`feedback-log -> feedback-dataset -> prompt-candidate-audit -> prompt-eval -> prompt-distill`。
- 交付铁律：只交最终结果，不交半成品，不拿路线表演冒充执行。
- Codex Use Cases 精华点已吸收为少数硬机制：持续目标直跑到底、复杂代码先出最小代码地图、重复工作优先沉成 skill/CLI、外部与批量操作必须留证据、前端默认真实浏览器验收、安全结果必须证据分级。

---

## v10.6 的关键变化 / Key Change

v10.6 新增 `pilot-page` 快速闭环：PPT/HTML 视觉成品不再一口气批量试错，三张表锁路后先生成 1 页最高风险/最高密度/最能代表风格的 pilot page，记录 `pilot-score`。首套新模板/新路线/高风险风格变更必须用户明确批准后再批量；成熟同路线默认允许自动批准继续批量，但仍必须留下 `pilot-approval` 工件。执行底座新增 `pptx-batch-gate`，缺 `pilot-page`、`pilot-preview`、`pilot-score`、`pilot-approval`，或 preview 发白/低对比，直接 `NO-GO`。

PPT 主链同时完成整流：不再把 Go 门禁当成 PPT 生产引擎。模板续写/补页固定切到官方 `Presentations`；从零高颜值固定补进 `HTML-first -> editable PPTX`；PowerPoint COM/MCP 只保留为最后一公里精修；原有 Go 链降到后置 QA 和安全护栏。这里的 `HTML-first` 当前主引擎已切到 `dom-to-pptx`，属于“高保真静态 HTML/CSS 导出”，不再是旧的低保真重画链。

教程、课程、说明、方案类 PPT 的默认内容顺序也已收口：`source -> outline -> speaker-notes -> slide-spec/design-brief -> reference-frame-map -> reusable-asset-map -> illustration-plan -> pilot-page -> pptx-batch-gate -> full PPTX -> notes -> QA`。除非用户明确要先审大纲，否则内部自动直跑；逐字稿默认进备注区，不再硬塞进正文。

教学页现在不允许只剩大框架：如果页面包含步骤、操作、界面、按钮、导入导出等教学信号，系统会默认生成 `outline / speaker-notes / illustration-plan` 工件，并把“截图 / 步骤示意图 / image2 教学插图 / 复用参考图框”写成显式策略；`pptx-batch-gate` 会拦截缺视觉策略的教学页。

这次又把两条高频跑偏点彻底写死进了 PPT 开发逻辑。第一，`style-lock` 会把整套 deck 的风格名、背景深浅、霓虹/高光语言、插图语言和禁止项固化下来；像“霓虹赛博卡通风”这种用户明确点名的风格，后续必须原样进入 image2 / 配图提示，不能再静默洗成白底或写实风。第二，`page-role-policy` 会把首页、目录页、单元页、总结页、结尾页这些固定页型锁住，后续不得再拿这些页去塞普通内容页。

为压速度，PPT 主链还额外固化了几条默认跳过规则：不再重复跑 `pptx-preflight`；同路线成功过后默认复用已有 inspect，不再重复 template inspect；html-first 默认只渲染 pilot 所需最小预览，不做整套 layout 体检；模板续写默认跳过无必要的 final preview 重渲染；同任务里已经验证过的工具链，不再为“确认能不能做”重复探路。

当前无极执行底座主链路统一为 Go：`wuji-cli` 已覆盖 `reference-guard`、`workflow-guard`、`claim-guard`、`time-guard`、`task`、`sync`、`audit`、`bench`、`bench-report`、`canon-report`、`route-task`、`context-pack`、`preview`、`asset-map`、`pptx-audit`、`pptx-preflight`、`pptx-batch-gate`、`mcp-guard`、`mcp-distill`、`feedback-log`、`feedback-dataset`、`prompt-candidate-audit`、`prompt-eval`、`prompt-distill`，并统一调度专项补位工具；但在 PPT 体系里，它只负责锁路线、批量放行和收口 QA，不再充当主生产器，`pptx-preflight` 只在定向探针场景触发。
模型档位当前收口为 3 档：默认 `gpt-5.4-mini`，中档 `gpt-5.4`，高档 `gpt-5.5`；由 `route-task` 基于 Go 内置路由骨架推荐 `tier + reasoning_effort`，再按需叠加 `config.json` 覆盖项，以“先低成本、再按需升级”为默认原则。
普通图像任务现在也加了确定性收口：普通 `生图/插图/海报/封面/配图` 直接走 `imagegen`，不再允许先查环境、先读系统 skill、先看 key、先试通道。
前台执行提示统一压成中文短句：执行前只报 `建议：5.4-mini 低` / `建议：5.4 中` / `建议：5.5 高` / `建议：5.5 超高`；任务完成后默认报 `建议切回：5.4-mini 低`。
提示词自进化当前已收口为离线闭环：`feedback-log -> feedback-dataset -> prompt-candidate-audit -> prompt-eval -> prompt-distill`，只沉淀脱敏偏好信号，不把运行时上下文越堆越肥。
`prompt-candidate-audit` 已把这类生图前探路话术列为失败项，防止规则说对了、执行又绕回去。

## v10.5 的关键变化 / Key Change

v10.5 针对 PPT 长时间空转和低质交付加前置硬门禁：口头说“直接生成”不算执行；非代码任务 10/15/30 分钟分级熔断；参考 PPTX 必须同时当作风格系统和素材库，批量生成前先锁定 `reference-frame-map`、`reusable-asset-map`、`illustration-plan`；优先复用表格、图标、卡片、流程箭头、章节页、背景装饰、教学插图、案例图、公式图和 image2 生图资产；真 PPTX 不得每页一张整图冒充可编辑成品。执行底座对占位符残留、空白页、模板碎片页和低对比 preview 做确定性硬拦截。

## v10.4 的关键变化 / Key Change

v10.4 新增 `执行底座 / 执行底座主帅`：把无极军团自身稳定、重复、可判定的动作下沉为确定性执行底座，负责 `wuji-cli`、guard、task、sync、audit、workflow、beep、bench、preview调度。它不替代开发主帅，不抢普通 Go/Tauri/业务代码，也不替代参谋本部、女娲、白帽、质检、安全、合规和进化判断。

## v10.3 的关键变化 / Key Change

v10.3 增加非代码交付先执行铁律：PPT、文档、图片、HTML演示稿、逐字稿等成品任务，默认先直接生成主成品；不允许开工前把时间耗在反复验证工具链上，除非遇到真实错误或安全风险。

## v10.2 的关键变化 / Key Change

v10.2 吸收 `DannyMac180/skills` 的动态工作流机制，但不新增入口：复杂 `LEGION_TASK` 必须留下最小可审计轨迹，包含目标、成功标准、任务切片、验证结果和最终收口；简单任务不启用，避免变慢。

## v10.1 的关键变化 / Key Change

v10.1 增加参考文件只读铁律：用户提供、点名、上传、要求“参考/借鉴/按照/对照”的文件默认只读；生成物、修复版、重做版必须另存为新文件，不得覆盖参考原件。

v10.0 修正 LEGION_TASK 触发口径：不再只看用户有没有喊“激活无极军团”。只要任务本身需要多能力协作，就必须进入 `LEGION_TASK`，并用参谋本部接管格式开场。

例如：根据大纲参考上节课 PPT，生成新 PPT 和逐字稿。

---

## v9.9 的关键变化 / Key Change

v9.9 修正“声音不是在最后响”的体感问题：由于工具只能在最终回复前运行，收尾提醒改为后台延迟响铃。

- 最终回复前调度：`.\scripts\beep.ps1 complete -SpawnDelayed -DelayMs 1200`
- 最终文字发出后约 1.2 秒响铃。
- 这样听感更接近微信/QQ的消息结束提醒。

---

## v9.8 的关键变化 / Key Change

v9.8 修正提示音触发时机：提示音不是“验证中响一下”，而是任务真正结束前的最后提醒。

- 非轻量对话任务收尾时，`beep.ps1` 必须尽量作为最终回复前的最后一个工具动作。
- 成功完成用 `.\scripts\beep.ps1 complete`。
- 阻塞或失败用 `.\scripts\beep.ps1 error`。

---

## v9.7 的关键变化 / Key Change

v9.7 增加任务收尾提示音：执行型任务完成、阻塞或失败收口前，会优先调用 [beep.ps1](E:\wuji-projects\wuji-legion-codex\scripts\beep.ps1) 生成临时 WAV 并播放提示音，避免多窗口工作时错过结果。

- 完成任务：`.\scripts\beep.ps1 complete`
- 阻塞或失败：`.\scripts\beep.ps1 error`
- 轻提醒：`.\scripts\beep.ps1 notify`

纯聊天/身份问答这类 `FAST_REPLY` 不强制响铃，避免日常对话太吵。

---

## v9.6 的关键变化 / Key Change

v9.6 做的是质量整流：把“能跑”继续推进到“干净、可安装、可验证”。

- 根 `SKILL.md` 已补齐 Codex skill frontmatter。
- 安装、恢复、同步、推送脚本已清掉旧路径、旧版本和破坏式逻辑。
- 远征军 Trae 工兵已从旧项目专用说明改为无极军团通用 handoff 工兵。
- 打靶场已支持 `skill` 目录扫描，可检查 frontmatter、乱码和占位残留。
- 本轮验证通过：专家生成 15 张；PowerShell 语法通过；Python 编译通过；打靶场 116/116 通过。

---

## v9.5 的关键变化 / Key Change

旧版专家库是“很多专家卡”。v9.5 改成：

```text
师团万能主帅入口
-> 内置多模式
-> 按任务切换
-> 独立白帽/质检/安全/合规审查
```

压缩的是单个师团内部入口，不是把整个无极军团压成一个超级大脑。

当前专家库从 `44` 张主责卡继续压缩为 `16` 张高密度卡：

- 参谋主帅
- 内容主帅
- 视觉主帅
- 提示词主帅
- 开发主帅
- 执行底座主帅
- ComfyUI主帅
- 情报主帅
- 安全主帅
- 合规审计官
- 白帽纠察官
- 质检主帅
- 性能基准官
- 进化主帅
- 交付主帅
- 归档主帅

---

## 现在实现了什么 / What It Enables

| 方向 | 主帅 | 内置能力 |
|---|---|---|
| 内容 | 内容主帅 | 小说、剧本、分镜、教程、计划书、营销方案、卖点提炼、短内容、人味改稿 |
| 视觉 | 视觉主帅 | 真PPTX、HTML演示、UI页面、数据可视化、信息图、配图 |
| 开发 | 开发主帅 | Go/Tauri、前端、小程序、ComfyUI插件、AI工程、自动化、原型 |
| 执行底座 | 执行底座主帅 | 无极执行底座、wuji-cli、guard、task、sync、audit、workflow、beep、bench、preview调度、pptx-preflight、pptx-batch-gate、pptx-audit、asset-map、time-guard |
| 情报 | 情报主帅 | 全网搜索、GitHub源码核验、趋势、用户研究、本地化 |
| 安全 | 安全主帅 | 威胁建模、漏洞验证、供应链、发布安全 |
| 审查 | 白帽/质检/合规/性能 | 前置封驳、最终验收、许可证/隐私、速度/token基准 |
| 进化 | 进化主帅 | 查源、裁决、实验、复盘、专家瘦身 |

---

## 借鉴了什么 / What It Distills

为避免误会，这里明确说明：无极军团 Codex 版参考、蒸馏、整流了多条开源 skill / 工作流，但没有把上游项目原样搬运进来。

它借走的是机制，不是名字：

- `openai/skills`、`anthropics/skills`：skill 结构、按需加载、资源分层、上下文节省。
- `addyosmani/agent-skills`：阶段识别、上下文工程、薄切片实现、五轴 review、安全门禁。
- `github/awesome-copilot`：大规模技能索引、bundled assets、Rust/QA/安全细分规则。
- `marketingskills`：产品、受众、定位先行，营销技能互相关联但有主线。
- `humanizer`：AI写作痕迹识别、作者声音匹配。
- `powerpoint-skill`：视觉优先、密度边界、重叠检查、预览验证。
- `ppt-master`、`elite-powerpoint-designer`、`slide-studio`、`Presentations`：真 PPTX 分阶段、可编辑交付、审美上限和验证闭环。
- `SkillClaw`、`Edict`：任务后进化、候选验证、前置封驳和状态审计。
- `DannyMac180/skills`：动态工作流工件、packet 切片、结果收集和验证脚手架。

来源、commit 和裁决记录见 [distillation.md](E:\wuji-projects\wuji-legion-codex\units\distillation.md)。

---

## 组织结构 / Organization

无极军团总结构仍然保留：

- 阿极：默认入口。
- 参谋本部：分拣、路由、验收标准。
- 女娲：按需补位，不默认组队。
- 各师团主帅：执行本师团任务。
- 白帽/质检/安全/合规：独立第三方审查。
- 进化部：查源、蒸馏、实验和规则整流。

---

## 快速开始 / Quick Start

```powershell
.\scripts\wuji-install.ps1
```

关键文件：

- [GLOBAL_AGENTS.md](E:\wuji-projects\wuji-legion-codex\GLOBAL_AGENTS.md)
- [SKILL.md](E:\wuji-projects\wuji-legion-codex\SKILL.md)
- [experts/INDEX.md](E:\wuji-projects\wuji-legion-codex\experts\INDEX.md)
- [distillation.md](E:\wuji-projects\wuji-legion-codex\units\distillation.md)
- [quick-imagegen.ps1](E:\wuji-projects\wuji-legion-codex\scripts\quick-imagegen.ps1)

---

## 更新日志 / Changelog

- `2026-06-01 v10.6`
  - 新增 pilot page 快速闭环：先做 1 页代表页，过线后才批量生成。
  - pilot 最多两轮，不过线必须换路线或短报阻塞。
  - 执行底座新增 `pptx-batch-gate`，缺 pilot 产物、`pilot-approval` 或 pilot-score 不允许批量生成；preview 发白/低对比也直接拦截。
- `2026-06-01 v10.5`
  - 非代码任务新增 10/15/30 分钟熔断，禁止“嘴上直接生成、实际继续绕路”。
  - PPT 参考任务必须在生成前锁定 `reference-frame-map`、`reusable-asset-map`、`illustration-plan`。
  - PPT 参考任务必须继承风格、框架、可复用元素和同等级 image2 教学插图表达。
  - 真 PPTX 禁止每页一张整图冒充可编辑成品。
  - 执行底座新增 `pptx-preflight`、`pptx-audit`、`asset-map`、`time-guard` 硬门禁方向。
- `2026-06-01 v10.4`
  - 新增 `执行底座 / 执行底座主帅`，作为无极军团通用确定性执行底座。
  - 明确普通 Go/Tauri/业务代码仍归开发主帅；`wuji-cli`、guard、sync、audit、workflow、beep、bench、preview调度归 执行底座主帅。
- `2026-06-01 v10.3`
  - 非代码成品任务先直接生成主成品，禁止开工前工具链连环验证。
  - PPT 顺序固定为读取输入、映射页型、生成完整成品、预览 QA、局部修复。
- `2026-06-01 v10.2`
  - 吸收 `DannyMac180/skills@5695fa1` 的动态工作流机制。
  - 复杂 `LEGION_TASK` 增加最小可审计轨迹；简单任务不启用，避免 token 噪音。
  - 新增 `scripts/wuji_workflow.py`，用于生成、切片、收集和验证无极工作流工件。
- `2026-05-31 v10.1`
- `2026-05-31 v10.0`
  - 修正 `LEGION_TASK` 触发口径：复杂多能力任务即使用户没喊“激活无极军团”，也必须进入参谋本部接管格式。
- `2026-05-31 v9.9`
  - `beep.ps1` 新增后台延迟模式 `-SpawnDelayed -DelayMs`。
  - 收尾提示音改为最终回复前调度、最终回复后响起，解决“不是结束时响”的体感问题。
- `2026-05-31 v9.8`
  - 修正提示音触发时机：提示音必须尽量作为最终回复前的最后一个工具动作，避免验证阶段提前响完。
- `2026-05-31 v9.7`
  - 增强 `scripts/beep.ps1`，用临时 WAV 播放提示音，支持 `complete`、`error`、`notify` 三种提示音。
  - 规则新增任务完成提示音：非 FAST_REPLY 的任务收尾前先响铃，再给最终结果。
- `2026-05-31 v9.6`
  - 补齐根 `SKILL.md` frontmatter。
  - 清理无用缓存，修复安装/恢复/同步/推送脚本。
  - Trae 工兵改为无极远征通用 handoff 执行层。
  - 修复打靶场 `skill` 扫描，验证结果 116/116 通过。
- `2026-05-31 v9.5`
  - 专家库从 `44` 张主责卡压缩为 `15` 张高密度卡。
  - 改为“师团万能主帅入口 + 内置多模式 + 独立质检”结构。
  - 内容、视觉、开发等执行师团合并入口；白帽、质检、安全、合规继续独立。
  - 新增全网源码核验来源：`github/awesome-copilot`、`marketingskills`、`humanizer`、`powerpoint-skill`、`ppt-master`。
- `2026-05-31 v9.4`
  - 专家库从 `70` 张卡蒸馏压缩到 `44` 张主责专家卡。
  - 合并重复人物和重复能力，新增 `experts/INDEX.md` 作为唯一专家索引。
  - 规则明确“专家不以量取胜、蒸馏不是叠加”。
- 详细记录见：[CHANGELOG.md](E:\wuji-projects\wuji-legion-codex\CHANGELOG.md)
