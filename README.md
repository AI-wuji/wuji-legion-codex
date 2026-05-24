# 无极军团 / Wuji Legion v4.0

**一句话**: 为 Codex AI 设计的 MoE 并行多部门协作框架。参谋本部(MoE中枢) + 14个部门 + 31位专家 + Reasonix缓存引擎。

**适配平台**: Codex CLI + Codex Desktop

[![GitHub](https://img.shields.io/badge/Codex-Skill-blueviolet)](https://github.com/AI-wuji/wuji-legion-codex)
[![Version](https://img.shields.io/badge/version-4.0-purple)]()
[![License](https://img.shields.io/badge/License-MIT-yellow)]()

---

## 🏛️ 架构总览

```
                    [用户指令]
                        │
                        ▼
              ┌─────────────────────┐
              │   MoE Gating (门控)  │ ← 拆解为N个子意图，并行评估
              └────────┬────────────┘
                       │
          ┌────────────┼────────────┐
          ▼            ▼            ▼
   ┌──────────┐ ┌──────────┐ ┌──────────┐
   │ 情报局    │ │ 内容师    │ │ 开发师    │  ← 14个部门并行评估
   │ 安全局    │ │ 视觉师    │ │ ComfyUI   │
   │ 质监局    │ │ 提示词局  │ │ 远征军    │
   └─────┬────┘ └─────┬────┘ └─────┬────┘
         │            │            │
         └────────────┼────────────┘
                      ▼
              ┌─────────────────────┐
              │   MoE加权汇总+路由   │ ← 同时激活多个部门并行执行
              └─────────────────────┘
```

### 核心特性

| 特性 | 说明 |
|------|------|
| 🧠 **MoE并行中枢** | 参谋本部作为Mixture of Experts，并行评估后路由 |
| 🔄 **多Agent协同** | 14个部门可并行执行，依赖关系自动编排 |
| ⚡ **Reasonix缓存引擎** | ImmutablePrefix + AppendOnlyLog + Auto-Compact，命中率>95% |
| 🔴 **红队审计** | 每次任务强制反对意见，不准顺着用户思路走 |
| 🎯 **七大融合方向** | PPT/UI/Rust/ComfyUI/文案/短视频/漫剧 — 优化融合非叠加 |
| 🎛️ **统一管理调度** | 所有skill/MCP统一纳管，新增需过五关 |
| 👥 **31位专家库** | 6组36位人格视角，按需动态组队 |

---

## 📥 安装

### 方法一：GitHub克隆（推荐）
```bash
git clone https://github.com/AI-wuji/wuji-legion-codex.git ~/.agents/skills/wuji-legion
```

### 方法二：全局规则（可选，推荐）
将 `GLOBAL_AGENTS.md` 复制为 `~/.codex/AGENTS.md`，铁律和缓存优化对所有对话生效：
```bash
cp ~/.agents/skills/wuji-legion/GLOBAL_AGENTS.md ~/.codex/AGENTS.md
```

---

## 🚀 快速开始

说「**阿极**」或「**无极军团**」即激活。

### 使用示例

| 你说 | 军团做什么 |
|------|-----------|
| "帮我做个Rust的PPT" | MoE拆解→visual.md(PPT)+dev.md(Rust)并行执行 |
| "搜一下最新的AI框架" | intel.md多引擎并行搜索→安全局审核→报告 |
| "审查这段代码" | qa.md红队模式→至少3个反对意见+风险评估 |
| "写个短视频脚本" | content.md文案融合+mvbeast钩子+humanizer去AI痕 |

---

## 📂 目录结构

```
~/.agents/skills/wuji-legion/
├── SKILL.md              # 主文件（硬性铁律+MoE架构+缓存引擎）
├── config.json           # 配置（providers/routing/cache/red_team）
├── GLOBAL_AGENTS.md      # 全局规则（复制到.codex/AGENTS.md）
├── CHANGELOG.md          # 更新日志
├── README.md             # 本文件
├── units/                # 14个部门文件
│   ├── staff.md          # 参谋本部(MoE中枢·常驻)
│   ├── nuwa.md           # 女娲(多Agent调度引擎)
│   ├── intel.md          # 情报局(多引擎并行搜索)
│   ├── security.md       # 安全局(L1-L5审计)
│   ├── qa.md             # 质监局+红队(双重审计)
│   ├── archive.md        # 档案局(备份回滚)
│   ├── content.md        # 第一师(文案融合)
│   ├── visual.md         # 第二师(PPT三件套+UI)
│   ├── comfyui.md        # 第三师(ComfyUI+插件)
│   ├── dev.md            # 第四师(Rust+CI/CD)
│   ├── prompt_engine.md  # 提示词局(新增)
│   ├── auto_evolve.md    # 进化部(新增·OODA循环)
│   ├── expedition.md     # 远征军(并行外派)
│   └── proving_ground.md # 试验场(沙箱测试)
└── scripts/              # 工具脚本
```

---

## 🏛️ 14个部门职责

| 部门 | 职责 | MoE并行 |
|------|------|---------|
| 参谋本部(staff.md) | MoE门控+加权路由+红队预检 | 中枢常驻 |
| 情报局(intel.md) | 多引擎搜索+可信度评估 | ✅ 可并行 |
| 安全局(security.md) | L1-L5安全审计+许可证检查 | ✅ 可并行 |
| 质监局+红队(qa.md) | 质量验收+强制反对意见 | ✅ 可并行 |
| 第一师-content.md | 文案融合+短视频流水线 | ✅ 可并行 |
| 第二师-visual.md | PPT三件套+HTML/UI美化 | 依赖content |
| 第三师-comfyui.md | ComfyUI+插件+图像生成 | 依赖visual |
| 第四师-dev.md | Rust编程+CI/CD自动化 | ✅ 可并行 |
| 提示词局-prompt_engine.md | 通用/图像/故事板/视频prompt | ✅ 可并行 |
| 进化部-auto_evolve.md | OODA自动进化循环 | 执行后触发 |
| 女娲-nuwa.md | 多Agent调度+31位专家库 | 按需 |
| 远征军-expedition.md | 并行外派+Handoff | ✅ 可并行 |
| 试验场-proving_ground.md | 沙箱化测试+对比评估 | 按需 |
| 档案局-archive.md | 备份回滚+崩溃恢复 | 自动 |

---

## ⚡ Prefix-Cache 引擎

融合 Reasonix(⭐6k) 四大支柱，实测 99.82% 缓存命中率：

| 支柱 | 原理 | 效果 |
|------|------|------|
| 🧊 ImmutablePrefix | 前缀全会话固定，不修改 | 缓存命中候选 |
| 📜 AppendOnlyLog | 只追加，不重排不重写 | 保持前缀连续性 |
| 📝 VolatileScratch | 草稿用完即丢 | 不进缓存前缀 |
| 🔄 Auto-Compact | 上下文>80%自动折叠 | 缓存存活 |

---

## 🔴 红队审计

每次任务强制执行：
1. **前提质疑** — 需求的前提是否成立？
2. **风险识别** — 哪里可能出问题？
3. **盲区检查** — 有没有不知道但该知道的信息？
4. **替代方案** — 有没有更好的做法？
5. **一票否决权** — 核心前提不成立时直接否决

---

## 📜 版本历史

| 版本 | 日期 | 关键更新 |
|------|------|---------|
| **v4.0** | **2026-05-24** | **MoE并行中枢+14部门+Reasonix缓存+红队审计+统一管理+5个虚文件补实** |
| v3.1 | 2026-05-19 | 参谋本部预加载+baoyu视觉融合 |
| v3.0 | 2026-05-14 | 5部门重构+女娲27专家+打靶场 |
| v1.0 | 2026-05-12 | 初始版本 |

---

**License**: MIT
