# 插件镜像

Mirror source: `kernel-source.json`

## 定位

本文件只保留三类真状态：

- `宿主可用`：当前 Codex 宿主环境已经提供
- `军团准入`：无极军团已经承认、可由主链按需挂载
- `边界裁决`：已明确只作来源池、按需单案审查，或明确拒绝

不再保留“先登记、以后再说”的候选台账。

插件只属于第二层 `capability-mount`，只做能力补位，不构成新入口、不替代主帅、不形成第二路由。

## 宿主可用插件

这些是当前 Codex 宿主已提供的插件能力，但“宿主可用”不等于“默认挂载”：

| 插件 | 宿主状态 | 默认归口 | 用途 |
|---|---|---|---|
| Browser | available | `visual-profile` / `development-profile` / `intelligence-profile` / `quality-inspection` | 打开、检查、交互、测试、截图浏览器目标 |
| Documents | available | `content-profile` / `intelligence-profile` / `audit` | 创建、编辑、整理文档工件 |
| Spreadsheets | available | `data-profile` / `intelligence-profile` / `content-profile` / `performance-benchmark-on-demand` | 表格、结构化数据、分析交付 |
| Presentations | available | `visual-profile` / `quality-inspection` | 创建、编辑、导出、预览 PPTX |
| PDF | available | `content-profile` / `visual-profile` / `quality-inspection` | PDF 读取、生成、核验 |
| Canva | available | `visual-profile` | 品牌演示稿、素材尺寸改写、设计翻译 |
| Figma | available | `visual-profile` / `development-profile` | 设计稿、组件库、设计系统映射 |
| HyperFrames by HeyGen | available | `visual-profile` | HTML/动画/视频合成与渲染 |
| Remotion | available | `visual-profile` | React 程序化视频 |

## 军团准入默认挂载面

下列插件已经被无极军团正式承认为“可按需挂载”的主链能力面：

| 插件 | 军团状态 | 规则 |
|---|---|---|
| Browser | admitted | 默认浏览器与前端验收能力，可由视觉/开发/情报/质检按需挂载 |
| Documents | admitted | 默认文档工件能力，可由内容/情报/审计按需挂载 |
| Spreadsheets | admitted | 默认结构化数据能力，可由数据/情报/内容/性能按需挂载 |
| Presentations | admitted | 默认原生 PPTX 能力，可由视觉/质检按需挂载 |

说明：

- “军团准入”表示已经进入当前主链默认插件集合。
- 其余宿主可用插件，如果没有被列在这里，就代表“当前不是默认挂载面”，只能按单案评审。

## 按需单案评审面

这些能力不是空候选，也不是默认挂载。它们的状态是：宿主或生态中有，但只有当任务真实缺口出现时，才允许由主链发起单案评审。

| 能力面 | 当前状态 | 边界 |
|---|---|---|
| Figma | restricted-boundary | 只在读取设计稿、组件库、Code Connect 或设计系统映射时启用 |
| Canva | restricted-boundary | 只在品牌演示稿、社媒尺寸改写、设计翻译等明确场景启用 |
| PDF | restricted-boundary | 只在 PDF 视觉核验、抽取、生成需求出现时启用 |
| HyperFrames by HeyGen | restricted-boundary | 只在明确视频/动画/HTML 合成需求时启用；不得变成通用运行时 |
| Remotion | restricted-boundary | 只在明确程序化视频需求时启用；不得变成默认视频主链 |

## 来源池或拒绝边界

以下对象不再保留“候选台账”，结论已经固定：

| 对象 | 裁决 | 边界 |
|---|---|---|
| `huashu-design` | source-pool-only | 只蒸馏为视觉原子，不作为本机插件或独立入口安装 |
| `OMX / oh-my-codex` | source-pool-only | 只保留 hook、HUD、sidecar 生命周期经验；拒绝外部 runtime shell |
| `baoyu-image-gen` | source-pool-only | 只保留提示词批处理和参考图规范；默认生图仍走主链模型路由 |
| `dbs-business-toolbox` | source-pool-only | 只保留商业诊断 recipe 碎片，归 content/intel/evolution 按需吸收；拒绝工具箱壳和第二业务分析运行时 |
| `baoyu-url-to-markdown` | restricted-boundary | 仅限 guarded capture 和选段摘要，不回灌整页 |
| `baoyu-youtube-transcript` | restricted-boundary | 仅限时间戳证据与章节摘要，不默认回放全文 |
| `baoyu-electron-extract` | restricted-boundary | 需要明确授权目标应用，并先过保卫科/合规 |
| `baoyu-post-to-x/wechat/weibo` | restricted-boundary | 只允许草稿、清单、格式适配；禁止静默登录、cookie/session 保留、自动发布 |
| `GitHub` 外部插件面 | reject | 当前宿主未作为军团默认插件准入；GitHub 相关需求优先由 `intelligence-profile` 搜索与现有工具处理 |
| `Supabase` / `Vercel` / `Sentry` / `Linear` / `Notion` / `Readwise` / `HeyGen` / `Cloudinary` / `Hugging Face` | reject | 当前不保留 standing 候选台账；未来如遇真实缺口，再做单案评审 |
| `ppt-mcp` / `mcp-server-ppt` / `dom-to-pptx` / `html2pptx` / `Marp` / `reveal.js` | reject | 当前不作为默认或预备挂载面；PPT 主链已由 `native-pptx-master-route` 统管 |

## 调用规则

- 插件不直接响应用户，必须回到单一主链收口。
- 未被军团准入的宿主插件，不得在路由报告里冒充“默认挂载”。
- 宿主可用但未准入的插件，只能在真实缺口出现后单案评审。
- 插件结果不得绕过 `white-hat`、`guard-office`、`audit`、`quality-inspection` 的既有门禁。
- `huashu-design` 永远不进入“已准入插件”；它只以视觉原子形式存在。
