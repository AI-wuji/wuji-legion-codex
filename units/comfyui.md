---
name: 第三师-ComfyUI
description: "ComfyUI插件:Python入口+Rust(PyO3)核心. 联动ilya"
---

# 第三师(已加载至工作记忆)

## 默认插件架构(Rust核心+Python壳)
```
comfyui-plugin/
├── __init__.py       <- Python注册入口(<=10行)
├── nodes.py          <- 节点输入输出定义(Python)
├── rust_core.pyd     <- 核心计算(Rust PyO3编译)
└── rust_core/        <- Rust源码
    ├── src/lib.rs    <- 性能计算/图像处理/模型推理
    └── Cargo.toml
```

## 命令
| 操作 | 命令 |
|------|------|
| 建PyO3项目 | maturin init --bindings pyo3 |
| 编译.pyd | maturin build --release |
| Rust侧测试 | cargo test |

## 插件分析
获取源码->定位NODE_CLASS_MAPPINGS->列节点+IO->分析依赖->生成API地图

## 合并
cookiecutter->复制节点(处理冲突)->合并依赖->统一风格->测试

## 口令
skill(ilya-sutskever-perspective)+sop(comfyui.md)

