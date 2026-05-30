# 无极军团 / Wuji Legion

> 阿极统一入口 + 无极军团总框架 + 其中一部分是 presentation 模块

## 简介

无极军团是给 Codex 使用的轻量调度框架。

它不追求“多角色越多越强”，而是追求两件事：

- 少走弯路
- 直接出最终结果

核心思路：

- 普通问题：阿极极速短答
- 明确任务：参谋本部只选一个主帅负责到底
- 主帅缺能力：女娲补位
- 路线明显错误：白帽提前否决

## 核心原则

- 省 token，高命中
- 高质高效
- 先分析透，再动手
- 只交最终结果，不交半成品

## 当前最终形态

PPT / HTML 这部分现在不再冒充整个无极军团。

正确层级是：

- **无极军团**：总框架
- **presentation 模块**：军团内部负责演示与视觉产出的子体系

presentation 模块内部再自动分：

- 真 PPTX
- HTML 演示稿
- 图片/配图

## 内部蒸馏结构

### 1. 真 PPTX 内核

文件：
- [pptx_master.md](E:\wuji-projects\wuji-legion-codex\units\pptx_master.md)

适用：
- 续写现有 `.ppt` / `.pptx`
- 套模板补页
- 从零做可编辑真 PPTX

吸收来源：
- `ppt-master`
- `elite-powerpoint-designer`
- `presentation-skill`
- `academic-pptx-skill`
- `guizang-ppt-skill`

明确拒绝：
- 逐字稿硬塞
- HTML 冒充 PPTX
- 缩字号救场
- “Word 投影感”

### 2. HTML 演示稿内核

文件：
- [html_slides_master.md](E:\wuji-projects\wuji-legion-codex\units\html_slides_master.md)

适用：
- 从零做浏览器演示稿
- 强风格、高观感、可讲可演示的 deck

吸收来源：
- `frontend-slides`
- `frontend-slides-editable`
- `html-ppt-skill`
- `guizang-ppt-skill`
- `huashu-design`
- `open-design`

明确拒绝：
- 伪装真 PPTX
- 泛 AI slop
- 抽象问风格不出预览
- 演讲稿进观众层

### 3. 图片内核

来源：

- `imagegen`
- 本地快速生图脚本

目标：

- 普通出图最短路径
- 成功直接交付图片
- 失败再短报阻塞

## 对外表现

用户只看到：

- 阿极
- 无极军团
- 一个最终成品

如果任务属于演示与视觉产出，再由内部的 presentation 模块自动分流。

## presentation 模块执行原则

| 任务 | 主执行链 |
|---|---|
| 普通生图 | `imagegen` 快速链 |
| 真 PPTX 续写 | `Presentations template-following` + `pptx_master` |
| 从零真 PPTX | `elite-powerpoint-designer` + `pptx_master` |
| 从零 HTML slides | `html_slides_master` |
| 搜索调研 | 情报局 |
| 代码开发 | 项目原生主线 |

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
- [pptx_master.md](E:\wuji-projects\wuji-legion-codex\units\pptx_master.md)
- [html_slides_master.md](E:\wuji-projects\wuji-legion-codex\units\html_slides_master.md)
- [quick-imagegen.ps1](E:\wuji-projects\wuji-legion-codex\scripts\quick-imagegen.ps1)
