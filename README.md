# 无极军团 Codex 版 / Wuji Legion for Codex

> 阿极统一入口 + 参谋本部调度 + 女娲按需补位 + 白帽前置封驳 + 多师局协同执行

## 最近更新

- `2026-05-31 v9.1`
  - 首页重新恢复为“无极军团商品介绍”口径
  - 明确 `presentation` 只是军团内部模块，不再代替整个军团
  - 补回军团用途、组织体系、师局结构、模块分层
- `2026-05-30 v9.0`
  - 完成军团主干整流
  - 收紧状态机、单主帅路由、成品交付规则
- 详细记录见：[CHANGELOG.md](E:\wuji-projects\wuji-legion-codex\CHANGELOG.md)

## 这是什么

无极军团 Codex 版，不是单一 skill，也不是单一提示词。

它是一个给 Codex 使用的轻量执行框架，用来把：

- 日常快答
- 搜索调研
- 代码开发
- PPT / HTML / 配图
- 文案内容
- QA / 安全 / 进化复盘

统一纳入同一套可调度、可纠偏、可交付的军团体系里。

它的目标不是“多角色越多越强”，而是：

- 省 token，高命中
- 高质高效
- 少走弯路
- 直接出最终结果

## 它主要解决什么问题

无极军团 Codex 版主要解决 5 类问题：

1. 普通 agent 容易一上来就长篇分析、空转查环境、消耗大量 token。
2. 多 skill 混用时，容易角色打架、流程冲突、输出半成品。
3. 复杂任务缺少统一调度，经常出现“工具很多，但没人负责到底”。
4. 成品任务容易变成“解释任务”，最后没有真正交付文件。
5. 旧规则不断叠补丁，最后前后冲突、命中率下降。

无极军团 Codex 版的解决方式是：

- 阿极统一对外
- 参谋本部只选主帅，不搞无效群聊
- 女娲只在缺能力时补位
- 白帽前置拦截错误路线
- 各师局按专业分工，但不抢对外入口

## 军团总结构

### 入口层

- **阿极**：统一对外入口，负责快答、短报、结果回执

### 调度层

- **参谋本部**：负责状态判断、主帅路由、封驳标准、验收口径

### 补位层

- **女娲**：负责能力融合、专家补位、冲突消解、最小组队

### 纠察层

- **白帽 / 质监局**：负责前置反对意见、质量审计、方向纠偏

### 执行层

- **各师局 / 专项模块 / 专家库 / 插件注册中心**

## 核心运行机制

### 1. 轻量状态机

只允许进入少数几个清晰状态：

- `FAST_REPLY`
- `CLARIFY`
- `SINGLE_COMMANDER`
- `LEGION_TASK`
- `BLOCKED`
- `DONE`

这套状态机的作用是：避免无意义开会，避免本来一句话能答却读半天环境，也避免复杂任务没有真正被接管。

### 2. 单主帅负责到底

无极军团 Codex 版默认不搞“多人轮流接管同一个成品”。

默认原则是：

- 一个任务先选一个主帅
- 主帅负责做到最终交付
- 其他能力只按需补位
- 如果路线错了，旧路线立即作废，不继续打补丁

### 3. 白帽前置

白帽不是事后装饰，而是开工前就能拦截错误路线。

它重点拦截：

- 分析没透就开干
- 半成品冒充成品
- 模板硬塞
- 错用路线
- 乱码
- 路径不明

## 军团组织架构

下面这些不是“装饰设定”，而是当前仓库里真实存在的组织单元。

### 参谋与总控

- **参谋本部**：路由中枢
- **女娲**：能力融合与补位
- **白帽 / 质监局**：审计、纠偏、验收

### 第一师：内容体系

- 文案融合引擎
- PPT / HTML 内容结构化
- 标题、钩子、节奏、人味优化

文件：
- [content.md](E:\wuji-projects\wuji-legion-codex\units\content.md)

### 第二师：视觉体系

- PPT / HTML / 配图生产线
- 臧老师美化链
- 图片、预览、版式、风格统一

文件：
- [visual.md](E:\wuji-projects\wuji-legion-codex\units\visual.md)

### 第四师：开发体系

- Rust / TS / Python 开发
- 自动化
- CI / CD
- 工程质量门禁

文件：
- [dev.md](E:\wuji-projects\wuji-legion-codex\units\dev.md)

### 情报局

- 多引擎并行搜索
- GitHub / 社区 / 文档检索
- 可信度评分
- 结构化情报报告

文件：
- [intel.md](E:\wuji-projects\wuji-legion-codex\units\intel.md)

### 远征军调度室

- 低成本外派
- 批处理任务分发
- handoff 标准化

文件：
- [expedition.md](E:\wuji-projects\wuji-legion-codex\units\expedition.md)

### 安全局

- 代码安全
- 依赖漏洞
- 许可证检查
- 发布前封装安全

文件：
- [security.md](E:\wuji-projects\wuji-legion-codex\units\security.md)

### 进化部

- OODA 复盘
- 失败模式记录
- 规则与 skill 持续进化

文件：
- [auto_evolve.md](E:\wuji-projects\wuji-legion-codex\units\auto_evolve.md)

### 插件注册中心

- Browser
- Documents
- Spreadsheets
- Presentations
- PPT / HTML 候选工具裁决

文件：
- [plugins.md](E:\wuji-projects\wuji-legion-codex\units\plugins.md)

## 专家库

除了组织单元，无极军团 Codex 版还挂了专家库，用来给女娲做按需补位。

当前专家目录包含：

- `content`
- `visual`
- `dev`
- `intel`
- `qa`
- `security`
- `prompt`
- `staff`
- `archive`
- `comfyui`
- `evolve`
- `expedition`
- `proving`

专家来源既有人物型，也有岗位型，例如：

- 臧老师(PPT)
- MrBeast
- Paul Graham
- Steve Jobs
- John Carmack
- Ilya Sutskever
- UX Architect
- Security Engineer
- Trend Researcher

## 模块体系

无极军团 Codex 版是总框架，不等于某一个模块。

当前已经单独定型的重点模块之一是：

- `presentation`

但它只是内部模块，不代表整个军团。

## Presentation 模块

`presentation` 模块负责：

- 真 PPTX
- HTML 演示稿
- 配图 / 封面 / 插图等视觉产出

当前三条入口文件：

- [pptx_master.md](E:\wuji-projects\wuji-legion-codex\units\pptx_master.md)
- [html_slides_master.md](E:\wuji-projects\wuji-legion-codex\units\html_slides_master.md)
- [quick-imagegen.ps1](E:\wuji-projects\wuji-legion-codex\scripts\quick-imagegen.ps1)

它的定位是：

- 对内：作为军团的演示与视觉专项模块
- 对外：不抢阿极入口，不冒充整个无极军团

## 典型能力地图

| 需求 | 主链 |
|---|---|
| 普通问答 | 阿极 |
| 搜索调研 | 情报局 |
| 代码开发 | 第四师（开发） |
| 真 PPTX 续写 | `presentation` -> `pptx_master` |
| 从零做 PPT | 第二师（视觉）+ `pptx_master` |
| HTML 演示稿 | `presentation` -> `html_slides_master` |
| 生图/插图/封面 | `presentation` -> `quick-imagegen.ps1` |
| 质量验收 | 白帽 / 质监局 |
| 安全检查 | 安全局 |
| 进化复盘 | 进化部 |

## 与普通 skill 的区别

普通 skill 更像一把单用途工具。

无极军团 Codex 版更像一个带调度能力的产品化框架，它强调：

- 统一入口
- 清晰状态机
- 单主帅到底
- 白帽前置
- 多组织协同
- 最终成品交付

所以它不是“再多装几个 skill”，而是把 skill、专家、插件、组织规则整流成一套能长期使用的执行系统。

## 适用场景

适合：

- 想让 Codex 长期稳定工作的人
- 想减少空转分析和 token 浪费的人
- 想把 PPT、HTML、代码、内容、调研统一纳入同一框架的人
- 想保留角色体系，但又不想真的让多个角色轮流表演的人

不适合：

- 只想要单一极简提示词的人
- 不需要组织层、不需要质量门禁的人

## 安装

```powershell
Copy-Item .\GLOBAL_AGENTS.md C:\Users\Administrator\.codex\AGENTS.md -Force
Copy-Item .\SKILL.md C:\Users\Administrator\.agents\skills\wuji-legion\SKILL.md -Force
Copy-Item .\config.json C:\Users\Administrator\.agents\skills\wuji-legion\config.json -Force
Copy-Item .\units C:\Users\Administrator\.agents\skills\wuji-legion\units -Recurse -Force
Copy-Item .\experts C:\Users\Administrator\.agents\skills\wuji-legion\experts -Recurse -Force
```

## 关键文件

- [GLOBAL_AGENTS.md](E:\wuji-projects\wuji-legion-codex\GLOBAL_AGENTS.md)
- [SKILL.md](E:\wuji-projects\wuji-legion-codex\SKILL.md)
- [content.md](E:\wuji-projects\wuji-legion-codex\units\content.md)
- [visual.md](E:\wuji-projects\wuji-legion-codex\units\visual.md)
- [dev.md](E:\wuji-projects\wuji-legion-codex\units\dev.md)
- [intel.md](E:\wuji-projects\wuji-legion-codex\units\intel.md)
- [security.md](E:\wuji-projects\wuji-legion-codex\units\security.md)
- [expedition.md](E:\wuji-projects\wuji-legion-codex\units\expedition.md)
- [auto_evolve.md](E:\wuji-projects\wuji-legion-codex\units\auto_evolve.md)
- [plugins.md](E:\wuji-projects\wuji-legion-codex\units\plugins.md)
- [pptx_master.md](E:\wuji-projects\wuji-legion-codex\units\pptx_master.md)
- [html_slides_master.md](E:\wuji-projects\wuji-legion-codex\units\html_slides_master.md)
- [quick-imagegen.ps1](E:\wuji-projects\wuji-legion-codex\scripts\quick-imagegen.ps1)
