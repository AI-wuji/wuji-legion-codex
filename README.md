# 无极军团 Codex 版 / Wuji Legion for Codex

> 阿极统一入口 + 参谋本部路由 + 女娲按需补位 + 白帽前置封驳

## 最近更新

- `2026-05-31 v9.1`
  - 层级纠偏完成
  - 明确：`presentation` 只是无极军团内部模块
  - 首页、规则、skill、日志统一回到“先军团、后模块”的口径
- `2026-05-30 v9.0`
  - 完成军团主干整流
  - 收紧状态机、单主帅路由、成品交付规则
- 详细记录见：[CHANGELOG.md](E:\wuji-projects\wuji-legion-codex\CHANGELOG.md)

## 简介

无极军团 Codex 版是给 Codex 使用的轻量执行框架。

它不追求“角色越多越强”，只追求两件事：

- 少走弯路
- 直接出最终结果

## 军团结构

- **阿极**：统一对外入口，负责快答、短报、结果回执
- **参谋本部**：只负责分拣、路由、封驳、验收标准
- **女娲**：只在主帅缺能力时补位，不默认进场
- **白帽**：前置否决低质路线、乱码、硬塞模板、空成品

## 运行方式

- 普通问题：阿极极速短答
- 明确任务：参谋本部只选一个主帅负责到底
- 复杂任务：按需补位，但不为“显得厉害”而开会
- 路线错了：立即作废旧路线，按新规划重走

## 核心原则

- 省 token，高命中
- 高质高效
- 先分析透，再动手
- 只交最终结果，不交半成品
- 知道就说知道，不知道就查，不确定就明说

## 当前内核

- 轻量状态机
- 单主帅执行
- 白帽前置封驳
- 工具懒加载
- 成品优先交付

## 内部模块

无极军团 Codex 版下面可以挂多种专项模块，`presentation` 只是其中一个。

当前已整流完成、并单独定型的模块有：

- `presentation`：负责 PPT / HTML / 配图类产出
- 其他模块：继续保留军团总框架接入，不被 presentation 覆盖

## Presentation 模块

这个模块只负责演示与视觉产出，不代表整个无极军团 Codex 版。

当前只保留三个入口：

- [pptx_master.md](E:\wuji-projects\wuji-legion-codex\units\pptx_master.md)
- [html_slides_master.md](E:\wuji-projects\wuji-legion-codex\units\html_slides_master.md)
- [quick-imagegen.ps1](E:\wuji-projects\wuji-legion-codex\scripts\quick-imagegen.ps1)

## 模块路由示例

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
