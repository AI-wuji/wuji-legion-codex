<div align="center">

# 无极军团 · Wuji Legion

> **「运筹帷幄之中，决胜千里之外。」**

[![GitHub](https://img.shields.io/badge/Codex-Skill-blueviolet)](https://github.com/AI-wuji/wuji-legion-codex)
[![Version](https://img.shields.io/badge/version-3.0-blue)]()
[![License](https://img.shields.io/badge/License-MIT-yellow)]()

<br>

**让 AI 拥有完整组织架构的协作系统——参谋部决策、情报局搜索、安全局护航、质监局审计、档案局归档。**

<br>

[演示](#效果展示) · [安装](#安装) · [组织架构](#组织架构) · [工作流程](#工作流) · [FAQ](#faq--常见问题) · [更新日志](#更新日志)

</div>

---

## 为什么需要无极军团？/ Why Wuji Legion?

大多数 AI Coding Agent 虽然有强大的编程能力，但存在一个根本问题：**单线程思维**。一个独立的 Agent 没有一个完整的组织来支撑——没有情报部门提前收集信息，没有质检部门事后审核，没有档案部门做版本管理，也没有安全部门做防护审查。

**无极军团**正是为了解决这个问题而生。它是一个为 Codex 生态量身定制的 **AI 协作框架**——将现代企业的组织架构和管理流程引入 Agent 协作中。

### 核心优势 / Core Advantages

| 优势 | 说明 |
|------|------|
| 🏛 **完整组织架构** | 参谋部/情报局/安全局/质监局/档案局 五大部门互相制衡 |
| 🤖 **27位专家角色** | 覆盖情报/安全/开发/设计/决策五大领域，顶级人才 |
| 🔄 **自动联动** | 各部门自动配合，无需手动切换激活 |
| 🛡 **多层安全** | 情报局一审+安全局二审+质监局验收，三重保障 |
| 💾 **双盘备份** | C盘日常开发+E盘灾难恢复，系统重做也能快速恢复 |
| 🎯 **省Token策略** | 工作记忆/名称引用/按需加载/死循环检测/错误DNA预检 |
| 🧬 **自我进化** | 发现新工具→安全审核→打靶测试→正式融合，持续进化 |
| 🔄 **多平台封装** | Rust核心+通用壳架构，支持Windows/macOS/Linux |

---

## 安装 / Installation

### 方式一：从GitHub安装（推荐）

```bash
git clone https://github.com/AI-wuji/wuji-legion-codex.git ~/.agents/skills/wuji-legion
```
或使用 Codex 内置技能安装器：对 Codex 说「安装无极军团技能」。

### 方式二：手动安装

1. 下载 [AI-wuji/wuji-legion-codex](https://github.com/AI-wuji/wuji-legion-codex)
2. 解压到 `~/.agents/skills/wuji-legion/`
3. 确保目录结构正确：

```
~/.agents/skills/wuji-legion/
├── SKILL.md          # 总纲+规则集
├── README.md         # 本文档
├── donate.svg        # 打赏
├── CHANGELOG.md      # 更新日志
├── units/            # 11个部门文件
│   ├── staff.md      # 参谋部
│   ├── nuwa.md       # 女娲(人事部)
│   ├── intel.md      # 情报局
│   ├── security.md   # 安全局
│   ├── qa.md         # 质监局
│   ├── archive.md    # 档案局
│   ├── content.md    # 第一师(内容)
│   ├── visual.md     # 第二师(视觉)
│   ├── comfyui.md    # 第三师(ComfyUI)
│   ├── dev.md        # 第四师(开发)
│   └── proving_ground.md  # 打靶场
└── scripts/          # 9个脚本工具
    ├── wuji-backup.py
    ├── wuji-e-sync.ps1
    ├── wuji-e-backup.ps1
    ├── wuji-restore.ps1
    ├── wuji-install.ps1
    ├── errors_db.py
    ├── target_range.py
    ├── beep.ps1
    └── push-to-github.ps1
```

### 激活方式

对 Codex 说 **「阿极」** 或 **「无极军团」** 即可激活。

> ✅ **暗号验证**：激活后出现暗号「运筹帷幄之中，决胜千里之外」即确认生效，且整轮会话持续有效。如未见此句，则技能未加载。

---

## 组织架构 / Organization

```
                   ┌─────────────────────────┐
                   │       用户指令           │
                   └───────────┬─────────────┘
                               │
                   ┌───────────▼─────────────┐
                   │     参 谋 本 部          │
                   │  · 战略决策/需求分析      │
                   │  · 拆分子任务/定方向风格   │
                   │  · 接收质监局结果 → 判断   │
                   │  · 女娲人事官(下属角色)    │
                   └───┬───────┬───────┬─────┘
                       │       │       │
              ┌────────▼─┐ ┌──▼────┐ ┌▼────────┐
              │  情报局   │ │安全局  │ │ 质监局   │
              │ · 搜索    │ │· 加密  │ │· 审计    │
              │ · 研判    │ │· 封装  │ │· 验收    │
              │ · 安全一审 │ │· 二审  │ │· 死循环检测│
              └──────────┘ └───────┘ └─────────┘
                       │
              ┌────────▼──────────────────┐
              │      女 娲 人 事 部         │
              │  接收需求 → 匹配专家        │
              │  → spawn并发派发 → 收集结果  │
              └────────┬──────────────────┘
                       │
       ┌───────────────┼───────────────┐
       │               │               │
  ┌────▼────┐   ┌─────▼─────┐   ┌────▼────┐
  │ 第一师   │   │ 第二师     │   │ 第四师   │
  │ (内容)   │   │ (视觉)     │   │ (开发)   │
  └─────────┘   └───────────┘   └─────────┘
       │
  ┌────▼────┐
  │ 第三师   │
  │ (ComfyUI)│
  └─────────┘

  ┌──────────────────────────────┐
  │        档 案 局               │
  │  · C盘日常开发 + 改动前备份    │
  │  · E盘镜像同步 + 灾难恢复     │
  │  · GitHub远程 + 版本回滚      │
  └──────────────────────────────┘

  ┌──────────────────────────────┐
  │        打 靶 场               │
  │  · 新工具融合前隔离测试        │
  │  · 兼容性/冲突/性能验证        │
  └──────────────────────────────┘
```

### 五大核心部门 / 5 Departments

| 部门 | 职责 | 独立性 |
|------|------|--------|
| **参谋部** | 战略决策、需求分析、拆分子任务 | 含女娲人事官 |
| **情报局** | 全网搜索、竞品分析、安全一审 | 完全独立于执行和质检 |
| **安全局** | 加密防护、不可逆封装、多平台打包 | 完全独立 |
| **质监局** | 第三方独立审计、错误DNA预检、死循环检测 | 完全独立，结果报参谋部 |
| **档案局** | 双盘互备、版本管理、灾难恢复 | 独立存储 |

### 四大作战师团 / 4 Divisions

| 师团 | 职责 | 技术栈 |
|------|------|--------|
| **第一师(内容)** | 文案/PPT/演示 | Slidev / Reveal.js / Markdown |
| **第二师(视觉)** | UI设计/数据可视化 | ECharts / 22种高级图表 |
| **第三师(ComfyUI)** | ComfyUI节点/插件 | Python入口 + Rust(PyO3)核心 |
| **第四师(开发)** | 全栈开发/封装 | Rust核心 + TS前端 / Tauri |

### 27位专家角色 / 27 Expert Roles

| 领域 | 角色 | 专长 |
|------|------|------|
| **情报/搜索** | Mitnick / Shimomura / Snowden / Bellard / Lamo / Swartz | 社会工程/OSINT/代码分析/抓取 |
| **安全/攻防** | Schneier / HD Moore / Hotz / 郭盛华 / 林勇 / Miller | 密码学/渗透/逆向/红客防御 |
| **开发/架构** | Torvalds / Carmack / Thompson / Bellard / 张一鸣 / Musk | 内核/引擎/编译器/产品架构 |
| **视觉/设计** | Steve Jobs / Paul Graham / Edward Tufte | 极简美学/文案/数据可视化 |
| **质量/决策** | 费曼 / 芒格 / 孙子 | 验证逻辑/决策偏误/战略情报 |

---

## 工作流 / Workflow

### 标准项目流程

```
指令 → 参谋部分析需求 → 拆分子任务 → 定方向风格 → 错误DNA预检
                                ↓
                 女娲人事部匹配专家 → spawn并发派发
                                ↓
                 各专家独立并行执行 → 返回结果
                                ↓
             ┌─── 质监局验收 + 死循环检测 ──┐
             ↓                              ↓
           通过 → 参谋部汇总 → 档案局归档   未通过 → 重新派发
```

### 自我进化流程

```
发现新工具 → 情报局搜索+研判+安全一审 → 安全局二审
    → 参谋部融合分析+省token优化 → 打靶场隔离测试
    → 质监局二审 → 档案局备份 → 正式融合
```

### 档案局双盘备份原则

- **C盘**：`C:\wuji-projects\{项目名}\` ← 日常开发 + 改动前备份
- **E盘**：`E:\wuji-projects\{项目名}\` ← robocopy /MIR 镜像同步
- 每次改动前 → 备份原文件到 `.wuji-backups/{日期}_{描述}/`
- 原地址文件删除 → 备份文件夹不动（除非用户要求）
- 系统重做 → GitHub拉无极军团 → 从E盘恢复所有项目

---

## 默认技术栈 / Tech Stack

| 层级 | 语言/框架 |
|------|-----------|
| 前端(用户) | TypeScript + React / HTML |
| 桌面壳 | Tauri(Rust) → EXE/DMG/AppImage |
| 后端核心 | Rust 单二进制 |
| ComfyUI入口 | Python (__init__.py, ≤10行) |
| ComfyUI核心 | Rust (PyO3 → .pyd) |

### 封装规范

| 平台 | 格式 |
|------|------|
| Windows | .exe / .msi |
| macOS Intel | .app / .dmg |
| macOS Apple Silicon | .app / .dmg |
| Linux | .AppImage / .deb |

---

## 运行时优化 / Runtime Optimization

| 策略 | 说明 |
|------|------|
| 🧠 **工作记忆** | 关键规则常驻context，不反复加载 |
| 🏷 **名称引用** | 工具库按名称而非路径引用 |
| 📦 **按需加载** | 只加载当前任务需要的模块 |
| ❌ **死循环检测** | 同子任务输出相似度>85%立即终止 |
| 🧬 **错误DNA预检** | 执行前查ERRORS.md，规避已知bug |
| 🔄 **跨任务缓存** | 同一会话内的结果复用 |

---

## FAQ / 常见问题

**问：和直接使用Codex有什么区别？**

答：普通Codex是一个「一个人」在工作。无极军团是一个「组织」在工作——情报部门提前收集信息、安全部门审查风险、质检部门验收质量、档案部门管理版本。各部门互相制衡，避免盲区。

**问：需要什么环境？**

答：Codex CLI 或 Codex Desktop，以及基础的 Rust/Python/Node.js 环境（取决于项目类型）。

**问：27位专家SKILL从哪里来？**

答：由女娲（huashu-nuwa）角色蒸馏系统创建的15位人物视角Skill + 情报/安全领域内置12位角色。这些专家不是角色扮演，而是在用他们的认知框架帮你分析问题。

**问：如何贡献？**

答：Fork 仓库，提 PR，或联系 wuji@legion.dev。

---

## English Summary

> **"A battle won by one who has calculated in the temple before the battle — such a general wins many victories."**

**Wuji Legion** is an AI collaboration framework for Codex CLI/Desktop that brings enterprise-grade organizational structure to AI-assisted development.

**Core Features:**
- **5 Independent Departments**: Strategy → Intelligence → Security → QA → Archives
- **27 Expert Roles**: Top talents in intelligence, security, development, design, decision-making
- **Automatic Activation**: All departments auto-link, no manual switching needed
- **Multi-Layer Security**: Intelligence review + Security review + QA verification
- **Dual-Disk Backup**: C-drive daily + E-drive disaster recovery
- **Token-Efficient**: Working memory, on-demand loading, infinite loop detection
- **Self-Evolution**: Discover → Review → Test → Integrate
- **Cross-Platform**: Rust core + universal shell (Tauri/TS/PyO3)

**Install:**
```bash
git clone https://github.com/AI-wuji/wuji-legion-codex.git ~/.agents/skills/wuji-legion
```
Say **"阿极"** or **"无极军团"** to activate.

---

## 更新日志 / Changelog

### 2026-05-14 V3.0 — 架构重构

**重大变更**
- 全新架构：参谋部/情报局/安全局/质监局/档案局五大核心互相制衡
- 省token策略：工作记忆/名称引用/按需加载/死循环检测/错误DNA预检
- 默认开发架构：Rust核心 + 通用壳 (Tauri/TS/PyO3)

**新增**
- 14个新人物视角Skill（情报/安全/开发/设计四类顶尖专家）
- 档案局双盘备份机制（C盘+E盘互备+灾难恢复）
- 打靶场流程（新工具融合前隔离测试）
- 女娲人事部管理27位专家角色池
- 完整的暗号验证机制（「运筹帷幄之中，决胜千里之外」）
- push-to-github.ps1 GitHub推送脚本

**修复**
- 脚本孤岛问题：9个脚本全部关联到对应部门
- 部门混搭：拆分qa_intel为质监局+情报局，拆分visual_security为视觉+安全局
- 女娲Skill存在但未联动的问题：修复女娲人事部自动匹配与spawn派发流程
- 删除过时文件
- 所有子技能互相联动，消灭空闲/未激活状态

**优化**
- 总大小紧凑控制在15KB以内
- 暗号验证机制确认真实激活状态

### 2026-05-12 V1.0 — 初始版本

- 基础规则集 + 5个初始unit文件
- 错误DNA数据库
- 本地+E盘备份

---

## 打赏支持 / Support the Legion

<div align="center">

![donate](donate.svg)

*如果无极军团对您有帮助，欢迎打赏支持持续开发。*
*If Wuji Legion helps your work, consider supporting its development.*

<br>

联系方式 / Contact: wuji@legion.dev

</div>

---

<div align="center">

**无极军团 · Wuji Legion**

*让AI拥有完整的组织协作能力*

<br>

MIT License © AI-wuji

[GitHub](https://github.com/AI-wuji/wuji-legion-codex) · [安装](#安装) · [组织架构](#组织架构) · [工作流](#工作流)

<br>

**「运筹帷幄之中，决胜千里之外。」**

</div>
