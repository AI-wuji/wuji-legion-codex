# 无极军团 Codex 版 / Wuji Legion for Codex

**一句话 / One Sentence**  
让 Codex 能稳定处理调研、代码、内容、PPT、HTML 与配图任务的一体化无极军团执行框架。  
An integrated Wuji Legion execution framework that enables Codex to reliably handle research, coding, content, PPT, HTML, and visual tasks.

---

## 它是干什么的 / What It Does

无极军团 Codex 版，本质上是把 Codex 常见的几类任务收进一套统一执行系统：

- 搜索调研
- 代码开发
- 内容生产
- PPT / HTML / 配图
- QA / 安全 / 复盘

它解决的不是“会不会答”，而是“能不能稳定交付最终结果”。

Wuji Legion for Codex is a unified execution system for Codex tasks such as research, coding, content, PPT/HTML/visual work, QA, security, and iteration review.  
Its real goal is not just answering, but reliably delivering final outputs.

---

## 借鉴了什么优点 / What It Distills

无极军团不是简单拼装，而是把开源 skill 和社区工作流里真正有效的部分拆出来，重新整流进自己的体系。

重点借鉴：

- `ppt-master`：真 PPTX 导出、分阶段执行
- `elite-powerpoint-designer`：高质量审美、重度美化
- `presentation-skill`：工作区化、QA 化流程
- `academic-pptx-skill`：一页一结论、信息压缩
- `guizang-ppt-skill`：风格收束
- `huashu-design`：先 brief 再设计
- `open-design`：设计系统视角、反 AI slop
- `frontend-slides` / `frontend-slides-editable` / `html-ppt-skill`：HTML deck 的预览、编辑、主题与布局
- `imagegen`：最短生图链路

无极军团自己的整流动作：

- 阿极统一入口
- 参谋本部单主帅路由
- 女娲按需补位
- 白帽前置封驳
- 状态机控制 token 和执行节奏

Wuji Legion distills the strongest parts of open-source skills and community workflows, then reorganizes them into its own system: unified A-Ji entry, single-commander routing, Nuwa augmentation, White Hat pre-checks, and state-machine control over execution and token use.

---

## 现在实现了什么 / What It Enables

当前已经落成的核心能力：

- 快答层：轻问题直接短答
- 调研层：GitHub / 社区 / 文档搜索与汇总
- 代码层：代码修改、自动化、质量门禁
- 内容层：文案结构化、标题钩子、人味优化
- Presentation 层：真 PPTX、HTML slides、插图与封面
- 质量层：白帽质疑、质监验收、安全审计
- 进化层：失败模式沉淀、规则持续整流

In practical terms, it already enables quick replies, research synthesis, coding workflows, structured content production, PPT/HTML/image delivery, QA/security gates, and ongoing system evolution.

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

---

## 融合来源 / Open-source Inspirations

为避免误会，这里明确说明：无极军团 Codex 版参考、蒸馏、整流了多条开源 skill / 工作流，但保留了自己的命名、状态机、路由规则和组织体系。

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
- [quick-imagegen.ps1](E:\wuji-projects\wuji-legion-codex\scripts\quick-imagegen.ps1)

---

## 更新日志 / Changelog

- `2026-05-31 v9.1`
  - README 收缩为核心价值优先版
  - 重点讲清：它是干什么的、借鉴了什么、实现了什么
- 详细记录见：[CHANGELOG.md](E:\wuji-projects\wuji-legion-codex\CHANGELOG.md)
