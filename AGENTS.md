# 无极军团 2.0

- `SKILL.md` 是热路径规则；`capabilities/*/manifest.json` 是能力真相源。
- 阿极是唯一主脑、唯一合并者和唯一写权限。
- 没有女娲，没有第二路由，没有默认会审。
- 所有任务按小任务直达执行；先做最简单有效的方法，再做必要验证。
- 专家是任务契约加完整冷能力包，不是长驻人格提示词。
- 当 `wuji route` 输出 `workers` 时，必须使用每个 worker 的精确 `model` 和 `fallback_models` 真正委派执行；只输出路由 JSON 而不创建 worker，不算完成模型路由。阿极仍是唯一的路由、合并、写入和收口主脑。
- 每个 worker 完成时必须返回 `requested_model`、有序 `attempts`、`effective_model`、`result_handle`、`context_handle_ids`、`context_bytes_sent`、`task_contract_bytes` 和 `delegation_gate_reason`；没有宿主执行证据，不得把分支标记为完成。
- Sol、Terra、Luna 之间不得假设共享提示缓存。代码任务只有在同一查询生成的内容寻址上下文工件通过源文件哈希、实际字节数和重放成本门禁后，才允许委派一个独立 Terra 实现分支；依赖实现的验证由阿极顺序完成。门禁失败就留在阿极，不把“降级模型”本身当作节省费用的证据。
- “已融合”必须同时具备真实调用入口和通过的行为测试。
- smoke/mount 探针只能证明 callable，不得把能力说成已融合。
- `behavior-verified` 必须依赖 `WUJI_PROBE_EVIDENCE_DIR` 中真实文件及验证器独立 SHA-256；不接受自报 receipt、signature 或退出码。`primary` 还必须有演化生成的内容寻址 promotion receipt 和归档基线。
- 默认只挂载 primary 来源；secondary/optional 需点名或要求完整能力。
- 多意图时读取 SecondaryCapabilities，写权限仍只属于阿极。
- 旧仓库 `E:/wuji-projects/wuji-legion-codex` 只读参考，不回退、不覆盖。
- 修改页面必须替换真实入口并清理旧实现，不建兼容版或平行 v2 页面。
- 不在仓库中保存 API Key、Token 或会话内容。
