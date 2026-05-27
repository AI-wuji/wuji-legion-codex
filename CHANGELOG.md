# 更新日志 / Changelog

## 2026-05-27 v5.7 — MoE按需门控 + PPT/HTML/Image2 + 质量门禁整合

### PPT/HTML/Image2增强
- **PPT生产线升级**: GPT内容结构 → 臧老师顶层设计 → image2配图 → pptx-generation生成 → slide-studio微调。
- **HTML生产线升级**: GPT信息架构 → impeccable审美规则 → 组件/样式系统 → Browser验证。
- **图像提示词增强**: 新增 `image-spec.json`，由GPT先理解页面意图再扩写给 image2/imagegen。
- **外部候选裁决**: PptxGenJS/Marp/reveal.js/shadcn 按条件保留；html2pptx/daisyUI 丢弃默认接入，不再长期悬空。
- **软件质量门禁**: Rust/Tauri、HTML/前端、Python脚本、PowerShell脚本、ComfyUI插件按技术栈自动跑门禁。
- **UI硬验收**: 软件/网页必须经过视觉层级、响应式、Browser预览、反模板味检查。
- **ComfyUI插件硬验收**: compileall、import smoke、节点注册、类型契约、路径安全成为默认验收项。
- **质量门禁脚本**: 新增 `scripts/wuji-quality-gate.ps1`，统一执行非破坏性质量检查。

### 修正
- **专家计数统一**: 当前 `experts/` 实际为69位专家、13个专家目录，清理历史错误计数残留。
- **白帽纠察分级**: 简单任务不再强制完整纠察，复杂/高风险/改文件任务仍强执行。
- **插件纳管写实**: 区分“已纳管路由”和“已安装/已授权”，避免假装插件可用。
- **Codex内置插件融合**: browser/documents/spreadsheets/presentations 显式接入路由。

### 优化
- **MoE轻量门控**: 项目启动不默认跑完整无极军团，只加载命中的部门、专家和插件。
- **阿极语义纠偏**: “阿极/启动无极军团”只启动 MoE 参谋本部，不等于完整军团全量启动。
- **快答路由**: 普通沟通默认 1-3 句话，Paul Graham/费曼/Reality Checker 组成短答裁剪器。
- **成品输出统一**: 图像、PPT、文档类默认只保留预览 + “文件在……”两个入口。
- **配置纠偏**: `active_project` 对齐当前仓库，`iron_rules_version` 对齐当前总纲。

## 2026-05-24 v4.0 — MoE并行中枢 + 全融合版

### 重大变更
- **架构重构**: 从串行决策树 → MoE(混合专家)并行评估中枢。
- **参谋本部升级**: MoE Gating门控+并行拆解+加权汇总路由。
- **女娲升级**: 多Agent并行调度引擎+动态组队算法+标准组队方案。

### 新增
- **硬性铁律**: 6条全局常驻铁律（实事求是/白帽纠察/先结论后原因/交付必报/诚实透明/自动备份）。
- **Reasonix缓存引擎**: ImmutablePrefix + AppendOnlyLog + VolatileScratch + Auto-Compact。
- **工具调用修复**: flatten/scavenge/truncation/storm 四种修复策略。
- **跨部门并行协作协议**: 14个部门的依赖关系+并行执行规则。
- **统一管理调度**: 所有 skill/MCP 统一纳管，新增需过五关。
- **Codex官方能力融合**: imagegen/openai-docs/plugin-creator/skill-creator/skill-installer/browser/documents/spreadsheets/presentations。
- **提示词局**: 通用/图像/故事板/视频 prompt 工程中心。
- **进化部**: OODA自动进化循环+失败模式库+经验积累。

### 补实
- **情报局**: 多引擎并行搜索+可信度评分+具体API命令。
- **安全局**: L1-L5审计+许可证矩阵+红线。
- **远征军**: 并行外派协议+Spec/Handoff模板。
- **试验场**: 沙箱化测试+5维评估+融合决策树。
- **档案局**: 完整备份/回滚/崩溃恢复/保留策略。

### 人才库
- 当前专家文件总数: 69。
- 当前专家目录总数: 13。
- 所有专家由女娲统一索引和按需组队。
