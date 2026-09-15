# 专家与路由来源目录

本目录把“产品行为参考”和“可融合代码/技能”分开。豆包工作、WorkBuddy 的内部专家实现不可审读，因此只能作为用户体验目标：预置领域流程、按需静默唤醒、自动绑定工具、持续优化；不能作为已验证代码来源。

## 已审读的公开候选

| 来源 | 许可证 | 可蒸馏切片 | 不融合部分 | 证据状态 |
| --- | --- | --- | --- | --- |
| [DeHor-Labs/agent-capability-router](https://github.com/DeHor-Labs/agent-capability-router) | MIT | 任务形状分类、最小能力路径、触发/反触发、路由结果验证；含 `SKILL.md`、`scripts/validate-skill.py`、`tests/test_route_task.py` | 不引入其第二路由器、自动化/记忆外壳或未知宿主适配 | 源码、脚本、测试已审读；待 Wuji 宿主行为探针 |
| [shinpr/sub-agents-skills](https://github.com/shinpr/sub-agents-skills) | MIT | 专家定义与执行器分离、按名发现、结构化回执、执行测试 | 不运行外部 CLI 代理，不引入其权限/会话壳，不创建第二用户入口 | 源码、`pyproject.toml`、SKILL、测试已审读；待 Wuji 宿主行为探针 |

## 融合门禁

- README 只能导航；必须审读源码入口、配置、测试/探针和许可证。
- 蒸馏先落为领域契约或冷参考，不改变运行时状态。
- 只有 Wuji 真实调用入口、任务级模型路由、原生回执，以及 `WUJI_PROBE_EVIDENCE_DIR` 中独立 SHA-256 验证文件齐全，才能标记 `callable` 或 `behavior-verified`。
- 来源能力若与阿极唯一入口、参谋部确定性调度、预算门禁或白帽判断冲突，则分类为“参考/不融合”。

## 专家团形成方式

自然语言任务先由阿极抽取目标、约束、风险和验收条件，再匹配领域契约；契约绑定已验证 Skill/MCP/插件工作流，参谋部只负责任务图和回执，执行节点完成制品，独立验证节点决定是否通过。领域内只有出现稳定互斥任务簇且对照数据证明误触发或失败率下降，才拆分专家；专家数量不是优化指标。

