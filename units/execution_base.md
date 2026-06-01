# 执行底座 - Go 主链路 + 专项补位

## 核心定位

`执行底座`只有一个执行入口：`执行底座主帅`。

执行底座负责把无极军团里稳定、重复、可判定的动作沉淀成高可靠本地工具链。主实现语言统一为 Go；C#、Python、Node 等只作为专项补位工具，由 Go 执行底座统一调度和验收。

```text
规则和判断保持 Markdown / skill / 主帅机制
稳定动作下沉 Go / CLI / 单二进制工具
专项生态能力按需补位，但不替代 Go 主链路
独立白帽、质检、安全、合规继续外部放行
```

## 职责范围

执行底座主帅负责无极军团自身工具链：

- `wuji-cli`：无极军团统一本地命令入口。
- `guard`：路径、只读参考文件、输出边界、危险操作防线。
- `task`：任务开始、结束、耗时、产物路径、阻塞点记录。
- `sync`：同步到 `.codex`、`.agents`、技能目录和版本一致性检查。
- `audit`：乱码、占位符、版本号、规则冲突、空专家、无用文件扫描。
- `audit-sarif`：把确定性审计结果同步输出为 SARIF，便于安全、质检和 CI 统一读取。
- `workflow`：工作流 contract、packet、result、final-report 的确定性生成与校验。
- `beep`：完成、失败、提醒音的跨任务调度。
- `bench`：速度、token、耗时、命中率、失败率基准记录。
- `bench-report`：从 `bench.jsonl` 汇总 TPM、QA 通过率、平均耗时和重试率。
- `quality-gate`：只扫描项目源码边界，避免把工具链、输出物和缓存当作项目问题。
- `vulncheck`：存在 Go 模块和官方 `govulncheck` 时执行漏洞检查；缺工具时记为 SKIP，不伪造通过。
- `mcp-guard`：检查 MCP/插件 manifest、权限边界、网络传输和明文密钥风险。
- `route-task`：按 `config.json` 做确定性任务路由。
- `route-task`：同时给出推荐模型档位和 `reasoning_effort`，默认先走低成本档。
- `context-pack`：把稳定前缀、动态任务和 cache key 装配成缓存友好上下文包。
- `preview调度`：PPT、HTML、图片等预览导出命令的统一调度壳。
- `pptx-audit`：真 PPTX 可编辑性、整页图片占比、文字/形状/图片比例、参考素材复用率检查。
- `asset-map`：从参考 PPTX 抽取页型、可复用元素、教学插图、案例图、公式图、背景装饰和 image2 生图资产清单。
- `pptx-preflight`：在生成前检查 `reference-frame-map`、`reusable-asset-map`、`illustration-plan` 是否存在且可执行。
- `pptx-batch-gate`：在批量生成前检查 pilot page、pilot preview、pilot-score 是否存在且可执行。
- `time-guard`：记录非代码任务是否 10/15/30 分钟熔断，防止长时间口头执行和无效绕路。

## 语言策略

- Go 是执行底座唯一主链路语言。
- 执行底座只有一条 Go 主链路，不保留第二套同级实现。
- C# 可用于 Office/PPTX/COM 深度专项工具。
- Python 可用于 AI、数据、图片和临时转换专项工具。
- Node/TypeScript 可用于浏览器、HTML/UI 和前端专项工具。
- 专项工具不能直接成为默认入口，必须由 Go `wuji-cli` 调度、记录和验收。

## 内置模式

| 模式 | 适用任务 | 默认门禁 |
|---|---|---|
| guard模式 | 参考文件只读、输出路径安全、工作区边界 | dry-run + path report |
| task模式 | 任务生命周期、产物路径、耗时和阻塞记录 | start/end consistency |
| sync模式 | 规则、skill、config、专家卡同步 | version + file presence |
| audit模式 | 乱码、占位符、版本漂移、无用文件 | deterministic report |
| sarif模式 | 审计结果给安全、质检、CI 统一消费 | SARIF 2.1.0 report |
| workflow模式 | contract/packet/result/final-report | schema + required fields |
| beep模式 | 完成/错误/提醒音调度 | non-blocking fallback |
| bench模式 | 速度、成本、命中率、失败率 | reproducible command |
| bench-report模式 | TPM、QA通过率、平均耗时、重试率汇总 | reproducible report |
| quality-gate模式 | Go/PowerShell/Python 项目源码质量检查 | project-boundary only |
| vulncheck模式 | Go 官方漏洞检查补位 | govulncheck when available |
| mcp-guard模式 | MCP/插件进入任务链路前的权限和 manifest 检查 | no-go before tool use |
| route-task模式 | 配置驱动的确定性任务路由 | rule match + priority |
| context-pack模式 | 稳定前缀和动态任务分层装配 | stable-prefix-first |
| preview调度模式 | PPT/HTML/图片预览导出统一入口 | output existence |
| pptx-audit模式 | 可编辑 PPTX、参考素材复用、整页图片伪 PPT 检查 | hard fail report |
| asset-map模式 | 参考 PPTX 的页型、可复用资产和插图资产抽取 | reference-frame-map + reusable-asset-map + illustration-plan |
| pptx-preflight模式 | 生成前路线锁定，不让错误方案开跑 | no-go before pilot |
| pptx-batch-gate模式 | pilot page 过线后才允许批量生成 | no-go before batch |
| time-guard模式 | 非代码任务时间熔断和空转检测 | 10/15/30 min gates |

## 与开发主帅的边界

- 普通业务软件、前端、小程序、ComfyUI插件、AI工程和自动化仍归 `开发主帅`。
- 开发主帅可以实现 Go/C#/Python/Node 代码，但不拥有无极军团全局执行底座规划权。
- 执行底座主帅只接无极军团自身 `wuji-cli`、guard、task、sync、audit、workflow、beep、bench、preview调度。
- 如果一个工具服务于外部业务项目，而不是无极军团自身执行稳定性，默认仍归开发主帅。

## 与上层机制的边界

执行底座不替代：

- `阿极`：默认用户入口和短报。
- `参谋本部`：任务判断、主帅路由和验收标准。
- `女娲`：能力补位和 skill 融合判断。
- `白帽/质检/安全/合规`：独立第三方审查。
- `进化主帅`：官方源核验、蒸馏裁决和规则升级。
- `内容主帅/视觉主帅`：写作、设计、PPT、HTML、图像创意。

## 执行顺序

```text
识别稳定动作
-> 定 CLI 子命令和输入输出
-> Go 最小实现
-> 最小门禁
-> 专项工具按需补位
-> 交白帽/质检/安全按需审查
```

## 已落地工具

- `tools/wuji_cli.go`：Go 版 `wuji-cli` 主源码。
- `scripts/build-wuji-cli.ps1`：构建 Go 版 `wuji-cli`。
- `scripts/wuji-pptx-preflight.ps1`：自动构建并执行 PPTX 门禁。
- `scripts/test-wuji-cli.ps1`：回归验证通用门禁与 PPTX 门禁。

当前 `wuji-cli` 子命令：

| 子命令 | 硬控内容 |
|---|---|
| `reference-guard` | 参考文件存在；输出不得覆盖参考原件或写入参考目录 |
| `workflow-guard` | 工作流 contract/state/final-report/packets/results 完整性 |
| `claim-guard` | 声称完成、通过、成功、已融合、已生成时必须附证据文件 |
| `time-guard` | 非代码任务 10/15/30 分钟无可验产物时给出 NO-GO |
| `task` | 任务开始、心跳、阻塞、结束的 JSONL 记录 |
| `sync` | 源/目标规则树的关键文件存在性与版本一致性检查 |
| `audit` | 乱码、占位符、主链路残留冲突的代码级扫描 |
| `audit --sarif` | 同一审计结果输出 SARIF 2.1.0，供安全、质检或 CI 汇总 |
| `bench` | token、耗时、重试、QA 结果的 JSONL 基准记录 |
| `bench-report` | 从 `bench.jsonl` 汇总 TPM、QA 通过率、平均耗时和重试率 |
| `mcp-guard` | MCP/插件 manifest、capabilities、permissions、filesystem、network 和 secret marker 检查 |
| `route-task` | 从 `config.json` 路由任务到对应执行链 |
| `context-pack` | 输出 `stable_prefix`、`dynamic_context` 和 `cache_key` |
| `preview` | 预览导出命令调度并验证目标文件已生成 |
| `asset-map` | 从参考 PPTX 自动生成三张表 |
| `pptx-audit` | 真 PPTX 可编辑性和整页图片伪 PPT 风险检查 |
| `pptx-preflight` | 三张表存在且生成器未命中整页图片伪 PPT 路线 |
| `pptx-batch-gate` | 三张表、pilot page、pilot preview、pilot-score 齐全后才允许批量 |

PPT 参考任务批量生成前，先运行：

```powershell
.\scripts\wuji-pptx-preflight.ps1 -Workspace <任务工作区> -Generator <生成脚本路径>
.\scripts\wuji-pptx-preflight.ps1 -Workspace <任务工作区> -Generator <生成脚本路径> -Mode batch
```

任一返回 `NO-GO` 时不得启动对应阶段。

通用门禁示例：

```powershell
.\.wuji-tools\wuji-cli.cmd reference-guard --reference <参考文件> --output <输出路径>
.\.wuji-tools\wuji-cli.cmd workflow-guard --workspace <工作流目录> --stage final
.\.wuji-tools\wuji-cli.cmd claim-guard --claim completed --evidence <验证记录>
.\.wuji-tools\wuji-cli.cmd time-guard --kind non-code --elapsed-minutes 15 --phase prototype
```

## 禁止

- 不把规则、蒸馏、创意、审查判断硬编译进 Go。
- 不为追求速度绕过参考文件只读、路径安全和用户授权。
- 不把执行底座做成另一个“全能军团大脑”。
- 不抢开发主帅的普通业务代码和产品开发任务。
- 不用专项工具绕过 Go 主链路的记录和验收。
- 没有基准测试，不宣称“更快”。
- 没有可复现命令，不宣称“更稳”。
- 执行底座可以硬控“能不能交付”，但不替视觉主帅决定“怎么设计更美”。

## 当前专家

- `执行底座主帅`：无极军团通用执行底座、`wuji-cli`、guard、sync、audit、workflow、beep、bench 和 preview 调度入口。
