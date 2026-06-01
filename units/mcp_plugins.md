# MCP / 插件协同

## 定位

MCP 和插件是工具层补位，不是新部门，也不是默认入口。

```text
skill = 任务方法和规则
MCP = 标准化外部工具上下文
plugin = Codex 本地能力包
wuji-cli = 执行底座门禁和记录
```

## 路由

- 参谋本部先判断任务是否真的需要外部工具。
- 已有主帅能完成时，不启用 MCP 或插件。
- 需要工具补位时，女娲只负责匹配最小工具组合。
- MCP/插件调用结果必须回到对应主帅收口。
- 白帽、安全、合规可否决高权限、网络、隐私、许可证不清的工具。

## MCP 门禁

MCP 进入任务链路前必须通过：

```powershell
.\.wuji-tools\wuji-cli.cmd mcp-guard --manifest <manifest.json> --workspace <workspace>
```

网络型 MCP 必须显式允许：

```powershell
.\.wuji-tools\wuji-cli.cmd mcp-guard --manifest <manifest.json> --workspace <workspace> --allow-network true
```

## 必查项

- manifest 必须有 `name`、`version`、`transport`。
- capabilities 必须声明 `tools`、`resources`、`prompts`。
- permissions 必须声明 `network` 和 `filesystem`。
- filesystem 权限不得越出任务 workspace。
- http/sse 网络传输必须显式允许。
- manifest 不得包含明文 token、secret、password、api_key。

## 吸收边界

| 类型 | 可以吸收 | 不吸收 |
|---|---|---|
| MCP | tools/resources/prompts 分层、权限声明、外部工具上下文 | 新组织层、新默认入口 |
| 插件 | Browser、Documents、Spreadsheets、Presentations 等本地能力 | 绕过主帅直接交付 |
| skill | 方法论、任务规则、输出结构 | 大段外部文本和重复角色 |

## MCP 蒸馏

候选 MCP 不直接进入主链路，先进入候选台账：

```powershell
.\.wuji-tools\wuji-cli.cmd mcp-distill --catalog .\units\mcp_catalog.json --report .\outputs\mcp-distill-report.json
```

裁决含义：

- `absorb`：本地、低权限、能直接减少失败模式，可作为工具层候选。
- `defer`：网络、账号、OAuth、外部写操作，需要任务级授权。
- `reject`：高权限、明文密钥、许可证不清、来源不清或会扩大攻击面。

当前台账：`units/mcp_catalog.json`。

白帽默认结论：

| MCP 类型 | 裁决 | 原因 |
|---|---|---|
| Filesystem / Git / Time / Sequential Thinking | absorb | 本地为主，权限边界可控，能提升交付稳定性 |
| Fetch / GitHub / Browser 类 | defer | 有网络或账号权限，必须按任务授权 |
| Postgres / Slack / 任意带 secrets 的 MCP | reject/defer | 高隐私和凭据风险，不进默认链路 |

## 默认协同

| 任务 | 主帅 | 可补位工具 |
|---|---|---|
| 网页检查、前端验收 | 视觉主帅 / 开发主帅 | Browser 插件 |
| 文档、Word、资料整理 | 内容主帅 | Documents 插件 |
| 表格、数据分析 | 情报主帅 / 内容主帅 | Spreadsheets 插件 |
| PPTX 生成和修复 | 视觉主帅 | Presentations 插件 |
| 外部系统工具调用 | 对应主帅 | 通过 `mcp-guard` 的 MCP |

## 红线

- 不因工具热门就接入。
- 不让 MCP/插件替代参谋本部路由。
- 不让 MCP/插件替代白帽、安全、合规判断。
- 不把未安装、未授权、未验证的工具说成可用。
- 不把网络工具默认放进主链路。
