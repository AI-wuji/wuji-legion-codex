# 提示词局 - 提示词主帅 + prompt/spec 模式

## 核心定位

提示词局只有一个入口：`提示词主帅`。
它负责把模糊需求压成可执行 prompt、image-spec、video-spec、storyboard-spec 或 tool-spec。
普通出图必须快，不无限追问；复杂任务则先压规格，再交对应主线执行。

## 内置模式

| 模式 | 适用任务 | 输出 |
|---|---|---|
| 生图模式 | 插图、封面、海报、PPT 配图 | prompt + negative + size/style |
| 视频模式 | 文生视频、图生视频、动效说明 | subject/action/scene/camera/duration |
| 分镜模式 | 短剧、动画、AI 视频故事板 | 镜号、景别、画面、对白、时长 |
| 工具模式 | CLI/API/脚本参数 | tool-spec 或命令参数 |
| 蒸馏模式 | 提示词候选优化、缓存友好重写、离线评测 | candidate-audit / eval / distill report |

## 普通生图规则

- 用户要求出图时，直接走 `imagegen`。
- 默认按用户原描述补足主体、风格、构图、尺寸和负面约束。
- 除非缺少关键画幅或安全边界，不追问。
- 除非用户只要提示词，否则不得只交提示词。
- 图片里的关键长文本默认拆到 PPT/HTML 层，不让生图模型承担可读排版。

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
- 日常使用中的偏好、纠偏和禁忌词，先用 `feedback-log` 做脱敏沉淀，再由 `feedback-dataset` 蒸馏成评测集。
- 提示词主帅先把候选压成稳定前缀 + 动态任务，再交执行底座做 `prompt-candidate-audit`、`prompt-eval`、`prompt-distill`。
- 候选只有在评测集覆盖率、缓存友好度和简洁度同时过线时，才允许晋升。
- 规则见 [prompt_optimization.md](E:\wuji-projects\wuji-legion-codex\units\prompt_optimization.md)。

## 当前专家

- `提示词主帅`：唯一 prompt 入口，内部包含生图、视频、分镜、工具和蒸馏模式。
