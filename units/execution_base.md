# 执行底座

## 定位

`执行底座主帅` 只负责无极军团自身的确定性执行底座，不负责内容创作、视觉设计、业务软件主线或审查裁决。

一句话：

```text
把稳定、重复、可判定的动作下沉为 Go 命令和硬门禁。
```

## 边界

执行底座负责：

- `wuji-cli`
- `guard / task / sync / audit / workflow / beep / bench / preview`
- `asset-map / pptx-audit / pptx-preflight / pptx-batch-gate / time-guard`
- `mcp-guard / mcp-distill`
- `canon-report / route-task / context-pack`

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
| `workflow-guard` | 工作流工件完整性 |
| `claim-guard` | “已完成/已通过”类声明的证据检查 |
| `time-guard` | 非代码任务 10/15/30 分钟熔断 |
| `task` | 任务生命周期记录 |
| `sync` | 关键规则树一致性检查 |
| `audit` | 乱码、占位符、残留旧链路、空壳内容扫描 |
| `bench` / `bench-report` | 成本、速度、重试、QA 基准 |
| `code-map` | 复杂代码任务的最小入口/依赖/风险/验证图 |
| `closeout-check` | 收口前检查是否仍有明显缺口未顺手完成 |
| `preview` | 预览导出调度与产物存在校验 |
| `asset-map` | 从参考 PPTX 生成三张表 |
| `pptx-audit` | 真 PPTX 可编辑性与伪 PPT 风险检查 |
| `pptx-preflight` | 新生成路线或可疑 generator 的定向预检，不是默认批量主链 |
| `pptx-batch-gate` | pilot、教学内容工件和预览质量的默认批量放行门 |
| `mcp-guard` | MCP/插件权限与 manifest 检查 |
| `mcp-distill` | MCP 候选蒸馏裁决辅助 |
| `canon-report` | 导出 Go 底座内置静态规范源 |
| `route-task` | Go 内置路由骨架 + config overlay 的任务路由 |
| `context-pack` | 稳定前缀与缓存友好上下文打包 |

## 从 Codex Use Cases 吸收的底座原则

- 默认优先把重复工作沉成 `wuji-cli` 命令、固定工件或 skill，而不是继续堆长提示词。
- 自动化链路默认要求“内部完成判据 + 逐项结果 + 验证证据”；这里的完成判据只用于系统自判是否真正收口，不得拿它当中途停下等用户继续的借口。
- 涉及多文件代码改动时，底座优先支持“代码地图 -> 变更 -> 验证”节奏，而不是直接跳到编辑。
- 收口前默认应跑一次 `closeout-check`，确认不存在同目标下显而易见、低风险、高收益且应顺手收完的缺口。
- 前端、网页、可视化交付的收口优先依赖真实浏览器预览与检查，不把静态描述当验收。
- 安全与审计类任务的底座输出应支持证据分级和候选验证，不只给一个模糊风险结论。

## 图像路由硬规则

- 普通 `生图/插图/海报/封面/配图` 视为 `imagegen` 直出链，不进 `ComfyUI`
- `ComfyUI` 只接明确工作流、节点、插件、批量生成、视频管线和技术美术
- `route-task` 命中 `imagegen` 时，默认强制推荐低档、低推理，避免为普通出图浪费 token
- `prompt-candidate-audit` 会拦截“先查环境/先读 skill/先看 key/先试通道”的生图前探路话术

## PPTX 专属硬门

- `asset-map` 必须产出：`reference-frame-map`、`reusable-asset-map`、`illustration-plan`
- 教学页如果在 `illustration-plan` 里出现 `requires_visual=true` 或教学信号，批量前还必须具备 `outline` 和 `speaker-notes`
- `pptx-preflight` 只负责新路线或可疑 generator 的定向拦截，不再作为默认主链必经
- `pptx-batch-gate` 缺 `pilot-page`、`pilot-preview`、`pilot-score`、`pilot-approval`、三张表或教学内容工件时，一律 `NO-GO`
- `pptx-batch-gate` 负责拦截发白洗底、低对比、近空白 preview
- `pptx-audit` 负责拦截占位符残留、空白页、模板碎片页、整页图片区伪装可编辑 PPT
- 模板续写时，Go 底座只负责检查三张表、pilot 和成品审计，不负责替代 `Presentations` 主生产链
- 从零高颜值路线时，Go 底座只负责检查是否走错成整页图/假可编辑，不负责决定版式审美

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
