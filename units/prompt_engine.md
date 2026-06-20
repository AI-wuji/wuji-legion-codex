# 提示词局 - 提示词主帅 + prompt/spec 模式

## 核心定位

提示词局的专项入口是 `提示词主帅`。
它只在主链判定当前任务确实需要 prompt/spec 压缩时显性挂载，不是顶层 owner profile，也不是独立官。
它负责把模糊需求压成可执行 prompt、image-spec、video-spec、storyboard-spec 或 tool-spec。
普通出图必须快，不无限追问；复杂任务则先压规格，再交对应主线执行。

## 内置模式

| 模式 | 适用任务 | 输出 |
|---|---|---|
| 生图模式 | 插图、封面、海报、PPT 配图 | prompt + negative + size/style |
| 视频模式 | 文生视频、图生视频、动效说明 | subject/action/scene/camera/duration |
| 分镜模式 | 短剧、动画、AI 视频故事板 | 镜号、景别、画面、对白、时长 |
| 工具模式 | CLI/API/脚本参数 | tool-spec 或命令参数 |
| 蒸馏模式 | 提示词候选优化、缓存友好重写、离线评测 | 仅离线候选审计与晋升记录 |

## 普通生图规则

- 用户要求出图时，直接走 `imagegen`。
- 默认按用户原描述补足主体、风格、构图、尺寸和负面约束。
- 除非缺少关键画幅或安全边界，不追问。
- 除非用户只要提示词，否则不得只交提示词。
- 图片里的关键长文本默认拆到 PPT/HTML 层，不让生图模型承担可读排版。

执行约束
- Agnes 图像或视频镜像失败时，只允许显式回退到默认 GPT 路径。
- 不得静默改成本地绘制、本地占位图，或把本地结果冒充模型直出。

## Prompt 结构

```text
目标/用途
-> 主体
-> 动作/场景
-> 风格/镜头
-> 构图/画幅
-> 约束/负面
-> 输出格式
```

## 提示词蒸馏

- DSPy / GEPA / MIPROv2 只吸收“离线优化机制”，不作为运行时主框架进入默认主链路。
- `headroom learn` 只吸收“失败会话 -> 候选修正 -> 离线晋升”机制，不吸收自动改主规则。
- 日常使用中的偏好、纠偏和禁忌词，必要时才进入 `feedback-log` 脱敏沉淀，再由 `feedback-dataset` 蒸馏成评测集。
- 提示词主帅只在做离线候选晋升时，才把候选交执行底座做 `prompt-candidate-audit`、`prompt-eval`、`prompt-distill`。
- 候选只有在评测集覆盖率、缓存友好度和简洁度同时过线时，才允许晋升。
- `feedback-log / feedback-dataset / prompt-eval / prompt-distill` 默认属于离线治理链，不属于每次执行任务都要经过的运行时主链。
- 规则见 [prompt_optimization.md](E:\wuji-projects\wuji-legion-codex\units\prompt_optimization.md)。

## 当前专家

- `提示词主帅`：当前 prompt 专项入口，内部包含生图、视频、分镜、工具和蒸馏模式。
