# 第三师（ComfyUI） - ComfyUI主帅 + 节点/流程模式

## 核心定位

第三师只有一个入口：`ComfyUI主帅`。

普通生图不进第三师，直接走 `imagegen`。只有当任务明确涉及 ComfyUI 流程、节点插件、批量生成、视频管线或技术美术时，第三师才进入。

---

## 内置模式

| 模式 | 适用任务 | 验证 |
|---|---|---|
| 流程模式 | 搭建/修改 ComfyUI 流程 | 参数可复跑、种子和模型记录 |
| 插件模式 | 自定义节点、ComfyUI插件开发 | import smoke、节点映射完整 |
| 视频动效模式 | AnimateDiff、ControlNet、批量帧 | 输入输出明确、失败可定位 |
| 技术美术模式 | 风格统一、批量资产、图像/视频质量 | 样例输出和参数表 |

---

## 插件红线

- `NODE_CLASS_MAPPINGS` / `NODE_DISPLAY_NAME_MAPPINGS` 必须完整。
- `INPUT_TYPES`、`RETURN_TYPES`、`FUNCTION`、`CATEGORY` 必须一致。
- 不写死用户目录、模型目录或绝对路径。
- 不静默下载模型或执行安装脚本，除非用户明确批准。
- 处理 IMAGE/LATENT/MASK/TENSOR 必须检查维度、batch、dtype。
- 节点失败必须返回清晰错误，不吞异常。

## 流程交付

必须包含：

- 流程文件或脚本路径。
- 模型、节点、参数、seed。
- 输入输出说明。
- 样例输出路径。
- 失败排查点。

## 当前专家

- `ComfyUI主帅`：唯一 ComfyUI 入口，内部包含流程、插件、视频动效和技术美术模式。
