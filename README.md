# 🦞 无极军团 Codex 版 / Wuji Legion for Codex CLI v1.0

**真并发多Agent作战系统 — 专为 OpenAI Codex CLI 打造**

**True Parallel Multi-Agent Combat System — Built for OpenAI Codex CLI**

[![Codex CLI](https://img.shields.io/badge/Codex_CLI-Optimized-4a9eff)](https://github.com/openai/codex)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

📺 [在线展板 / Live Dashboard](https://ai-wuji.github.io/wuji-legion-codex/) | 📋 [更新日志 / Changelog](./CHANGELOG.md)

---

## 🎯 这是什么？ / What is this?

🦞 **无极军团 Codex 版** 是基于龙虾军团架构 + wuji-dev 方法论 + 社区精华融合而成的 **Codex CLI 专属多Agent作战系统**。

🦞 **Wuji Legion for Codex** is a fusion of the Lobster Legion architecture, wuji-dev methodology, and community best practices — purpose-built for **Codex CLI**.

### 核心能力 / Core Capabilities

| 特性 Feature | 中文 | English |
|-------------|------|---------|
| 🎖️ | spawn_agent 真并发执行 | True Parallel Execution via spawn_agent |
| ⚔️ | 七大师团协同作战 | 7-division coordinated combat |
| 🛡️ | 打靶场沙盒安全测试 | Range (Sandbox) Security Testing |
| 🔒 | 爆炸半径权限管控 | Blast Radius Permission Control |
| 🧬 | 错误DNA防重复bug | Error DNA — Never repeat a bug |
| 🔄 | 智能备份10份轮换 | Smart Backup with 10-version rotation |
| 🔔 | 任务完成提示音 | Task Completion Beep Notification |
| 📊 | 全网5路情报并行搜索 | 5-channel Parallel Intelligence Search |
| 🔧 | ComfyUI 插件合并/反编译 | ComfyUI Plugin Merge & Reverse Engineering |
| 💻 | HTML→EXE 打包加密加壳 | HTML→EXE Packaging with Encryption |
| 💰 | 五层Token优化体系 | 5-Layer Token Optimization (-99.6%) |

---

## 🏛️ 核心架构 / Core Architecture

```
用户 User (Codex CLI)
    │
    ▼
阿极 A-Ji（总指挥 / Commander）
    │
    ▼
参谋本部 General Staff (P-01) → 情报先行 → 安全审核
    │
    ▼ True Parallel (spawn_agent)
┌──────────┬──────────┬──────────┬──────────┐
│  第一师   │  第二师   │  第三师   │  第四师   │
│  内容作战  │  视觉作战  │ ComfyUI   │  软件开发  │
│ Content  │  Visual  │ 插件作战   │ HTML→EXE │
└──────────┴──────────┴──────────┴──────────┘
    │
    ▼
质监局 QA → 安全局 Security → 档案局 Archives
    │
    ▼
参谋本部评估 → 多轮迭代（≤3轮，重复检测自动停）
    │
    ▼
阿极汇报 + 🔔 提示音
```

---

## 🚀 快速安装 / Quick Start

### 安装 / Installation

```bash
# Clone 仓库
git clone https://github.com/AI-wuji/wuji-legion-codex.git

# 复制到 Codex skills 目录
cp -r wuji-legion-codex/skills/* ~/.codex/skills/
# 或 (Windows PowerShell):
Copy-Item -Path "wuji-legion-codex\skills\*" -Destination "$env:USERPROFILE\.codex\skills\" -Recurse -Force
```

### 使用 / Usage

直接对 Codex CLI 说：

> 「阿极，开始工作 — 我要合并 A ComfyUI 插件和 B 插件的某某功能」

或：

> 「全网搜索 HTML→EXE 打包方案」

系统自动激活完整作战流程。

---

## 📁 文件结构 / File Structure

```
wuji-legion-codex/
├── SKILL.md                   # 主系统（15章完整作战手册）
├── README.md                  # 本说明文件
├── index.html                 # GitHub Pages 展板
├── CHANGELOG.md               # 版本历史
│
├── skills/
│   ├── wuji-orch/SKILL.md     # 参谋本部 + 各师 SOP
│   └── wuji-security/SKILL.md # 安全局 + 质监局 + 情报
│
├── scripts/
│   ├── beep.ps1               # 提示音
│   ├── errors_db.py           # 错误DNA数据库
│   ├── wuji-backup.py         # 智能备份系统
│   ├── target_range.py        # 打靶场沙盒
│   └── wuji-e-sync.ps1        # E盘实时备份
│
└── units/
    ├── staff.md               # 参谋本部 SOP
    ├── comfyui.md             # 第三师 ComfyUI
    ├── dev.md                 # 第四师 软件开发
    ├── visual_security.md     # 第二师视觉 + 安全局
    └── qa_intel.md            # 质监局 + 第五师情报
```

---

## 🔗 融合来源 / Fusion Sources

| # | 来源 / Source | 融合内容 / What's Fused |
|---|------|---------|
| 1 | 🦞 龙虾军团 v7.1 | 指挥链、动态流水线、打靶场、爆炸半径权限 |
| 2 | ⚡ 无极开发 v3.0 | 12模块方法论、工作流、安全体系、UI设计 |
| 3 | 🪨 caveman (59K⭐) | 精准输出，-75% Token |
| 4 | 💰 token_saver | 5层优化体系，-85% Token |
| 5 | ⚡ mcp-code-execution | CLI优先，99.6% 工具定义缩减 |
| 6 | 🏛 spec_driven_develop | S.U.P.E.R 架构原则 |
| 7 | 📦 agent-skill-creator | 技能注册表模型 |
| 8 | 🎨 impeccable | 20命令UI审查体系 |
| 9 | 🧠 女娲 | 思维蒸馏、多角度分析 |
| 10 | 📊 drawio-skill | 架构图生成 |

---

## 📊 版本历史 / Version History

| Version | Date | Major Update |
|---------|------|-------------|
| **v1.0** | 2026-05-13 | 首版发布：完整15章+4脚本+5单位+E盘备份 |

---

**🦞 阿极在此，全军待命。请下达指令！**
**🦞 A-Ji here, all units standing by. Awaiting your orders!**

**License**: MIT
