# 专家领域谱系矩阵

这份矩阵定义“需要什么专业能力”，不定义专家数量。历史专家先经过聚类、去重和边界审查，再映射到少量领域契约。人格、重复协调角色和第二控制平面不重新实体化。

## 当前领域

| 领域契约 | 历史资产信号 | 当前绑定 | 专业边界 | 主要缺口 | 晋级证据 |
| --- | --- | --- | --- | --- | --- |
| `code-repair` | agentic coding、Claude loops、disciplined debug、patch debt | code、code-review | 只处理可复现代码问题和变更验证；不代替架构决策 | 真实宿主路由、失败回执、回归探针 | 修复前后测试证据、独立 SHA-256、失败分类 |
| `research` | AnySearch、research-mode、URL/视频证据、competitive teardown | search、writing | 先证据再结论；不把搜索结果当事实，不默认全网无限检索 | 官方/GitHub/社区预检的真实调用链 | 来源句柄、冲突检查、引用完整性 |
| `document-deck` | ppt-master、ppt-keynote、paper-deck、内容结构原则 | documents、presentation、writing | 产出文档或演示制品；不拥有发布账号或第二写作入口 | 专家专用工作流和渲染验收 | 可打开制品、渲染检查、内容检查 |
| `data-analysis` | SheetMagic、data-profile、异常与统计原则 | data | 只在有数据对象和分析问题时唤醒；不常驻索引数据 | 真实数据探针和异常证据格式 | profile、计算复核、异常证据 |
| `frontend-visual` | taste、CopilotKit、动态看板、设计画布 | frontend、visual、image | 负责界面/视觉实现与检查；不成为泛化 QA 或品牌人格 | 真实浏览器渲染、交互和移动端探针 | 截图/像素、交互回执、响应式检查 |
| `incident-diagnosis` | trivy、pg_durable、storage analyzer、性能/安全原则 | search、code-review、security | 处理故障、依赖、性能和安全风险；不自动扩大权限或修改治理规则 | 有界 preflight 与根因证据 | preflight、根因、回归和安全证据 |

## 数量与质量规则

- 领域覆盖不足时，先扩充该领域的 Skill/MCP 工作流和验证探针，不先新增专家。
- 只有当一个领域出现稳定、互斥、可验证且持续降低误触发或失败率的任务簇，才允许拆分契约。
- 每个领域默认一个主专家，必要时最多一个辅助专家；专家平时缄默。
- `doctrine-only` 只表示契约和原则存在；没有宿主调用入口及 `WUJI_PROBE_EVIDENCE_DIR` 独立哈希证据，不得升级为 `callable` 或 `behavior-verified`。
- 专家版本的晋级必须有真实任务对照、失败/复用/来源/验证事件、内容寻址 promotion receipt 和边界复核。数量变化本身不是晋级理由。

## 历史资产处置

`migration/legacy-verdict-ledger.json` 中的 `core-doctrine` 继续归入阿极/PonyTail 的判断和调度原则；`presentation`、`writing`、`search`、`visual` 等真实领域资产只作为对应领域专家的冷能力来源。`exclude` 对象不因“专家化”重新进入运行时。

