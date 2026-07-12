---
name: 高顿 ComfyUI 流式直播课件
description: 一套以流体舞台、单命题页面和清晰演讲节奏构成的科技直播演示系统
colors:
  deep-space: "#030712"
  midnight: "#081426"
  navy-depth: "#0d1f35"
  panel: "#08111fc7"
  text-primary: "#f5f8ff"
  text-muted: "#9cb0d0"
  signal-cyan: "#84e7ff"
  signal-blue: "#7c8dff"
  signal-mint: "#56edc4"
  signal-gold: "#ffd986"
  signal-rose: "#ff92bc"
typography:
  display:
    fontFamily: "Microsoft YaHei, PingFang SC, Segoe UI, sans-serif"
    fontSize: "52px"
    fontWeight: 900
    lineHeight: 1.08
    letterSpacing: "0"
  headline:
    fontFamily: "Microsoft YaHei, PingFang SC, Segoe UI, sans-serif"
    fontSize: "34px"
    fontWeight: 800
    lineHeight: 1.16
    letterSpacing: "0"
  body:
    fontFamily: "Microsoft YaHei, PingFang SC, Segoe UI, sans-serif"
    fontSize: "14px"
    fontWeight: 400
    lineHeight: 1.75
    letterSpacing: "0"
  label:
    fontFamily: "Segoe UI, Microsoft YaHei, sans-serif"
    fontSize: "10px"
    fontWeight: 600
    lineHeight: 1.2
    letterSpacing: "0"
rounded:
  sm: "8px"
  md: "14px"
  lg: "20px"
  xl: "26px"
spacing:
  xs: "8px"
  sm: "16px"
  md: "24px"
  lg: "40px"
  xl: "64px"
components:
  stage-shell:
    backgroundColor: "{colors.panel}"
    textColor: "{colors.text-primary}"
    rounded: "{rounded.xl}"
    padding: "40px"
  stage-pill:
    backgroundColor: "{colors.midnight}"
    textColor: "{colors.signal-cyan}"
    rounded: "{rounded.xl}"
    padding: "12px 18px"
  action-button:
    backgroundColor: "{colors.signal-cyan}"
    textColor: "{colors.deep-space}"
    rounded: "{rounded.sm}"
    padding: "12px 18px"
---

# Design System: 高顿 ComfyUI 流式直播课件

## Overview

**Creative North Star: "流体直播舱"**

页面像一个为科技课程搭建的直播舞台：外层环境有连续流动的光和空间深度，内层 16:9 内容面稳定、清晰、可被直播截取。设计不模拟软件后台，而是借用控制舱的秩序感来组织演讲节奏。

每个页面只有一个视觉主语。章节页留白充足；洞察页使用不对称的文字与图形；对比页让差异直接可见；流程页让路径实际流动；行动页只留下可执行动作。

## Colors

主背景由深空蓝到海军蓝构成，青色代表新信号，薄荷色代表增长，金色代表被记住的高价值节点，玫瑰色只用于风险和阻断。

**The Signal Rule.** 高饱和颜色只出现在标题关键词、路径、峰值和当前状态，禁止整页所有边框同时发光。

## Typography

中文主标题使用高重量无衬线体，依靠字号和留白制造权威感。正文保持 14px 以上的逻辑字号，英文仅作为短标签，不承担核心信息。

**The Two-Second Rule.** 模糊视线后仍应先看到主命题，再看到解释图形，最后才看到辅助说明。

## Elevation

深度主要依靠流体背景、低对比面板和局部环境阴影建立。内容卡片不层层嵌套；舞台外壳是最高层级，内部模块只使用细微明度差。

**The Frozen Frame Rule.** 任意动画帧暂停时，文字和图形仍应清楚，不允许动效成为可读性的前提。

## Components

- **舞台外壳：** 固定 16:9 安全区，圆角 26px，四角信号标记和底部进度负责演出感。
- **章节转场：** 一个大标题、一句引导、一个低能量图形，不放数据卡片。
- **洞察页：** 左侧结论与短证据，右侧单一图形或关系图，避免四等分卡片网格。
- **对比页：** 两个明确世界并置，中间用一道可见的断层或切换轨迹连接。
- **流程页：** 节点之间必须有实际方向和动态路径，当前节点比其他节点更亮。
- **交互工作流：** tab 使用明确选择态；切换时内容淡入和位移，不改变整体布局尺寸。

## Do's and Don'ts

### Do

- 使用参考稿的舞台感、流体氛围、章节节奏和逐层揭示。
- 让标题足够大，让页面在直播缩略画面中仍可读。
- 用图形承担比例、关系和流程解释。
- 保留减弱动效模式，并保证键盘可用。

### Don't

- 禁止回到“密集小字、细线边框和重复 HUD 卡片组成的课程提纲”。
- 禁止每页都使用同一网格和同一信息密度。
- 禁止无意义的英文状态词、伪数据和持续扫描线抢占注意力。
- 禁止把所有内容包成卡片，更禁止卡片嵌套卡片。
- 禁止只换颜色不换叙事结构。
