# 第四师（开发） — Rust/TS/Python 编程 + 自动化 + CI/CD

## 核心理念：规范先行，自动兜底

```
不是：写代码 → 出bug → 修bug
而是：规范(rules) → 编码 → 自动检查 → 自动修复
```

---

## 模块一：默认技术栈（保留）

- 前端: TypeScript + React / HTML
- 桌面壳: Tauri(Rust) → EXE
- 后端核心: Rust单二进制
- 自动化脚本: Python/PowerShell

---

## 模块二：Rust 编程规范（融合强化）

### 编码原则

| 原则 | 说明 |
|------|------|
| 1. 类型驱动设计 | 先定义类型，再写逻辑 |
| 2. 所有权清晰 | 理解所有权/借用/生命周期 |
| 3. 错误处理 | 用Result + anyhow/thiserror，不用unwrap |
| 4. 测试先行 | 公共函数必写测试 |
| 5. 文档即测试 | doc test 覆盖API用法 |

### 编译优化

| 优化项 | Cargo.toml配置 |
|--------|---------------|
| 发布模式 | `opt-level = 3`, `lto = true`, `codegen-units = 1` |
| 二进制瘦身 | `strip = true`, `panic = "abort"` |
| 加速编译 | `mold` 链接器, `sccache` 缓存 |

### 工具链

```
cargo check       # 快速检查（比build快）
cargo clippy      # lint检查（必跑）
cargo fmt         # 格式化（必用）
cargo test        # 测试
cargo audit       # 安全审计
cargo deny        # 许可证检查
```

---

## 模块三：自动化流水线（新增融合）

### 提交前自动检查（pre-commit）

```
[准备提交]
    ↓
① cargo fmt --check
② cargo clippy -- -D warnings
③ cargo test
④ cargo audit
⑤ 白帽纠察审计（代码质量）
    ↓
[通过才能提交]
```

### CI/CD 原则（参考cicd-expert）
| 阶段 | 检查项 |
|------|--------|
| CI | 编译+测试+lint+安全扫描 |
| CD | 自动构建+自动发布 |
| 监控 | 构建失败通知 |

---

## 模块四：repomix集成（保留升级）

当需要「分析代码」「打包项目」「外派任务」时:

```bash
npx repomix --output repomix-output.xml
```

### 使用场景
- 全量打包: `npx repomix`
- 远征军外派: spec附带repomix-output.xml
- 白帽纠察审计: 打包后做全量代码审查

---

## 模块五：远征军外派标记（保留）

| 可外派 | 不可外派 |
|--------|---------|
| 样板代码/CRUD | 架构设计 |
| 单元测试 | 核心算法 |
| 重构（有spec） | 安全代码 |
| 文档生成 | 白帽纠察审计 |

---

## 模块六：技术债务管理（新增）

| 类型 | 处理策略 |
|------|---------|
| 编译警告 | 立即修复，不允许积压 |
| 过时依赖 | 每月检查更新 |
| 无测试代码 | 每次修改时补测试 |
| 硬编码值 | 提取为配置 |


---

## 模块七：
---


---

## 与各部门协作
- 与security.md协作：代码提交前安全检查
- 与expedition.md协作：可外派任务交远征军
- 与qa.md协作：代码交白帽纠察质疑

## 模块七：新增领域专精专家（女娲统一调度）

| 专家 | 专长 | 所属师团 |
|------|------|---------|
| ⚡ Rapid Prototyper (快速原型师) | 快速验证、MVP | 第四师 |
| 🔍 Code Reviewer (代码审查师) | 代码审查 | 第四师 |
| 🚀 DevOps Automator (DevOps自动化师) | CI/CD自动化 | 第四师 |
| 🏗️ Software Architect (软件架构师) | 系统架构 | 第四师 |
| 📝 Technical Writer (技术作家) | 技术文档 | 第四师 |
| 🤖 AI Engineer (AI工程师) | AI模型集成 | 第四师 |
