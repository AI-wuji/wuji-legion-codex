# 执行底座

## 定位

`执行底座主帅` 只负责无极军团自身的确定性执行底座，不负责内容创作、视觉设计、业务软件主线或审查裁决。

一句话：

```text
把稳定、重复、可判定的动作下沉为 Go 命令和硬门禁。
```

默认执行姿势：

```text
先执行主线
-> 真报错再诊断
-> 只有零思考硬门允许前置
```

## 边界

执行底座负责：

- `wuji-cli`
- 默认硬门与审计：`guard / audit`
- `asset-map / pptx-audit / pptx-preflight / pptx-batch-gate / time-guard`
- `mcp-guard / mcp-distill`
- `canon-report`
- 低频支持工具：`bench / preview / beep / sync`

执行底座不负责：

- 参谋判断、主帅路由、蒸馏裁决
- 内容写作、视觉设计、PPT 审美决策
- 普通业务项目的 Go/Tauri/前端/插件开发
- 白帽、质检、安全、合规的独立放行

## 语言策略

- Go 是执行底座唯一主链语言。
- Python、Node、PowerShell、C# 只作为专项补位工具存在。
- 专项工具不能绕过 Go 主链的调度、记录和验收。

## 核心命令

| 命令 | 作用 |
|---|---|
| `reference-guard` | 参考文件只读与输出路径安全 |
| `claim-guard` | “已完成/已通过”类声明的证据检查 |
| `time-guard` | 非代码任务 10/15/30 分钟熔断 |
| `audit` | 乱码、占位符、残留旧链路、空壳内容扫描 |
| `bench` / `bench-report` | 成本、速度、重试、QA 基准（低频支持） |
| `code-map` | 复杂代码任务的最小入口/依赖/风险/验证图 |
| `bugfix-guard` | 修 bug 的复现、自测、回归、浏览器/程序复验门禁 |
| `qa-guard` | 质检是否已做浏览器/程序/命令/MCP独立复验的门禁 |
| `migration-guard` | 旧项目迁移/改编的功能对齐、可运行性、关键流程证据门禁 |
| `closeout-check` | 收口前检查是否仍有明显缺口未顺手完成 |
| `finish-or-block` | 明确目标后，只允许完成或真实阻塞 |
| `repeat-candidates` | 从反馈或任务日志中挖出应沉成 CLI/skill 的重复动作 |
| `evidence-grade` | 把结果统一标成 `candidate/checked/verified/shipped` |
| `preview` | 预览导出调度与产物存在校验（低频支持） |
| `asset-map` | 从参考 PPTX 生成三张表 |
| `pptx-audit` | 真 PPTX 可编辑性与伪 PPT 风险检查 |
| `pptx-preflight` | 新生成路线或可疑 generator 的定向预检，不是默认批量主链 |
| `pptx-batch-gate` | pilot、教学内容工件和预览质量的默认批量放行门 |
| `mcp-guard` | MCP/插件权限与 manifest 检查 |
| `mcp-distill` | MCP 候选蒸馏裁决辅助 |
| `canon-report` | 导出 Go 底座内置静态规范源 |
| `sync` | 关键规则树一致性检查（低频支持） |

## 从 Codex Use Cases 吸收的底座原则

- 默认优先把重复工作沉成 `wuji-cli` 命令、固定工件或 skill，而不是继续堆长提示词。
- 自动化链路默认要求“内部完成判据 + 逐项结果 + 验证证据”；这里的完成判据只用于系统自判是否真正收口，不得拿它当中途停下等用户继续的借口。
- `CacheAligner` 思路进入底座：动态值后移、稳定前缀前置、重复头部固定化，优先提高缓存命中而不是事后省几句字。
- 涉及多文件代码改动时，底座优先支持“代码地图 -> 变更 -> 验证”节奏，而不是直接跳到编辑。
- 修 bug 时，底座优先要求“复现证据 -> 修复后自测 -> 回归/浏览器/程序复验 -> 再收口”，不允许直接跳到“已修好”。
- 旧项目迁移/跨栈改编时，底座优先要求“功能对照表 -> 可运行证据 -> 关键流程证据”，不接受只交视觉壳子。
- `ContentRouter` 思路进入底座：代码、日志、JSON、工具输出、长文本应允许走不同压缩/摘要策略，不再假设一种通用压缩对所有内容都有效。
- 收口前默认应跑一次 `closeout-check`，确认不存在同目标下显而易见、低风险、高收益且应顺手收完的缺口。
- `finish-or-block` 用于把“一干到底”从文案要求变成底座硬门：有剩余步骤且无真实阻塞理由时，直接 `NO-GO`。
- 同类反馈、同类任务、同类命令链重复出现时，应优先用 `repeat-candidates` 挖出可沉淀候选，而不是继续靠人脑记忆。
- 证据默认统一走 `candidate / checked / verified / shipped` 四级；越高等级，越必须附真实工件。
- `truth-state` 用于约束结果陈述只能落在 `fact / inference / todo` 三类状态；未验证推测不得冒充已完成事实。
- `bugfix-guard` 用于封死“修完就报、用户一测还是原样”的假修复闭环。
- `qa-guard` 用于封死“质检只看报告、没独立复验却放行”的弱质检闭环。
- `migration-guard` 用于封死“旧项目迁移任务里首页像了、内部是假页、项目还跑不起来却宣称完成”这类假交付。
- `bench-report` 现在会给出 `decision + evidence_level`，可以直接参与后续蒸馏放行判断。
- `repeat-candidates` 现在会顺手写出固定的 `distill-queue.json`，让重复动作候选直接进入蒸馏台账，而不是停留在一次性报告里。
- `IntelligentContext` 思路进入底座：上下文装配优先保高价值片段、证据摘要和可追回引用，不把“压缩”误写成“删掉原文再赌模型没事”。
- 前端、网页、可视化交付的收口优先依赖真实浏览器预览与检查，不把静态描述当验收。
- 质检链默认应具备最小必要的浏览器/MCP/插件复验能力；能独立验证就不把第一轮验收交给用户。
- 安全与审计类任务的底座输出应支持证据分级和候选验证，不只给一个模糊风险结论。
- 执行底座默认只暴露少数高价值硬门；离线治理链能不进当前任务就不进，避免把底座写成新的管理脑。
- `headroom learn` 只吸收“失败模式离线学习”机制，不开放“自动直写主规则”权限；任何 learn 结果都必须先过白帽、质检或进化主帅裁决。

## 图像路由硬规则

- 普通 `生图/插图/海报/封面/配图` 视为 `imagegen` 直出链，不进 `ComfyUI`
- `ComfyUI` 只接明确流程、节点、插件、批量生成、视频管线和技术美术
- `prompt-candidate-audit` 会拦截“先查环境/先读 skill/先看 key/先试通道”的生图前探路话术
- `prompt-candidate-audit` 也会拦截“分阶段停机、参谋进入套话、等用户回复继续后再推进”的管理表演式候选

## PPTX 专属硬门

- 默认不在开工前跑分析型预检；`asset-map`、`pptx-batch-gate`、`pptx-audit` 服务于产出链和批量放行，不服务于“先确认一下能不能做”。
- `pptx-preflight` 只在新路线、可疑 generator、白帽明确封驳、或已经出现真实异常后才触发；不得把它重新扩散成日常任务的默认第一步。
- `audit` 现在会专扫 `task-log.jsonl` 里的执行节奏违规：如果起手就在 `preflight/probe/research` 打转，且没有主产物，直接记为执行层失败模式。
- `asset-map` 必须产出：`reference-frame-map`、`reusable-asset-map`、`illustration-plan`
- `asset-map` 现在默认还要补出 `motion-plan.md/json`，避免动态任务继续靠人工补规则
- 教学页如果在 `illustration-plan` 里出现 `requires_visual=true` 或教学信号，批量前还必须具备 `outline` 和 `speaker-notes`
- `pptx-preflight` 只负责新路线或可疑 generator 的定向拦截，不再作为默认主链必经
- `pptx-batch-gate` 缺 `pilot-page`、`pilot-preview`、`pilot-score`、`pilot-approval`、三张表或教学内容工件时，一律 `NO-GO`
- 任务如果声明需要动效，`pptx-preflight / pptx-batch-gate` 还必须看到可用的 `motion-plan`；其中 `required=true` 时必须同时存在 `live-demo-source.html` 或等效动态源，否则一律 `NO-GO`
- `pptx-batch-gate` 负责拦截发白洗底、低对比、近空白 preview
- `pptx-batch-gate` 现在还会校验 `pilot-preview-layout.json`：`overflow_count > 0` 或 `unsafe_count > 0` 直接拦截，重点封死真实越界和底部危险区贴边
- `pptx-audit` 负责拦截占位符残留、空白页、模板碎片页、整页图片区伪装可编辑 PPT
- 模板续写时，Go 底座只负责检查三张表、pilot 和成品审计，不负责替代 `Presentations` 主生产链
- 从零高颜值路线时，Go 底座只负责检查是否走错成整页图/假可编辑，不负责决定版式审美
- `HTML-first` 现在会同步导出 `htmlfirst-preview-layout.json` 和布局报告，作为批量前后的布局证据；包装脚本会自动落地 `motion-plan`、`live-demo-source.html`、`pilot-preview-layout.json`

## 默认主链降级

- 以下能力不再视为默认主链必经项：`workflow-guard`、`task`、`route-task`、`context-pack`、`feedback-log`、`feedback-dataset`、`prompt-eval`、`prompt-distill`
- 这些能力如需保留，定位为离线治理、演化或专项分析工具，而不是日常执行必经层
- `beep` 保留，但仅作为工具层习惯，不进入顶层原则判断
- 如果某个能力不能直接减少返工、缩短总耗时或提前拦住高频错误，就不应再往默认执行面前摆。

## 软件迁移硬门

- 旧项目改 Rust、桌面改编、跨语言重写、跨框架迁移，必须把“原项目功能是否真正对齐”单独当成门禁项。
- 没有 `功能对照表`、没有构建或启动证据、没有关键流程证据时，不得宣称迁移完成。
- 发现 `假页 / 空壳页 / 占位按钮 / 只有首页像 / 内部页未按原项目功能实现 / 根本无法运行` 时，`migration-guard` 必须直接 `NO-GO`。

## Bug 修复硬门

- 修 bug 默认必须同时具备：复现说明、自测证据、至少一条回归或独立复验证据。
- 如果任务本身可调用浏览器、本地程序、测试命令、启动命令、接口回放或其他电脑资源，执行链和质检链都应先自己跑；不过则继续修，不把第一轮验证交给用户。
- 发现“用户一测仍复现”“浏览器验收仍失败”“程序仍起不来”“关键流程仍不通”时，`bugfix-guard` 必须直接 `NO-GO`。
- 如果复验依赖浏览器、PowerPoint COM、已过门禁的 MCP/插件或本地程序资源，质检链默认应主动调用；只有资源不可用、权限不足或工具未过门禁时，才允许报真实阻塞。

## 质检硬门

- 质检收口前默认至少留下以下任一类独立复验证据：`browser-check / program-check / command-check / mcp-check`
- 没有独立复验证据却声称“已验收通过”，`qa-guard` 必须直接 `NO-GO`
- 如果独立复验结果里已明确仍失败，`qa-guard` 必须直接 `NO-GO`

## 执行顺序

```text
识别稳定动作
-> 设计命令输入输出
-> Go 最小实现
-> 加硬门禁
-> 允许专项工具补位
-> 交白帽/质检/安全按需审查
```

## 现成入口

- [tools/wuji_cli.go](E:\wuji-projects\wuji-legion-codex\tools\wuji_cli.go)
- [scripts/build-wuji-cli.ps1](E:\wuji-projects\wuji-legion-codex\scripts\build-wuji-cli.ps1)
- [scripts/test-wuji-cli.ps1](E:\wuji-projects\wuji-legion-codex\scripts\test-wuji-cli.ps1)
- [scripts/wuji-ppt-pipeline.ps1](E:\wuji-projects\wuji-legion-codex\scripts\wuji-ppt-pipeline.ps1)
- [scripts/test-wuji-cli.ps1](E:\wuji-projects\wuji-legion-codex\scripts\test-wuji-cli.ps1)

## 禁止

- 把执行底座做成另一个“全能大脑”
- 用专项工具绕过 Go 主链记录和验收
- 把审美、路线判断、蒸馏裁决硬塞进执行底座
- 没有可复现命令就宣称“更稳”“更快”“已落地”
