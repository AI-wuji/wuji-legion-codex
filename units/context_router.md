# 上下文路由与缓存友好装配

## 目标

把“规则很长、任务很多、工具很多”的无极军团压成可缓存、可路由、可评测的执行链。

## 固定原则

- 稳定前缀和动态任务分开。
- 动态值尽量后移：时间、日期、session id、临时路径、随机数、统计字段不要污染稳定前缀。
- 先路由，再装配上下文。
- 只装最小必要规则，不整包塞给模型。
- 评测看命中率、token、耗时、重试和 QA，不靠感觉。
- 能沉成固定命令、固定工件、固定 skill 的，不继续依赖长上下文描述。
- 路由不仅决定“谁来做”，还要尽量决定“做到什么才算收口”，减少执行中反复问用户。
- 不同内容走不同压缩路径：`json / code / logs / text / tool-output` 不默认共用同一种压缩方式。

## Go 底座命令

以下命令主要用于离线治理、路由分析、缓存装配和专项评测，不是每次执行任务都要逐条经过的运行时主链。

```powershell
.\.wuji-tools\wuji-cli.cmd canon-report --report .\outputs\canon-report.json
.\.wuji-tools\wuji-cli.cmd route-task --config .\config.json --query "<任务描述>" --report .\outputs\route-report.json
.\.wuji-tools\wuji-cli.cmd context-pack --config .\config.json --workspace .\outputs --query "<任务描述>" --artifact <文件> --report .\outputs\context-pack.json
.\.wuji-tools\wuji-cli.cmd bench-report --workspace <任务目录> --report .\outputs\bench-report.json
```

## 含义

- `canon-report`：导出 Go 底座内置的固定模型三档、核心路由骨架、内置插件归口和 MCP 默认裁决。
- `route-task`：优先使用 Go 内置路由骨架和模型三档，再叠加 `config.json` 的环境与覆盖项。
- `route-task`：除了选执行链，还会给出推荐模型档位和推理强度。
- `route-task`：普通 `生图/插图/海报/封面/配图` 直接归到 `imagegen` 低档直出链；只有明确 `ComfyUI/pipeline/node/plugin/batch/video pipeline` 才进入 `comfyui` 技术链。
- `route-task`：命中复杂代码任务时会显式给出 `code_map_required=true`，把代码地图要求前推到路由层。
- `context-pack`：把任务拆成 `stable_prefix`、`dynamic_context` 和 `cache_key`，提高缓存命中。
- `context-pack`：装配时优先保留高价值上下文，低价值大块内容只保留压缩结果、证据摘要或可追回引用。
- `bench-report`：从 `bench.jsonl` 汇总速度、token、TPM、重试和 QA 通过率，并给出 `decision + evidence_level` 供蒸馏放行参考。
- 这些命令更适合作为离线治理、路由分析或专项装配工具，不应被包装成每次执行任务的默认必经层。

## 结果边界

- `route-task` 只做路由，不替代参谋本部判断。
- 模型选档以总返工成本最省为准：简单高频任务可走 `gpt-5.4-mini`，但复杂推理、复杂编码、重验证任务应直接升到 `gpt-5.4` 或 `gpt-5.5`。
- `context-pack` 只做装配，不替代主帅执行。
- `bench-report` 只做评测闭环，不美化结果。
- 任务一旦已明确，路由和装配应尽量支持连续执行到收口，而不是把模型推回“先做一步再问继续”的节奏。
- 默认不要求先跑 `route-task` 或 `context-pack`；只有任务足够复杂、重复率足够高，或确实需要缓存/归档证据时才值得调用。
- 如果压缩会损伤验证、审计或事实判断，就宁可多带一点上下文，也不为省 token 盲压。
