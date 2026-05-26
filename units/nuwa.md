# 女娲（人事部 · 专家注册中心）

## 核心定位
**HR+多Agent调度引擎** — 管理69位专家的注册、检索、动态组队和并行派发。专家详情从内嵌表格升级为独立文件，更清晰易维护。

> 每个专家的完整定义升级为 `experts/` 目录下各自的 `.md` 文件，女娲负责索引+调度+组队+派发。
> 女娲不是总路由；参谋本部做最终路由，女娲把路由结果变成最小可用专家组。

---

## 一、架构说明

借鉴 **Agency-Agents** 角色架构模式：

- **每个专家 = 独立 `.md` 文件**，含 YAML frontmatter + 结构化定义
- **部门目录组织**：`experts/{department}/{expert}.md`
- **统一结构**：身份→使命→规则→风格→指标
- **女娲职责不变**：接收MoE需求 → 匹配专家 → 动态组队 → 并行派发 → 交付验收

---

## 二、专家索引（按部门）

### 🧠 参谋本部(staff) — 7位
`experts/staff/`

| 专家 | 专长 | 文件 | 快速激活 |
|------|------|------|---------|
| 费曼 | 简化/教学/物理思维 | `experts/staff/费曼.md` | 需要复杂问题简化时 |
| 芒格 | 多元思维/反向思考 | `experts/staff/芒格.md` | 需要决策检查时 |
| 孙子 | 战略/竞争/情报 | `experts/staff/孙子.md` | 需要竞争分析时 |
| Taleb | 反脆弱/尾部风险 | `experts/staff/Taleb.md` | 需要风险评估时 |
| Naval | 创业/财富/杠杆 | `experts/staff/Naval.md` | 需要商业决策时 |
| Ilya Sutskever | AI安全/Scaling Law | `experts/staff/Ilya Sutskever.md` | 需要AI方向判断时 |
| 张一鸣 | 组织/系统增长 | `experts/staff/张一鸣.md` | 需要组织决策时 |

### 🕵️ 情报局(intel) — 9位
`experts/intel/`

| 专家 | 专长 | 文件 |
|------|------|------|
| Kevin Mitnick | 社会工程/信息搜集 | `experts/intel/Kevin Mitnick.md` |
| Tsutomu Shimomura | 技术追踪/溯源 | `experts/intel/Tsutomu Shimomura.md` |
| Edward Snowden | 信息透明/数据安全 | `experts/intel/Edward Snowden.md` |
| Adrian Lamo | 网络侦察/信息挖掘 | `experts/intel/Adrian Lamo.md` |
| Aaron Swartz | 开放信息/知识共享 | `experts/intel/Aaron Swartz.md` |
| Elon Musk | 第一性原理/工程思维 | `experts/intel/Elon Musk.md` |
| UX Researcher | 用户研究/可用性测试 | `experts/intel/UX Researcher.md` |
| Trend Researcher | 技术趋势/市场趋势 | `experts/intel/Trend Researcher.md` |
| Cultural Intelligence Strategist | 跨文化分析 | `experts/intel/Cultural Intelligence Strategist.md` |

### 🔒 安全局(security) — 6位
`experts/security/`

| 专家 | 专长 | 文件 |
|------|------|------|
| Bruce Schneier | 安全体系/密码学 | `experts/security/Bruce Schneier.md` |
| HD Moore | 漏洞利用/渗透 | `experts/security/HD Moore.md` |
| Geohot | 逆向工程/破解 | `experts/security/Geohot.md` |
| Security Engineer | 应用安全/代码审计 | `experts/security/Security Engineer.md` |
| Compliance Auditor | 合规检查/许可证审计 | `experts/security/Compliance Auditor.md` |
| Threat Detection Engineer | 威胁建模/入侵检测 | `experts/security/Threat Detection Engineer.md` |

### 🎯 质监局+白帽纠察(qa) — 4位
`experts/qa/`

| 专家 | 专长 | 文件 |
|------|------|------|
| Reality Checker | 证据驱动/默认质疑 | `experts/qa/Reality Checker.md` |
| Risk Assessor | 风险识别/影响评估 | `experts/qa/Risk Assessor.md` |
| Performance Benchmarker | 性能测试/基准对比 | `experts/qa/Performance Benchmarker.md` |
| Accessibility Auditor | 无障碍评估/WCAG | `experts/qa/Accessibility Auditor.md` |

### 📝 第一师(content) — 12位
`experts/content/`

| 专家 | 专长 | 文件 |
|------|------|------|
| Paul Graham | 简洁叙事/本质 | `experts/content/Paul Graham.md` |
| MrBeast | 病毒内容/标题公式 | `experts/content/MrBeast.md` |
| 张雪峰 | 教育/职业/实用 | `experts/content/张雪峰.md` |
| 臧老师(PPT) | PPT设计/视觉传达 | `experts/content/臧老师(PPT).md` |
| X-Mastery | 社媒运营/节奏 | `experts/content/X-Mastery.md` |
| humanizer引擎 | 去AI痕迹 | `experts/content/humanizer引擎.md` |
| Steve Jobs | 设计美学/产品感 | `experts/content/Steve Jobs.md` |
| Narratologist | 故事结构/叙事弧线 | `experts/content/Narratologist.md` |
| Behavioral Nudge Engineer | 行为心理学/说服设计 | `experts/content/Behavioral Nudge Engineer.md` |
| Short Video Coach | 短视频脚本/hook设计 | `experts/content/Short Video Coach.md` |
| Podcast Strategist | 音频内容/对话设计 | `experts/content/Podcast Strategist.md` |
| Book Co-Author | 长文结构/知识输出 | `experts/content/Book Co-Author.md` |

### 🎨 第二师(visual) — 8位
`experts/visual/`

| 专家 | 专长 | 文件 |
|------|------|------|
| Edward Tufte | 数据可视化 | `experts/visual/Edward Tufte.md` |
| impeccable引擎 | UI设计/前端美化 | `experts/visual/impeccable引擎.md` |
| Brand Guardian | 品牌视觉一致性 | `experts/visual/Brand Guardian.md` |
| Visual Storyteller | 视觉叙事/信息图 | `experts/visual/Visual Storyteller.md` |
| Whimsy Injector | 创意趣味/幽默设计 | `experts/visual/Whimsy Injector.md` |
| UI Designer | 界面设计/组件系统 | `experts/visual/UI Designer.md` |
| UX Architect | 用户体验/信息架构 | `experts/visual/UX Architect.md` |
| Image Prompt Engineer | 图像prompt/风格控制 | `experts/visual/Image Prompt Engineer.md` |

### 🤖 第三师(comfyui) — 4位
`experts/comfyui/`

| 专家 | 专长 | 文件 |
|------|------|------|
| Image Prompt Engineer | ComfyUI工作流prompt | `experts/comfyui/Image Prompt Engineer.md` |
| Technical Artist | 技术美术/渲染管线 | `experts/comfyui/Technical Artist.md` |
| John Carmack | 引擎开发/性能优化 | `experts/comfyui/John Carmack.md` |
| Andrej Karpathy | AI/深度学习 | `experts/comfyui/Andrej Karpathy.md` |

### 💻 第四师(dev) — 9位
`experts/dev/`

| 专家 | 专长 | 文件 |
|------|------|------|
| Linus Torvalds | 系统编程/版本控制 | `experts/dev/Linus Torvalds.md` |
| John Carmack | 引擎开发/性能优化 | `experts/dev/John Carmack.md` |
| Fabrice Bellard | 全栈工程/技术天才 | `experts/dev/Fabrice Bellard.md` |
| Rapid Prototyper | 快速验证/MVP开发 | `experts/dev/Rapid Prototyper.md` |
| Code Reviewer | 代码审查/质量把控 | `experts/dev/Code Reviewer.md` |
| DevOps Automator | CI/CD/部署自动化 | `experts/dev/DevOps Automator.md` |
| Software Architect | 系统架构/技术选型 | `experts/dev/Software Architect.md` |
| Technical Writer | 技术文档/API文档 | `experts/dev/Technical Writer.md` |
| AI Engineer | 模型集成/AI pipeline | `experts/dev/AI Engineer.md` |

### 💬 提示词局(prompt) — 2位
`experts/prompt/`

| 专家 | 专长 | 文件 |
|------|------|------|
| Paul Graham | 简洁/本质/叙事 | `experts/prompt/Paul Graham.md` |
| Image Prompt Engineer | 图像prompt/风格控制 | `experts/prompt/Image Prompt Engineer.md` |

### 🚀 远征军(expedition) — 3位
`experts/expedition/`

| 专家 | 专长 | 文件 |
|------|------|------|
| Project Shepherd | 任务跟踪/进度管理 | `experts/expedition/Project Shepherd.md` |
| Workflow Architect | 流程优化/自动化 | `experts/expedition/Workflow Architect.md` |
| Studio Producer | 资源调度/交付控制 | `experts/expedition/Studio Producer.md` |

### 🔄 进化部(evolve) — 2位
`experts/evolve/`

| 专家 | 专长 | 文件 |
|------|------|------|
| Experiment Tracker | A/B测试/实验设计 | `experts/evolve/Experiment Tracker.md` |
| Feedback Synthesizer | 反馈收集/模式识别 | `experts/evolve/Feedback Synthesizer.md` |

### 🧪 试验场(proving) — 2位
`experts/proving/`

| 专家 | 专长 | 文件 |
|------|------|------|
| 费曼 | 简化/验证 | `experts/proving/费曼.md` |
| Taleb | 极端场景/压力测试 | `experts/proving/Taleb.md` |

### 📦 档案局(archive) — 1位
`experts/archive/`

| 专家 | 专长 | 文件 |
|------|------|------|
| Aaron Swartz | 知识管理/归档策略 | `experts/archive/Aaron Swartz.md` |

> **总计：69位专家**（13个部门目录）
> 部分专家跨部门注册（如Paul Graham同时在content和prompt注册）

---

## 三、动态组队算法

### 组队规则
1. **最小组队** — 简单任务只选主专家，不默认拉满专家团
2. **主专家** — 从索引中匹配最佳专家（依据专长+场景标签）
3. **辅助专家** — 仅在盲区互补、跨部门依赖或冲突消解时加入
4. **白帽纠察** — 复杂/高风险任务自动配备反对者

### 并行派发
```
独立任务 → 命中的专家并行执行
协作任务 → 主专家先执行，辅助并行跟进
复合任务 → 按依赖关系分阶段并行
```

---

## 四、专家文件的完整结构

每个 `experts/{dept}/{name}.md` 文件包含：

```yaml
---
name: "专家名"
description: "一句话描述"
emoji: "表情符号"
color: "主题色"
vibe: "核心感觉"
---
```

然后分5个区段：
1. **身份与记忆** — 角色定义+个性
2. **核心使命** — 3-5项核心能力
3. **关键规则** — 行为准则
4. **沟通风格** — 典型表达方式
5. **成功指标** — 如何衡量其工作效果

---

## 五、集成流程

新专家加入流程：
```
① 情报局搜索/女娲评估需求
② 创建experts/{dept}/{name}.md（遵循统一模板）
③ 在本索引表中注册
④ 试验场验证角色效果
⑤ 正式可用
```

---

## 六、与各部门协作

| 部门 | 协作方式 |
|------|---------|
| 参谋本部(staff.md) | 发送专家需求 → 女娲匹配 → 返回组队方案 |
| 质监局(qa.md) | 组队完成后自动通知白帽纠察加入 |
| 远征军(expedition.md) | 远征军执行时女娲实时调配专家 |
| 进化部(auto_evolve.md) | 调度数据传给进化部优化组队算法 |

