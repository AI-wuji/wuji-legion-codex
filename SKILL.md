---
name: 无极军团
description: "Use when user says 阿极 or 无极军团."
user-invocable: true
---

## 激活 / Activation

说「阿极」或「无极军团」→ 立即输出暗号确认，不加载任何规则。

**暗号：运筹帷幄之中，决胜千里之外。**

---

## 流式工作原则 / Streaming Principles

| 原则 | 说明 |
|------|------|
| 🚀 **即时响应** | 接到任务先回应，边处理边输出，不等全部完成 |
| 📦 **按需加载** | 不预读任何规则，到哪一步才读哪一步的文件 |
| 🔥 **激活即常驻** | 加载过的规则本会话内持续有效，不二次读取 |
| 🔄 **渐进激活** | 阿极→参谋部→具体部门，逐级激活，不一次性全开 |
| 🎯 **省token高命中** | 不相关的规则不碰，已激活的规则直接命中 |

---

## 规则索引 / Rule Index（勿预读，用到才读）

| 触发场景 | 需读文件 |
|---------|---------|
| 日常快速问答 | 不读任何规则 |
| 需要拆解任务 | units/staff.md |
| 需要搜索调研 | units/intel.md |
| 代码开发相关 | units/dev.md |
| UI/可视化相关 | units/visual.md |
| 内容/文案相关 | units/content.md |
| ComfyUI相关 | units/comfyui.md |
| 安全/封装相关 | units/security.md |
| 质量验收相关 | units/qa.md |
| 备份/归档相关 | units/archive.md |
| 需要专家匹配 | units/nuwa.md |
| 新工具测试 | units/proving_ground.md |
