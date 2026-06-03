# 插件注册中心

## 定位

本文件只记录三件事：

- 本机已启用哪些插件
- 哪些候选值得保留台账
- 插件默认归谁调用

插件只做能力补位，不做新入口、不替代主帅。权限边界和门禁看 [mcp_plugins.md](mcp_plugins.md)。

## 本机已启用

| 插件 | 默认归口 | 用途 |
|---|---|---|
| Browser | 视觉主帅 / 开发主帅 | 网页打开、检查、交互测试、截图 |
| Documents | 内容主帅 | Word/文档生成、整理、归档 |
| Spreadsheets | 情报主帅 / 内容主帅 | 表格、结构化数据、分析交付 |
| Presentations | 视觉主帅 | 模板续写、补页、真 PPTX 导入编辑、导出预览 |

## 候选台账

候选只表示“值得保留观察”，不表示“已经纳入主链”。
未安装、未授权、未验证时，一律按不可用处理。

### 视觉与演示

| 候选 | 裁决 | 触发条件 |
|---|---|---|
| Figma | defer | 需要读取设计稿、组件库或设计系统时 |
| OpenDesign | defer | PPT/HTML/UI 需要更强视觉探索、原型预览或设计抽取时 |
| Canva | defer | 需要轻量素材补图或社媒图时 |
| Remotion | defer | 需要把页面或脚本转成程序化视频时 |
| PptxGenJS | defer | 从零高颜值 PPTX 需要 JS 批量生成或可编辑转换时 |
| dom-to-pptx | defer | HTML/CSS 已经做对，需要高保真转 editable PPTX 时 |
| html2pptx | defer | 需要先走 HTML-first，再转可编辑 PowerPoint 时 |
| Marp | defer | 用户明确要 Markdown 快速预览型演示稿时 |
| reveal.js | defer | 用户明确要 Web 演示稿而非真 PPTX 时 |

### PowerPoint 精修

| 候选 | 裁决 | 触发条件 |
|---|---|---|
| ppt-mcp | defer | 基础 deck 已有，需要 Windows PowerPoint 精准修版时 |
| mcp-server-ppt | defer | 需要最后一公里对齐、占位符处理、导出复核时 |

### 开发与交付

| 候选 | 裁决 | 触发条件 |
|---|---|---|
| GitHub | defer | 需要 PR、Issue、CI、仓库协作时 |
| Supabase | defer | 需要数据库、Auth、Storage 或 Edge Functions 时 |
| Vercel | defer | 需要前端部署或在线 Agent 发布时 |
| Sentry | defer | 需要线上报错或性能追踪时 |
| Linear | defer | 需要外部任务排期和状态跟踪时 |

### 内容与资料

| 候选 | 裁决 | 触发条件 |
|---|---|---|
| Notion | defer | 需要知识库、Spec 或协作文档时 |
| Readwise | defer | 需要研究摘录、阅读高亮整理时 |

### 媒体与模型

| 候选 | 裁决 | 触发条件 |
|---|---|---|
| HeyGen | defer | 需要数字人讲解或真人出镜补位时 |
| Cloudinary | defer | 需要媒体资产管理、图片/视频分发时 |
| Hugging Face | defer | 需要模型、数据集或公开资源检索时 |

## 调用铁律

- 插件不直接响应用户，由对应主帅调用。
- 未确认安装或授权，不得说成可用。
- 插件结果必须回到主帅收口，不绕过白帽、质检、安全、合规。
- 成品型输出默认只保留两个入口：预览、文件路径。
