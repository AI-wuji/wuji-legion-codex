# 专家与专家团外部基准

日期：2026-09-15

目标：寻找类似豆包工作、WorkBuddy 的预置专家和专家团机制，提取单项优势，并决定哪些可蒸馏进入无极军团。该报告不把产品宣传当成行为验证。

## 来源与结论

| 来源 | 观察到的优势 | 白帽限制 | 处置 |
| --- | --- | --- | --- |
| [oh-my-claudecode](https://github.com/Yeachan-Heo/oh-my-claudecode)（MIT） | 预置多角色、自动编排、用户低学习成本；README 另提供 Codex 适配项目入口 | 角色数量和编排复杂度可能扩大上下文与调用成本；本次未将其运行时安装到 Wuji | 蒸馏“预置角色契约 + 静默选择”；不整包融合 |
| [OpenAI Swarm](https://github.com/openai/swarm)（MIT） | Agent=指令+工具，handoff 原语轻量、可测试 | 官方 README 标注 experimental/educational，并建议生产迁移 Agents SDK；无状态 handoff 不满足 Wuji 的需求图/证据门禁 | 仅参考轻量 handoff；不引入运行时 |
| [Agent Orchestrator](https://github.com/Untrivial-ai/agent-orchestrator)（Apache-2.0） | 每项编码任务独立工作区、反馈循环、PR/CI/review 可见 | 桌面编排器较重，面向编码流程；不能替代 Wuji 的阿极需求图与确定性参谋部 | 蒸馏“任务隔离 + 反馈闭环”；不整包融合 |
| [awesome-claude-agents](https://github.com/vijaythecoder/awesome-claude-agents)（MIT） | 24 个预置专长角色、自动配置团队、领域分工明确 | README 明确警告 token-intensive，复杂功能约 10k–50k tokens；角色集合质量未由 Wuji 独立验证 | 作为角色设计参考；禁止默认全员启动 |
| [agent-capability-router](https://github.com/DeHor-Labs/agent-capability-router)（MIT） | 任务形状识别、最小能力路径、触发/反触发、验证测试 | 仍需宿主适配和 Wuji 证据探针 | 可蒸馏并优先落地 |
| [sub-agents-skills](https://github.com/shinpr/sub-agents-skills)（MIT） | 专家定义与执行器分离，结构化执行回执，带测试 | 外部 CLI 执行模型不属于 Wuji 原生执行面 | 可蒸馏契约/回执，不运行其外部 CLI |

## 最值得采用的单项能力

1. **专家契约**：每个专家预先声明触发条件、反触发条件、输入字段、工具/Skill/MCP、输出结构、验收探针和预算。
2. **专用工作流**：专家不是一句人格提示词，而是固定的步骤编译器。例如图像专家把结构化需求编译成正向提示词、负面提示词、参数、生成和检查步骤。
3. **静默 MoE 路由**：默认不启动专家；阿极完成需求整理和白帽质疑后，参谋部只派发命中的最小专家集合。
4. **任务隔离与反馈闭环**：每个执行节点有独立上下文、写边界、结果句柄和验证反馈；失败只重做受影响节点。
5. **证据门禁**：角色定义、smoke 或自报“完成”都不能晋级；必须有真实宿主调用和独立验证文件。

## 不能照搬的部分

- 大型常驻专家 roster、默认多专家并行和全局 agent-team 配置：与低 token、快速完成和防无限循环冲突。
- 第二个主路由器、第二个用户入口或专家自行 handoff：破坏阿极唯一沟通入口和参谋部确定性调度。
- 外部 CLI/桌面工作区/账号会话运行时：超出当前 Codex 宿主边界，且难以满足内容寻址回执。
- 仅凭 README、star 数或演示质量宣称“已融合”：必须通过 Wuji 行为探针。

## 对无极军团的落地判断

这些来源支持“预置高质量专家流程”方向，但不支持恢复几十个常驻专家。应先为高频领域建立候选专家包，再用相同任务对照：任务触发准确率、误触发率、输出验收通过率、Token/上下文字节、失败恢复次数和独立验证通过率。只有对照结果改善，候选才可升级；否则保留为冷参考。

