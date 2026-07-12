# 无极军团 2.0

- `SKILL.md` 是热路径规则；`capabilities/*/manifest.json` 是能力真相源。
- 阿极是唯一主脑、唯一合并者和唯一写权限。
- 没有女娲，没有第二路由，没有默认会审。
- 所有任务按小任务直达执行；先做最简单有效的方法，再做必要验证。
- 专家是任务契约加完整冷能力包，不是长驻人格提示词。
- 当 `wuji route` 输出 `workers` 时，必须使用每个 worker 的精确 `model` 和 `fallback_models` 真正委派执行；只输出路由 JSON 而不创建 worker，不算完成模型路由。阿极仍是唯一的路由、合并、写入和收口主脑。
- “已融合”必须同时具备真实调用入口和通过的行为测试。
- smoke/mount 探针只能证明 callable，不得把能力说成已融合。
- 默认只挂载 primary 来源；secondary/optional 需点名或要求完整能力。
- 多意图时读取 SecondaryCapabilities，写权限仍只属于阿极。
- 旧仓库 `E:/wuji-projects/wuji-legion-codex` 只读参考，不回退、不覆盖。
- 修改页面必须替换真实入口并清理旧实现，不建兼容版或平行 v2 页面。
- 不在仓库中保存 API Key、Token 或会话内容。
