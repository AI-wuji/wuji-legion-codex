# 蒸馏师团 — 官方源核验 + 能力蒸馏 + 反叠加闸门

## 核心定位

蒸馏师团隶属进化部，负责把外部 skill、工具、工作流和社区经验转化为无极军团自己的能力。

它不是安装器，也不是收藏夹。它只做五件事：

1. 查官方源和最新版。
2. 判断有没有必要吸收。
3. 抽取可执行机制，不照搬组织名和长篇说明。
4. 映射到唯一主责师团或专家。
5. 经试验场和白帽通过后再入库。

## 铁律

- 没看官方源，不得说已经蒸馏。
- 没看源码或规则正文，不得说已经看明白。
- 不知道上游是否最新，就标注“待核验”，不能冒充结论。
- 不复制上游完整 skill，不把别人的组织编制搬进无极军团。
- 不为了显得强大而叠加入口、角色、流程或口号。
- 每次蒸馏必须能回答：解决了哪个失败模式，提升了哪个交付指标。

## 六步蒸馏链

```text
需求触发
-> Source Scan 官方源核验
-> Necessity Gate 必要性判断
-> Essence Extract 精华抽取
-> Owner Map 主责归位
-> Sandbox Verify 试验验证
-> Publish 入库与日志
```

## Source Scan 官方源核验

每个候选来源必须登记：

| 字段 | 要求 |
|---|---|
| source_id | 简短唯一名 |
| upstream_url | 官方仓库、官网或文档地址 |
| checked_at | 核验日期 |
| checked_commit | Git commit / 版本号；没有则写 `n/a` |
| license | MIT / Apache-2.0 / source-available / unknown |
| files_read | 实际读过的源码、规则或 README |
| status | verified / pending / rejected |

优先级：

1. 官方 GitHub 仓库源码。
2. 官方文档或规范。
3. 作者文章和演示。
4. 社区二手介绍。

二手介绍只能作为线索，不能作为蒸馏依据。

## Necessity Gate 必要性判断

只允许三种结论：

| 结论 | 条件 | 动作 |
|---|---|---|
| absorb | 明确补齐失败模式、质量短板或效率短板 | 进入精华抽取 |
| defer | 有价值但当前没有高频触发或证据不足 | 登记，不入主规则 |
| reject | 只增加复杂度、重复现有能力、无法验证或授权不清 | 记录原因，停止 |

必要性至少命中一项：

- 修复已发生的系统性失败。
- 明显降低 token、耗时或误触发。
- 明显提高最终交付质量。
- 给现有师团补上缺失能力。
- 把不稳定人工步骤变成可验证流程。

## Essence Extract 精华抽取

只抽四类东西：

- 工作流：例如阶段路由、薄切片、候选验证。
- 验收门：例如前置审核、质量阈值、发布门禁。
- 资产结构：例如 metadata、references、scripts、assets 分层。
- 失败反模式：例如过度加载、硬套模板、无验证即发布。

不抽：

- 上游组织名称。
- 大段营销文案。
- 与无极军团状态机冲突的入口。
- 没有验证路径的“看起来很厉害”机制。

## Owner Map 主责归位

每个蒸馏结果必须只进入一个主责位置：

| 类型 | 主责 |
|---|---|
| 路由、状态机、阶段判断 | 参谋本部 |
| skill / 专家 / 插件补位 | 女娲 |
| 蒸馏、升级、失败复盘 | 进化部 |
| 前置否决、质量门禁 | 白帽 / 质监局 |
| 沙箱验证、A/B、回归测试 | 试验场 |
| 来源、版本、日志 | 档案局 |
| 具体领域能力 | 对应师团或专家 |

禁止同一机制在多个 unit 重复写一遍。跨 unit 只写“调用关系”。

## Sandbox Verify 试验验证

入库前至少做一种验证：

- 对比现有方案是否更省 token、更快或更稳。
- 用真实失败案例复测是否改善。
- 用小样本任务跑一次可执行路径。
- 由白帽检查是否增加冲突、误触发或半成品风险。

无法验证的能力只登记为 `defer`。

## Publish 入库与日志

通过后必须更新：

- 对应 `units/*.md` 或 `experts/**/*.md`。
- `CHANGELOG.md`。
- 如涉及公开来源，README 的来源说明或本文件的来源台账。

如果用户要求同步 GitHub，再提交并推送。

## 白帽蒸馏否决项

以下任一情况直接否决：

- 来源不是官方源，且没有其他强证据。
- 许可证不清，还要复制上游内容。
- 只是在无极军团外面再套一个别人的流程。
- 让简单任务变慢、变啰嗦、变多角色表演。
- 需要用户手动选择大量内部路线。
- 没有落到某个具体师团或专家。
- 没有验收方式。

## 已核验来源台账

| source_id | upstream | checked_at | commit/version | license | 已读重点 | 蒸馏结论 |
|---|---|---|---|---|---|---|
| openai-skills | https://github.com/openai/skills | 2026-05-31 | a8924c2 | per-skill LICENSE.txt | README、Codex skill 结构 | absorb：按需加载、skill 自包含、来源许可证登记 |
| anthropics-skills | https://github.com/anthropics/skills | 2026-05-31 | da20c92 | Apache-2.0 + source-available 混合 | README、template、skill 结构 | absorb：metadata、清晰 description、自包含与测试提醒 |
| addyosmani-agent-skills | https://github.com/addyosmani/agent-skills | 2026-05-31 | 6ce0298 | MIT | using-agent-skills、工程阶段、review/TDD/security | 已入第四师：阶段路由、验证门、五轴审查 |
| SkillClaw | https://github.com/AMAP-ML/SkillClaw | 2026-05-31 | 1f96ec8 | MIT | README、进化服务、验证候选、dashboard 线索 | absorb：任务后进化、候选验证、去重、版本轨迹 |
| Edict | https://github.com/cft0808/edict | 2026-05-31 | 14a2075 | MIT | README、agents.json、状态/审核/看板结构 | absorb：强制封驳、状态审计、可视化追踪；不吸收官制命名 |
| vercel-agent-skills | https://github.com/vercel-labs/agent-skills | 2026-05-31 | 1801156 | 待逐 skill 核验 | AGENTS、README、测试脚本线索 | absorb：500 行以内、触发描述、按需 reference；具体技能先 defer |

## 待核验来源

这些来源只能作为后续线索，不能宣称已融合：

- PPT/HTML 相关社区 skill 的最新版逐仓库源码。
- 写作、短剧、卖点提炼相关社区 skill 的官方源。
- OpenDesign / 设计类仓库的可复用机制和授权边界。

## 与进化部协作

进化部负责发现“需要升级”的信号，蒸馏师团负责判定“能不能吸收、怎么吸收”。

```text
失败模式 / 用户反馈 / 新来源
-> 进化部归类
-> 蒸馏师团核验与裁决
-> 试验场验证
-> 白帽放行
-> 对应师团更新
```
