# PPTX Master

## 定位

只做真 PPTX。

适用：

- 续写现有 `.ppt/.pptx`
- 基于模板补页
- 从零生成可编辑 PowerPoint

不适用：

- HTML 演示稿
- 图片版 PPT
- 用网页截图、整页大图或整页位图冒充 PPTX

## 核心目标

- 保持可编辑
- 继承参考 deck 的视觉语法、页型节奏和信息结构
- 优先复用参考素材，不低配重画
- 先用最短闭环确认方向，再批量生成
- 把生产主链和 Go 门禁拆开，避免“门很多，成品很弱”
- 默认追求“动态感先成立，再落到可编辑 PPTX”，避免只剩静态排版

## 四层路线

### Route A: 模板续写 / 跟版主线

- 只要用户给了现有 `.ppt/.pptx`、模板页或参考成品页，默认进入这条线。
- 主生产链固定为官方 `Presentations` 的 `template-following exact clone/edit`。
- 必须先读源 deck，再做 `template-frame-map`，再复制源页并在继承元素上原位编辑。
- 不允许跳过源页复制，直接从空白页重搭一个“像模板”的新 deck。
- 不允许用 HTML、整页截图、整页位图或 OOXML 生拼硬改来冒充模板续写主链。

### Route B: 从零高颜值主线

- 用户没有给模板，但要求高颜值、高完成度、高商品感时，默认走 `HTML-first -> editable PPTX`。
- 先把版式、层级、配图策略在 HTML/CSS 舞台上做对，再转成可编辑 PPTX。
- 允许补位：`html2pptx`、`dom-to-pptx`、`PptxGenJS`。
- 当前本地 `HTML-first` 已落地主引擎是 `Playwright + dom-to-pptx` 的浏览器计算样式导出；它能高保真保留静态 HTML/CSS 视觉，但不会自动把 HTML/CSS 动画、过渡和动态组件转成 PowerPoint 动画。
- 这条线适合“先把视觉做漂亮”，不允许拿来冒充模板续写。

### Route C: Windows 最后一公里精修

- Route A 或 Route B 已经产出基础 deck，但还需要细节对齐时，才进入 PowerPoint COM / MCP 精修。
- 这层只负责：备注写入、占位符清理、局部修版、导出校验和最后修复。
- 这层不是主生产器，不能反过来替代 Route A / Route B。

### Route D: Go 门禁与审计

- `asset-map`、`pptx-preflight`、`pptx-batch-gate`、`pptx-audit` 只负责门禁、审计和收口。
- Go 门禁不负责“把 PPT 做漂亮”。
- 默认主线先产出，再由 Go 负责锁路线、批量放行和收口质检。
- `pptx-preflight` 保留为定向探针，只在新生成路线、可疑 generator 或白帽明确要求时触发；日常成熟主链默认不跑。

## 默认生产顺序

```text
source
-> outline
-> speaker-notes
-> slide-spec
-> design-brief
-> reference-frame-map
-> reusable-asset-map
-> illustration-plan
-> pilot-page
-> pptx-batch-gate
-> layout-plan
-> preview
-> full PPTX
-> notes
-> 质检
```

- 教程、课程、说明、方案类 PPT，默认先出 `outline`，再出逐页 `speaker-notes`，再进入页面生产。
- `speaker-notes` 默认写入 PowerPoint 备注区，不得把逐字稿直接塞进观众可见正文。
- 如果用户明确说“先给我审大纲”，就停在 `outline`；否则默认内部继续直跑，不为大纲单独卡住主线。
- 教学型页如果仅靠文字无法让人学会操作，`illustration-plan` 必须明确是补 `软件截图 / 步骤示意图 / image2 教学插图 / 复用参考图框`，不能只做空框架。

## 动态优先策略

- 凡是课程、教程、发布会、演示型 deck，默认同时规划两层成品：`editable PPTX` 和 `live HTML demo`。
- `live HTML demo` 是动态体验源：负责动效、节奏、科技感、信息显隐和氛围验证。
- `editable PPTX` 是交付源：负责可编辑、可讲解、可复用模板资产。
- 静态 PPT 只负责承接版式、结构、备注、配图和模板元素，不得谎称自己已经承接了 HTML 动态效果。
- 只要任务要求“大量动态效果”“科技感演示”“看板动效”“赛博氛围”，`design-brief` 里就必须显式写出 `motion-plan`，至少包含：页面动效角色、节奏、可接受的静态降级方式。
- `motion-plan` 不是软建议，而是批量前硬工件；如果标记 `required=true`，就必须同时产出 `live-demo-source.html` 或等效动态源，否则不得进入批量。
- 当前默认事实边界必须写死：HTML 动画、过渡、呼吸光效、旋转扫描、浮动卡片等，可先在 HTML 演示稿实现；PowerPoint 侧默认只承接静态可编辑表达，除非后续明确接入 PPT 原生动画编排链。

## 三张表

批量前必须锁死：

- `reference-frame-map`
- `reusable-asset-map`
- `illustration-plan`

同时必须补齐 2 个路线护栏工件：

- `style-lock`
- `page-role-policy`

要求：

- 三张表必须短、可执行、能约束生成器。
- 缺任一项，不得进入批量生成。
- `illustration-plan` 必须逐页判断：无需插图、复用参考图、复用图框换图、按参考风格新生图。
- 教学页如果被标成 `requires_visual=true` 或出现步骤/按钮/界面/导入导出等教学信号，批量前还必须具备 `outline` 和 `speaker-notes`。
- `style-lock` 必须写清：整体风格名、背景深浅极性、高光语言、插图语言、禁止项。
- `page-role-policy` 必须写清：首页、目录页、单元页、总结页、结尾页哪些是固定页型，以及这些页型不得挪作普通内容页。
- `page-role-policy` 默认必须优先锁定模板里已经定好的首页、目录页、单元页、总结页、结尾页；没有明确授权，不得自创替代页型去覆盖这些角色。
- 如果模板或用户已经明确点名风格，例如“霓虹赛博卡通风”，该风格名必须原样进入配图 / image2 / 截图提示；不得临时自由发挥改成白底、写实或电商风。

## Pilot 规则

- pilot page 不是工具链测试，只验风格、结构、素材复用、插图策略和可编辑路线。
- pilot 必须选最有代表性、最高风险、最高密度的一页。
- pilot 必须导出 preview，并记录 `pilot-score`。
- pilot 现在还必须尽量附带 `pilot-preview-layout.json`；一旦出现真实越界或底部危险区贴边，`pptx-batch-gate` 直接拦截，不允许带病批量。
- pilot 后默认先走内部 gate 判断是否放行；只有命中真实审批点、用户明确要求先看中间结果，或路线风险高到无法内部判定取舍时，才停下给用户看 preview；其余情况默认自动放行，不得为了展示过程额外停顿。
- 成熟同路线默认允许自动批准批量，但仍必须留下 `pilot-approval` 工件。
- pilot 最多两轮；两轮仍不过线，必须换路线或短报阻塞。
- `pptx-batch-gate` 是默认批量放行门；缺三张表、`pilot-page`、`pilot-preview`、`pilot-score` 或 `pilot-approval`，一律 `NO-GO`。

## 参考 Deck 继承规则

- 用户给现有 `.ppt/.pptx`、模板页、前几页成品时，统一视为模板续写。
- 继承的是：标题层级、留白逻辑、图文比例、配色气质、页型节奏、信息结构。
- 参考 deck 同时是素材库：表格、图标、卡片、图框、箭头、章节页、背景装饰、插图资产优先复用。
- 参考 deck 里的固定模板元素默认优先“借位复用”，不是看到空白就自创方框、自创卡片、自创底板。
- 如果参考 deck 是重插图、重案例、重信息图风格，新 deck 不能退化成纯文字框架。
- 用户说“参考风格/框架/版式”时，默认不复制原内容，只继承视觉与结构。
- 如果参考 deck 已经固定了首页、目录页、单元页、总结页、结尾页，对应页型默认锁死，不得拿这些页去承载普通操作步骤页。
- 用户明确说“目录页用哪张、单元页用哪张、总结页用哪张、结尾页用哪张”后，同任务后续默认继承，不再重新判断。

## 速度与执行原则

- 非代码成品默认先跑主链，遇错再修，不做连环前测。
- 同一路线在当前任务里成功过一次，不得重复验证同类能力。
- html-first 的 inspect 默认只取 pilot 所需最小集，不做整套预览和整套 layout 体检。
- html-first 虽然不做整套 layout 体检，但现在会同步产出 `htmlfirst-preview-layout.json` 作为布局证据；重点只拦真实越界和底部危险区，不扩大成过度体检。
- template-following 默认复用已有 inspect，并跳过不必要的 final preview 渲染。
- 备注逐字稿在机器支持 PowerPoint COM 时默认自动写入，不让用户再额外提醒。
- 默认跳过：开工前工具可用性连环测试、环境探测表演、重复 template inspect、重复 preflight、无必要的全量 preview 重渲染。
- 同一路线如果已经成功过，不再为“确认能不能做”去空跑试通道。
- 用户明确要 `image2`、截图型教学配图或同等级参考插图时，不允许静默降级成 generic 方框、临时本地图示或纯文字凑数；真拿不到就应阻塞，不得伪装完成。

## 硬失败

以下情况一律视为失败：

- 每页一张整图、整页截图、网页截图冒充可编辑 PPT
- 占位符残留、`Click to add`、模板碎片页、空白页
- 发白洗底、文字低对比不可读、风格明显偏离参考深浅和氛围
- 能复用参考元素却全部低配自创
- 参考 deck 明明重插图，结果新页只剩文字框
- 把逐字稿硬塞进页面，靠缩字号救场
- 用 Route B 假装模板续写
- 用 Route C 充当主生产器
- 用 Go 门禁替代真正的视觉生产链
- 首页/目录页/单元页/总结页/结尾页被挪用成普通内容页
- 用户明确锁定的深色/霓虹/赛博风格被洗成白底、浅灰底或写实宣传风
- 用户要求 `image2` / 截图风格页，结果静默降级成 generic 方框、低配示意图或自创丑元素
- 用户明确要动态感或 HTML 演示体验，结果只交静态 PPT 却宣称“已实现动态”

## 质检核心问题

- 是否仍可在 PowerPoint 中编辑
- 是否继承了参考 deck 的页型节奏和素材系统
- 是否有遮挡、溢出、字号过小、残留占位符
- 是否出现 Word 投影感或纯文字堆砌
- 是否先通过 pilot 再批量
- 教学页是否真的具备截图、示意图或同等级插图表达

## 外部能力裁决

- `Presentations`：`absorb`，作为模板续写真主线
- `html2pptx`：`targeted-absorb`，适合从零高颜值路线
- `dom-to-pptx`：`targeted-absorb`，适合本地 HTML 转可编辑 PPTX
- `PptxGenJS`：`targeted-absorb`，适合 JS 生态批量可编辑生成，尤其是原生超链接、目录跳转和 action button 这类真交互 PPTX
- `mcp-server-ppt` / `ppt-mcp`：`defer`，适合 Windows 最后一公里精修
- `python-pptx`：保留为补位，不作为模板续写主线

## 与执行底座配合

- 三张表由 `asset-map` 或人工规划锁定
- 默认批量放行走 `pptx-batch-gate`
- 成品收口必须过 `pptx-audit`
- `pptx-preflight` 只做定向路线探测，不再作为默认主链必经
- Go 只做门禁和审计，不替代 Route A / Route B / Route C
