# 无极军团 / Wuji Legion v9.1

> 阿极统一入口 + 无极军团总框架 + 内部 presentation 模块

## 一句话

无极军团不是 PPT skill，也不是 HTML skill。

它是总框架。

PPT / HTML 只是无极军团里的一个专业模块，负责演示与视觉产出。

## 模块划分

无极军团总层：

- 阿极入口
- 参谋本部路由
- 女娲补位
- 白帽封驳

其中一个子模块：

- `presentation`

这个 presentation 模块内部再分三条产线：

- 真 PPTX
- HTML 演示稿
- 图片/配图

## presentation 模块定位

presentation 模块只负责：

- 真 PPTX
- HTML 演示稿
- 配图与封面等演示配套视觉

它不等于整个无极军团。

## presentation 模块蒸馏来源

### 真 PPTX 线蒸馏来源

- `ppt-master`
- `elite-powerpoint-designer`
- `presentation-skill`
- `academic-pptx-skill`
- `guizang-ppt-skill`
- `huashu-design`
- `open-design`

### HTML 演示线蒸馏来源

- `frontend-slides`
- `frontend-slides-editable`
- `html-ppt-skill`
- `guizang-ppt-skill`
- `huashu-design`
- `open-design`

### 图片线来源

- `imagegen`
- 已实测的本地快速通道

## presentation 模块统一入口

- 用户只需要描述演示任务，不需要自己判断该走哪条子线。
- presentation 模块内部自动判断：
  - 有现成 `.ppt/.pptx`、模板页、前几页成品：走真 PPTX 续写
  - 目标是浏览器展示、动画演示、网页 slides：走 HTML 演示线
  - 只是配图、封面、插图：走图片线

## 真 PPTX 内核

保留的优点：

- 来自 `ppt-master` 的分阶段执行、真 PPTX 导出、规范锁定
- 来自 `elite-powerpoint-designer` 的高等级视觉层级与品牌感
- 来自 `presentation-skill` 的工作区、QA、证据化工作流
- 来自 `academic-pptx-skill` 的行动标题、证据优先、一页一结论
- 来自 `guizang-ppt-skill` 的风格收束
- 来自 `huashu-design` / `open-design` 的方向先行、brief 先锁、反 AI slop

统一后主线：

```text
input
-> artifact detect
-> slide-spec
-> design-brief
-> layout-plan
-> key-page preview
-> full PPTX
-> QA
```

统一后硬门：

- 逐字稿不是页面正文
- 模板不是填字容器
- 不允许 HTML 冒充 PPTX
- 不允许缩字号救场
- 不允许占位符残留
- 不允许 Word 投影感

## HTML 演示内核

保留的优点：

- 来自 `frontend-slides` 的 16:9 固定舞台、3 版预览、show don’t tell
- 来自 `frontend-slides-editable` 的生成后继续编辑
- 来自 `html-ppt-skill` 的主题库、布局库、presenter 模式
- 来自 `guizang-ppt-skill` 的有限主题和风格纪律
- 来自 `huashu-design` 的品牌资产优先与反模板味
- 来自 `open-design` 的 brief 锁定、方向选择、design system 视角

统一后主线：

```text
input
-> artifact detect
-> audience + purpose + density
-> 3 style previews
-> style lock
-> section outline
-> full HTML deck
-> browser QA
```

统一后硬门：

- 不伪装成真 PPTX
- 不移动端重排 slide 内容
- 预览必须像真实页面
- 讲稿只进 notes
- 不允许泛 AI slop

## 图片内核

保留的优点：

- 直接执行
- 最短链路
- 不先空转分析

统一后主线：

```text
prompt
-> quick imagegen
-> preview
-> file path
```

统一后硬门：

- 不先给长解释
- 不先只给提示词
- 不失败前乱切通道

## presentation 模块对外表现

虽然内部仍然区分真 PPTX、HTML、图片三条执行线，但这些分流对用户不可见。

用户看到的永远只是：

- 阿极
- 无极军团
- presentation 模块产出的最终成品

## 统一白帽规则

- 简单任务不展开白帽
- 复杂任务开工前只报 1-3 个关键风险
- 白帽可直接否决：
  - 分析未透就开干
  - 半成品冒充成品
  - 模板硬塞
  - 错用 HTML / PPTX 线路
  - 编码乱码
  - 路径不明

## 安装

```powershell
Copy-Item .\GLOBAL_AGENTS.md C:\Users\Administrator\.codex\AGENTS.md -Force
Copy-Item .\SKILL.md C:\Users\Administrator\.agents\skills\wuji-legion\SKILL.md -Force
Copy-Item .\config.json C:\Users\Administrator\.agents\skills\wuji-legion\config.json -Force
Copy-Item .\units C:\Users\Administrator\.agents\skills\wuji-legion\units -Recurse -Force
Copy-Item .\experts C:\Users\Administrator\.agents\skills\wuji-legion\experts -Recurse -Force
```

## 当前版本定位

- `v9.1`：纠正层级，presentation 退回为无极军团内部模块
- `v9.0`：曾错误把 presentation 提升成总 skill，现已作废
