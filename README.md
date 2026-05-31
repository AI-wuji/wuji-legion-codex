# 无极军团 Codex 版 / Wuji Legion for Codex

**一句话 / One Sentence**  
无极军团 Codex 版是一个 MoE 总调度执行框架：阿极统一入口，参谋本部选主帅，各师团主帅按模式执行，白帽/质检/安全/合规独立审查，把 Codex 从“会回答”推进到“能稳定交付”。
Wuji Legion for Codex is an MoE-style execution framework: Aji is the single user-facing entrance, Staff HQ selects one commander, each division commander runs task-specific modes, and independent QA/security/compliance gates keep delivery reliable.

---

## 它是干什么的 / What It Does

无极军团不是“再装很多 skill”。

它解决的是：让 Codex 在调研、代码、内容、PPT、HTML、配图、插件、规则升级这些任务里，不再乱调工具、不再堆流程口号，而是按一个稳定执行框架交付结果。

核心能力：

- 一个入口：默认永远是阿极。
- 一个主帅：每个任务先选单主帅负责到底。
- 一个师团主帅：每个师团内部尽量压成万能主帅入口。
- 多个模式：小说、剧本、PPT、HTML、小程序、ComfyUI 插件、wuji-cli/Rust执行底座等能力进入主帅内置模式。
- 独立审查：白帽、质检、安全、合规不并入执行者。
- 硬门禁：能用 Rust师硬控的路径、PPTX可编辑性、参考素材复用、时间熔断，不靠嘴上反复提醒。
- 持续进化：外部 skill 先查官方源和源码，再蒸馏机制，不叠加入口。

---

## v10.6 的关键变化 / Key Change

v10.6 新增 `pilot-page` 快速闭环：PPT/HTML 视觉成品不再一口气批量试错，预检和三张表通过后先生成 1 页最高风险/最高密度/最能代表风格的 pilot page，记录 `pilot-score`，过线后才批量生成。Rust师新增 `pptx-batch-gate`，缺 `pilot-page`、`pilot-preview`、`pilot-score` 直接 `NO-GO`。

## v10.5 的关键变化 / Key Change

v10.5 针对 PPT 长时间空转和低质交付加前置硬门禁：口头说“直接生成”不算执行；非代码任务 10/15/30 分钟分级熔断；参考 PPTX 必须同时当作风格系统和素材库，批量生成前先锁定 `reference-frame-map`、`reusable-asset-map`、`illustration-plan`；优先复用表格、图标、卡片、流程箭头、章节页、背景装饰、教学插图、案例图、公式图和 image2 生图资产；真 PPTX 不得每页一张整图冒充可编辑成品。Rust师新增 `pptx-preflight`、`pptx-audit`、`asset-map`、`time-guard` 方向，用确定性检查硬控这些问题。

## v10.4 的关键变化 / Key Change

v10.4 新增 `Rust师 / Rust主帅`：把无极军团自身稳定、重复、可判定的动作下沉为确定性执行底座，负责 `wuji-cli`、guard、task、sync、audit、workflow、beep、bench、preview调度。它不替代开发主帅，不抢普通 Rust/Tauri/业务代码，也不替代参谋本部、女娲、白帽、质检、安全、合规和进化判断。

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
- Rust主帅
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
| 开发 | 开发主帅 | Rust/Tauri、前端、小程序、ComfyUI插件、AI工程、自动化、原型 |
| Rust师 | Rust主帅 | 无极执行底座、wuji-cli、guard、task、sync、audit、workflow、beep、bench、preview调度、pptx-preflight、pptx-batch-gate、pptx-audit、asset-map、time-guard |
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
  - Rust师新增 `pptx-batch-gate`，缺 pilot 产物或 pilot-score 不允许批量生成。
- `2026-06-01 v10.5`
  - 非代码任务新增 10/15/30 分钟熔断，禁止“嘴上直接生成、实际继续绕路”。
  - PPT 参考任务必须在生成前锁定 `reference-frame-map`、`reusable-asset-map`、`illustration-plan`。
  - PPT 参考任务必须继承风格、框架、可复用元素和同等级 image2 教学插图表达。
  - 真 PPTX 禁止每页一张整图冒充可编辑成品。
  - Rust师新增 `pptx-preflight`、`pptx-audit`、`asset-map`、`time-guard` 硬门禁方向。
- `2026-06-01 v10.4`
  - 新增 `Rust师 / Rust主帅`，作为无极军团通用确定性执行底座。
  - 明确普通 Rust/Tauri/业务代码仍归开发主帅；`wuji-cli`、guard、sync、audit、workflow、beep、bench、preview调度归 Rust主帅。
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
