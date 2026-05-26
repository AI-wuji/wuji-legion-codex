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

### Rust交付硬门禁

Rust项目默认不允许“能编译就交付”。除非用户明确要求跳过，修改后必须按可用工具执行以下门禁：

| 门禁 | 命令 | 失败处理 |
|------|------|----------|
| 格式 | `cargo fmt --check` | 先运行 `cargo fmt`，再复查 |
| 编译 | `cargo check --all-targets` | 必须修到通过 |
| Lint | `cargo clippy --all-targets -- -D warnings` | 必须修到无警告 |
| 测试 | `cargo test` | 必须修失败测试，不允许忽略 |
| 安全 | `cargo audit` / `cargo deny check` | 工具存在则执行，高危必须处理 |

允许例外：
- 仓库没有 `Cargo.toml`
- 依赖无法下载或环境缺工具，但必须如实报告
- 用户明确要求只做草稿/原型

### Rust Bug防线

- 禁止新增 `unwrap()` / `expect()` 到业务路径，测试代码例外但要合理
- 所有外部输入必须有错误处理：文件、网络、CLI参数、JSON、路径
- 涉及路径操作时优先用规范化路径，避免路径遍历
- 并发代码必须说明共享状态、锁粒度和失败恢复
- 修改 bug 时必须补最小回归测试；没有测试框架时至少补可复现命令

---

## 模块三：跨技术栈质量门禁（新增强化）

### 项目类型自动识别

修改项目后，先识别技术栈，再执行对应门禁，不允许只看一种语言。

| 识别文件 | 项目类型 | 必跑门禁 |
|----------|----------|----------|
| `Cargo.toml` | Rust/Tauri后端 | fmt/check/clippy/test/audit |
| `package.json` | HTML/前端/Node | install/typecheck/lint/test/build |
| `pyproject.toml` / `requirements.txt` | Python/ComfyUI插件 | compile/lint/test/import smoke |
| `*.ps1` | PowerShell自动化 | 语法解析/非破坏性dry-run/路径安全检查 |
| `src-tauri/` + `package.json` | Tauri软件 | Rust门禁 + 前端门禁 + Tauri build/check |
| `ComfyUI`插件结构 | ComfyUI插件 | 节点注册/import/依赖/路径安全/最小工作流测试 |

仓库自带统一入口：

```powershell
.\scripts\wuji-quality-gate.ps1 -Path <项目路径>
```

该脚本只做非破坏性检查；能跑的门禁会跑，缺工具或缺脚本会标记为 SKIP，不会伪装成 PASS。

### HTML/前端 Bug防线

- 禁止只写静态页面不验证：重要页面必须用 Browser 打开或截图验证
- 禁止“桌面能看、手机崩掉”：必须检查移动端断点
- 禁止未处理空状态、加载态、错误态
- 禁止不可点击假按钮，除非明确标注为原型
- 禁止 console error、明显布局溢出、横向滚动
- 组件修改后优先跑 typecheck/lint/build

### ComfyUI插件 Bug防线

- 插件必须能被 Python import，不允许启动时炸节点注册
- `NODE_CLASS_MAPPINGS` / `NODE_DISPLAY_NAME_MAPPINGS` 必须完整
- `INPUT_TYPES`、`RETURN_TYPES`、`FUNCTION`、`CATEGORY` 必须一致
- 不允许在节点执行中写死用户目录、模型目录或绝对路径
- 不允许静默下载模型或执行安装脚本，除非用户明确批准
- 处理 IMAGE/LATENT/MASK/TENSOR 时必须检查维度、batch、dtype
- 节点失败必须返回清晰错误，不吞异常
- 修改后至少做 import smoke test；有工作流样例时跑最小工作流

### Python/脚本 Bug防线

- Python修改后至少跑 `python -m compileall` 或对应测试
- CLI脚本必须有 `--help` 或最小无害运行验证
- PowerShell脚本修改后必须做语法解析，涉及删除/移动前必须验证目标路径
- 所有脚本不得把密钥、token、用户隐私写入日志

---

## 模块四：自动化流水线（融合）

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

### HTML/前端交付硬门禁

前端或网页项目修改后，默认按项目实际脚本执行：

| 类型 | 优先命令 |
|------|----------|
| 安装检查 | `npm ci` / `pnpm install --frozen-lockfile` / `yarn install --frozen-lockfile` |
| 类型检查 | `npm run typecheck` / `pnpm typecheck` |
| Lint | `npm run lint` / `pnpm lint` |
| 测试 | `npm test` / `pnpm test` |
| 构建 | `npm run build` / `pnpm build` |

如果脚本不存在，不要编造“已通过”，要报告“项目未提供该脚本”。

### ComfyUI插件交付硬门禁

```bash
python -m compileall .
python - <<'PY'
import importlib.util, pathlib
p = pathlib.Path('__init__.py')
assert p.exists(), 'ComfyUI插件缺少 __init__.py'
spec = importlib.util.spec_from_file_location('wuji_plugin_smoke', p)
m = importlib.util.module_from_spec(spec)
spec.loader.exec_module(m)
assert hasattr(m, 'NODE_CLASS_MAPPINGS'), '缺少 NODE_CLASS_MAPPINGS'
print('ComfyUI plugin import smoke OK')
PY
```

如果插件不是标准ComfyUI结构，要说明无法执行标准smoke test，并改跑项目提供的测试或最小导入验证。

### CI/CD 原则（参考cicd-expert）
| 阶段 | 检查项 |
|------|--------|
| CI | 编译+测试+lint+安全扫描 |
| CD | 自动构建+自动发布 |
| 监控 | 构建失败通知 |

---

## 模块五：repomix集成（保留升级）

当需要「分析代码」「打包项目」「外派任务」时:

```bash
npx repomix --output repomix-output.xml
```

### 使用场景
- 全量打包: `npx repomix`
- 远征军外派: spec附带repomix-output.xml
- 白帽纠察审计: 打包后做全量代码审查

---

## 模块六：远征军外派标记（保留）

| 可外派 | 不可外派 |
|--------|---------|
| 样板代码/CRUD | 架构设计 |
| 单元测试 | 核心算法 |
| 重构（有spec） | 安全代码 |
| 文档生成 | 白帽纠察审计 |

---

## 模块七：技术债务管理（新增）

| 类型 | 处理策略 |
|------|---------|
| 编译警告 | 立即修复，不允许积压 |
| 过时依赖 | 每月检查更新 |
| 无测试代码 | 每次修改时补测试 |
| 硬编码值 | 提取为配置 |

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

---

## 七、Codex插件融合

| 插件 | 用途 | 比什么好 | 融合方式 |
|------|------|---------|---------|
| **GitHub** | PR/Issue/CI管理 | 原生git | 项目协作入口 |
| **Supabase** | 数据库/Auth/Storage | Neon/Convex | 后端基础设施 |
| **Vercel** | 前端部署 | Netlify/Render | 发布通道 |
| **CircleCI** | CI/CD | — | 自动化流水线 |
| **Sentry** | 错误追踪 | — | QA质量反馈 |
| **CodeRabbit** | AI代码审查 | — | PR自动审查 |
| **Hugging Face** | 模型/数据集 | — | AI资源查询 |
| **Game Studio** | 浏览器游戏 | — | 创意输出通道 |

**调用链路：** dev.md收到任务 → 判断需要哪个插件 → 调用对应Codex技能 → 交付
