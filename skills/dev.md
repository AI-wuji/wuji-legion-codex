# 🦞 无极军团 — 第四师：软件开发师 SOP

---
name: wuji-dev
description: "HTML→EXE 软件开发、图表生成、仪表盘、电子展牌、加密加壳、反编译"
---

# 💻 第四师 — 软件开发师

## 技术栈
| 需求 | 首选 | 备选 |
|------|------|------|
| 打包 | Electron (electron-builder) | Tauri (更小) |
| 图表 | ECharts (66k⭐) | Chart.js |
| Excel | SheetJS (xlsx) | — |
| 仪表盘 | Tabler (39k⭐) | CoreUI |
| 展牌 | ECharts + 全屏模式 | Plotly.js |

## 打包加密流程
```
HTML/CSS/JS 开发完成
    ↓
SheetJS 读取 Excel 数据
    ↓
ECharts 渲染图表
    ↓
Tabler 布局仪表盘
    ↓
Electron 打包 (asar 加密)
    ↓
UPX 加壳压缩
    ↓
✅ 最终加密 EXE
```

## 反编译同类软件
| 类型 | 工具 | 效果 |
|------|------|------|
| Electron (asar) | npx asar extract | 完整源码 |
| .NET (C#) | dnSpy | 近乎源码级 |
| .NET 加壳 | de4dot + dnSpy | 脱壳后还原 |
| UPX | upx -d | 直接解压 |
| PyInstaller | pyinstxtractor | Python 源码 |

## 保护自己的软件
- L1: asar 加密（基础）
- L2: UPX 压缩壳
- L3: JS Obfuscator 混淆
- L4: VMProtect/Themida 商业加壳
- L5: 反调试 + 完整性校验
