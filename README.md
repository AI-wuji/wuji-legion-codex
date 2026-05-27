# 无极军团 / Wuji Legion v5.7

> **全球首款为 Codex AI 设计的 MoE 并行多部门协作框架 · 全局自动激活 · 白帽纠察全程监督 · 69位专家协同 · Reasonix 缓存引擎**

[![GitHub](https://img.shields.io/badge/Codex-Skill-blueviolet)](https://github.com/AI-wuji/wuji-legion-codex)
[![Version](https://img.shields.io/badge/version-5.7-purple)]()
[![License](https://img.shields.io/badge/License-MIT-yellow)]()
[![Agents](https://img.shields.io/badge/Experts-69-green)]()
[![MoE](https://img.shields.io/badge/Architecture-MoE-blue)]()

---

## 📋 一句话

为 Codex AI 设计的 MoE 并行多部门协作框架。**参谋本部(MoE中枢)** 自动拆解指令 → 按需激活14个部门 + **69位专家** → 白帽纠察分级监督 → PPT/HTML/Image2/Rust/UI/ComfyUI 质量门禁 → **Reasonix缓存引擎** 保Token。**全局自动生效，无需口令，但不会每次项目启动都跑完整军团。**

---

## 🧩 核心亮点

| 亮点 | 说明 | 自动激活 |
|------|------|---------|
| 🧠 **MoE 并行中枢** | 参谋本部自动拆解指令，并行评估后路由，多部门可同时执行 | ✅ 全局永久 |
| ⛑️ **白帽纠察** | 全局常驻、分级触发；复杂/高风险任务完整质疑，简单任务不拖慢交付 | ✅ 全局永久 |
| 👥 **69位专家** | 69位专家统一管理（按部门归并，按需动态组队） | 按需 |
| 🔄 **多Agent并行** | Promise.allSettled 调度，独立任务并行，依赖任务自动编排 | ✅ MoE路由 |
| ⚡ **Reasonix缓存** | ImmutablePrefix + AppendOnlyLog + VolatileScratch + Auto-Compact | ✅ 全局永久 |
| 🎯 **七大融合方向** | PPT/UI/Rust/ComfyUI/文案/短视频/漫剧 — 优化融合不是简单叠加 | 按需 |
| 🎛️ **统一管理调度** | 所有 skill/MCP 统一纳管，新增需过五关 | 自动 |
| 🏗️ **Agency-Agents架构** | 每位专家独立`.md`文件，YAML frontmatter+结构化定义，13个专家目录 | 已融入 |
| 🧪 **质量门禁** | Rust/Tauri、HTML/UI、Python、PowerShell、ComfyUI 插件按栈验收 | 按需 |

---

## 🏛️ 架构总览

```
                    [用户指令]
                        │
                        ▼
              ┌─────────────────────────┐
              │  ⚙️ 白帽纠察预检 ⚙️      │ ← 全局常驻，先质疑前提
              └────────┬────────────────┘
                       │
                       ▼
              ┌─────────────────────────┐
              │   MoE Gating (门控)      │ ← 参谋本部拆解为N个子意图
              └────────┬────────────────┘
                       │
          ┌────────────┼────────────────┐
          ▼            ▼                ▼
   ┌──────────┐ ┌──────────┐   ┌──────────┐
   │ 情报局    │ │ 内容师    │   │ 开发师    │  ← 14个部门并行评估
   │ 安全局    │ │ 视觉师    │   │ ComfyUI   │
   │ 质监局    │ │ 提示词局  │   │ 远征军    │
   └─────┬────┘ └─────┬────┘   └─────┬────┘
         │            │              │
         └────────────┼──────────────┘
                      ▼
              ┌─────────────────────────┐
              │  MoE加权汇总+路由        │ ← 女娲动态组队，并行执行
              └────────┬────────────────┘
                       │
                       ▼
              ┌─────────────────────────┐
              │  白帽纠察复审+质监局验收 │ ← 全程监督到交付
              └─────────────────────────┘
```

---

## 🏛️ 14个部门一览

| 部门 | 文件 | 职责 | 并行 |
|------|------|------|------|
| 🧠 **参谋本部** | `units/staff.md` | MoE门控+加权路由+白帽纠察预检 | 中枢常驻 |
| 🕵️ **情报局** | `units/intel.md` | 多引擎搜索+可信度评估 | ✅ |
| 🛡️ **安全局** | `units/security.md` | L1-L5安全审计+许可检查 | ✅ |
| 🎯 **质监局** | `units/qa.md` | 质量验收+白帽纠察 | ✅ |
| 📝 **第一师(内容)** | `units/content.md` | 文案融合+短视频流水线 | ✅ |
| 🎨 **第二师(视觉)** | `units/visual.md` | PPT三件套+HTML/UI美化 | 依赖内容 |
| 🤖 **第三师(ComfyUI)** | `units/comfyui.md` | ComfyUI+插件+图像生成 | 依赖视觉 |
| 💻 **第四师(开发)** | `units/dev.md` | Rust编程+CI/CD自动化 | ✅ |
| 💬 **提示词局** | `units/prompt_engine.md` | 通用/图像/故事板/视频prompt | ✅ |
| 🔄 **进化部** | `units/auto_evolve.md` | OODA自动进化循环 | 执行后 |
| 👩 **女娲(人事部)** | `units/nuwa.md` | 69位专家动态组队+并行派发 | 按需 |
| 🚀 **远征军** | `units/expedition.md` | 并行外派+Handoff+状态协议 | ✅ |
| 🧪 **试验场** | `units/proving_ground.md` | 沙箱测试+AB对比+评估 | 按需 |
| 📦 **档案局** | `units/archive.md` | 备份回滚+崩溃恢复 | 自动 |

---

## 👥 69位专家体系


**69位专家（按部门归并）** — 不再区分人物视角与领域专精，统一调度：

| 部门 | 专家数 | 代表 |
|------|--------|------|
| 🧠 参谋本部 | 7位 | 费曼/芒格/孙子/Taleb/Naval/Ilya/张一鸣 |
| 🕵️ 情报局 | 9位 | Mitnick/Snowden/Elon + UX/Trend/Cultural |
| 🔒 安全局 | 9位 | Schneier/Moore/Geohot + Security/Compliance/Threat |
| 📝 第一师(内容) | 12位 | PGraham/MrBeast/臧老师/张雪峰 + Narratologist/Coach |
| 🎨 第二师(视觉) | 12位 | Tufte/impeccable + Brand/UI/UX/Prompt |
| 🤖 第三师(ComfyUI) | 4位 | Carmack/Karpathy + Technical Artist/Image Prompt |
| 💻 第四师(开发) | 9位 | Linus/Bellard + Prototyper/Reviewer/Architect |
| 🎯 质监局 | 4位 | Reality Checker/Risk Assessor/Performance/Accessibility |
| 🚀 远征军 | 5位 | Project Shepherd/Workflow/Producer |
| 其他 | 7位 | 提示词局/进化部/试验场/档案局 |

> 女娲(Human Resources)根据experts/目录下69个独立.md文件统一匹配，按部门+领域双维度检索，确保最佳团队配置。每个专家有完整的身份→使命→规则→风格→指标定义


---

## ⛑️ 白帽纠察（全局常驻·永久生效）

**不是审计，是按风险分级触发的反对意见者。** 写入全局 `AGENTS.md`，所有对话自动生效：

1. **前提质疑** — 你的需求前提是否成立？
2. **风险识别** — 哪里可能出问题？
3. **盲区检查** — 有没有不知道但该知道的信息？
4. **替代方案** — 有没有更好的做法？
5. **一票否决权** — 核心前提不成立时直接否决

> 🎯 目标：不在错误的路上越走越远

---

## ⚡ Reasonix 缓存引擎（全局常驻）

融合 Reasonix 开源项目四大支柱，省Token + 高命中：

| 支柱 | 原理 | 效果 |
|------|------|------|
| 🧊 ImmutablePrefix | 前缀全会话固定，不修改 | 缓存命中候选 |
| 📜 AppendOnlyLog | 只追加，不重排不重写 | 保持前缀连续性 |
| 📝 VolatileScratch | 思考草稿用完即丢 | 不进缓存前缀 |
| 🔄 Auto-Compact | 上下文>80%自动折叠为摘要追加 | 缓存持续存活 |

### 省Token行为准则
- 不重复已说内容，不输出废话
- 能用一句话说完不用两句
- 上一步结果直接传下一步，不重复描述
- 不提未激活的部门/技能

---

## 🗂️ 文件结构

```
wuji-legion-codex/
├── SKILL.md               # 完整体系总纲
├── GLOBAL_AGENTS.md       # 全局规则（复制到 .codex/AGENTS.md）
├── CHANGELOG.md           # 版本历史
├── README.md              # 本文件
├── config.json            # 全局配置
├── commander/
│   └── SKILL.md           # Commander 技能
├── units/                 # 14个部门
│   ├── staff.md           # 参谋本部(MoE中枢·常驻)
│   ├── nuwa.md            # 女娲(HR+69位专家，按需组队)
│   ├── intel.md           # 情报局
│   ├── security.md        # 安全局
│   ├── qa.md              # 质监局+白帽纠察
│   ├── archive.md         # 档案局
│   ├── content.md         # 第一师(文案)
│   ├── visual.md          # 第二师(视觉/PPT/UI)
│   ├── comfyui.md         # 第三师(ComfyUI)
│   ├── dev.md             # 第四师(Rust/CI-CD)
│   ├── prompt_engine.md   # 提示词局
│   ├── auto_evolve.md     # 进化部
│   ├── expedition.md      # 远征军
│   └── proving_ground.md  # 试验场
└── scripts/               # 工具脚本
```

---

## 🚀 快速开始

### 安装

```bash
# 1. 克隆仓库到 E 盘
cd E:\wuji-projects\
git clone https://github.com/AI-wuji/wuji-legion-codex.git

# 2. 复制全局规则到 Codex 配置
cp .\wuji-legion-codex\GLOBAL_AGENTS.md $env:USERPROFILE\.codex\AGENTS.md

# 3. 复制 skill 到 Agents 目录
cp -Recurse .\wuji-legion-codex $env:USERPROFILE\.agents\skills\wuji-legion

# 4. 运行安装脚本
.\scripts\wuji-install.ps1
```

### 使用

```
方式一：全局自动激活（推荐）
  铁律+白帽纠察+MoE中枢+Cache已写入 .codex/AGENTS.md，所有对话自动生效
  
方式二：手动激活（备用）
  说「阿极」或「启动无极军团」= 启动MoE参谋本部，仍然按需门控，不全量启动14部门+69专家
```

### 工作示例

**"帮我做一个 Rust CLI 工具的 PPT + 写短视频脚本"**

1. ⛑️ **白帽纠察**：质疑前提 — 目标受众是谁？PPT用途是演示还是培训？
2. 🧠 **MoE参谋部**：拆解为2个子任务 → 并行路由
3. 📝 **内容师(第一师)** → 写PPT文案 + 短视频脚本
4. 🎨 **视觉师(第二师)** → 根据文案做PPT设计
5. 💻 **开发师(第四师)** → 并行写Rust CLI代码
6. ⛑️ **白帽纠察复审** + 质监局验收
7. 📦 **档案局**自动备份

---


## 💾 灾备恢复

**系统重装或 Codex 重装后，三步恢复：**

### 一键恢复
```bash
cd E:\wuji-projects\
git clone https://github.com/AI-wuji/wuji-legion-codex.git
cd wuji-legion-codex
.\scripts\wuji-restore.ps1
```

### 手动恢复
| 步骤 | 操作 | 说明 |
|------|------|------|
| 1 | `git clone https://github.com/AI-wuji/wuji-legion-codex.git` | 拉取最新版本 |
| 2 | 复制 `GLOBAL_AGENTS.md` → `.codex/AGENTS.md` | 恢复铁律+白帽纠察+MoE+Cache |
| 3 | 复制整个仓库 → `.agents/skills/wuji-legion/` | 恢复14个部门+69位专家+所有配置 |
| 4 | Codex Desktop → 设置 → 插件 → 搜名称安装 | 恢复市场插件；内置 browser/documents/spreadsheets/presentations 以 `config.toml` 为准 |
| 5 | 新建对话，说"阿极" | 验证MoE参谋本部门控 |

**备份策略：** 仓库本身就是备份。所有配置都在 GitHub 上，重装后 git clone 即可。市场插件可能需要重新安装或授权，但 `units/plugins.md` 记录了纳管清单；Codex 内置插件以本地 `config.toml` 启用状态为准。

## 📜 版本历史

| 版本 | 日期 | 关键更新 |
|------|------|---------|
| **v5.7** | **2026-05-27** | **MoE按需门控 + PPT/HTML/Image2生产线 + Rust/UI/ComfyUI质量门禁 + 插件/专家一致性修正** |
| **v4.2** | **2026-05-25** | **白帽纠察全局化 + 69位专家 + MoE执行计划器 + 冲突解决协议 + 并行状态协议** |
| v4.0 | 2026-05-24 | MoE并行中枢 + 14部门 + Reasonix缓存 + 统一管理 |
| v3.1 | 2026-05-19 | 参谋本部预加载 + baoyu视觉融合 |
| v3.0 | 2026-05-14 | 5部门重构 + 女娲27专家 + 打靶场 |
| v1.0 | 2026-05-12 | 初始版本 |

---

## 📄 License

MIT — 自由使用，欢迎 Star ⭐

---

**Made with 🧠 by AI-Wuji · [GitHub](https://github.com/AI-wuji/wuji-legion-codex)**






