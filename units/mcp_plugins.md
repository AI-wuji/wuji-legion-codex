# MCP / 插件协同

## 定位

这里只管工具层边界：什么时候该补工具、怎么过门禁、什么默认不进主链。
已启用插件和候选台账看 [plugins.md](plugins.md)。

## 进入条件

- 参谋本部先判断任务是否真的需要外部工具。
- 已有主帅能完成时，不启用 MCP 或插件。
- 需要补位时，只补最小工具组合。
- 工具结果必须回到对应主帅收口。
- 高权限、联网、账号、隐私或许可证边界不清的工具，先过白帽、安全、合规。

## MCP 门禁

MCP 进入任务链路前必须通过：

```powershell
.\.wuji-tools\wuji-cli.cmd mcp-guard --manifest <manifest.json> --workspace <workspace>
```

网络型 MCP 必须显式允许：

```powershell
.\.wuji-tools\wuji-cli.cmd mcp-guard --manifest <manifest.json> --workspace <workspace> --allow-network true
```

候选 MCP 的蒸馏台账命令：

```powershell
.\.wuji-tools\wuji-cli.cmd mcp-distill --catalog .\units\mcp_catalog.json --report .\outputs\mcp-distill-report.json
```

## 必查项

- manifest 必须有 `name`、`version`、`transport`
- capabilities 必须声明 `tools`、`resources`、`prompts`
- permissions 必须声明 `network` 和 `filesystem`
- filesystem 权限不得越出任务 workspace
- http/sse 网络传输必须显式允许
- manifest 不得包含明文 `token`、`secret`、`password`、`api_key`

## 默认裁决

| 类型 | 默认结论 | 说明 |
|---|---|---|
| 本地、低权限、边界清楚 | `absorb` | 例如 Filesystem、Git、Time、Sequential Thinking |
| 联网、账号、OAuth、外部写操作 | `defer` | 需要任务级授权，不进默认主链 |
| 高权限、带 secrets、来源或许可证不清 | `reject/defer` | 不默认接入，必要时单独审查 |

## 默认协同

| 任务 | 主帅 | 工具补位 |
|---|---|---|
| 网页检查、前端验收 | 视觉主帅 / 开发主帅 | Browser |
| 文档、Word、资料整理 | 内容主帅 | Documents |
| 表格、数据分析 | 情报主帅 / 内容主帅 | Spreadsheets |
| 模板续写、补页、跟版真 PPTX | 视觉主帅 | Presentations |
| 从零高颜值 PPTX | 视觉主帅 | HTML-first 转 editable PPTX 工具链 |
| PowerPoint 最后一公里精修 | 视觉主帅 | 通过 `mcp-guard` 的 PowerPoint COM/MCP |
| 外部系统工具调用 | 对应主帅 | 通过 `mcp-guard` 的 MCP |

## PPT 专项归口

- `Presentations` 是模板续写/补页的主生产链，不是可有可无的补位。
- `HTML-first -> editable PPTX` 是从零高颜值路线，不得冒充模板续写。
- PowerPoint COM / MCP 只做最后一公里精修、验证和局部修复。
- Go 执行底座在 PPT 体系内只负责 QA、审计和放行，不替代上面三条生产路线。

## 红线

- 不因工具热门就默认接入。
- 不让 MCP/插件替代参谋本部路由。
- 不让 MCP/插件替代白帽、安全、合规判断。
- 不把未安装、未授权、未验证的工具说成可用。
- 不把网络工具默认放进主链路。
