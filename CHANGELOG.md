# 更新日志 / Changelog

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
