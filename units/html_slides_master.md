# HTML Slides Master

## 定位

只做浏览器演示型成品。

适用范围：

- 从零做高观感 HTML slides
- 做可讲、可演示、可预览的 deck
- 需要强风格、强节奏、强展示感的浏览器演示稿

不负责：

- 续写现有 PowerPoint 模板
- 冒充真 PPTX

## 蒸馏后的保留能力

### 来自 `frontend-slides`

- 固定 16:9 舞台
- 先出 3 版风格预览
- show, don't tell
- 低密度 / 高密度分流

### 来自 `frontend-slides-editable`

- 生成后仍可继续浏览器内编辑
- 适合评审后继续微调
- 不是只读死产物

### 来自 `html-ppt-skill`

- 主题库
- 版式库
- 演讲者模式
- notes 容器

### 来自 `guizang-ppt-skill`

- 有限主题
- 风格收束
- 先锁用途和受众
- 结构化模板意识

### 来自 `huashu-design`

- 反 AI slop
- 先看方向再扩写
- 品牌资产优先
- “先给可见物，再快速迭代”

### 来自 `open-design`

- 交互式方向选择
- 设计系统意识
- skill + design system 分层
- 先锁 brief，再出 artifact

## 强制主线

```text
内容输入
-> audience + purpose + density
-> 3 style previews
-> style pick
-> section outline
-> full HTML deck
-> browser preview
-> QA
```

## 硬门

- 交付物必须明确是 HTML deck
- 固定 16:9 stage，禁止移动端重排 slide 内容
- 默认先给 3 个首屏风格预览
- 风格预览必须像真实首页
- 演讲稿放 notes，不放在观众可见区

## 密度模式

### `speaker-led`

- 一页一个主观点
- 大字、大留白、少字
- 更强调节奏和演讲感

### `reading-first`

- 允许更高信息密度
- 但不允许拥挤、滚动、重叠
- 更强调自解释性

## 风格策略

默认走三选一：

- 1 个安全主题
- 1 个大胆主题
- 1 个 wildcard

禁止默认落入：

- 紫色白底 AI 套路
- 千篇一律 SaaS 卡片网格
- 系统字体
- 为动而动

## 内容规则

- 讲稿进 notes
- 页面正文只保留观众需要看到的内容
- 一屏必须完整看完，不允许内部滚动
- 风格必须鲜明，不能像模板拼接

## QA 清单

- 是否是单文件或自洽工程
- 是否固定 16:9
- 是否无溢出、无滚动、无重叠
- 是否桌面和手机视口都可正常预览
- 是否风格足够鲜明
- 是否 notes 与观众层分离

## 明确剔除

- HTML 冒充 PPTX
- 预览阶段只给抽象描述不给视觉样本
- 演讲稿直接露出给观众
- 过度通用、过度保守、模板味太重
