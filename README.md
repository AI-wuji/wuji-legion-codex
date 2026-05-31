# 无极军团 Codex 版 / Wuji Legion for Codex

**一句话 / One Sentence**  
通过 MoE 总调度统一编排多方向 skill 与多 agent 协同工作的无极军团执行框架，把 Codex 从“会回答”推进到“能交付”。  
A Wuji Legion execution framework that uses MoE-style orchestration to coordinate multi-domain skills and multi-agent collaboration, pushing Codex from “can answer” to “can deliver”.

---

## 它是干什么的 / What It Does

无极军团 Codex 版，干的不是“再装几个 skill”这件事。

它真正做的是三件事：

1. 总调度，把不同方向的 skill、agent、插件和专家统一编排进同一套执行链  
2. 协同工作，让调研、代码、内容、PPT、HTML、配图这些能力不再各自为战  
3. 逼近交付，把“给建议”推进成“给成品、给文件、给路径”

所以它解决的核心问题，不是“会不会答”，而是“能不能稳定交付”。

Wuji Legion for Codex is not about stacking more skills. It is about orchestrating many skills and agents under one command system, making them collaborate instead of fragment, and pushing Codex toward reliable delivery instead of endless explanation.

---

## 借鉴了什么优点 / What It Distills

无极军团不是简单拼装，而是把开源 skill 里真正有用的“长处”拆出来，再按自己的体系重新整流。

它借走的，不是名字，是能力：

- `ppt-master`：真 PPTX、分阶段执行、结构先行
- `elite-powerpoint-designer`：审美上限、重度美化、成品感
- `presentation-skill`：工作区、QA、证据化流程
- `academic-pptx-skill`：一页一结论、信息压缩
- `guizang-ppt-skill`：风格收束，不让页面失控
- `huashu-design` / `open-design`：先定 brief、后做设计，减少 AI slop
- `frontend-slides` / `frontend-slides-editable` / `html-ppt-skill`：HTML deck 的预览、编辑和布局组织
- `imagegen`：最短出图链路
- `addyosmani/agent-skills`：工程生命周期、meta-skill 路由、薄切片实现、TDD、五轴 review、安全门禁

无极军团自己的增量，是把这些能力重新装进自己的执行体系：

- 阿极统一入口
- 参谋本部单主帅路由
- 女娲按需补位
- 白帽前置封驳
- 进化部蒸馏师团负责官方源核验、必要性判断和版本登记
- 状态机控制 token 和执行节奏，不让系统失控

Wuji Legion borrows strengths, not identities. It pulls out what actually works, then rebuilds those strengths inside its own execution system.

---

## 现在实现了什么 / What It Enables

现在已经真正落地的能力，不是抽象概念，而是这些能直接用的东西：

- 快答层：轻问题直接短答，不先读半天环境
- 调研层：GitHub / 社区 / 文档搜索、汇总、提炼结论
- 代码层：代码修改、自动化、质量门禁
- 内容层：文案结构化、标题钩子、人味优化
- Presentation 层：真 PPTX、HTML slides、插图、封面、配图
- 工程层：DEFINE / PLAN / BUILD / VERIFY / REVIEW / SHIP 阶段判断、薄切片实现、TDD、五轴审查
- 质量层：白帽质疑、质监验收、安全审计
- 进化层：失败模式沉淀、官方源核验、蒸馏裁决、规则持续整流

换句话说，它已经不是“概念军团”，而是一套能直接干活的执行框架。

In practice, it already delivers quick replies, research synthesis, coding workflows, structured content, PPT/HTML/image output, QA/security gates, and ongoing iteration.

---

## 组织与模块 / Organization and Modules

军团核心编制：

- 阿极
- 参谋本部
- 女娲
- 白帽 / 质监局
- 第一师内容
- 第二师视觉
- 第四师开发
- 情报局 / 安全局 / 进化部 / 远征军 / 插件注册中心

重点内部模块：

- `presentation`
- `content`
- `visual`
- `dev`
- `intel`
- `security`
- `distillation`

---

## 融合来源 / Open-source Inspirations

为避免误会，这里明确说明：无极军团 Codex 版参考、蒸馏、整流了多条开源 skill / 工作流，但没有把上游项目原样搬运进来。

它保留的是自己的命名、状态机、路由规则和组织体系。

Key inspirations include:

- `ppt-master`
- `elite-powerpoint-designer`
- `presentation-skill`
- `academic-pptx-skill`
- `guizang-ppt-skill`
- `huashu-design`
- `open-design`
- `frontend-slides`
- `frontend-slides-editable`
- `html-ppt-skill`
- `imagegen`
- `addyosmani/agent-skills`
- `openai/skills`
- `anthropics/skills`
- `AMAP-ML/SkillClaw`
- `cft0808/edict`
- `vercel-labs/agent-skills`

来源使用原则：

- 官方源优先，社区文章只作为线索。
- 记录 checked commit / version / license。
- 只吸收机制，不复制上游组织编制。
- 未完成核验的来源只进入待核验台账，不进入默认执行链。

---

## 快速开始 / Quick Start

```powershell
Copy-Item .\GLOBAL_AGENTS.md C:\Users\Administrator\.codex\AGENTS.md -Force
Copy-Item .\SKILL.md C:\Users\Administrator\.agents\skills\wuji-legion\SKILL.md -Force
Copy-Item .\config.json C:\Users\Administrator\.agents\skills\wuji-legion\config.json -Force
Copy-Item .\units C:\Users\Administrator\.agents\skills\wuji-legion\units -Recurse -Force
Copy-Item .\experts C:\Users\Administrator\.agents\skills\wuji-legion\experts -Recurse -Force
```

关键文件：

- [GLOBAL_AGENTS.md](E:\wuji-projects\wuji-legion-codex\GLOBAL_AGENTS.md)
- [SKILL.md](E:\wuji-projects\wuji-legion-codex\SKILL.md)
- [pptx_master.md](E:\wuji-projects\wuji-legion-codex\units\pptx_master.md)
- [html_slides_master.md](E:\wuji-projects\wuji-legion-codex\units\html_slides_master.md)
- [distillation.md](E:\wuji-projects\wuji-legion-codex\units\distillation.md)
- [quick-imagegen.ps1](E:\wuji-projects\wuji-legion-codex\scripts\quick-imagegen.ps1)

---

## 更新日志 / Changelog

- `2026-05-31 v9.3`
  - 新增进化部蒸馏师团
  - 增加官方源核验、必要性裁决、试验场验证和白帽退回红线
  - 已核验 OpenAI skills、Anthropic skills、Addy agent-skills、SkillClaw、Edict、Vercel agent-skills 当前源码/规则
- `2026-05-31 v9.1`
  - README 收缩为核心价值优先版
  - 重点讲清：它是干什么的、借鉴了什么、实现了什么
- 详细记录见：[CHANGELOG.md](E:\wuji-projects\wuji-legion-codex\CHANGELOG.md)
