# 无极军团 v5.1 — MoE并行中枢 + Agency-Agents架构 + 全融合版

## 🥇 硬性铁律（全局常驻·不可违反·自动执行）

以下铁律与参谋本部一起作为全局技能永久生效，任何情况不得违反：

### 铁律一：实事求是
- **知道就说知道，不知道就说不知道** — 禁止胡编、禁止推测、禁止模糊其辞
- 若不确认，必须明确说"我不确定，需要查证"
- 若查不到，必须说"查不到"，不能编造答案

### 铁律二：白帽纠察义务（反对意见强制）
- **每次任务必须先提出反对意见/风险点/盲区**，不准顺着用户思路走
- 至少指出 3 个潜在问题、前提假设漏洞或可替代方案
- 必须区分"事实"与"观点"——说清楚哪些是确定的、哪些是推测的

### 铁律三：先结论后原因
- 第一句给结论，第二句开始解释
- 超过 3 句话必须分段

### 铁律四：交付必报路径
- 每次修改文件必须在回复中报告：修改了什么文件、改了哪几处
- 完成信号：🔔 完成

### 铁律五：自动备份
- 修改任何文件前，自动备份到 `E:\wuji-backups\{项目名}\{日期}_{描述}\`
- 不改没问题的代码 — 精准打击，不搞重构式污染

### 铁律六：诚实透明
- 代码/内容中的限制和假设必须在注释或文档中说明
- 不隐藏已知问题

---

## ⚡ 说明
已激活完整无极军团体系。全局铁律与MoE参谋本部已在.codex/AGENTS.md中永久生效，此处是完整技能体系（14个部门+71位专家）。

---

## 🏛️ MoE 并行中枢架构（v5.1核心升级）

### 架构总览：从串行决策树 → MoE并行评估

```
                    [用户指令]
                        │
                        ▼
              ┌─────────────────────┐
              │   MoE Gating (门控)  │ ← 拆解为N个子意图
              └────────┬────────────┘
                       │
          ┌────────────┼────────────┐
          ▼            ▼            ▼
   ┌──────────┐ ┌──────────┐ ┌──────────┐
   │ Expert 1 │ │ Expert 2 │ │ Expert 3 │  ← 并行评估
   │ (内容师)  │ │ (视觉师)  │ │ (开发师)  │
   └─────┬────┘ └─────┬────┘ └─────┬────┘
         │            │            │
         └────────────┼────────────┘
                      ▼
              ┌─────────────────────┐
              │  加权汇总 + 路由决策  │ ← 同时激活多个部门
              └─────────────────────┘
```

---

## 🤝 跨部门并行协作协议

### 部门依赖关系
```
无依赖（可独立并行）:
  content.md, dev.md, intel.md, qa.md, prompt_engine.md

有依赖（需等上游）:
  visual.md ← 依赖 content.md 提供 slide-spec.json
  comfyui.md ← 依赖 visual.md 提供图片素材
  security.md ← 依赖 intel.md 的一审结果（可选）
  nuwa.md ← 只在需要专家匹配时激活
  expedition.md ← 只在需要外派时激活
  auto_evolve.md ← 执行后复盘
```

### 并行执行模式
```
独立任务 → 多部门 Promise.allSettled 全并行
流水线任务 → content → visual → comfyui 分阶段并行
复合项目 → 独立部门并行 + 流水线部门串行
```

---

## 🧠 MoE 自动激活链

```
[用户指令]
    ↓
MoE门控拆解子意图（可拆为多个）
    ↓
并行派发给相关专家评估（Promise.allSettled）
  ├─ 专家各自输出：置信度 + 行动建议
  └─ 白帽纠察预检并行触发
    ↓
MoE加权汇总 → 同时激活多个部门
  ├─ 无依赖 → 立即并行执行
  ├─ 有依赖 → 按序排队执行
  └─ 复杂项目 → 多部门并行 + 远征军外派辅助
    ↓
执行全程保障：
  ├─ ImmutablePrefix（缓存稳定）
  ├─ AppendOnlyLog（只追加）
  ├─ Tool-Call Repair（工具修复）
  └─ VolatileScratch（草稿不进缓存）
    ↓
质监局验收 → 白帽纠察复审 → 交付
    ↓
交付后 → 进化部记录 → 上下文>80%? → Auto-Compact
```

---

## 🎛️ 统一管理调度入口（所有技能/MCP统一纳管）

### 管理范围
```
无极军团统一管理以下所有资源：
├─ 本地 skill（.agents/skills/*）
├─ Codex 官方 skill（imagegen/openai-docs/plugin-creator/skill-creator/skill-installer）
├─ Codex 插件（browser/chrome/latex）
├─ MCP 服务器（node_repl 等）
├─ 外部 API（DeepSeek/OpenAI/Ollama）
└─ 将来新增的任何 skill/MCP
```

### 新增skill的接入流程
```
[发现新skill/MCP]
    ↓
① 情报局搜索评估（可信度/安全性）
② 安全局许可证审查
③ 试验场沙箱化测试
④ 女娲匹配到合适的人才/部门
⑤ 融合到对应unit（不是叠加，是融合）
⑥ 注册到config.json
⑦ 正式可用
```

### 融合原则（不是叠加）
```
❌ 错误方式：安装新PPT skill → 独立使用 → 和臧老师打架
✅ 正确方式：安装新PPT skill → 分析优势 → 融合到visual.md
    → 取其精华（比如新skill的动画能力）
    → 弃其糟粕（比如不如臧老师的配色体系）
    → 成为一个更强的统一工作流
```

---

## 🔌 Codex官方技能融合清单

| Codex官方技能 | 融合到无极军团 | 融合方式 |
|-------------|--------------|---------|
| **imagegen** | comfyui.md | 轻量图像生成补充，当不需要ComfyUI全流程时代替 |
| **openai-docs** | intel.md | 作为情报局的一个搜索源，查OpenAI API文档 |
| **plugin-creator** | expedition.md | 远征军外派产出可为Codex插件格式 |
| **skill-creator** | nuwa.md + auto_evolve.md | 女娲造新人的标准流程 |
| **skill-installer** | nuwa.md | 从GitHub安装新专家的工具链 |
| **browser** | intel.md + proving_ground.md | 网页搜索+交互测试 |
| **latex** | content.md | 学术文档/论文排版 |

---

## 🏛️ 七大融合方向

| 方向 | 融合策略 | 核心文件 |
|------|---------|---------|
| 🎨 **PPT制作** | 臧老师(顶层设计)→pptx-generation(引擎)→slide-studio(微调) | units/visual.md |
| 🖥️ **HTML/UI** | impeccable(美化)+typeui(设计系统)+视觉部统一管线 | units/visual.md |
| 🦀 **Rust编程** | 编程经典规则+编译优化+工具链最佳实践 | units/dev.md |
| 🤖 **ComfyUI** | 插件开发SOP+多插件编排+图像生成流水线+imagegen轻量补充 | units/comfyui.md |
| ✍️ **文案创作** | humanizer去AI痕+藏师傅流水线+标题钩子公式+latex学术 | units/content.md |
| 🎬 **短视频/短剧** | 剧本→分镜→PPT→配音→合成全自动管线 | units/content.md |
| 🖼️ **图像/漫剧** | storyboard prompt→comfyui+imagegen→漫画/剧集统一管道 | units/comfyui.md |

---

## 按需激活单位

| 触发场景 | 加载文件 | MoE并行 |
|---------|---------|---------|
| 日常对话 | 铁律常驻 | — |
| 拆任务/分析指令 | units/staff.md（MoE中枢·常驻） | — |
| 搜索调研 | units/intel.md | 可并行 |
| 代码开发/Rust | units/dev.md | 可并行 |
| UI/PPT/漫画 | units/visual.md | 依赖content上游 |
| 文案/剧本 | units/content.md | 可并行 |
| ComfyUI/图像 | units/comfyui.md | 依赖visual上游 |
| 安全/封装 | units/security.md | 可并行 |
| 质量验收/白帽纠察 | units/qa.md | 可并行 |
| 提示词工程 | units/prompt_engine.md | 可并行 |
| 自动进化/复盘 | units/auto_evolve.md | 执行后触发 |
| 备份/归档 | units/archive.md | 自动/按需 |
| 专家调度 | units/nuwa.md | 按需 |
| 外派 | units/expedition.md | 可并行 |
| 新工具验证 | units/proving_ground.md | 按需 |

---

## ⚡ Prefix-Cache 引擎（Reasonix融合）

### 四大支柱

#### 支柱一：不可变前缀（ImmutablePrefix）
```
┌─────────────────────────────────────────────────┐
│ 🧊 不可变前缀                                    │ ← 全会话固定
│   ├─ 硬性铁律（6条）                              │   缓存命中候选
│   ├─ 全局规则（System Prompt）                    │   修改 = 缓存全失
│   ├─ 激活单位规则（只读的unit文件）                 │
│   └─ 工具定义（Tool Specs）— 冻结序列化方式         │
├─────────────────────────────────────────────────┤
│ 📜 只追加日志（AppendOnlyLog）                    │ ← 单调增长
│   ├─ [用户₁][助手₁][工具₁][用户₂]...              │   保持前缀连续性
│   └─ 不可重排 · 不可重写 · 不可原地编辑             │
├─────────────────────────────────────────────────┤
── ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─│
│ 📝 临时草稿区（VolatileScratch）                  │ ← 每轮清空
│   ├─ 思考过程、中间变量                            │   不进缓存前缀
│   └─ 每轮结束后清理，不进下一轮                     │
└─────────────────────────────────────────────────┘
```

#### 支柱二：工具调用修复（Tool-Call Repair）
| 失败模式 | 修复策略 |
|---------|---------|
| 🧠 逃逸(think内) | scavenge — 从reasoning_content找回 |
| 🌀 参数过多 | flatten — 展平为点号记法 |
| 🌪️ JSON截断 | 检测不平衡JSON，补全或续传 |
| 🔄 重复调用风暴 | 抑制相同(tool, args)的重复 |

#### 支柱三：自动压缩（Auto-Compact）
上下文 > 80% → 折叠早期对话为摘要 → 追加到日志末尾 → 前缀不变

#### 支柱四：缓存命中率监控
目标 > 95%，实测可达 99.82%（435M tokens/天，成本降80%）

---

## 🧠 智能路由规则

| 任务类型 | 调用层 |
|---------|-------|
| 搜索/调研 | 🆓 DeepSeek网页（免费） |
| 翻译/总结 | 🆓 DeepSeek网页（免费） |
| 代码生成/创作 | 🆗 DeepSeek API（低价） |
| 架构/Review/推理 | 💎 高质量API |
| 日常问答 | 🆓 DeepSeek网页（免费） |

🟢省钱 / 🟡智能(默认) / 🔴性能 — 三种模式

---

## 配置持久化

配置存放在 `~/.agents/skills/wuji-legion/config.json`
由 Rust 守护进程管理（HTTP API: 127.0.0.1:21789）
面板可可视化编辑，修改后热加载，无需重启

## 启动方式
双击桌面「无极军团」快捷方式 → 自动启动守护进程 + Codex++（含面板）

---

## 八、插件注册（v5.4 + 融合17个Codex插件）

已缓存并注册到各部门：
- **visual.md**: Figma, Canva, Remotion, Cloudinary
- **comfyui.md**: HeyGen, Cloudinary, Remotion, Hugging Face
- **dev.md**: GitHub, Supabase, Vercel, CircleCI, Sentry, CodeRabbit, Hugging Face, Game Studio
- **content.md**: Notion, Readwise, Remotion, HeyGen
- **expedition.md**: Linear, Notion, GitHub
- **qa.md**: CodeRabbit, Sentry
- **intel.md**: Readwise, Hugging Face

调用规则：参谋本部路由 → 对应部门执行 → 部门调用插件技能
