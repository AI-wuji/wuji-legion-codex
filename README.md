# 无极军团 2.0 / Wuji Legion 2.0

面向最新 ChatGPT Codex 的系统级总 Skill：由一个最强阿极主脑按需挂载完整能力包，调度可并行专家，并以真实产物和行为测试完成闭环。

A system-level Skill for modern ChatGPT Codex: one strongest Aji brain cold-mounts complete capability packages, dispatches bounded parallel experts, and closes work with real artifact and behavior verification.

## 关键变化 / What Changed

- 单主脑、单写权限，不再保留女娲或第二工作流。
- Skill 保留完整脚本、模板、资产和入口；摘要不再冒充蒸馏。
- 专家按任务冷启动，可并行，但只接收紧凑契约和必要证据句柄。
- 进化主帅用原版对照测试决定吸收、替换或淘汰。
- 同名能力只有在现有路线与候选路线声明同一 `fixture`、均通过真实探针且候选不降级时才允许替换；应用前会归档旧 manifest。
- 项目检索使用小预算上下文选择；命令输出优先在上下文外压缩。
- 冷来源、Go 工具链和可选上下文工具均由 `sources.lock.json` 固定版本、路径与完整性证据。
- PPT 对外只保留两个新 Skill：`wuji-web-deck` 与 `wuji-editable-deck`。HTML-PPT、Slidev、流体舞台、PPT Master、Huashu、Humanize PPT、Baoyu 等已成为内部模板、组件、脚本和规划层。
- 写作、视觉、研究、前端、数据、文档、安全、图像和视频也各自收束为一个场景型统一 Skill。
- 旧版 104 个对象和 52 条工作树路径已逐项重新裁决；每项明确标记为更新、蒸馏、补落地或剔除，详见 `migration/legacy-verdict-ledger.json` 与 `migration/legacy-worktree-ledger.json`。

- One brain and one write authority; no Nuwa or parallel workflow.
- Skills retain complete scripts, templates, assets, and entrypoints.
- Experts are cold, bounded, parallel when independent, and receive compact contracts.
- Evolution uses upstream-vs-integrated behavioral comparison.
- Replacement requires the same declared fixture on both routes, passing probes, and no lifecycle regression; the previous manifest is archived before apply.
- Retrieval is budgeted; command output is compressed outside the model context.
- Cold sources, the Go toolchain, and optional context tools are pinned by `sources.lock.json`.
- Presentation exposes only unified web-deck and editable-deck Skills; upstream projects are internal atoms.
- Other domains expose one scenario-oriented suite each.
- All 104 legacy objects and 52 legacy worktree paths are reclassified by evidence.



## Capability Honesty

- Lifecycle: `known -> doctrine-only -> assets-retained -> callable -> behavior-verified -> primary`.
- Say “fused” only for `behavior-verified` or `primary`.
- Current verified/primary surfaces: `presentation` (primary) and `context` (primary).
- Other suites, including `code-review`, are `callable` with smoke/mount probes; they are host-usable but not claimed fused.
- Sparse mount: primary sources mount by default; secondary/optional only when named or when the query asks for 完整能力.
- Multi-intent: `SecondaryCapabilities` may list follow-on suites (e.g. writing + image).
- Install CLI on PATH: `./scripts/install-wuji-path.ps1` (session) or `-User` (persistent).
- Fast audit: `./scripts/audit.ps1 -Mode fast` (skips slow presentation full probe set when used intentionally).
- Fast regression: `./scripts/test.ps1`; full acceptance: `./scripts/test.ps1 -Full`.

## 受控更新 / Controlled Updates

- 只检查上游：`./scripts/update-cold-sources.ps1`。
- 更新指定冷源：`./scripts/update-cold-sources.ps1 -SourceId ppt-master -Apply -Transport TreeDelta`。小型仓库可使用 `Archive`；更新成功后会写入精确 commit 与完整树哈希。
- Slidev 与流体舞台是经过筛选的 2.0 内置运行资产，位于 `capabilities/presentation/assets/`；验证时在临时目录安装和构建，不提交 `node_modules` 或 `dist`。
- Open Design 的新 daemon/auth/provider 平台明确不接入；旧仓库仅用于迁移证据，绝不作为 2.0 运行时。
- Xiaobai Image2 仅通过 `capabilities/image/providers/xiaobai-image2/invoke.ps1` 按次调用，凭据只从当前进程环境读取，不提供常驻 bridge、GUI 或服务。

## Quick Start

```powershell
./scripts/build.ps1
./bin/wuji.exe route --query "修改登录页并验证真实路由"
./bin/wuji.exe context-select --workspace . --query "capability verification" --max-bytes 12288
./bin/wuji.exe verify --capability presentation
./bin/wuji.exe evolve --candidate ./candidate-manifest.json
```

`evolve` 默认只评估；确认结果后加 `--apply` 才会写入。运行 `./scripts/test.ps1 -Full` 完成全套验收。
