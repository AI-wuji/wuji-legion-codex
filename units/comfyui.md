# 第三师（ComfyUI） — 图像生成 + 插件开发 + 多插件编排 + 视频渲染

## 核心理念：插件不是零散装，是流水线

## 交付准则

- 普通生图默认最短路径：`prompt → imagegen → 保存 PNG → 预览`
- 不需要复杂工作流、角色一致性、节点编排或视频动效时，不进入 ComfyUI
- imagegen 失败前不排查环境、不查代理、不读技能文件
- 只要用户要的是图片成品，或者 PPT/海报/封面/截图的页面图，就不要先展开长篇分析
- 成功后先给预览图，再给文件路径
- 只有在缺少关键约束且会明显影响画面时，才补一句确认
- PPT/HTML所需配图优先接收 `image-spec.json`，保证与页面信息和视觉系统一致

```
不是：插件A + 插件B + 插件C 各用各的
而是：插件A → 插件B → 插件C 串联成工作流
      图像生成 → 动效 → 合成 统一管线
```

---

## 模块一：架构（保留升级）

```
comfyui-plugin/
├── __init__.py          # 节点注册入口
├── nodes.py             # Python自定义节点（动效/过渡/合成）
├── workflows/           # 预设工作流JSON
│   ├── ppt-to-video.json      # PPT→视频
│   ├── image-to-animation.json # 图转动效
│   └── comic-to-video.json    # 漫画→视频
├── rust_core.pyd        # Rust高性能核心（图像处理）
└── rust_core/           # Rust源码
    ├── src/
    │   ├── lib.rs       # 核心算法
    │   ├── effects.rs   # 动效引擎
    │   └── pipeline.rs  # 流水线编排
    └── Cargo.toml
```

---

## 模块二：插件开发标准（新增）

### 自定义节点开发规范

```python
# 标准节点模板
class WujiCustomNode:
    """命名规范：Wuji功能名"""
    
    @classmethod
    def INPUT_TYPES(cls):
        return {
            "required": {
                "输入参数": ("IMAGE", {"default": None}),
                "配置项": ("STRING", {"default": ""}),
            },
            "optional": {}
        }
    
    RETURN_TYPES = ("IMAGE",)
    FUNCTION = "process"
    CATEGORY = "WujiLegion"
    
    def process(self, 输入参数, 配置项):
        """Rust核心加速"""
        import rust_core
        result = rust_core.process(输入参数, 配置项)
        return (result,)
```

### 插件开发SOP
```
① 需求分析 → 确定节点类型/输入输出
② 写Python壳（nodes.py）
③ 性能敏感部分 → Rust实现（rust_core/）
④ 写工作流JSON测试
⑤ 白帽纠察审计（安全检查：路径遍历/注入/资源泄露）
⑥ 注册到工作流库
```

### ComfyUI插件质量门禁

插件修改后不允许只看代码不运行验证。默认执行：

| 门禁 | 检查内容 |
|------|----------|
| 语法 | `python -m compileall .` |
| 导入 | 插件 `__init__.py` 可被导入 |
| 节点注册 | 存在并导出 `NODE_CLASS_MAPPINGS` |
| 类型契约 | `INPUT_TYPES` / `RETURN_TYPES` / `FUNCTION` 对齐 |
| 依赖 | `requirements.txt` 不含可疑源，不静默安装 |
| 路径 | 不写死用户目录，不越权读写 |
| 最小工作流 | 有样例 workflow 时必须跑或至少校验JSON结构 |

### ComfyUI常见退回原因

- 节点启动时报 import error
- 输入输出类型和实际返回不一致
- batch维度处理错误
- 依赖缺失但没有声明
- 异常被吞掉导致前端只显示空结果
- 插件安装脚本执行高风险命令
- 生成结果路径不可控或污染用户目录

---

## 模块三：多插件融合编排（新增核心）

### 插件编排原则

| 原则 | 说明 |
|------|------|
| 1. 单节点单职责 | 一个节点只做一件事 |
| 2. 标准接口 | 统一IMAGE/LATENT/CONDITIONING类型 |
| 3. 工作流即配置 | 用JSON编排，不改代码 |
| 4. 流水线思维 | 每个节点输出=下个节点输入 |

### 融合编排示例

```
[文本prompt]
    ↓
WujiPromptEnhancer(提示词增强)
    ↓
WujiImageGen(图像生成) ← 可切换：SDXL/FLUX/SD3
    ↓
WujiUpscale(放大) + WujiFaceRestore(人脸修复)
    ↓
WujiAnimation(动效) ← Rust加速
    ↓
WujiVideoCompose(视频合成)
    ↓
[输出视频]
```

### 与外部插件的融合

| 外部插件 | 融合方式 | 用途 |
|---------|---------|------|
| ComfyUI-Manager | 插件管理、安装、更新 | 基础依赖 |
| WAS NS | 图像处理节点 | 批量处理 |
| Efficiency Nodes | 工作流简化 | 减少节点量 |
| AnimateDiff | 视频动效 | 动画生成 |
| ControlNet | 精准控制 | 姿势/深度/边缘 |
| IP-Adapter | 图像风格 | 风格迁移 |

---

## 模块四：图像生成工作流（新增）

### 统一图像生成管线

```
[用户需求]
    ↓
提示词局(prompt_engine.md) 生成图像prompt
    ↓
选择模型：
  ├─ 写实 → SDXL/FLUX
  ├─ 二次元 → NovelAI/Animagine
  └─ 漫画 → comic-specific models
    ↓
ComfyUI工作流执行：
  ├─ 基础生成（txt2img/img2img）
  ├─ 优化（放大/修复/增强）
  └─ 后处理（抠图/调色/排版）
    ↓
对接下游：
  ├─ visual.md → PPT/漫画
  ├─ content.md → 视频/短剧
  └─ 直接交付
```

### PPT/HTML配图快路径

```
slide-spec.json / html-design-spec.json
    ↓
prompt_engine.md 生成 image-spec.json
    ↓
image2 / imagegen 生成素材
    ↓
保存到 output/assets/{project}/images/
    ↓
visual.md 插入PPT或HTML
    ↓
渲染预览图并验收
```

规则：
- 图片不要承担关键文字，文字由PPT/HTML原生文本层承载
- 多张配图必须共享风格锚点：色调、光线、构图密度、质感
- 普通配图用轻量 image2/imagegen；需要复杂工作流、动效或一致角色时再升级到ComfyUI
- 失败时不拿占位图冒充正式图

---

## 模块五：与藏师傅流水线对接（保留升级）

第二师交付PPT单页图片 → ComfyUI动效/过渡 → video_clip → 配音合成 → 出片

### 渲染方案对比

| 方案 | 速度 | 质量 | 适用场景 |
|------|------|------|---------|
| HyperFrames | ⚡快 | ⭐⭐ | 快速演示/草稿 |
| ComfyUI | 🐢慢 | ⭐⭐⭐⭐⭐ | 高质量出片 |
| 即梦CLI | ⚡快 | ⭐⭐⭐⭐ | 片段补全 |
| ComfyUI+AnimateDiff | 🐢慢 | ⭐⭐⭐⭐⭐ | 动画/动效 |

---

## 与各部门协作
- 与visual.md协作：接收PPT图片做动效渲染
- 与content.md协作：接收视频脚本需求
- 与prompt_engine.md协作：图像生成prompt输入
- 与qa.md协作：输出交白帽纠察质疑

## 模块六：主责专家库（女娲统一调度）

图像 prompt 与 ComfyUI 技术美术能力已蒸馏进 `ComfyUI Pipeline Engineer`。普通生图仍走 `imagegen`，只有需要节点工作流、批量管线或视频动效时才进入第三师。

| 专家 | 专长 | 所属师团 |
|------|------|---------|
| 🤖 ComfyUI Pipeline Engineer | ComfyUI工作流、技术美术、图像/视频生成管线 | 第三师 |

---

## 七、Codex插件融合

| 插件 | 用途 | 融合方式 |
|------|------|---------|
| **HeyGen** | AI数字人视频 | 短视频流水线的数字人出镜环节 |
| **Cloudinary** | 媒体资产管理 | 生成图片/视频的CDN分发 |
| **Remotion** | React视频生成 | 程序化动画+ComfyUI出图组合 |
| **Hugging Face** | 模型/数据集查询 | ComfyUI工作流模型选型参考 |
