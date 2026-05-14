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

