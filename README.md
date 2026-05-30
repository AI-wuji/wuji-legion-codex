# 无极军团 Codex 版 / Wuji Legion for Codex

**一句话 / One Sentence**:
面向 Codex 的轻量执行军团，用单主帅调度、白帽前置和多师局协同，把复杂任务收束成稳定交付。 / A lightweight execution legion for Codex that turns complex work into reliable delivery through single-commander routing, preemptive white-hat checks, and coordinated divisions.

**适配架构 / Architecture**:
Codex Desktop + AGENTS.md + Skills + Plugins + GitHub workflow / Codex Desktop + AGENTS.md + Skills + Plugins + GitHub workflow

---

## 这是什么？ / What Is This?

**无极军团 Codex 版 / Wuji Legion for Codex** 不是单一 skill，也不是单一提示词。

它是一套给 Codex 使用的轻量执行框架，把快答、调研、代码、内容、PPT、HTML、配图、QA、安全和复盘，统一纳入同一套可调度、可纠偏、可交付的体系。

**Wuji Legion for Codex** is not a single skill or a single prompt.

It is a lightweight execution framework for Codex that brings quick replies, research, coding, content, PPT, HTML, visuals, QA, security, and iteration review into one coordinated system.

### 核心目标 / Core Goals

| 中文 | English |
|---|---|
| 省 token，高命中 | Save tokens, maximize hit rate |
| 高质高效 | High quality, high efficiency |
| 少走弯路 | Reduce wasted motion |
| 直接出最终结果 | Deliver final results directly |

---

## 它主要解决什么问题？ / What Problems Does It Solve?

| 痛点 | Problem |
|---|---|
| 普通 agent 容易长篇分析、空转查环境、浪费 token | Ordinary agents often overthink, over-inspect the environment, and waste tokens |
| 多 skill 混用时容易角色打架、流程冲突 | Multiple skills often conflict with each other and create workflow drift |
| 复杂任务经常没人负责到底 | Complex tasks often lack a true end-to-end owner |
| 成品任务容易被做成“解释任务” | Delivery tasks often degrade into explanation-only tasks |
| 规则持续打补丁后会前后冲突 | Patch-stacked rules eventually conflict and reduce reliability |

无极军团 Codex 版通过阿极统一入口、参谋本部单主帅路由、女娲按需补位、白帽前置封驳来解决这些问题。

Wuji Legion for Codex solves these issues through A-Ji as the unified interface, General Staff single-commander routing, Nuwa on-demand augmentation, and preemptive white-hat veto.

---

## 核心架构 / Core Architecture

```text
用户 User
    ↓
阿极 A-Ji（统一入口 / Unified Interface）
    ↓
参谋本部 General Staff（状态判断 + 主帅路由 / State Routing + Commander Selection）
    ↓
主帅 Commander（负责到底 / Owns Final Delivery）
    ↓
女娲 Nuwa（按需补位 / On-demand Augmentation）
    ↓
白帽 + 质监局 White Hat + QA（前置封驳 / Quality Gate）
    ↓
各师局 / 模块 / 专家库 Divisions / Modules / Expert Pool
    ↓
最终交付 Final Delivery
```

---

## 编制总览 / Organization Overview

| 单位 Unit | 职责 Role | 状态 Status |
|---|---|---|
| 阿极 A-Ji | 统一入口、快答、短报 / Unified interface, quick replies, concise reporting | 常驻 Resident |
| 参谋本部 General Staff | 状态机、路由、选主帅 / State machine, routing, commander selection | 常驻 Resident |
| 女娲 Nuwa | 能力融合、专家补位 / Capability fusion, expert augmentation | 按需 On-demand |
| 白帽 / 质监局 White Hat / QA | 前置反对、质检验收 / Preemptive challenge, QA review | 强制 Mandatory |
| 第一师 Div.1 Content | 内容、文案、结构化表达 / Content, writing, structural output | 按需 On-demand |
| 第二师 Div.2 Visual | PPT、HTML、配图、视觉交付 / PPT, HTML, visuals, delivery | 按需 On-demand |
| 第四师 Div.4 Dev | 代码、自动化、工程质量 / Code, automation, engineering quality | 按需 On-demand |
| 情报局 Intel Bureau | 搜索调研、情报研判 / Search, research, intelligence synthesis | 情报先行 Intel First |
| 远征军 Expedition | 低成本外派、批处理 / Low-cost delegation, batch work | 按需 On-demand |
| 安全局 Security Bureau | 安全审计、许可证与漏洞 / Security, licenses, vulnerabilities | 强制 Mandatory |
| 进化部 Evolution Bureau | 复盘、进化、经验沉淀 / Review, evolution, memory | 按需 On-demand |
| 插件注册中心 Plugin Registry | 插件纳管与裁决 / Plugin governance and routing | 常驻 Resident |

---

## 运行机制 / Execution Model

### 轻量状态机 / Lightweight State Machine

仅允许少数清晰状态：`FAST_REPLY`、`CLARIFY`、`SINGLE_COMMANDER`、`LEGION_TASK`、`BLOCKED`、`DONE`。

Only a small set of explicit states is allowed: `FAST_REPLY`, `CLARIFY`, `SINGLE_COMMANDER`, `LEGION_TASK`, `BLOCKED`, and `DONE`.

### 单主帅负责到底 / Single Commander Ownership

默认由参谋本部只选一个主帅负责到底，其他能力只做补位，不做轮流接管。

By default, General Staff appoints one commander to own the task end-to-end. Other capabilities only augment; they do not take turns hijacking delivery.

### 白帽前置 / Preemptive White-Hat Gate

白帽负责在开工前拦截错误路线，而不是等成品出来再事后找补。

White Hat blocks bad routes before execution rather than patching problems after delivery fails.

---

## 各师局与组织 / Divisions and Bureaus

### 第一师（内容） / Division 1 Content

- 文案融合引擎 / Writing fusion engine
- PPT / HTML 内容结构化 / Structured content for PPT and HTML
- 标题、钩子、节奏、人味优化 / Hooks, rhythm, humanization

文件 / File:
- [content.md](E:\wuji-projects\wuji-legion-codex\units\content.md)

### 第二师（视觉） / Division 2 Visual

- PPT / HTML / 配图生产线 / PPT, HTML, and visual production line
- 臧老师美化链 / high-end deck polish chain
- 预览、版式、风格统一 / preview, layout, style consistency

文件 / File:
- [visual.md](E:\wuji-projects\wuji-legion-codex\units\visual.md)

### 第四师（开发） / Division 4 Development

- Rust / TS / Python 开发 / Rust, TS, Python development
- 自动化与工程质量门禁 / Automation and engineering quality gates

文件 / File:
- [dev.md](E:\wuji-projects\wuji-legion-codex\units\dev.md)

### 情报局 / Intel Bureau

- 多引擎并行搜索 / Multi-engine search
- GitHub / 社区 / 文档研判 / GitHub, community, docs analysis

文件 / File:
- [intel.md](E:\wuji-projects\wuji-legion-codex\units\intel.md)

### 远征军调度室 / Expedition Office

- 低成本外派 / Low-cost delegation
- 批处理与 handoff / Batch processing and handoff

文件 / File:
- [expedition.md](E:\wuji-projects\wuji-legion-codex\units\expedition.md)

### 安全局 / Security Bureau

- 代码安全 / Code security
- 依赖漏洞 / Dependency vulnerabilities
- 许可证检查 / License checks

文件 / File:
- [security.md](E:\wuji-projects\wuji-legion-codex\units\security.md)

### 进化部 / Evolution Bureau

- OODA 复盘 / OODA review loop
- 失败模式沉淀 / Failure pattern capture

文件 / File:
- [auto_evolve.md](E:\wuji-projects\wuji-legion-codex\units\auto_evolve.md)

### 插件注册中心 / Plugin Registry

- Browser / Documents / Spreadsheets / Presentations 纳管
- PPT / HTML 候选工具裁决 / governance of PPT and HTML candidate tools

文件 / File:
- [plugins.md](E:\wuji-projects\wuji-legion-codex\units\plugins.md)

---

## 专家库 / Expert Pool

无极军团 Codex 版不只靠组织，还挂有专家库供女娲按需调度。

Wuji Legion for Codex relies not only on divisions, but also on an expert pool that Nuwa can call on demand.

当前专家目录 / Current expert domains:

- `content`
- `visual`
- `dev`
- `intel`
- `qa`
- `security`
- `prompt`
- `staff`
- `archive`
- `comfyui`
- `evolve`
- `expedition`
- `proving`

代表性专家 / Representative experts:

- 臧老师(PPT)
- MrBeast
- Paul Graham
- Steve Jobs
- John Carmack
- Ilya Sutskever
- UX Architect
- Security Engineer
- Trend Researcher

---

## 模块体系 / Module System

无极军团 Codex 版是总框架，不等于任何一个单模块。

Wuji Legion for Codex is the overall framework, not any single module.

当前已重点定型的内部模块之一是：

One of the currently distilled internal modules is:

- `presentation`

---

## Presentation 模块 / Presentation Module

`presentation` 只负责演示与视觉产出，不代表整个无极军团 Codex 版。

The `presentation` module is only responsible for presentation and visual output. It does not represent the whole Wuji Legion for Codex.

负责范围 / Scope:

- 真 PPTX / True editable PPTX
- HTML 演示稿 / HTML slide decks
- 配图、封面、插图 / illustrations, covers, supporting visuals

入口文件 / Entry files:

- [pptx_master.md](E:\wuji-projects\wuji-legion-codex\units\pptx_master.md)
- [html_slides_master.md](E:\wuji-projects\wuji-legion-codex\units\html_slides_master.md)
- [quick-imagegen.ps1](E:\wuji-projects\wuji-legion-codex\scripts\quick-imagegen.ps1)

---

## 典型能力地图 / Capability Map

| 需求 Need | 主链 Primary Path |
|---|---|
| 普通问答 General replies | 阿极 A-Ji |
| 搜索调研 Research | 情报局 Intel Bureau |
| 代码开发 Coding | 第四师 Div.4 Dev |
| 真 PPTX 续写 PPTX continuation | `presentation` -> `pptx_master` |
| 从零做 PPT PPT from scratch | 第二师 Visual + `pptx_master` |
| HTML 演示稿 HTML slides | `presentation` -> `html_slides_master` |
| 生图/插图/封面 Images and covers | `presentation` -> `quick-imagegen.ps1` |
| 质量验收 QA | 白帽 / 质监局 White Hat / QA |
| 安全检查 Security | 安全局 Security Bureau |
| 复盘进化 Evolution | 进化部 Evolution Bureau |

---

## 与普通 Skill 的区别 / How It Differs from Ordinary Skills

普通 skill 更像单用途工具。

Ordinary skills behave more like single-purpose tools.

无极军团 Codex 版更像产品化执行框架，强调：

Wuji Legion for Codex behaves more like a productized execution framework, emphasizing:

- 统一入口 / unified interface
- 清晰状态机 / explicit state machine
- 单主帅到底 / single-owner delivery
- 白帽前置 / preemptive white-hat gate
- 多组织协同 / coordinated divisions
- 最终成品交付 / final artifact delivery

---

## 适用场景 / Best Fit

适合 / Good fit:

- 想让 Codex 长期稳定工作的人 / people who want Codex to work reliably over time
- 想减少空转分析和 token 浪费的人 / people who want less wasted analysis and token burn
- 想把 PPT、HTML、代码、内容、调研统一纳入一套框架的人 / people who want one framework for PPT, HTML, code, content, and research

不适合 / Not ideal:

- 只想要单一极简提示词的人 / people who only want a single minimal prompt
- 不需要组织层和质量门禁的人 / people who do not want organizational layers or quality gates

---

## 安装 / Installation

```powershell
Copy-Item .\GLOBAL_AGENTS.md C:\Users\Administrator\.codex\AGENTS.md -Force
Copy-Item .\SKILL.md C:\Users\Administrator\.agents\skills\wuji-legion\SKILL.md -Force
Copy-Item .\config.json C:\Users\Administrator\.agents\skills\wuji-legion\config.json -Force
Copy-Item .\units C:\Users\Administrator\.agents\skills\wuji-legion\units -Recurse -Force
Copy-Item .\experts C:\Users\Administrator\.agents\skills\wuji-legion\experts -Recurse -Force
```

---

## 关键文件 / Key Files

- [GLOBAL_AGENTS.md](E:\wuji-projects\wuji-legion-codex\GLOBAL_AGENTS.md)
- [SKILL.md](E:\wuji-projects\wuji-legion-codex\SKILL.md)
- [content.md](E:\wuji-projects\wuji-legion-codex\units\content.md)
- [visual.md](E:\wuji-projects\wuji-legion-codex\units\visual.md)
- [dev.md](E:\wuji-projects\wuji-legion-codex\units\dev.md)
- [intel.md](E:\wuji-projects\wuji-legion-codex\units\intel.md)
- [security.md](E:\wuji-projects\wuji-legion-codex\units\security.md)
- [expedition.md](E:\wuji-projects\wuji-legion-codex\units\expedition.md)
- [auto_evolve.md](E:\wuji-projects\wuji-legion-codex\units\auto_evolve.md)
- [plugins.md](E:\wuji-projects\wuji-legion-codex\units\plugins.md)
- [pptx_master.md](E:\wuji-projects\wuji-legion-codex\units\pptx_master.md)
- [html_slides_master.md](E:\wuji-projects\wuji-legion-codex\units\html_slides_master.md)
- [quick-imagegen.ps1](E:\wuji-projects\wuji-legion-codex\scripts\quick-imagegen.ps1)

---

## 更新日志 / Changelog

- `2026-05-31 v9.1`
  - 首页恢复双语商品介绍结构
  - 更新日志移动到文末
  - 明确 `presentation` 只是内部模块，不代替整个军团
- `2026-05-30 v9.0`
  - 完成军团主干整流
  - 收紧状态机、单主帅路由、成品交付规则
- 详细记录见：[CHANGELOG.md](E:\wuji-projects\wuji-legion-codex\CHANGELOG.md)
