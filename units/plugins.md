# 无极军团 · 插件注册中心 v1.2

> 本文件记录“无极军团纳管和路由”的插件/skill，不把未验证的安装状态写死。
> 市场插件是否已安装，以 Codex Desktop 本地插件页和 `~/.codex/config.toml` 为准；内置插件目前已在配置中启用 browser/documents/spreadsheets/presentations。

---

## 筛选逻辑

**入选标准：** 与无极七方向直接相关，去重取最优。
**排除：** 金融/HR/销售/法律/会议/邮件类，除非项目明确需要。
**融合原则：** 插件只做能力补足，不做新部门；参谋本部路由，女娲按需组队。

---

## Codex 内置插件（本机已启用）

| 插件 | 路由到 | 用途 |
|------|--------|------|
| Browser | intel.md + proving_ground.md | 网页打开、检查、交互测试、截图 |
| Documents | content.md + archive.md | Word/文档生成、整理、归档 |
| Spreadsheets | intel.md + content.md | 表格、结构化数据、分析交付 |
| Presentations | visual.md + content.md | PPT生成、修改、导出预览 |

---

## PPT/HTML候选最终裁决

| 候选 | 裁决 | 原因 | 触发条件 |
|------|------|------|----------|
| PptxGenJS | 保留，按需试验接入 | 程序化PPT能力强，但与现有 `pptx-generation` 重叠 | 现有PPT引擎无法满足复杂图表/JS项目批量生成时启用试验 |
| Marp | 保留，轻量草稿链路 | Markdown转演示快，但设计上限不如主PPT链路 | 用户要快速从Markdown出草稿PPT/PDF时使用 |
| reveal.js | 保留，Web演示专用 | 适合HTML演示和动效，不适合替代PPTX交付 | 用户明确要网页演示、演讲页面或在线deck时使用 |
| OpenDesign | 保留，设计增强链 | 擅长设计系统、Deck/原型/HTML预览和多方向探索，但不替代PPTX主交付 | PPT/HTML/UI需要更强视觉探索、设计系统抽取、交互预览或Deck预览时接入 |
| shadcn/ui | 保留，组件参考 | 组件质量高，但不强行引入依赖 | React/Tailwind项目需要高质量组件参考时使用 |
| html2pptx | 丢弃默认接入 | HTML转PPT布局稳定性不可控，容易和PPTX主线冲突 | 不进入默认链路；只有用户明确要求HTML转PPT时临时评估 |
| daisyUI | 丢弃默认接入 | 快速但模板味重，容易降低无极视觉质量 | 不进入默认链路；只允许低保真原型临时参考 |

裁决原则：没有“永久候选”。保留项有明确触发条件；丢弃项不再占主路由，只在用户明确指定时临时评估。

---

## 市场插件纳管清单

### 第二师 visual.md

| 插件 | 用途 | 融合方式 |
|------|------|---------|
| Figma | 设计稿转代码、UI组件库、设计系统 | 专业UI优先入口 |
| Canva | 轻量设计、社媒图、PPT素材 | 臧老师PPT素材补充 |
| OpenDesign | 设计系统、Deck、原型、HTML预览 | 第二师设计增强；由参谋本部判定，女娲按需接入 |
| Remotion | React视频、动态图形 | 文字/页面视频化 |

### 第三师 comfyui.md

| 插件 | 用途 | 融合方式 |
|------|------|---------|
| HeyGen | AI数字人视频、头像讲解 | 短视频流水线数字人环节 |
| Cloudinary | 媒体资产管理、图片/视频CDN | 档案局媒体管理补充 |
| Hugging Face | 模型库、数据集、Spaces | 模型选型和资源查询 |

### 第四师 dev.md

| 插件 | 用途 | 融合方式 |
|------|------|---------|
| GitHub | PR/Issue/CI管理、仓库协作 | 开发协作入口 |
| Supabase | 数据库、Auth、Edge Functions、Storage | 全栈后端补充 |
| Vercel | 前端部署和Agent发布 | 发布通道 |
| CircleCI | CI/CD流水线 | DevOps自动化补充 |
| Sentry | 错误追踪、性能监控 | QA线上质量反馈 |
| CodeRabbit | AI代码审查 | PR审查自动化 |
| Game Studio | 浏览器游戏开发 | 创意交互输出 |

### 第一师 content.md

| 插件 | 用途 | 融合方式 |
|------|------|---------|
| Notion | 文档协作、Spec、知识库 | 内容素材和项目文档管理 |
| Readwise | 阅读高亮、研究笔记 | 情报局调研信息源 |
| Remotion | 程序化视频 | 内容视频化输出 |
| HeyGen | 数字人讲解 | 真人出镜/课程讲解补充 |

### 远征军 expedition.md

| 插件/skill | 用途 | 融合方式 |
|------------|------|---------|
| Linear | Issue跟踪、项目管理 | 任务拆分、排期和状态跟踪 |
| Notion | Spec、协作文档 | Handoff和需求文档 |
| GitHub | PR/Issue管理 | 开发项目外派协作 |
| Superpowers | 任务拆解、TDD、复审纪律 | 执行纪律补充，不覆盖MoE路由 |

---

## 调用原则

1. 插件不直接响应用户指令，通过对应部门调用。
2. 参谋本部先判断是否需要插件，女娲只在需要跨部门组队或冲突消解时介入。
3. 未确认安装的市场插件，只能提示需要安装/授权，不得假装可用。
4. 成品型输出默认只保留两个入口：预览 + “文件在……”
