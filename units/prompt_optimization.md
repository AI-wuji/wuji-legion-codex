# 提示词蒸馏与离线优化

## 核心定位

这里不新开部门，也不照搬 DSPy 名词。
只吸收一条有效机制：反馈沉淀 -> 候选评测 -> 达线晋升。

```text
daily feedback
-> feedback-log
-> feedback-dataset
-> candidate prompt
-> prompt-candidate-audit
-> prompt-eval
-> prompt-distill
-> absorb / defer / reject
-> promote stable prompt kernel
```

## 当前链路

- `feedback-log`：脱敏记录用户偏好、纠偏和禁忌词
- `feedback-dataset`：把多次反馈蒸馏成固定 `cases`
- `prompt-candidate-audit`：检查模板、变量、缓存前缀和密钥风险
- `prompt-eval`：用评测集对比 baseline 与 candidate
- `prompt-distill`：只把达线候选晋升主链
- `context-pack`：固化稳定前缀，减少重复 token

## 当前命令

```powershell
.\.wuji-tools\wuji-cli.cmd feedback-log --workspace .\outputs --task "日常答复质量" --prefer "keep the answer concise" --prefer "state uncertainty explicitly" --avoid "stub text"
.\.wuji-tools\wuji-cli.cmd feedback-dataset --log .\outputs\feedback\feedback-log.jsonl --report .\outputs\prompt-dataset.json
.\.wuji-tools\wuji-cli.cmd prompt-candidate-audit --candidate .\outputs\prompt-candidate.json --report .\outputs\prompt-candidate-audit.json
.\.wuji-tools\wuji-cli.cmd prompt-eval --candidate .\outputs\prompt-candidate.json --dataset .\outputs\prompt-dataset.json --report .\outputs\prompt-eval.json
.\.wuji-tools\wuji-cli.cmd prompt-distill --baseline .\outputs\prompt-baseline.json --candidate .\outputs\prompt-candidate.json --dataset .\outputs\prompt-dataset.json --report .\outputs\prompt-distill.json
```

## 铁律

- 运行时不堆长记忆；只记录可蒸馏的偏好信号。
- 提示词优化先离线评测，再晋升主链路。
- 不把框架名词当能力本身。
- 不把 runtime token 噪音包装成“提示词升级”。
- 不允许候选里出现密钥、账号、cookie、token 或占位假话。
- 未经过 `prompt-distill` 过线的候选，不得宣称“已蒸馏完成”。
