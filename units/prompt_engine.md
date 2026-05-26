# 提示词局 — 通用/图像/故事板/视频 Prompt 工程

## 核心理念：prompt不是玄学，是工程

```
好的prompt = 清晰的角色 × 具体的任务 × 明确的格式 × 约束条件
```

---

## 模块一：通用Prompt原则（适用于所有场景）

### 结构公式
```
[角色] + [任务] + [上下文] + [格式要求] + [约束条件]
```

### 写作规范
| 原则 | 说明 |
|------|------|
| 1. 角色明确 | "你是资深XX专家" 比 "请帮我" 效果好 |
| 2. 任务具体 | "写一篇300字的产品介绍" > "介绍一下这个产品" |
| 3. 给例子 | one-shot/few-shot 显著提升质量 |
| 4. 限格式 | "输出Markdown" / "输出JSON" / "逐条列出" |
| 5. 约束兜底 | "不知道就说不知道" / "不要编造" |

---

## 模块二：图像生成Prompt（新增融合）

### GPT提示词理解与扩写分级

图像生成默认需要GPT做需求理解，但扩写强度按任务风险分级，避免普通生图变慢。

| 场景 | 扩写等级 | 处理方式 |
|------|----------|----------|
| 普通单图 | 轻扩写 | 补主体、风格、构图、尺寸、负面约束 |
| PPT封面/章节图 | 中扩写 | 结合PPT风格、留白、色调、文字承载区 |
| 品牌/产品图 | 深扩写 | 加品牌调性、材质、镜头、场景一致性 |
| 含文字图 | 深扩写+拆层 | 图片不直接承载关键文字，文字回到PPT/HTML层 |
| 多页PPT配图 | 批量一致性扩写 | 生成统一 `image-spec.json`，保证风格一致 |

### 通用结构
```
[主体] + [动作/姿态] + [环境/背景] + [风格] + [光线] + [构图] + [质量词]
```

### 风格词汇表
| 风格 | 关键词（中/英） |
|------|----------------|
| 写实 | 照片级/photorealistic, 8K, 细节丰富 |
| 二次元 | 动漫/anime, 赛璐珞/cel shading, 平涂 |
| 水墨 | 水墨风/ink wash, 留白/negative space |
| 油画 | 油画/oil painting, 厚涂/impasto |
| 赛博朋克 | cyberpunk, 霓虹/neon, 光污染 |
| 像素 | pixel art, 8-bit, 复古/retro |

### 质量修饰词
```
正向：masterpiece, best quality, highly detailed, 8K, sharp focus
负向：nsfw, lowres, bad anatomy, bad hands, extra fingers, blurry
```

### ComfyUI Prompt工作流
```
用户需求 → 结构填充 → 中英双语 → style alignment → 输入ComfyUI
```

### image2 / imagegen Prompt工作流
```
用户需求/slide-spec.json
    ↓
GPT理解页面意图：这张图服务哪一句结论？
    ↓
生成 image-spec.json：
  - use_case
  - asset_type
  - slide_id/page_id
  - subject
  - scene
  - style
  - composition
  - lighting
  - palette
  - negative_constraints
  - text_policy
    ↓
交给 image2 / imagegen 生成
    ↓
视觉部检查是否匹配整套PPT/HTML风格
```

### image-spec.json 最小格式

```json
{
  "slide_id": "S03",
  "use_case": "ppt-section-visual",
  "asset_type": "16:9 slide background",
  "page_message": "这一页要让用户理解AI可以统一调度多工具",
  "subject": "抽象的多agent协同网络",
  "style": "premium technical keynote, clean, not sci-fi cliché",
  "composition": "wide 16:9, center-left focal point, right side negative space for title",
  "palette": "match deck palette, dark graphite with muted cyan accent",
  "text_policy": "no embedded text, no letters, no logos",
  "avoid": "plastic gradients, random UI text, watermark, clutter"
}
```

---

## 模块三：故事板/分镜Prompt（新增融合）

### 分镜Prompt结构
```
[场景编号] - [镜头类型] - [画面描述] - [情绪/氛围] - [对白/VO]
```

### 镜头类型表
| 类型 | 效果 | 适用场景 |
|------|------|---------|
| 远景(LS) | 交代环境 | 开场/过渡 |
| 全景(FS) | 展示主体全貌 | 角色出场 |
| 中景(MS) | 对话/动作 | 常规叙事 |
| 特写(CU) | 表情/细节 | 情绪高潮 |
| 过肩(OTS) | 对话视角 | 双人对话 |
| 俯拍/仰拍 | 权力关系 | 紧张/压迫场景 |

### 分镜串联格式
```markdown
## 分镜序列

### Scene 1: [场景名]
**画面**: [详细视觉描述]
**镜头**: [镜头类型]  
**情绪**: [氛围/色调]
**对白**: "..."
**时长**: X秒

### Scene 2: [场景名]
...
```

---

## 模块四：短视频Prompt（新增）

### 视频脚本Prompt结构
```
[视频类型] + [目标受众] + [核心信息] + [时长] + [风格/调性]
```

### Hook开头Prompt公式
```
[前3秒] 反常识陈述/悬念问题/惊人事实
[3-15秒] 快速引入主题/说明价值
[15-45秒] 核心内容/步骤/故事
[45-60秒] CTA/总结/引导互动
```

---


---

## 与各部门协作
- 与content.md协作：提供文案的prompt优化
- 与visual.md协作：提供图像设计prompt
- 与comfyui.md协作：提供ComfyUI工作流prompt
- 与qa.md协作：prompt策略交白帽纠察质疑

## 模块五：Prompt质量检查清单（白帽纠察审计用）

| 检查项 | 说明 |
|--------|------|
| ✅ 角色是否明确 | 模型知道自己扮演什么角色吗？ |
| ✅ 任务是否具体 | 输出内容/格式/长度是否明确？ |
| ✅ 约束是否设置 | 不知道时是否说了"不编造"？ |
| ✅ 是否有示例 | 复杂任务是否给了few-shot？ |
| ✅ 是否有冗余 | 有没有重复/矛盾的指令？ |
| ✅ 是否符合铁律 | 是否要求了实事求是？ |
