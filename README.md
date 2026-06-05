# 无极军团 Codex 版 / Wuji Legion for Codex

**当前正式版：v10.8**  
当前仓库的设计目标是：`以总成本最省为准，高命中优先`、`高质高效`。这里主要记录系统结构、工具边界和版本变化，不作为运行时行为注入文本。

**让 Codex 从“会回答”升级成“真交付”。**  
无极军团是一套面向真实交付的执行系统：把调研、代码、内容、PPT、HTML、配图、插件和规则升级收口到可验证结果。

**一句话卖点**  
以更少返工换更低总成本；复杂任务走真实生产链，简单任务走低噪音直达链。

## 使用定位

- 你不用自己学复杂工具链，阿极就是统一入口。
- 你可以按需使用门禁、底座、PPT 生产链和独立审查位。
- 这里描述的是仓库能力，不是运行时必须逐条展开给用户的话术。

## 能力范围

- 调研：查得更准，结论更短，更少信息噪音。
- 代码：Go 执行底座负责硬门禁，功能链路可构建、可测试、可审计。
- 内容：写作、脚本、方案、教程不再像 AI 拼接稿。
- PPT / HTML / UI：先走真正的生产引擎，再做预览、校验和修复，减少返工。
- PPT：固定首页/目录页/单元页/总结页/结尾页角色，优先复用模板元素；动态感先在 HTML 演示稿成立，再落到可编辑 PPTX。
- 安全：白帽、质检、安全、合规独立，不让执行者自己给自己放行。
- 进化：只蒸馏有用能力，不做无穷叠加。

## 最适合谁

- 想把 Codex 当长期生产力系统的人
- 不会编程、但要高质量成品的人
- 讨厌半成品、试验版、待开发的人
- 在乎成本、稳定、命中率和最终质量的人

## 当前收口能力

- Go 执行底座：`wuji-cli` 当前保留的核心硬门包括 `reference-guard`、`claim-guard`、`time-guard`、`audit`、`code-map`、`bugfix-guard`、`qa-guard`、`migration-guard`、`closeout-check`、`mcp-guard` 以及 PPT 专属门禁；其他能力更适合作为离线治理或专项分析工具，而不是默认主链必经项。
- 模型分档：低档 `gpt-5.4-mini`，中档 `gpt-5.4`，高档 `gpt-5.5`；按任务难度、验证成本和总返工成本选档，不把 `mini` 写死成一切任务默认档。
- 静态骨架：固定模型三档、核心路由骨架、内置插件归口和 MCP 默认裁决已沉入 Go 底座，`config.json` 只保留环境信息和覆盖项。
- 图像直出：普通 `生图/插图/海报/封面/配图` 已从 `ComfyUI` 重链剥离，默认直走 `imagegen`。
- 执行节奏：普通任务默认不做开工前预检，不扫全仓、不试环境、不探接口；直接做主线，只有文件安全、外部同步、批量破坏性动作保留零思考硬门。
- Go 底盘主要承担可判定硬门，不承担默认管理编排。
- PPT 主链：模板续写走 `Presentations template-following exact clone/edit`；从零高颜值走 `HTML-first -> editable PPTX`；Windows PowerPoint COM/MCP 只做最后一公里精修；Go 负责锁三张表、`style-lock`、`page-role-policy`、pilot 放行和收口 QA，默认主线不再先跑 `pptx-preflight`。
- 当前 `HTML-first` 的真实边界已明确：它优先走 `Playwright + dom-to-pptx` 的浏览器计算样式导出，能高保真保留静态 HTML/CSS 视觉，但仍不把 HTML/CSS 的动画、过渡和动态组件自动转成 PowerPoint 动画。
- 动态交付主张：演示型 PPT 默认双轨成品，`live HTML demo` 承接动态体验，`editable PPTX` 承接可编辑交付；静态 PPT 不再冒充动态成品。
- 动态硬门现已落地：`motion-plan` 成为批量前必检工件；如果任务要求动效却没有 `live-demo-source.html` 或等效动态源，`pptx-batch-gate` 直接 `NO-GO`。
- 布局硬门现已落地：`pilot-preview-layout.json` 会确定性拦截真实越界和底部危险区贴边；`HTML-first` 导出链也会同步生成 `htmlfirst-preview-layout.json` 作为布局证据。
- 自进化闭环：保留为离线治理能力，不建议作为每次执行的默认主链。
- 交付铁律：只交最终结果，不交半成品，不拿路线表演冒充执行。
- Codex Use Cases 吸收重点：持续目标直跑到底、复杂代码先出最小代码地图、重复工作沉成 skill/CLI、外部与批量操作保留证据、前端优先真实浏览器验收。
- Headroom 吸收重点：稳定前缀缓存对齐、按内容类型压缩路由、重要性优先上下文装配、失败模式离线学习。

---

## 当前关键变化 / Key Changes

- 执行底座当前只强调少而硬的门禁与审计能力，不再把 `workflow-guard / task / route-task / context-pack / feedback-* / prompt-distill` 这类治理链路当作日常执行默认必经层。
- `truth-state`、`finish-or-block`、`closeout-check` 已把“胡说/假完成/收口后再问继续”从文案要求推进到 Go 硬门。
- 白帽显式出场与“同任务默认持续执行到收口”已继续写入全局规则和审查规则，不再只靠前台口头承诺。
- `bugfix-guard` 已加入修 bug 专属门禁：没有复现、自测、回归/独立复验证据，或浏览器/程序/关键流程仍失败时，一律不得宣称修好。
- `qa-guard` 已加入质检专属门禁：没有浏览器/程序/命令/MCP 独立复验证据，或复验已明确仍失败时，一律不得宣称质检已通过。
- `migration-guard` 已加入旧项目迁移专属门禁：没有功能对照表、可运行证据、关键流程证据，或仍存在假页/空壳页时，一律不得宣称完成。
- 质检能力边界已增强：质检不再只看报告，默认可调用浏览器、启动命令、本地程序及必要的已过门禁 MCP/插件做独立复验，不过就直接打回继续修。
- `prompt-candidate-audit` 现在会拦截：
  - 生图前探路话术
  - 收口后再问继续
  - 参谋进入套话 + 分阶段停机 + 等用户继续
- `audit` 现在会拦截任务日志里的“等待用户继续”假阻塞。
- PPT 主链当前收口为：模板续写走 `Presentations template-following exact clone/edit`；从零高颜值走 `HTML-first -> editable PPTX`；Go 负责三张表、pilot 放行和 QA，不再把 `pptx-preflight` 扩散成默认第一步。
- 模型分档当前只保留三档事实：`gpt-5.4-mini / gpt-5.4 / gpt-5.5`；按任务复杂度、验证负担和返工风险选档，不写死某一档为一切任务默认答案。
- 普通图像任务默认直走 `imagegen`，不再允许先探环境、先读系统 skill、先看 key、先试通道。
- 提示音保留，但只作为工具层习惯，不承载原则层含义。
- Headroom 已被定向蒸馏为四个机制，不引入整套 proxy/wrap 外壳，也不开放自动改主规则。

完整历史变更请看：[CHANGELOG.md](E:\wuji-projects\wuji-legion-codex\CHANGELOG.md)

---

## 现在实现了什么 / What It Enables

| 方向 | 主帅 | 内置能力 |
|---|---|---|
| 内容 | 内容主帅 | 小说、剧本、分镜、教程、计划书、营销方案、卖点提炼、短内容、人味改稿 |
| 视觉 | 视觉主帅 | 真PPTX、HTML演示、UI页面、数据可视化、信息图、配图 |
| 开发 | 开发主帅 | Go/Tauri、前端、小程序、ComfyUI插件、AI工程、自动化、原型 |
| 执行底座 | 执行底座主帅 | 无极执行底座、wuji-cli、少而硬的 guard/audit 门禁、PPT 批量门禁、提示音与专项审计 |
| 情报 | 情报主帅 | 全网搜索、GitHub源码核验、趋势、用户研究、本地化 |
| 安全 | 安全主帅 | 威胁建模、漏洞验证、供应链、发布安全 |
| 审查 | 白帽/质检/合规/性能 | 前置封驳、最终验收、许可证/隐私、速度/token基准 |
| 进化 | 进化主帅 | 查源、裁决、实验、复盘、专家瘦身 |

---

## 借鉴了什么 / What It Distills

为避免误会，这里明确说明：无极军团 Codex 版参考、蒸馏、整流了多条开源 skill / 流程，但没有把上游项目原样搬运进来。

它借走的是机制，不是名字：

- `openai/skills`、`anthropics/skills`：skill 结构、按需加载、资源分层、上下文节省。
- `addyosmani/agent-skills`：阶段识别、上下文工程、薄切片实现、五轴 review、安全门禁。
- `github/awesome-copilot`：大规模技能索引、bundled assets、Rust/QA/安全细分规则。
- `marketingskills`：产品、受众、定位先行，营销技能互相关联但有主线。
- `humanizer`：AI写作痕迹识别、作者声音匹配。
- `powerpoint-skill`：视觉优先、密度边界、重叠检查、预览验证。
- `ppt-master`、`elite-powerpoint-designer`、`slide-studio`、`Presentations`：真 PPTX 分阶段、可编辑交付、审美上限和验证闭环。
- `SkillClaw`、`Edict`：任务后进化、候选验证、前置封驳和状态审计。
- `DannyMac180/skills`：复杂任务最小审计工件、切片回收、结果收集和验证脚手架。

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

详细历史记录见：[CHANGELOG.md](E:\wuji-projects\wuji-legion-codex\CHANGELOG.md)
