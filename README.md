# 无极军团 / Wuji Legion v6.0

> 阿极极速秘书层 + 参谋本部轻量状态机 + 单主帅执行 + 女娲按需补位 + 白帽封驳

## 一句话

无极军团是给 Codex 使用的轻量调度框架。默认由阿极快速沟通；任务成型后，参谋本部只做分拣和路由，优先选择一个主帅负责到底。女娲只在主帅缺能力时补位，白帽负责提前封驳错误路线。

核心原则：

- 省 token 高命中。
- 高质高效。
- 保留无极军团命名，不改成三省六部、内阁或其他外部编制。

## 状态机

| 状态 | 触发 | 动作 |
|---|---|---|
| `FAST_REPLY` | 普通沟通、轻量问题 | 阿极 1-3 句短答，零工具 |
| `CLARIFY` | 目标或交付物不清 | 阿极问最少关键问题 |
| `SINGLE_COMMANDER` | 一个主帅可完成 | 参谋本部选主帅，主帅负责到底 |
| `LEGION_TASK` | 明确激活无极军团或复杂多能力任务 | 参谋本部路由，女娲按需补位 |
| `BLOCKED` | 缺文件、权限、API、环境或关键信息 | 短报阻塞和 1-3 个选项 |
| `DONE` | 成品通过最低验收 | 阿极短报结果和路径 |

## 架构

```text
用户
  ↓
阿极秘书层：快答 / 澄清 / 任务规划
  ↓
参谋本部：分拣 / 路由 / 封驳标准
  ↓
单主帅：一个 skill、插件或部门负责到底
  ↓
女娲：仅在主帅缺能力时补专家、skill、MCP、插件
  ↓
白帽/质监：提前封驳 + 交付抽检
  ↓
阿极短报：结果 + 路径
```

## 单主帅路由

| 任务 | 主帅 |
|---|---|
| 普通生图、插图、海报、封面 | imagegen |
| 模板续写、套模板 PPT | Presentations template-following |
| 从零 PPT、重度美化 | 臧老师 / elite-powerpoint-designer |
| HTML/UI | 项目原生前端主线 |
| 搜索调研 | 情报局 |
| 代码开发 | 对应技术栈主线 |

## PPT 硬门

PPT 必须按顺序交付：

```text
slide-spec.json 或等价逐页结构
→ design-brief.md 或等价视觉策略
→ layout-plan.json 或等价版式计划
→ PPTX
→ 预览/抽检
```

禁止：

- 把逐字稿硬塞进模板。
- 在 `slide-spec` 前先查 unzip、工具链、skill 路径或模板脚本。
- 用 OpenDesign、imagegen、slide-studio 抢主线。
- 用解释、页稿或半成品冒充 PPT 成品。

## 图像硬门

- 普通生图直接 imagegen。
- 失败前不查环境、不读 skill、不写长计划。
- 成功：预览 + `文件在...`
- 失败：短报错误 + `重试 / 换通道 / 排查`

## 安装

```powershell
Copy-Item .\GLOBAL_AGENTS.md C:\Users\Administrator\.codex\AGENTS.md -Force
Copy-Item .\SKILL.md C:\Users\Administrator\.agents\skills\wuji-legion\SKILL.md -Force
Copy-Item .\config.json C:\Users\Administrator\.agents\skills\wuji-legion\config.json -Force
Copy-Item .\units C:\Users\Administrator\.agents\skills\wuji-legion\units -Recurse -Force
Copy-Item .\experts C:\Users\Administrator\.agents\skills\wuji-legion\experts -Recurse -Force
```

## 版本

- `v6.0`：轻量状态机 + 单主帅内核。
- `v5.27`：触发词收窄。
- `v5.26`：普通生图直连。
- `v5.25`：极速秘书硬门。
- `v5.24`：单主帅制初版。
