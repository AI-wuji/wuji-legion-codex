# ☯️ 无极军团 — 第三师：ComfyUI 插件作战师 SOP

---
name: wuji-comfyui
description: "ComfyUI 插件合并、开发、反向工程、融合。插件合并、功能提取、工作流生成"
---

# ⚙️ 第三师 — ComfyUI 插件作战师

## 插件分析流程
1. 获取源码（GitHub clone / pip 安装目录 / 手动下载）
2. 定位 NODE_CLASS_MAPPINGS
3. 列出所有节点 + 输入输出类型
4. 分析依赖（requirements.txt）
5. 生成 API 地图

## 插件合并
1. 创建新项目 → cookiecutter 脚手架
2. 复制节点类 → 处理命名冲突
3. 合并依赖
4. 统一代码风格（遵循原插件习惯）
5. 创建 __init__.py 注册所有节点
6. 测试加载

## 反向工程
- 90% 插件是纯 Python → 直接读
- .pyc → 用 uncompyle6
- PyInstaller → pyinstxtractor + uncompyle6
- ⚠️ 提取功能思想，重写实现，不直接复制

## 关键陷阱
- 进口生命周期：__init__.py 的 import 顺序
- VRAM 管理：显存缓存不要泄漏
- 序列化：确保节点状态可保存/恢复
- 冲突：同名节点要加前缀

