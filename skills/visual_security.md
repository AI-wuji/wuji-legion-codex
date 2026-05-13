# ☯️ 无极军团 — 第二师：视觉作战师 + 安全局 SOP

---
name: wuji-visual-security
description: "UI设计审查（联动impeccable）、加密加壳、反编译、安全审计"
---

# 🎨 第二师 — 视觉作战师

## UI 设计审查流程
1. 检测到 UI 任务 → 自动激活 impeccable
2. 跑 impeccable teach → 了解项目背景
3. 跑 impeccable shape → 确定设计方向
4. 执行针对性审查命令
5. 应用八大反模式自查

## 反模式快速清单
□ AI塑料渐变 → HSL 精确定义
□ 默认字体 → Clash Display / Satoshi / Noto Sans SC
□ 灰色地狱 → 加色调倾向
□ 卡片套娃 → 间距代替嵌套
□ 纯黑纯白 → #0a0a0f / #fafaf9
□ 弹性缓动 → cubic-bezier(0.4, 0, 0.2, 1)
□ 图标网格 → 不对称布局
□ 无层级阴影 → 多层阴影叠加

---

# 🛡️ 安全局 SOP

## 进攻（反向工程）
| 目标 | 方法 | 注意 |
|------|------|------|
| Python 插件 | 直接读源码 | 开源检查许可证 |
| .pyc | uncompyle6/decompyle3 | 可能不完整 |
| Electron EXE | asar extract | 最易破解的格式 |
| .NET EXE | dnSpy | 还原度极高 |
| 加壳 .NET | de4dot → dnSpy | 先脱壳 |
| PyInstaller | pyinstxtractor | 可还原 |

## 防御（保护方案）
| 级别 | 措施 | 难度 |
|------|------|------|
| L1 | asar 加密 + UPX | ★ |
| L2 | 代码混淆 | ★★ |
| L3 | VMProtect/Themida | ★★★★★ |
| L4 | 反调试 + 完整性校验 | ★★★★ |
| L5 | Tauri (Rust原生编译) | ★★★★★ |

## 安全红线检查
- ❌ 不直接复制受版权保护的代码
- ❌ 提取功能后要重写实现
- ❌ 注意原项目的许可证要求

