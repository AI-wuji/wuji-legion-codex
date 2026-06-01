# 提示词蒸馏与离线优化

## 核心定位

这不是再加一个新部门，也不是把 DSPy 一整套框架直接搬进来。
无极军团吸收的是它背后的有效机制：先建评测集，再做候选对比，再只提升被验证过的提示词内核。

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

## 已吸收的机制

- 离线候选审计：检查目标、模板、变量、缓存前缀和密钥暴露风险。
- 使用中沉淀：先把用户偏好、纠偏、禁忌词脱敏记入 `feedback-log`。
- 离线评测集：再由 `feedback-dataset` 把多次反馈蒸馏成固定 `cases`，检查 required / forbidden terms 覆盖率。
- 蒸馏晋升：候选相对 baseline 提升达线才 `absorb`，否则 `defer` 或 `reject`。
- 缓存友好内核：稳定前缀单独固化，动态任务部分后装，减少重复 token。

## 当前命令

```powershell
.\.wuji-tools\wuji-cli.cmd feedback-log --workspace .\outputs --task "日常答复质量" --prefer "keep the answer concise" --prefer "state uncertainty explicitly" --avoid "placeholder"
.\.wuji-tools\wuji-cli.cmd feedback-dataset --log .\outputs\feedback\feedback-log.jsonl --report .\outputs\prompt-dataset.json
.\.wuji-tools\wuji-cli.cmd prompt-candidate-audit --candidate .\outputs\prompt-candidate.json --report .\outputs\prompt-candidate-audit.json
.\.wuji-tools\wuji-cli.cmd prompt-eval --candidate .\outputs\prompt-candidate.json --dataset .\outputs\prompt-dataset.json --report .\outputs\prompt-eval.json
.\.wuji-tools\wuji-cli.cmd prompt-distill --baseline .\outputs\prompt-baseline.json --candidate .\outputs\prompt-candidate.json --dataset .\outputs\prompt-dataset.json --report .\outputs\prompt-distill.json
```

## 评测集格式

```json
{
  "summary": {
    "cases": 2,
    "prefer_terms": ["cite primary sources", "keep the answer concise"],
    "avoid_terms": ["placeholder", "todo"]
  },
  "cases": [
    {
      "id": "feedback-01",
      "task": "source discipline",
      "required_terms": ["cite primary sources", "state uncertainty explicitly"],
      "forbidden_terms": ["todo", "placeholder"]
    }
  ]
}
```

## 铁律

- 运行时不堆长记忆；只记录可蒸馏的偏好信号，再离线转成评测集。
- 提示词优化先离线评测，再晋升主链路。
- 不把框架名词当能力本身。
- 不把 runtime token 噪音变成“提示词工程升级”。
- 不允许提示词候选里出现密钥、账号、cookie、token 或占位假话。
- 未经 `prompt-distill` 过线的候选，不得宣称“已蒸馏完成”。
