# 更新日志 / Changelog

## 2026-05-24 v4.0 — MoE并行中枢 + 全融合版

### 重大变更
- **架构重构**: 从串行决策树 → MoE(混合专家)并行评估中枢
- **参谋本部升级**: MoE Gating门控+并行拆解+加权汇总路由
- **女娲升级**: 多Agent并行调度引擎+动态组队算法+12种标准组队方案

### 新增
- **硬性铁律**: 6条全局常驻铁律（实事求是/红队义务/先结论后原因/交付必报/诚实透明/自动备份）
- **Reasonix缓存引擎**: ImmutablePrefix + AppendOnlyLog + VolatileScratch + Auto-Compact
- **工具调用修复**: flatten/scavenge/truncation/storm 四种修复策略
- **红队审计**: 每次任务强制反对意见+前提质疑+风险识别+一票否决权
- **跨部门并行协作协议**: 14个部门的依赖关系+并行执行规则
- **统一管理调度**: 所有skill/MCP统一纳管，新增需过五关
- **Codex官方技能融合**: imagegen/openai-docs/plugin-creator/skill-creator/skill-installer/browser/latex
- **提示词局**: 通用/图像/故事板/视频 prompt 工程中心
- **进化部**: OODA自动进化循环+失败模式库+经验积累

### 补实(5个虚文件重写)
- **情报局**: 18行→70行，多引擎并行搜索+可信度评分+具体API命令
- **安全局**: 18行→62行，L1-L5审计+许可证矩阵+红线
- **远征军**: 19行→87行，并行外派协议+Spec/Handoff模板
- **试验场**: 15行→66行，沙箱化测试+5维评估+融合决策树
- **档案局**: 23行→78行，完整备份/回滚/崩溃恢复/保留策略

### 缓存优化
- 预期缓存命中率 >95%（Reasonix实测 99.82%）
- 预期成本降低 80%（$60/天 → $12/天，435M tokens场景）
- 全局AGENTS.md写入铁律+Cache优化原则

### 人才库
- Perspective技能: 17→31个（新增14个，从女娲例子库正式安装）
- 功能skill: 28个
- 总可用技能: 59个
- 所有新建模视图已安装到 .agents/skills/

### 其他
- 所有unit文件最低62行，最高271行，零虚文件
- config.json新增 cache_config/red_team_enabled/iron_rules_version
- 全局AGENTS.md同时包含铁律+Cache优化
