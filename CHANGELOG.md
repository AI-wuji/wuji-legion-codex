# 更新日志 / Changelog

## 2026-06-01 v10.6 hard-gate completion

- `wuji-cli` 新增 `reference-guard`：输出路径不得覆盖参考原件或写入参考目录。
- `wuji-cli` 新增 `workflow-guard`：确定性检查复杂任务工作流工件完整性。
- `wuji-cli` 新增 `claim-guard`：完成、成功、通过、已融合、已生成声明必须附验证证据。
- `wuji-cli` 新增 `time-guard`：非代码任务 10/15/30 分钟无可验产物时直接 `NO-GO`。
- `wuji-cli` 补齐 `task`、`sync`、`audit`、`bench`、`preview`、`asset-map`、`pptx-audit`，执行底座文档声明的核心命令全部落成可执行实现。
- 新增 `scripts/test-wuji-cli.ps1`，回归验证四个通用 gate 与两道 PPTX gate。
- 文档明确区分已落地子命令和后续方向，禁止把规划能力冒充已实现能力。

## 2026-06-01 v10.6

### Pilot Page 快速闭环

- PPT/HTML 视觉成品采用“先试投一页，再批量”的机制，避免整套生成完才发现风格和结构错误。
- 三张表通过后先生成 1 页最高风险/最高密度/最能代表风格的 `pilot-page`，导出 `pilot-preview` 并记录 `pilot-score`。
- pilot 只检验风格、结构、素材复用、插图策略和可编辑路线，不做工具链连环测试。
- pilot 最多两轮；两轮不过线，必须换路线或短报阻塞，不得继续批量生成。
- 执行底座新增 `pptx-batch-gate`：缺 `pilot-page`、`pilot-preview`、`pilot-score` 时直接 `NO-GO`，不允许开始批量生成。
- 顶层 `GLOBAL_AGENTS.md` / `SKILL.md` 明确写入批量前必须过 `pptx-batch-gate`，避免新窗口只记得 pilot 规则、忘记硬门禁。

## 2026-06-01 v10.5

### PPT前置硬门禁

- 非代码任务新增 10/15/30 分钟分级熔断：10 分钟无主成品要短报或切回，15 分钟仍在查接口/试 API/找字体/小原型则熔断，30 分钟无可验成品判定节奏失败。
- 口头说“直接生成、马上开做、不再验证”不算执行；实际动作仍在读文档、查接口、试 API、小原型时，按绕路处理。
- 参考 PPTX 任务生成前必须锁定 `reference-frame-map`、`reusable-asset-map`、`illustration-plan`，缺任一项不得批量生成。
- 参考 PPTX 同时视为风格系统和素材库：表格、图标、卡片、图框、流程箭头、章节页、背景装饰、教学插图、案例图、公式图、image2 生图资产都要纳入复用或同风格补图判断。
- 真 PPTX 路线必须在开工前锁定为可编辑 PPTX；每页渲染成整图再塞回 PPT 的方案不得启动。
- 执行底座新增 `pptx-preflight`、`pptx-audit`、`asset-map`、`time-guard` 方向，用确定性门禁在生成前拦截错误路线。

## 2026-06-01 v10.4

### 执行底座

- 新增 `执行底座 / 执行底座主帅`，定位为无极军团通用确定性执行底座师团。
- 执行底座负责 `wuji-cli`、guard、task、sync、audit、workflow、beep、bench、preview调度等稳定工具链。
- 明确边界：执行底座不替代开发主帅，不做普通业务软件主线，不负责内容、视觉、PPT设计、写作、蒸馏判断或参谋决策。
- 普通 Go/Tauri/后端/插件开发仍归开发主帅；无极军团自身执行器和稳定工具链归 执行底座主帅。

## 2026-06-01 v10.3

### 非代码交付先执行铁律

- PPT、文档、图片、HTML演示稿、逐字稿、海报、课程材料等非代码成品任务，默认先直接生成成品。
- 禁止开工前反复验证“能不能做”；只有遇到真实错误、缺文件、缺权限、乱码、模板损坏、路径冲突或可能覆盖参考原件时，才允许局部验证修复。
- PPT 类任务顺序固定为：读取输入 -> 映射页型 -> 直接生成完整成品 -> 导出预览/QA -> 必要时局部修复。
- 同一能力通过一次后不得重复验证；10 分钟没有主成品文件，必须切回主线生成或短报阻塞。

## 2026-06-01 v10.2

### 可审计工作流机制

- 源码级核验 `DannyMac180/skills`，checked commit `5695fa19b9d39b8270025e79633b49a8b863f9a2`，许可证 MIT。
- 裁决为 `absorb`：吸收动态工作流工件、packet 切片、结果收集和验证脚手架；不新增外部入口，不照搬上游命名。
- 复杂 `LEGION_TASK` 不能只口头说“参谋本部已接管”，必须留下目标、成功标准、任务切片、验证结果和最终收口。
- 简单任务、普通生图、轻量问答、单文件小改不启用工作流工件，避免 token 噪音。
- 新增 `scripts/wuji_workflow.py`，用于生成、切片、收集和验证无极工作流工件。

## 2026-05-31 v10.1

### 参考文件只读铁律

- 新增全局文件安全铁律：用户提供、点名、上传、要求“参考/借鉴/按照/对照”的文件，默认全部只读。
- 允许复制参考文件到任务工作区做审计、预览或格式转换，但不得反向覆盖原件。
- 生成物、修复版、重做版必须另存为新文件；白帽发现输出路径指向参考原件时必须否决。

## 2026-05-31 v10.0

### LEGION_TASK 触发口径修正

- 修正参谋本部接管条件：进入 `LEGION_TASK` 的首条回复必须使用接管格式。
- 不再只依赖用户明确喊“激活无极军团”；只要任务本身需要多能力协作，也必须进入 `LEGION_TASK`。
- 典型场景：根据大纲参考旧 PPT，生成新 PPT 和逐字稿。

## 2026-05-31 v9.9

### 提示音后台延迟

- `beep.ps1` 新增 `-SpawnDelayed` 和 `-DelayMs` 参数。
- 收尾提示音改为最终回复前调度后台进程，最终回复发出后延迟响铃。
- 默认推荐：`.\scripts\beep.ps1 complete -SpawnDelayed -DelayMs 1200`。

## 2026-05-31 v9.8

### 提示音时机修正

- 修正任务完成提示音规则：提示音必须尽量作为最终回复前的最后一个工具动作。
- 避免在验证、扫描或中间步骤提前响铃，后续又继续处理，导致用户误以为任务已经结束。

## 2026-05-31 v9.7

### 任务完成提示音

- 增强 [beep.ps1](E:\wuji-projects\wuji-legion-codex\scripts\beep.ps1)，用临时 WAV 播放提示音，并支持 `complete`、`error`、`notify` 三种模式。
- 全局规则新增收尾提醒：非 `FAST_REPLY` 任务在完成、阻塞或失败收口前，先尝试播放提示音，再输出最终结果。
- 提示音失败不阻塞交付，避免因为系统音频不可用影响最终回复。

## 2026-05-31 v9.6

### 质量整流审查

- 为根 `SKILL.md` 增加标准 frontmatter，保证作为 Codex skill 安装时可被正确识别。
- 清理 `scripts/__pycache__` 生成缓存，并确保安装/恢复/同步脚本排除缓存、日志和临时产物。
- 将 `trae/commander/SKILL.md` 从旧项目专用工兵改为无极军团远征军通用工兵，避免旧项目语境污染当前体系。
- 重写 `push-to-github.ps1`，删除旧 C 盘硬编码、破坏式覆盖和过期 V3 提交信息，改为当前仓库内安全提交/推送。
- 重写 `wuji-install.ps1`、`wuji-restore.ps1`、`wuji-e-sync.ps1`、`wuji-e-backup.ps1`，统一 v9.6 路径策略和干净复制规则。
- 修复 `target_range.py skill <dir>` 把目录当代码文件读的问题，新增 skill 目录、frontmatter、乱码和占位扫描。
- 清理 `units/proving_ground.md` 重复协作段，并把 Marp 定位从“草稿链路”改为“快速预览链路”。

### 验证结果

- 专家生成：15 张高密度专家卡。
- PowerShell 语法：全部通过。
- Python 编译：全部通过。
- 打靶场 skill 扫描：116/116 通过。
- 旧项目残留扫描：无旧项目关键词、旧版本号、占位拉丁文、待办标记、修复标记和替换乱码符残留。

## 2026-05-31 v9.5

### 师团主帅内核蒸馏

- 将专家库从 44 张主责专家卡继续压缩为 15 张高密度专家卡。
- 新结构定型为：`师团万能主帅入口 -> 内置多模式 -> 独立白帽/质检/安全/合规审查`。
- 明确边界：压缩的是单个师团内部入口，不是把整个无极军团压成一个超级大脑。
- 内容师团合并为 `内容主帅`，内部覆盖小说、剧本、分镜、教程、计划书、商业营销方案、卖点提炼、短内容和人味改稿。
- 视觉师团合并为 `视觉主帅`，内部覆盖真 PPTX、HTML 演示、UI 页面、数据可视化、视觉叙事和配图。
- 开发师团合并为 `开发主帅`，内部覆盖 Go/Tauri、前端、小程序、ComfyUI 插件、AI 工程、自动化和原型。
- 白帽、质检、安全、合规、性能基准保持独立，不并入执行主帅，避免自己写自己审。
- 重写 `content`、`visual`、`dev`、`qa`、`security`、`intel`、`comfyui`、`nuwa`、`staff`、`prompt_engine`、`expedition`、`archive`、`auto_evolve`、`distillation` 规则文件，避免补丁式叠加。

### 本轮源码核验

- `openai/skills`：commit `a8924c2`
- `anthropics/skills`：commit `da20c92`
- `addyosmani/agent-skills`：commit `6ce0298`
- `github/awesome-copilot`：commit `9b74459`
- `marketingskills`：commit `7f4af1e`
- `humanizer`：commit `a2ace14`
- `powerpoint-skill`：commit `a39cd8c`
- `ppt-master`：commit `232415d`

### 蒸馏裁决

- 吸收官方 skill 规范的轻入口、按需加载、bundled resources 和上下文节省机制。
- 吸收 Addy agent-skills 的阶段识别、上下文工程、薄切片、五轴 review 和安全门禁。
- 吸收 GitHub awesome-copilot 的大规模技能索引、Rust/QA/安全细分规则和资源分层。
- 吸收 marketingskills 的产品/受众/定位先行和营销技能互相关联机制。
- 吸收 humanizer 的 AI 写作痕迹识别和作者声音匹配机制。
- 吸收 powerpoint-skill 的视觉优先、密度边界、重叠检查和预览验证机制。
- 不吸收外部组织命名，不新增重复入口，不让简单任务变慢。

## 2026-05-31 v9.4

### 专家库瘦身蒸馏

- 将专家库从 70 张卡压缩为 44 张主责专家卡。
- 新增 [experts/INDEX.md](E:\wuji-projects\wuji-legion-codex\experts\INDEX.md)，作为女娲补位和人工查找的唯一专家索引。
- 重写 [gen_experts.py](E:\wuji-projects\wuji-legion-codex\scripts\gen_experts.py)，默认删除旧重复专家卡并重新生成“短操作合约”。
- 合并重复人物和重复能力：
  - `Image Prompt Engineer` 合并进 `Prompt Architect` 与 `ComfyUI Pipeline Engineer`
  - `Risk Assessor` 合并进 `Reality Checker`
  - `Book Co-Author` / `Podcast Strategist` / `Narratologist` 合并进 `Narrative Architect`
  - `Short Video Coach` 合并进 `MrBeast`
  - `Brand Guardian` / `UI Designer` / `UX Architect` 合并进 `impeccable引擎`
  - `Whimsy Injector` 合并进 `Visual Storyteller`
- 全局规则新增铁律：专家不以量取胜；蒸馏不是叠加，新增入口前必须先判断能否合并进现有主责位。
- 同步更新 `nuwa`、`content`、`visual`、`comfyui`、`dev`、`intel`、`security`、`qa` 的专家索引，避免旧专家名继续制造误触发。
- 清理 ignored `output/` 生成产物，减少工作区占用。

## 2026-05-31 v9.3

### 进化部蒸馏师团

- 新增 [distillation.md](E:\wuji-projects\wuji-legion-codex\units\distillation.md)
- 新增进化审计入口，后续已收口并并入 [进化主帅](E:\wuji-projects\wuji-legion-codex\experts\evolve\进化主帅.md)
- 进化部从“复盘学习”升级为“复盘 -> 查源 -> 蒸馏 -> 验证 -> 入库”
- 女娲补位前必须检查蒸馏台账，未核验外部能力不得进默认主链路
- 白帽新增 skill 蒸馏退回红线：没看官方源、没看源码、没记录版本、没验证，都不能说已融合
- 试验场新增蒸馏候选验证表，防止 skill 叠加污染主流程

### 已核验官方源

- `openai/skills`：commit `a8924c2`
- `anthropics/skills`：commit `da20c92`
- `addyosmani/agent-skills`：commit `6ce0298`
- `AMAP-ML/SkillClaw`：commit `1f96ec8`
- `cft0808/edict`：commit `14a2075`
- `vercel-labs/agent-skills`：commit `1801156`

### 蒸馏裁决

- 吸收官方 skill 规范里的按需加载、metadata、许可证登记和 progressive disclosure
- 吸收 SkillClaw 的任务后进化、候选验证、去重、版本轨迹
- 吸收 Edict 的强制封驳、状态审计和可视化追踪思路
- 不吸收外部组织命名，不新增重复入口，不让简单任务变慢

## 2026-05-31 v9.2

### 第四师工程流程线

- 源码级查看并蒸馏 `addyosmani/agent-skills`
- 参考源码提交：`6ce0298`
- 吸收进无极军团：
  - meta-skill 阶段路由
  - DEFINE / PLAN / BUILD / VERIFY / REVIEW / SHIP 生命周期
  - 薄切片实现
  - TDD / Prove-It bug 修复模式
  - 五轴 code review
  - 安全与发布门禁
- 保留无极军团命名与编制，不照搬上游命令体系

## 2026-05-31 v9.1

### 层级纠偏

- 纠正 `v9.0` 的错误收口
- 明确：
  - 无极军团是总框架
  - PPT / HTML 只是军团内部的 presentation 模块
- 不再把 presentation 模块冒充成整个无极军团

## 2026-05-30 v9.0

### 统一总 skill

- 此版本的“统一总 skill”提法已作废
- 问题：错误把 presentation 模块提升成整个无极军团本体

### 再次蒸馏

- 真 PPTX 内核继续吸收：
  - `ppt-master`
  - `elite-powerpoint-designer`
  - `presentation-skill`
  - `academic-pptx-skill`
  - `guizang-ppt-skill`
  - `huashu-design`
  - `open-design`
- HTML 演示内核继续吸收：
  - `frontend-slides`
  - `frontend-slides-editable`
  - `html-ppt-skill`
  - `guizang-ppt-skill`
  - `huashu-design`
  - `open-design`
- 图片内核统一收口到快速生图主链

## 2026-05-30 v8.0

### 整流重写

- 重写 [GLOBAL_AGENTS.md](E:\wuji-projects\wuji-legion-codex\GLOBAL_AGENTS.md)，去掉补丁式叠加冲突
- 重写 [SKILL.md](E:\wuji-projects\wuji-legion-codex\SKILL.md)，统一阿极入口、状态机、单主帅路由
- 重写 [README.md](E:\wuji-projects\wuji-legion-codex\README.md)，明确真 PPTX 与 HTML 两条母 skill

### 母 skill 定型

- 定型 [pptx_master.md](E:\wuji-projects\wuji-legion-codex\units\pptx_master.md)
- 定型 [html_slides_master.md](E:\wuji-projects\wuji-legion-codex\units\html_slides_master.md)
- 明确 PPT 任务先分流：
  - 真 PPTX
  - HTML 演示稿
  - 图片版 PPT

### 融合依据

- 真 PPTX 线吸收：
  - `ppt-master`
  - `elite-powerpoint-designer`
  - `presentation-skill`
  - `academic-pptx-skill`
  - `guizang-ppt-skill`
- HTML 演示线吸收：
  - `frontend-slides`
  - `frontend-slides-editable`
  - `html-ppt-skill`
  - `guizang-ppt-skill`
  - `huashu-design`
  - `open-design`

### 新铁律

- 先分析透，再动手
- 只交最终结果，不交半成品
- 全局规则和主 skill 禁止补丁式叠加
- 用户明确要求最终版时，不给 `v1`、`先用着`、`先来一版`

## 2026-05-30 v6.1

- Added `units/pptx_master.md`
- Added `units/html_slides_master.md`
- Split presentation work into two artifact lines:
  - true PPTX
  - HTML slide deck
- HTML slide skills can no longer stand in for existing PPTX template continuation

## 2026-05-28 v6.0

- 轻量状态机
- 单主帅内核
- 普通生图直连
- PPT 主链路前置
