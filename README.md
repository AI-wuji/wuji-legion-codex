# 无极军团 / Wuji Legion v5.23

> **阿极秘书层 + MoE参谋本部 + 女娲组队调度的 Codex 多部门协作框架 · 快答优先 · 按需执行 · 白帽可进群纠察 · 69位专家协同**

[![GitHub](https://img.shields.io/badge/Codex-Skill-blueviolet)](https://github.com/AI-wuji/wuji-legion-codex)
[![Version](https://img.shields.io/badge/version-5.23-purple)]()
[![License](https://img.shields.io/badge/License-MIT-yellow)]()
[![Agents](https://img.shields.io/badge/Experts-69-green)]()
[![MoE](https://img.shields.io/badge/Architecture-MoE-blue)]()

---

## 📋 一句话

为 Codex AI 设计的分层协作框架。所有项目和新会话中，用户默认只和 **阿极秘书层** 交流：快聊、问答、澄清需求、省Token；任务成型后，阿极把任务规划书交给 **参谋本部(MoE中枢)** 拆解路由，再由 **女娲** 组建最小可用团队，按需激活14个部门和69位专家。白帽、女娲等成员可被用户点名加入群聊，但阿极始终是默认对话入口和短报出口。

---

## 🧩 核心亮点

| 亮点 | 说明 | 自动激活 |
|------|------|---------|
| 🧑‍💼 **阿极秘书层** | 日常快聊、问答、澄清、整理任务规划书，不默认启动MoE | ✅ 默认 |
| 🧠 **MoE 并行中枢** | 任务成型后由参谋本部拆解路由，多部门可同时执行 | 按需 |
| 👩 **女娲组队** | 根据MoE结果匹配专家/skill/MCP/插件，去重融合 | 按需 |
| ⛑️ **白帽纠察** | 可被用户点名进群；复杂/高风险/执行节点内部纠察，外部短报 | 按需/点名 |
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
              │  阿极秘书层              │ ← 快聊/问答/澄清/任务规划书
              └────────┬────────────────┘
                       │
             任务未成型│任务成型
                 ┌─────┴─────┐
                 ▼           ▼
            继续沟通   ┌─────────────────────────┐
                       │   MoE参谋本部            │ ← 拆解子意图/路由能力
                       └────────┬────────────────┘
                                │
                                ▼
                       ┌─────────────────────────┐
                       │   女娲组队调度           │ ← 专家/skill/MCP/插件去重融合
                       └────────┬────────────────┘
                                │
                ┌───────────────┼────────────────┐
                ▼               ▼                ▼
         ┌──────────┐    ┌──────────┐     ┌──────────┐
         │ 情报局    │    │ 内容/视觉 │     │ 开发/安全 │
         │ 质监局    │    │ ComfyUI  │     │ 远征军    │
         └─────┬────┘    └─────┬────┘     └─────┬────┘
               └───────────────┼────────────────┘
                               ▼
                       ┌─────────────────────────┐
                       │  质监/白帽复核           │ ← 内部纠察，外部短报
                       └────────┬────────────────┘
                                ▼
                       ┌─────────────────────────┐
                       │  阿极向用户短报           │
                       └─────────────────────────┘
```

白帽可被用户点名加入日常讨论：

```
你 ↔ 阿极
    ↘ 白帽（1-3条短反对意见，不自动启动MoE）
```

女娲也可被用户点名加入讨论：

```
你 ↔ 阿极
    ↘ 女娲（组队/能力融合/专家选择，不自动启动MoE）
```

---

## 🏛️ 14个部门一览

| 部门 | 文件 | 职责 | 并行 |
|------|------|------|------|
| 🧠 **参谋本部** | `units/staff.md` | 任务成型后的MoE门控+加权路由 | 按需 |
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
| 🚀 **远征军** | `units/expedition.md` | Trae/免费模型等低成本外派算力，当前为备选 | 按需 |
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
  阿极秘书层+铁律+Cache已写入 .codex/AGENTS.md，普通沟通默认快答
  
方式二：手动激活（备用）
  任务成型后，阿极整理任务规划书 → 参谋本部MoE → 女娲组队 → 部门执行
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
| 5 | 新建对话，普通沟通直接问；执行任务时说“启动无极军团” | 验证阿极默认入口与MoE参谋本部门控 |

**备份策略：** 仓库本身就是备份。所有配置都在 GitHub 上，重装后 git clone 即可。市场插件可能需要重新安装或授权，但 `units/plugins.md` 记录了纳管清单；Codex 内置插件以本地 `config.toml` 启用状态为准。

## 📜 版本历史

| 版本 | 日期 | 关键更新 |
|------|------|---------|
| **v5.23** | **2026-05-28** | **PPT设计先行门禁：臧老师/OpenDesign产物必须可见** |
| **v5.22** | **2026-05-28** | **PPT审美硬验收：禁止文字硬塞和模板滥用** |
| **v5.21** | **2026-05-27** | **PPT预检禁行 + 主链路前置** |
| **v5.20** | **2026-05-27** | **PPT执行门禁 + 可观察链路** |
| **v5.19** | **2026-05-27** | **OpenDesign 按需融合 + PPT/HTML/UI 设计增强链** |
| **v5.18** | **2026-05-27** | **女娲角色分工表 + 默认多Agent并行派发** |
| **v5.17** | **2026-05-27** | **参谋本部先拆解、女娲后组队 + 工具环境懒加载** |
| **v5.16** | **2026-05-27** | **启动回执闸门 + 快答零工具 + 老会话强制重载口令** |
| **v5.15** | **2026-05-27** | **规则瘦身：执行纠偏总纲 + GitHub按需同步** |
| **v5.14** | **2026-05-27** | **PPT禁止逐字稿直灌模板 + 内容重构优先** |
| **v5.13** | **2026-05-27** | **启动/激活无极军团显式触发 MoE + 中文PPT编码铁律** |
| **v5.12** | **2026-05-27** | **PPT/HTML/文档成品强制路由 + PPT环境降级协议** |
| **v5.11** | **2026-05-27** | **阿极全局身份层 + 身份问答规则 + 群聊成员接入** |
| **v5.10** | **2026-05-27** | **并行分路归属纠偏 + 远征军低成本外派定位 + Trae CLI实测备选** |
| **v5.9** | **2026-05-27** | **阿极秘书层 + MoE任务交接 + 女娲组队 + 白帽群聊短报** |
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






