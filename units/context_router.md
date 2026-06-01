# 上下文路由与缓存友好装配

## 目标

把“规则很长、任务很多、工具很多”的无极军团压成可缓存、可路由、可评测的执行链。

## 固定原则

- 稳定前缀和动态任务分开。
- 先路由，再装配上下文。
- 只装最小必要规则，不整包塞给模型。
- 评测看命中率、token、耗时、重试和 QA，不靠感觉。

## Go 底座命令

```powershell
.\.wuji-tools\wuji-cli.cmd route-task --config .\config.json --query "<任务描述>" --report .\outputs\route-report.json
.\.wuji-tools\wuji-cli.cmd context-pack --config .\config.json --workspace .\outputs --query "<任务描述>" --artifact <文件> --report .\outputs\context-pack.json
.\.wuji-tools\wuji-cli.cmd bench-report --workspace <任务目录> --report .\outputs\bench-report.json
```

## 含义

- `route-task`：按 `config.json` 的规则和关键词做确定性路由。
- `route-task`：除了选执行链，还会给出推荐模型档位和推理强度。
- `context-pack`：把任务拆成 `stable_prefix`、`dynamic_context` 和 `cache_key`，提高缓存命中。
- `bench-report`：从 `bench.jsonl` 汇总速度、token、TPM、重试和 QA 通过率。

## 结果边界

- `route-task` 只做路由，不替代参谋本部判断。
- 模型默认走低成本档：`gpt-5.4-mini`；只有任务信号更强时，才升到 `gpt-5.4` 或 `gpt-5.5`。
- `context-pack` 只做装配，不替代主帅执行。
- `bench-report` 只做评测闭环，不美化结果。
