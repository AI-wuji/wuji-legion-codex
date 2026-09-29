# 来源到单一专家方法的归属

运行时权威目录只有 `capabilities/experts/manifest.json`。每个专家的 `methods` 是从多个互补来源提炼后的具体任务方法；`sources` 是出处，不是要安装的包，也不授权外部 API、账号或写入。按任务词面唯一选中专家后，只注入最多两条匹配方法；未选中时这些内容保持冷态。

55 个共享入口逐项的融合方法或不准入理由见 [来源决议账](source-decisions.json)。课程模板与当前能力的组合规则见 [W5/W6 模板映射](w5-w6-template-map.md)；后续新增和升级走 [统一准入流程](evolution-admission.md)。入口审查记录见 [共享来源审查](shared-skill-audit-2026-09-27.md)，它不是完整包行为验证。

可调用原子与提炼方法必须区分：

- 现有 capability 的 primary/optional 来源继续按其生命周期、哈希和行为证据挂载，不因为共享入口有同名包就被覆盖。
- `skill_sources` 和 `methods.sources` 仅记录来源；真实可执行来源须先在所属 capability manifest 准入，再由专家 `source_ids` 精确引用。
- 文档排版不等于远端草稿，写文章不等于发布，分镜不等于视频，provider 名称不等于真实 API。每一步按当前宿主回执验收。
