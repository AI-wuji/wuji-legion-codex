# 无极军团 Codex 版 / Wuji Legion for Codex

**一句话 / One Sentence**  
面向 Codex 的军团式执行框架：阿极统一入口，参谋本部路由主帅，女娲按需补位，白帽前置封驳，把复杂任务收束成稳定交付。  
A legion-style execution framework for Codex: A-Ji as the unified interface, General Staff routing a single commander, Nuwa augmenting on demand, and White Hat vetoing bad paths early to turn complex work into reliable delivery.

**适配架构 / Architecture**  
Codex Desktop + AGENTS.md + Skills + Plugins + GitHub workflow

---

## 这是什么 / What Is This

无极军团 Codex 版不是单一 skill，也不是单条提示词。  
它是一套给 Codex 使用的轻量执行框架，用来统一调度：

- 日常快答
- 搜索调研
- 代码开发
- 内容生产
- PPT / HTML / 配图
- QA / 安全 / 复盘进化

Wuji Legion for Codex is not a single skill or a single prompt.  
It is a lightweight execution framework for Codex that coordinates quick replies, research, coding, content, PPT/HTML/visuals, QA, security, and evolution.

---

## 核心优势 / Core Advantages

| 特性 Feature | 中文 | English |
|---|---|---|
| 阿极入口 | 对外只保留一个人格入口 | Single user-facing interface |
| 单主帅机制 | 一个任务默认一个主帅负责到底 | One commander owns delivery |
| 白帽前置 | 开工前拦错，不等失败后补救 | Catch bad routes before execution |
| 轻量状态机 | 少空转、少乱查、少浪费 token | Less drift, less tool thrash, fewer wasted tokens |
| 多师局协同 | 按需补位，不为表演而组队 | Coordinated divisions without role theater |
| 成品导向 | 默认追最终交付，不拿半成品充数 | Final artifact delivery by default |

---

## 核心架构 / Core Architecture

```text
用户 User
    ↓
阿极 A-Ji（统一入口 / Unified Interface）
    ↓
参谋本部 General Staff（状态判断 + 主帅路由）
    ↓
主帅 Commander（负责到底 / Owns Delivery）
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
| 阿极 A-Ji | 统一入口、快答、短报 | 常驻 Resident |
| 参谋本部 General Staff | 状态机、路由、选主帅 | 常驻 Resident |
| 女娲 Nuwa | 能力融合、专家补位 | 按需 On-demand |
| 白帽 / 质监局 White Hat / QA | 前置反对、质检验收 | 强制 Mandatory |
| 第一师 Div.1 Content | 内容、文案、结构化表达 | 按需 On-demand |
| 第二师 Div.2 Visual | PPT、HTML、配图、视觉交付 | 按需 On-demand |
| 第四师 Div.4 Dev | 代码、自动化、工程质量 | 按需 On-demand |
| 情报局 Intel Bureau | 搜索调研、情报研判 | 情报先行 Intel First |
| 远征军 Expedition | 低成本外派、批处理 | 按需 On-demand |
| 安全局 Security Bureau | 安全审计、许可证、漏洞 | 强制 Mandatory |
| 进化部 Evolution Bureau | 复盘、进化、经验沉淀 | 按需 On-demand |
| 插件注册中心 Plugin Registry | 插件纳管与裁决 | 常驻 Resident |

---

## 完整工作流 / Complete Workflow

```text
User Request
    ↓
[A-Ji] Fast reply or task intake
    ↓
[General Staff] State decision + commander routing
    ↓
[Intel / Content / Visual / Dev] Main execution path
    ↓
[Nuwa] Expert augmentation only if needed
    ↓
[White Hat / QA] Preemptive check and acceptance
    ↓
[Security] Audit when applicable
    ↓
[Evolution] Capture lessons and update patterns
    ↓
Final Artifact + Path
```

---

## 军团模块 / Legion Modules

### 内容模块 / Content
- 文案融合引擎
- 标题、钩子、节奏、人味优化
- PPT / HTML 结构化内容输出

File:
- [content.md](E:\wuji-projects\wuji-legion-codex\units\content.md)

### 视觉模块 / Visual
- PPT / HTML / 配图生产线
- 臧老师美化链
- 预览、版式、风格统一

File:
- [visual.md](E:\wuji-projects\wuji-legion-codex\units\visual.md)

### 开发模块 / Development
- Rust / TS / Python
- 自动化与工程门禁

File:
- [dev.md](E:\wuji-projects\wuji-legion-codex\units\dev.md)

### 情报模块 / Intelligence
- 多引擎搜索
- GitHub / 社区 / 文档研判

File:
- [intel.md](E:\wuji-projects\wuji-legion-codex\units\intel.md)

### 安全模块 / Security
- 代码安全
- 依赖漏洞
- 许可证检查

File:
- [security.md](E:\wuji-projects\wuji-legion-codex\units\security.md)

### 进化模块 / Evolution
- OODA 复盘
- 失败模式沉淀

File:
- [auto_evolve.md](E:\wuji-projects\wuji-legion-codex\units\auto_evolve.md)

---

## 特色模块：Presentation / Featured Module: Presentation

`presentation` 是无极军团内部的演示与视觉产出模块，不代表整个军团。  
The `presentation` module handles presentation and visual delivery, but it does not represent the whole legion.

负责范围 / Scope:
- 真 PPTX / True editable PPTX
- HTML 演示稿 / HTML slide decks
- 插图、封面、配图 / visuals, covers, illustrations

入口文件 / Entry Files:
- [pptx_master.md](E:\wuji-projects\wuji-legion-codex\units\pptx_master.md)
- [html_slides_master.md](E:\wuji-projects\wuji-legion-codex\units\html_slides_master.md)
- [quick-imagegen.ps1](E:\wuji-projects\wuji-legion-codex\scripts\quick-imagegen.ps1)

---

## 专家库 / Expert Pool

当前专家域 / Current Expert Domains:

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

代表性专家 / Representative Experts:

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

## 适用场景 / Best Fit

适合 / Good Fit:
- 想让 Codex 长期稳定工作
- 想减少空转分析和 token 浪费
- 想把代码、内容、调研、PPT、HTML 放进同一套框架
- 想保留角色体系，但不想角色表演

不适合 / Not Ideal:
- 只想要单一极简提示词
- 不需要组织层和质量门禁

---

## 快速开始 / Quick Start

### 安装 / Installation

```powershell
Copy-Item .\GLOBAL_AGENTS.md C:\Users\Administrator\.codex\AGENTS.md -Force
Copy-Item .\SKILL.md C:\Users\Administrator\.agents\skills\wuji-legion\SKILL.md -Force
Copy-Item .\config.json C:\Users\Administrator\.agents\skills\wuji-legion\config.json -Force
Copy-Item .\units C:\Users\Administrator\.agents\skills\wuji-legion\units -Recurse -Force
Copy-Item .\experts C:\Users\Administrator\.agents\skills\wuji-legion\experts -Recurse -Force
```

### 使用 / Usage

1. 安装 `AGENTS.md` 和 skill 文件  
   Install the `AGENTS.md` and skill files
2. 在 Codex 中直接下达任务  
   Give tasks directly in Codex
3. 需要完整调度时，明确说“激活无极军团”  
   Say "激活无极军团" when you want full legion routing

---

## 文件结构 / File Structure

```text
wuji-legion-codex/
├── GLOBAL_AGENTS.md
├── SKILL.md
├── CHANGELOG.md
├── scripts/
│   └── quick-imagegen.ps1
├── units/
│   ├── content.md
│   ├── visual.md
│   ├── dev.md
│   ├── intel.md
│   ├── security.md
│   ├── auto_evolve.md
│   ├── plugins.md
│   ├── pptx_master.md
│   └── html_slides_master.md
└── experts/
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
- [auto_evolve.md](E:\wuji-projects\wuji-legion-codex\units\auto_evolve.md)
- [plugins.md](E:\wuji-projects\wuji-legion-codex\units\plugins.md)
- [pptx_master.md](E:\wuji-projects\wuji-legion-codex\units\pptx_master.md)
- [html_slides_master.md](E:\wuji-projects\wuji-legion-codex\units\html_slides_master.md)
- [quick-imagegen.ps1](E:\wuji-projects\wuji-legion-codex\scripts\quick-imagegen.ps1)

---

## 版本历史 / Version History

| Version | Date | Major Update |
|---|---|---|
| `v9.1` | 2026-05-31 | README 重排为产品化双语结构，明确 presentation 只是内部模块 |
| `v9.0` | 2026-05-30 | 军团主干整流，收紧状态机、单主帅路由、成品交付 |
| `v8.0` | 2026-05-30 | presentation 子体系蒸馏与模块边界纠偏 |

---

## 更新日志 / Changelog

- `2026-05-31 v9.1`
  - README 改为参考产品化 skill 的介绍结构
  - 保留无极军团原有命名、编制和模块体系
  - 更新日志放到文末
  - 保持双语介绍
- `2026-05-30 v9.0`
  - 完成军团主干整流
  - 收紧状态机、单主帅路由、成品交付规则
- 详细记录见：[CHANGELOG.md](E:\wuji-projects\wuji-legion-codex\CHANGELOG.md)
