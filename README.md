# 无极军团 / Wuji Legion v3.1

**一句话 / One Sentence**:
为 Codex AI 设计的完整组织协作框架，参谋部决策/情报局搜索/安全局护航/质监局审计/档案局归档 / Full-stack AI collaboration framework for Codex with 5 independent departments

**适配平台 / Platforms**:
Codex CLI + Codex Desktop 双端适配 / Dual compatible

[![GitHub](https://img.shields.io/badge/Codex-Skill-blueviolet)](https://github.com/AI-wuji/wuji-legion-codex)
[![Version](https://img.shields.io/badge/version-3.1-purple)]()
[![License](https://img.shields.io/badge/License-MIT-yellow)]()

---

## 🎯 这是什么？ / What is this?

⚔️ **无极军团 / Wuji Legion** 是一个为 **Codex CLI / Desktop** 设计的 AI 多部门协作框架。

将现代企业的组织架构引入 AI 协作：参谋部决策、情报局搜索、安全局护航、质监局审计、档案局归档，五大部门互相制衡。

A multi-department AI collaboration framework for Codex with checks-and-balances: Strategy Command, Intelligence Bureau, Security Bureau, Quality Assurance Bureau, and Archives Bureau.

### 核心优势 / Core Advantages

| 优势 / Advantage | 中文说明 / Chinese | English |
|-----------------|-------------------|---------|
| 🏛️ | 五大部门互相制衡 | 5 departments with checks & balances |
| 🤖 | 27位专家角色池 | 27 expert roles pool |
| 🔄 | 部门自动联动 | Auto cross-department linkage |
| 🛡️ | 情报一审+安全二审+质检验收 | 3-layer security: Intel+Security+QA |
| 💾 | C盘+E盘双盘互备 | Dual-disk backup (C: + E:) |
| 🎯 | 省Token运行策略 | Token-efficient runtime |
| 🧬 | 自我进化能力 | Self-evolution capability |
| 🔧 | Rust核心+多平台封装 | Rust core + cross-platform packaging |
| 🎨 | **新** baoyu漫画/PPT/信息图 | baoyu comic/slide-deck/infographic |

---

## 📦 安装 / Installation

### 方法一：GitHub安装（推荐）/ Method 1: GitHub Install (Recommended)

```bash
git clone https://github.com/AI-wuji/wuji-legion-codex.git ~/.agents/skills/wuji-legion
```

### 方法二：手动安装 / Method 2: Manual Install

下载/解压到 `~/.agents/skills/wuji-legion/`，确保目录结构：

Download and extract to `~/.agents/skills/wuji-legion/`, ensure this structure:

```
~/.agents/skills/wuji-legion/
├── SKILL.md              # 总纲+规则集 / Master rules
├── README.md             # 本文档 / This doc
├── donate.svg            # 打赏 / Donation QR
├── CHANGELOG.md          # 更新日志 / Changelog
├── units/                # 11个部门文件 / 11 unit files
│   ├── staff.md          # 参谋部 / General Staff
│   ├── nuwa.md           # 女娲(人事部) / Nuwa HR
│   ├── intel.md          # 情报局 / Intel Bureau
│   ├── security.md       # 安全局 / Security Bureau
│   ├── qa.md             # 质监局 / QA Bureau
│   ├── archive.md        # 档案局 / Archives Bureau
│   ├── content.md        # 第一师(内容) / 1st Div (Content)
│   ├── visual.md         # 第二师(视觉) / 2nd Div (Visual)
│   ├── comfyui.md        # 第三师(ComfyUI) / 3rd Div (ComfyUI)
│   ├── dev.md            # 第四师(开发) / 4th Div (Dev)
│   └── proving_ground.md # 打靶场 / Proving Ground
└── scripts/              # 9个脚本 / 9 scripts
    ├── wuji-backup.py    # 本地备份 / Local backup
    ├── wuji-e-sync.ps1   # E盘同步 / E-drive sync
    ├── wuji-e-backup.ps1 # E盘备份 / E-drive backup
    ├── wuji-restore.ps1  # 灾难恢复 / Disaster recovery
    ├── wuji-install.ps1  # 首次安装 / First install
    ├── errors_db.py      # 错误DNA库 / Error DNA DB
    ├── target_range.py   # 打靶范围 / Target range
    ├── beep.ps1          # 完成提示音 / Completion beep
    └── push-to-github.ps1 # GitHub推送 / GitHub push
```

### 激活方式 / Activation

对 Codex 说 **「阿极」** 或 **「无极军团」** 即可激活。

Say **"A-Ji"** or **"Wuji Legion"** to activate.

> ✅ **暗号验证 / Shibboleth**: 激活后出现「运筹帷幄之中，决胜千里之外」即确认生效 / After activation, if you see "A battle won by one who has calculated in the temple" the skill is truly loaded.

---

## 🏛️ 组织架构 / Organization

```
                   用户指令 / User Command
                         ↓
               ┌──────────────────────┐
               │    参谋本部            │
               │    General Staff      │
               │  · 战略决策 Strategy    │
               │  · 拆任务 Split tasks  │
               └──┬──────┬──────┬──────┘
                   │      │      │
        ┌──────────▼─┐ ┌──▼───┐ ┌▼──────────┐
        │ 情报局      │ │安全局 │ │ 质监局     │
        │ Intel      │ │Security│ │ QA        │
        │ · 搜索 Search│ │·加密  │ │·审计 Audit │
        │ · 研判 Eval │ │·封装  │ │·验收 Check │
        └────────────┘ └──────┘ └───────────┘
                   │
        ┌──────────▼────────────────────────┐
        │    女娲人事部 Nuwa HR              │
        │  匹配专家 → spawn并发派发           │
        │  Match experts → parallel spawn   │
        └──────────┬────────────────────────┘
                   │
    ┌──────────────┼──────────────┐
    │              │              │
 ┌──▼────┐   ┌────▼──────┐  ┌───▼───┐
 │ 第一师  │   │ 第二师    │  │ 第四师 │
 │ Div1   │   │ Div2      │  │ Div4  │
 │(内容)  │   │(视觉)     │  │(开发) │
 │Content│   │Visual     │  │ Dev   │
 └───────┘   └───────────┘  └───────┘
    │
 ┌──▼──────┐
 │ 第三师   │
 │ Div3    │
 │(ComfyUI)│
 └─────────┘

 ┌────────────────────────────────┐
 │  档案局 Archives Bureau        │
 │  · C盘开发+备份 C: dev+backup  │
 │  · E盘镜像 E: mirror sync      │
 │  · GitHub远程 GitHub remote    │
 └────────────────────────────────┘

 ┌────────────────────────────────┐
 │  打靶场 Proving Ground         │
 │  · 新工具隔离测试 Isolated test │
 │  · 兼容/冲突/性能验证 Validation│
 └────────────────────────────────┘
```

### 五大核心部门 / 5 Core Departments

| 部门 / Department | 职责 / Role | 独立性 / Independence |
|------------------|------------|----------------------|
| **参谋部 General Staff** | 战略决策+拆任务+接收质检结果判断 Strategy & split tasks | 含女娲人事官 + Nuwa HR |
| **情报局 Intel Bureau** | 搜索+研判+安全一审 Search, analyze, security L1 | 完全独立 Fully independent |
| **安全局 Security Bureau** | 加密+封装+二审 Encryption, packaging, L2 review | 完全独立 Fully independent |
| **质监局 QA Bureau** | 审计+错误DNA+死循环检测 Audit, error DNA, loop detection | 完全独立，结果报告参谋部 Independent, reports to Staff |
| **档案局 Archives Bureau** | 双盘备份+版本+回滚 Dual backup, version, rollback | 独立存储 Independent storage |

### 四大作战师团 / 4 Combat Divisions

| 师团 / Division | 职责 / Role | 技术栈 / Tech Stack |
|----------------|------------|-------------------|
| **第一师 Div.1** | 文案/PPT/演示 Copywriting, PPT, demos | Slidev / Reveal.js / Markdown |
| **第二师 Div.2** | UI设计/可视化 UI design, data viz | ECharts / 22种图表 22 chart types |
| **第三师 Div.3** | ComfyUI节点/插件 Nodes & plugins | Python入口+Rust核心 Python shim + Rust core |
| **第四师 Div.4** | 全栈开发+封装 Full-stack dev & packaging | Rust核心+TS前端+Tauri Rust core + TS frontend + Tauri |

### 27位专家角色 / 27 Expert Roles

| 领域 / Field | 角色 / Roles | 专长 / Expertise |
|------------|------------|-----------------|
| **情报/搜索 Intel** | Mitnick / Shimomura / Snowden / Bellard / Lamo / Swartz | 社会工程/OSINT/代码分析/Web抓取 Social eng, OSINT, code analysis, web scraping |
| **安全/攻防 Security** | Schneier / HD Moore / Hotz / 郭盛华 / 林勇 / Miller | 密码学/渗透/逆向/红客防御 Crypto, pentest, reverse eng, red-blue defense |
| **开发/架构 Dev** | Torvalds / Carmack / Thompson / Bellard / 张一鸣 / Musk | 内核/引擎/编译器/产品架构 Kernel, engines, compilers, product arch |
| **视觉/设计 Design** | Jobs / Paul Graham / Tufte | 极简美学/文案叙事/数据可视化 Aesthetics, copywriting, data viz |
| **质量/决策 QA** | 费曼 / 芒格 / 孙子 | 验证逻辑/决策偏误/战略情报 Verification, bias detection, strategy |

---

## ⚙️ 工作流 / Workflow

### 标准项目流程 / Standard Project Pipeline

```
[用户 / User] 指令 Command
    ↓
[参谋部 Staff] 分析需求 → 拆任务 → 定方向 → 错误DNA预检
               Analyze → Split → Style → Error DNA pre-check
    ↓ 需求单 / Requirement ticket
[女娲 HR] 匹配专家 → spawn并发派发
         Match experts → parallel dispatch
    ↓
[各专家 Experts] 独立并行执行 / Independent parallel execution
    ↓
[质监局 QA] 验收 Audit + 死循环检测 Loop detection
    ├─ ✅ 通过 Pass → 参谋部汇总 Staff summary → 档案局归档 Archives
    └─ ❌ 未通过 Fail → 重新派发策略调整 Retry with adjusted strategy
```

### 自我进化流程 / Self-Evolution Pipeline

```
新工具 / New tool discovered
    ↓
[情报局 Intel] 全网搜索+研判+安全一审 / Search + analyze + L1 security
    ↓
[安全局 Security] 安全二审 / L2 security review
    ↓
[参谋部 Staff] 融合分析+省token优化 / Integration analysis + token optimization
    ↓
[打靶场 Range] 隔离测试（兼容/冲突/性能）/ Isolated test (compat/conflict/perf)
    ↓
[质监局 QA] 二审 / L2 QA review
    ↓
[档案局 Archives] 备份+更新ERRORS.md / Backup + update error DB
    ↓
✅ 正式融合 / Formal integration
```

### 档案局备份原则 / Archives Backup Policy

| 原则 / Policy | 中文 | English |
|--------------|------|---------|
| 🖥️ **C盘 C: drive** | `C:\wuji-projects\{项目名}\` 日常开发+改动前备份 Daily dev + pre-change backup |
| 💾 **E盘 E: drive** | `E:\wuji-projects\{项目名}\` robocopy /MIR 镜像同步 Mirror sync |
| 📝 **改动前 Before change** | 备份到 `.wuji-backups/{日期}_{描述}/` Backup with date + description |
| 🗑️ **原文件删除 Src deleted** | 备份不动（除非用户要求）Backup stays (unless user requests) |
| ⚡ **系统重做 OS reinstall** | GitHub拉军团→E盘恢复所有项目 Pull legion from GitHub → restore from E: |

---

## 🔧 默认技术栈 / Default Tech Stack

| 层级 / Layer | 语言/框架 / Language/Framework |
|-------------|-------------------------------|
| 前端 Frontend | TypeScript + React / HTML |
| 桌面壳 Desktop shell | Tauri(Rust) → EXE/DMG/AppImage |
| 后端核心 Backend core | Rust 单二进制 Single binary |
| ComfyUI入口 Entry | Python (__init__.py, ≤10行 lines) |
| ComfyUI核心 Core | Rust (PyO3 → .pyd) |

### 封装规范 / Packaging Standards

| 平台 / Platform | 格式 / Format |
|----------------|--------------|
| Windows | .exe / .msi |
| macOS Intel | .app / .dmg |
| macOS Apple Silicon | .app / .dmg |
| Linux | .AppImage / .deb |

---

## ⚡ 运行时优化 / Runtime Optimization

| 策略 / Strategy | 中文说明 / Chinese | English |
|----------------|-------------------|---------|
| 🧠 **工作记忆 Working Memory** | 关键规则常驻context，不反复加载 Key rules stay in context |
| 🏷️ **名称引用 Name Ref** | 工具库按名称而非路径引用 Reference tools by name not path |
| 📦 **按需加载 On-demand Load** | 只加载当前任务需要的模块 Load only needed modules |
| ❌ **死循环检测 Loop Detection** | 同子任务相似度>85%立即终止 Same task >85% similarity = terminate |
| 🧬 **错误DNA预检 Error DNA** | 执行前查ERRORS.md，规避已知bug Pre-check known bugs from ERRORS.md |
| 🔄 **跨任务缓存 Cross-task Cache** | 同一会话内结果复用 Reuse results within session |

---

## ❓ FAQ / 常见问题

**问：和直接使用Codex有什么区别？/ What's the difference from plain Codex?**

答：普通Codex是「一个人」在工作，无极军团是一个「组织」在工作。情报部提前收集信息、安全部审查风险、质检部验收质量、档案部管理版本。

Plain Codex works alone; Wuji Legion works as an organization — Intel collects info beforehand, Security reviews risks, QA verifies quality, Archives manages versions.

**问：需要什么环境？/ What environment is needed?**

答：Codex CLI 或 Codex Desktop，以及基础的 Rust/Python/Node.js（取决于项目类型）。

Codex CLI or Codex Desktop, plus basic Rust/Python/Node.js (depends on project type).

**问：27位专家从哪来？/ Where do the 27 experts come from?**

答：女娲(huashu-nuwa)角色蒸馏系统创建的15位人物视角Skill + 情报/安全领域内置12位角色。

15 perspective skills from Nuwa distillation + 12 built-in intel/security roles.

---

## 📋 版本历史 / Version History

| 版本 Version | 日期 Date | 主要更新 Major Update |
|-------------|----------|---------------------|
| **v3.1** | **2026-05-19** | **参谋部预加载 + baoyu漫画/PPT/信息图融合 / Staff pre-load + baoyu comic/slide-deck/infographic** |
| **v3.0** | 2026-05-14 | 五大核心部门重构+女娲人事部+27专家+暗号验证+打靶场 / 5-department restructure + Nuwa HR + 27 experts + shibboleth + range |
| **v1.0** | 2026-05-12 | 初始版本：基础规则集+5个unit文件+错误DNA数据库+双盘备份 Initial: basic rules + 5 units + error DNA DB + dual backup |

---

## 💰 打赏支持 / Support the Legion

<div align="center">

<img src="https://raw.githubusercontent.com/AI-wuji/wuji-legion-codex/master/donate.svg" alt="donate" width="400">

*如果无极军团对您有帮助，欢迎打赏支持持续开发。*
*If Wuji Legion helps your work, consider supporting its development.*

<br>

联系方式 / Contact: wuji@legion.dev

</div>

---

**⚔️ 运筹帷幄之中，决胜千里之外。**
**⚔️ A battle won by one who has calculated in the temple — such a general wins many victories.**

**License**: MIT
