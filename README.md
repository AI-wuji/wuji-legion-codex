# ☯️ 无极军团 Codex 版 / Wuji Legion for Codex CLI v1.1

**真并发多Agent作战系统 — 专为 OpenAI Codex CLI 打造**

**True Parallel Multi-Agent Combat System — Built for OpenAI Codex CLI**

[![Codex CLI](https://img.shields.io/badge/Codex_CLI-Optimized-4a9eff)](https://github.com/openai/codex)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

📺 [在线展板](https://ai-wuji.github.io/wuji-legion-codex/) | 📋 [更新日志](./CHANGELOG.md)

---

## 🎯 这是什么？

☯️ **无极军团 Codex 版** 是一个融合了龙虾军团指挥链架构、wuji-dev 开发方法论、以及 15+ 社区精华的全方位 Codex CLI 作战系统。

安装后，你对 Codex CLI 说的每一句话都自动走完整流程：
```
你说 → 阿极提炼 → 参谋部分析 → 情报搜索 → 安全审核 → 
拆分子任务 → spawn_agent 真并发执行 → 质检验收 →
参谋部评估（≤3轮，重复输出自动停）→ 汇报 + 提示音
```

### 核心优势

| 特性 | 说明 |
|------|------|
| **真并发执行** | spawn_agent 同时调度多个子任务，互不阻塞 |
| **全场景覆盖** | ComfyUI插件 / HTML→EXE打包 / PPT演示 / 图表展牌 / 软件开发 / Bug修复 / 逆向工程 |
| **安全第一** | 打靶场沙盒测试（5种扫描）+ 四级权限管控 + 八条安全红线 |
| **不重复踩坑** | 错误DNA数据库自动记录每次Bug修复，改代码前自动检查历史 |
| **文件永不丢** | 改前智能备份（10份轮换）+ E盘实时同步 + 灾难恢复一键还原 |
| **Token 五层优化** | 基础配置(-85%) + 对话管理(-60%) + 子代理纪律(-80%) + Caveman精准输出(-40%) + 渐进工具加载(99.6%) |
| **情报先行** | 每次任务自动并行搜索网页/GitHub/社区/竞品/安全，只回摘要不进上下文 |
| **攻防一体** | 加密(asar+UPX+VMProtect) + 反编译(dnSpy+pycdc+uncompyle6) + 加壳脱壳 |

---

## 🏛️ 系统架构

```
用户 (Codex CLI)
    ↓
☯️ 阿极（总指挥）
    ↓
🎖️ 参谋本部 — 需求分析 → 情报先行 → 安全审核 → 制定计划 → 拆分子任务
    ↓  spawn_agent 真并发
┌──────────┬──────────┬──────────┬──────────┬──────────┬──────────┐
│ 📝 第一师 │ 🎨 第二师 │ ⚙️ 第三师 │ 💻 第四师 │ 📡 第五师 │ 🔧 第六师 │
│  内容作战  │  视觉作战  │ ComfyUI  │  软件开发 │  情报作战  │  支援作战  │
│ PPT/文案  │ UI/品牌   │ 插件合并  │ HTML→EXE │ 全网搜索  │ 部署/CI   │
└──────────┴──────────┴──────────┴──────────┴──────────┴──────────┘
    ↓
📋 质监局 → 🛡️ 安全局 → 🗄️ 档案局 → 🎯 打靶场
    ↓
🔄 参谋本部评估（≤3轮 · 重复输出>85%自动停）
    ↓
☯️ 阿极汇报（含文件路径 + 修改时间 + 🔔 提示音）
```

---

## 🚀 快速开始

### 安装

装好 Codex CLI 后，直接对它说：

> **安装github的无极军团**

系统会自动从 GitHub 克隆并安装全部文件到 `.agents/skills/wuji-legion/`。

### 恢复

如果重装系统，安装完成后说：

> **恢复**

系统自动从 E 盘备份还原所有任务进度。E 盘不在则从 GitHub 克隆基础版。

### 日常使用

不需要记住任何命令。你的每一句话都自动走完整作战流程。

---

## 📁 文件结构

```
wuji-legion-codex/
├── SKILL.md             3.5KB  主系统（17章完整作战手册）
├── README.md                   本说明文件
├── index.html                  GitHub Pages 在线展板
├── CHANGELOG.md                版本历史
│
├── scripts/
│   ├── errors_db.py            错误DNA数据库（add/check/search/dedup/list）
│   ├── target_range.py         打靶场沙盒（code/plugin/dependency/config/permission）
│   ├── wuji-backup.py          智能备份系统（10份轮换）
│   ├── beep.ps1                提示音
│   ├── wuji-e-sync.ps1         E盘实时备份同步
│   ├── wuji-e-backup.ps1       自动备份守护
│   ├── wuji-restore.ps1        灾难恢复
│   └── wuji-install.ps1        安装引导
│
└── skills/
    ├── staff.md                参谋本部 SOP
    ├── comfyui.md              第三师 — ComfyUI
    ├── dev.md                  第四师 — 软件开发
    ├── visual_security.md      第二师视觉 + 安全局
    └── qa_intel.md             质监局 + 第五师情报
```

---

## 🔗 融合来源

| 来源 | 融合内容 | ⭐ |
|------|---------|-----|
| 龙虾军团 v7.1 | 指挥链、动态流水线、打靶场、爆炸半径权限 | AI-wuji |
| 无极开发 v3.0 | 12模块方法论、工作流、安全体系、UI设计 | AI-wuji |
| caveman | 精准输出，-75% Token | 59K⭐ |
| token_saver | 5层Token优化体系，-85% | 社区 |
| mcp-code-execution | CLI优先，工具定义缩减99.6% | 社区 |
| spec_driven_develop | S.U.P.E.R 架构原则 | 715⭐ |
| agent-skill-creator | 跨14平台技能分发 | 911⭐ |
| impeccable | 20命令UI审查体系 | 已安装 |
| 女娲 | 思维蒸馏、多角度分析 | 已安装 |
| drawio-skill | 架构图生成 | 1.4K⭐ |
| comfyui-workflow | 360+节点库 | 235⭐ |
| ComfyUI-custom-node | V3 API开发规范 | 198⭐ |

---

## 🛡️ 灾难恢复

| 场景 | 方法 |
|------|------|
| C盘中毒清空 | `powershell E:\wuji-legion-backup\skills\wuji-legion\scripts\wuji-restore.ps1` |
| 文件改崩 | `python wuji-backup.py restore <文件> [版本]` |
| 换电脑 | 装Codex说"安装github的无极军团"→"恢复" |

---

**☯️ 阿极在此。全军待命。**
**License**: MIT
