# 提示词蒸馏与离线优化

## 核心定位

这里不新开部门，也不照搬 DSPy 名词。
只吸收一条有效机制：反馈沉淀 -> 候选评测 -> 达线晋升。
这条链路默认是离线治理能力，不是每次执行任务都要经过的运行时主链。

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

- 这是一条离线优化链，不是日常执行主链。
- `feedback-log`：脱敏记录用户偏好、纠偏和禁忌词
- `feedback-dataset`：把多次反馈蒸馏成固定 `cases`
- `prompt-candidate-audit`：检查模板、变量、缓存前缀和密钥风险
- `prompt-eval`：用评测集对比 baseline 与 candidate
- `prompt-distill`：只把达线候选晋升稳定提示词内核或离线候选层，不直接抬成每次执行都要经过的运行时主链
- `context-pack`：仅在确有重复装配价值时固化稳定前缀，减少重复 token

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
- 不把 `feedback-log / feedback-dataset / prompt-distill / context-pack` 包装成日常任务默认必经层。
- 提示词优化先离线评测，再晋升稳定内核；不把离线优化链误写成运行时主链。
- 能不启动离线链就不启动；没有明显复用价值的单次偏好，不值得为此增加治理动作。
- learn 类机制只能生成候选，不得自动写回主规则、主 skill 或默认用户面话术。
- 不把框架名词当能力本身。
- 不把 runtime token 噪音包装成“提示词升级”。
- 不允许候选里出现密钥、账号、cookie、token 或占位假话。
- 未经过 `prompt-distill` 过线的候选，不得宣称“已蒸馏完成”。

## 200k Cached Token Diagnosis

If backend usage shows about 200k cached/blue-hit tokens per request, treat it as long-context bloat until proven otherwise.

Likely causes:

- The conversation/thread carries too much old history.
- Stable prefix contains too many resident rules, role bodies, skill bodies, or repeated summaries.
- Tool outputs, logs, search pages, transcripts, or README bodies were replayed instead of referenced by handle.
- Too many officers/skills were mounted together.

Fix path:

```text
runtime-context-audit
-> locate cached/input/fresh/output p95
-> split or reset unrelated long thread
-> replace history with task-state summary
-> keep evidence handles, not full replay
-> mount one owner, one selected skill, triggered officers only
-> verify cached_tokens_p95, input_tokens_p95, fresh_input_tokens_p95, output_tokens_p95 all fall together
```

Do not lower cached volume by increasing uncached input or verbose output. The target is high hit rate with smaller total volume.
