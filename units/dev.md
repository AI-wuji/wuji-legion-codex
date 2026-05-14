---
name: 第四师-开发
description: "Rust核心+TS前端+Python ComfyUI. 多平台封装(win/mac/linux)"
---

# 第四师(工作记忆:默认Rust核心+通用壳架构)

## 默认技术栈
| 层级 | 语言/框架 |
|------|-----------|
| 前端(用户) | TypeScript+React/HTML |
| 桌面壳 | Tauri(Rust) -> EXE/DMG/AppImage |
| APP | Tauri移动端(待成熟) |
| 后端核心 | Rust单二进制 |
| ComfyUI入口 | Python(__init__.py,10行) |
| ComfyUI核心 | Rust(PyO3->.pyd) |

## 工具库(名称引用)
| 类别 | 命令 |
|------|------|
| Rust构建 | cargo build --release |
| Rust→Python | maturin build --release |
| Tauri打包(本平台) | tauri build |
| 格式化 | prettier --write **/*.{html,ts,tsx} |
| TS检查 | npx tsc --noEmit |
| 安全 | cargo audit |
| 测试 | cargo test + playwright |

## 脚手架
| 项目 | 命令 |
|------|------|
| Tauri+TS桌面 | npm create tauri-app@latest |
| Rust后端 | cargo init |
| PyO3插件 | maturin init --bindings pyo3 |
| TS前端 | npm create vite -- --template react-ts |

## 口令
skill(linus)+sop(dev.md)+mcp(codebadger)


## 外派远征标记

以下任务可外派给 **工兵L（本地35B）**（需配合读取 units/expedition.md）：

| 可外派场景 | 示例 |
|-----------|------|
| 样板代码 | CRUD、路由、模型定义、配置文件 |
| 明确spec的功能 | 接口实现、模块开发、组件撰写 |
| 代码重构 | 重命名、提取函数、文件拆分 |
| 文档生成 | README、注释、API文档 |
| 单元测试 | 按spec撰写测试用例 |
| 批量修改 | 多文件格式统一、升级迁移 |

不可外派：架构设计、核心算法、安全关键代码 —— 留在 Codex 本地。
