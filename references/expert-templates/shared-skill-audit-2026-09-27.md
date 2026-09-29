# `D:\下载\skill共享` 候选审查账（2026-09-27）

这是**55 个顶层目录的入口审查记录**，不是“55 个完整包已融合/已测试”。每项都有 `SKILL.md`；对入口内容、触发与工作流做了逐项筛查，重点核对若干脚本/资产与当前能力。后续逐项方法归属或不准入理由及入口哈希以 [来源决议账](source-decisions.json) 为准。嵌套 `ima-skills`/`lark-unified` 子 Skill、全部参考文件、脚本行为、许可证和 Windows 兼容性**尚未逐一完成**；提炼方法不等于安装完整包或取得外部授权。

## 已映射的 26 个冷来源：具体可借鉴片段与阻断

| 候选入口 | 取舍与继承方向（不是已融合声明） |
| --- | --- |
| `ai-comic-video-replication` | 已读其阶段依赖/确认规则：十阶段状态、已确认资产登记、定向重跑可给视频团队作增量；该流程第六步为**八列**小分镜，课程特定七列交付不能被它覆盖；原默认 GPT-Image/VideoGen 与 Agnes 偏好冲突，付费批量仍需授权。 |
| `ai-landing-page-generator` | 事实账、参考图忠实度、首屏审批和视觉 QA 补前端/视觉；不能把生产长图/PSD 固定为每个网页任务。 |
| `ai-studio-quote` | 简洁可转发的三档报价与待确认项；¥2,980/8,980/19,800 和案例是来源示例，不是用户报价。 |
| `colleague-skill` | Persona/Work Skill 分离、反馈版本演进可用于专家进化；拒绝默认采集飞书私聊、钉钉、邮件及 Claude Code 脚本。 |
| `customer-proposal-gen` | 需求诊断、资料完整度、先结论后展开；macOS 私有 Obsidian 路径、固定客户名单和“强制六图”不继承。 |
| `deck-generator` | 内容规格与风格预设可参考；Imagen/Google API 费用且整页图片不满足既定可编辑 PPTX 目标，不能替换当前入口。 |
| `expert-interview-extraction` | 六维追问、九字段知识单元、原话出处和演进分列可增强模板沉淀；不从不存在的访谈推断专家经验。 |
| `frontend-dev` | 设计/交互/QA 片段按场景借鉴；固定 React/Tailwind、字体禁令和强制生图不作为全军团规则。 |
| `html-ppt` | 模板、演讲者视图、键盘/FX 的具体互补资产需逐个许可和渲染检查；其 `render.sh` 写死 macOS Chrome。 |
| `image-to-video` | 定价快照、任务 ID、超时不重复计费的原则有价值；AI Hive 凭据/费用、固定模型不能偷偷取代 Agnes。 |
| `kdocs` | 金山云文档域路由和真实操作回执候选；需要独立鉴权/写权限，非通用知识库问答默认入口。 |
| `kdocs-skill` | 与 `kdocs` 近似但入口哈希不同，不能按名称视为同一版本；先做差异/平台核对，避免双路由。 |
| `lark-unified` | 飞书领域细分、预检与按需子能力可参考；项目已有官方 Lark Skill，不叠加第二个常驻 CLI/认证层。 |
| `minimax-docx` | OpenXML 元素顺序、模板保真、渲染检查补当前文档能力；C#/.NET 安装、引用样本和真实编辑探针未准入。 |
| `minimax-xlsx` | 原文件编辑完整性、公式/格式复核补数据/文档专家；逐单元 XML 修改和依赖未做 Windows 行为验证。 |
| `new-project-launcher` | 源数据盘点、假设标注、跨制品一致性值得继承；其 macOS/WorkBuddy 路径、默认 35% 业务阈值、合同条款不能泛化。 |
| `ppt-generator-skill` | 受众/目的/语言/版式采集可补 PPT 专家；来源中举例的市场数字不可直接采用，“每页 80–150 字”不应固定。 |
| `ppt-master` | 项目状态、版式/图表/品牌索引、SVG→PPTX 资产与质检已有融合路径；来源强制全程串行、禁止子代理与本军团条件并发冲突，不能覆盖当前能力。 |
| `precise-image-text-editor` | 目标区域限定、尺寸格式/非目标区域核验和一次重试可增强图像专家；需有实际编辑工具与图像证据。 |
| `qq-email` | UID 回读、环境变量凭据和发送回执可补自动化；不能从“写周报”推断用户授权代发。 |
| `smart-page` | 场景/叙事/皮肤分离、单文件与 reduced-motion 可补工作台；腾讯上云/MCP/后台常驻服务不默认启动。 |
| `storyboard-generator` | 镜头节拍、景别变化和画面可生成性可补分镜；其六列 Markdown 不得覆盖课程七列 Excel 的明确交付。 |
| `wechat-article-pro` | 案例—观点—结论结构可补写作；默认 3000–5000 字、模仿具名作者及自动保存草稿不成为通用约束/授权。 |
| `weekly-meeting-dashboard` | DOCX→指标 JSON→交互看板与脱敏可补复盘；macOS `textutil`、具体周会正则和 CDN 离线问题需适配。 |
| `xmind` | XMind 双格式读写/局部更新可成为按需资产；其 `/tmp` 会话缓存非 Windows 运行约定，不常驻记忆。 |
| `投放素材生成器` | 三平台差异、字数与画面建议可补内容运营；视频号默认“中老年”和特定价值主张是原案例，不泛化。 |

`html-ppt` 与 `html-ppt-skill` **入口 SHA-256 相同**（`289E1CF2C0FA…`），且两包有 MIT `LICENSE`，但包大小分别约 10.9 MB/0.14 MB，不能只凭入口哈希说资产完全重复。项目内蒸馏快照入口哈希为 `803788C9CF76…`，共享版与它也**不是同一版本**；`ppt-master` 项目快照入口 `5AFCC0F84E7…`（约 99.9 KB）与共享版 `E493180F479D…`（约 42.1 KB）亦不同。不能直接覆盖。共享入口提炼方法已进入单一专家目录，但 `source_ids` 只引用准入后的真实可调用资产。

## 其余 29 个入口：不能因“不在原 26 个映射”而不看

| 候选入口 | 初判与需继续核查的差异 |
| --- | --- |
| `做课` | 学员视角六块方案，可补内容团课程设计；不强制把讲师稿改成学员稿。 |
| `courseware-builder` | Show–Tell–Do–Test 与 WPS 输出；情绪化“90%”宣称、WPS 账号操作需改。 |
| `courseware-gen` | 七问后六段课件；“七项缺一即停”与 PonyTail/用户已给足信息时直接做冲突。 |
| `courseware-generator-pro` | 同一七问/STDT 近似变体；需差异合并，不另造三个课件专家。 |
| `excalidraw-diagram` | `.excalidraw` 结构与 PNG 渲染验收是可选专业原子；不令方案默认六图。 |
| `gog` | Google Workspace CLI 的服务边界；无本任务授权/凭据，不准入。 |
| `gzh-typeset` | 公众号内联 HTML 的红线校验、三主题组件有增量价值；与写作/发布分离，不自动发稿。 |
| `html-ppt-skill` | 与 `html-ppt` 同入口，不等于同资源；先比许可证、可用主题/脚本/模板，不双计融合。 |
| `humanizer` | 活人感与机械句式诊断可补当前写作能力；保留原意，不强制第一人称或凭空经历。 |
| `ima-skills` | 笔记/知识库模块分流、UTF-8 边界；嵌套模块与账号 API 需授权/测试，不能常驻。 |
| `impeccable` | 设计上下文和可访问性检查补当前视觉能力；“绝不用某字体”等偏好不是用户通用约束。 |
| `khazix-writer` | 来源账已归写作；真实体验不可编造是应保留的能力，不被课程文案覆盖。 |
| `mubu-integration` | 大纲读写与软删除可按明确幕布任务接入；密码、Token 缓存、Linux `fcntl` 不默认迁入。 |
| `nano-banana-pro` | 图片草稿→迭代→终稿经验；Gemini API Key/费用与当前媒体供应商策略需单独决议。 |
| `nano-banana-pro-product-image` | 产品图/参考图控制有针对性；AI Hive 费用/账号不默认激活。 |
| `notebooklm-studio` | 学习资料→测验/播客等任务可用候选；Google 会话、上传外部资料和长轮询需明确授权。 |
| `obe-course-outline` | 可测成果→教学活动→评价矩阵补课程专家；不是所有课程都默认 32 学时/高校场景。 |
| `obsidian` | Markdown vault 与链接维护候选；入口写死 macOS vault 位置，Windows 需定位用户指定库。 |
| `openai-image-gen` | 批量图/图库候选；API Key/费用，不替代 Agnes，也不自动批量。 |
| `pptx` | **实际 `name: find-skills`，不是 PPTX 制作**；目录名误导，不能当 PPT 资产。 |
| `pptx-generator` | JSON 驱动 11 种可编辑页型，可与当前 PPT 原生资产比较；需真实渲染/许可/依赖探针。 |
| `skill_2053084036212973568` | 腾讯文档 MCP 与幻灯片编辑分流；连接器/权限未验证，不能因目录名并入文档专家。 |
| `tencent-meeting-skill` | 会议读写能力候选；其“遇错即主动反馈”与须二次确认冲突，不能自动外发。 |
| `tencent-news` | 新闻/事实核查/天气/高考细分；需官方 CLI、Key、时效与高风险正确性门禁。 |
| `training-mindmap-generator` | 课程大纲→模拟练习数据→Office 制品的互补原子；模拟数据必须醒目标识，不混入真实事实。 |
| `tutor-skills` | 学习测验/掌握度可补知识专家；持久化 StudyVault 要用户指定，不默认写记忆。 |
| `wechat-article-search` | 微信文章标题/日期/链接检索候选；站点条款和反爬限制，不默认爬取。 |
| `wechat-publisher` | 微信草稿封面/标题校验候选；账号写入/图片上传需明确授权，不能与写作等同。 |
| `wecom-unified` | 企业微信多域路由候选；账号认证、跨域权限、机器 ID 输出规则不能整体覆盖阿极。 |

## 下一道门：不能把本表冒充完成

本表记录 55 个入口的初步差异；当前 55 项决议及 49 项方法提炼见 `source-decisions.json`，不表示 55 个完整包的逐资源、逐脚本、逐许可证审计。课程设计等方法只在所选专家内按需注入，不取代当前 `source_ids`，也不把上云/邮件/发布/生成等外部副作用随模板自动授权。
