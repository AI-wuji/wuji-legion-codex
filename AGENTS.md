# 无极军团 2.0

## Models

- 阿极默认使用 `gpt-5.6-terra`，推理强度遵从用户当前选择；阿极保留唯一合并和写权限。
- Luna 仅处理不依赖父上下文的机械提取与受限检索；生成前可因可用性错误升级至 Terra。
- Sol 仅处理显式高推理、只读的子代理裁决；一轮生成、无回退，不得作为默认主模型或 Luna 回退。
- 路由声明必须由宿主以精确模型创建真实子代理；只有有效执行回执才能证明已发生模型路由。

- `SKILL.md` 是热路径规则；`capabilities/*/manifest.json` 是能力真相源。
- 阿极是唯一主脑、唯一合并者和唯一写权限。
- 没有女娲，没有第二路由，没有默认会审。
- 所有请求默认只作为一个小任务处理；先判定任务类型，走最简单有效的路径，不创建无边界“大任务”计划。
- 非平凡的故障、API/SDK、依赖、框架、路由、缓存、架构、迁移、性能、安全或集成任务，先执行一个受限 `preflight_workers` 搜索：官方 -> GitHub -> 社区，最多 3 个来源、90 秒，发现决定性证据立即停止。确定性文案、重命名、格式化等修改不触发搜索；用户明确离线或禁止搜索时必须跳过。
- `preflight_workers` 必须在 `workers` 之前顺序完成，禁止并行启动。若预检证据改变方案，立即作废原执行计划并重新路由；不得让旧 worker 继续执行。
- 专家是任务契约加完整冷能力包，不是长驻人格提示词。
- 当 `wuji route` 输出 `preflight_workers` 或 `workers` 时，必须使用每个 worker 的精确 `model`、`session_key` 和 `fallback_models` 真正委派执行；只输出路由 JSON 而不创建 worker，不算完成模型路由。阿极仍是唯一的路由、合并、写入和收口主脑。
- 每个 worker 完成时必须返回路由声明的全部 `execution_evidence_fields`，包括有序尝试/实际模型、payload 哈希与字节数、token/cache、计费基线/实际成本/节省额和阿极验收；没有宿主执行证据，不得把分支标记为完成。
- Sol、Terra、Luna 之间不得假设共享提示缓存。每个 worker 在任务开始时选定一次模型，并在自己的 `session_key` 内保持粘性。代码任务只有在同一查询生成的内容寻址上下文工件通过源文件哈希、实际字节数、至少 60% 检索锚点覆盖率、代码摘录和总重放成本门禁后，才允许委派一个独立 Terra 实现分支。宿主必须按 `stable_capability_prefix -> context_payload -> task_contract` 发送确定性 prompt；演示与写作默认留在阿极。Luna 只允许因模型不可用或生成前 provider 错误向上升级到 Terra；Terra 才可在同样条件下升级到 Sol，最多一次模型切换、两次尝试；Sol 一次尝试且无回退。生成后不得降级、切换或因质量付费重试。
- “已融合”必须同时具备真实调用入口和通过的行为测试。
- smoke/mount 探针只能证明 callable，不得把能力说成已融合。
- `behavior-verified` 必须依赖 `WUJI_PROBE_EVIDENCE_DIR` 中真实文件及验证器独立 SHA-256；不接受自报 receipt、signature 或退出码。`primary` 还必须有演化生成的内容寻址 promotion receipt 和归档基线。
- 检索统一采用两类有界关系图：工作区图默认先定位 `workspace -> files -> symbols/tests`，陈旧即重建、每文件最多 512 个词、每词最多 256 个文件引用、无命中最多扫描 512 个源文件；经验图只在 `failure`、`reported-failure`、`explicit-reuse`、`capability-miss`、`verification-trace` 事件触发，正常任务启动禁止查询。
- 图谱只返回紧凑关系、候选路径和解决方案位置，正文按需读取；金字塔只优化路由，不是无限事实库。经验记录必须带显式 scope、根因和验证文件 SHA-256，并受查询预算约束。
- `internal_adversarial_pass` 只是阿极内部反方声明，不是真实白帽证据。只有实际启动独立 officer 并留下审查结果，才可报告白帽已执行；用户点名白帽或风险契约要求独立证据时不得只返回布尔值。
- 默认只挂载 primary 来源；secondary/optional 需点名或要求完整能力。
- 多意图时读取 SecondaryCapabilities，写权限仍只属于阿极。
- 旧仓库 `E:/wuji-projects/wuji-legion-codex` 只读参考，不回退、不覆盖。
- 修改页面必须替换真实入口并清理旧实现，不建兼容版或平行 v2 页面。
- 不在仓库中保存 API Key、Token 或会话内容。
