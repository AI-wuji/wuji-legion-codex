# Context Mode 兼容切片契约

本文件记录对 [context-mode](https://github.com/mksglu/context-mode) 的最小兼容边界。它是可选宿主适配参考，不是无极军团的第二路由器、第二知识库或必需依赖。

## 当前可调用切片

| 切片 | 无极军团中的归属 | 验收要求 |
| --- | --- | --- |
| 固定 JSONL 批处理 | `ctx_execute` 可选宿主执行器 | 只允许计数、字段等值过滤和字段抽取；Go 侧读取显式工作区文件并生成固定 JavaScript。 |

适配器输入最大 1 MiB，返回最大 4096 bytes，超时固定 10000 ms，最多返回 100 个位置或值。生成代码不接收用户 JavaScript，只嵌入 Go 侧已规范化且确认位于 workspace 内的单个输入路径，不含 shell、网络、后台任务或 intent 调用。脚本运行时重新限制文件大小并校验 SHA-256。CLI 输出 `prepared_only: true` 的调用契约；只有宿主真实调用 `ctx_execute` 并随后通过 `context-mode-validate` 校验，才构成执行证据。

## 明确不引入

- 不复制 context-mode 的 ELv2 源码、构建产物或整套 hooks。
- 不让外部 SQLite/FTS 数据库成为无极军团的状态权威；唯一状态仍由现有受限存储和证据目录管理。
- 不接受“98% 节省”等上游宣传数字作为验收结果，必须用本仓库的真实探针重新测量。
- 不把工具输出压缩结果直接升级为 `behavior-verified` 或 `primary`；仍需独立验证文件、SHA-256、对照和晋级回执。
- 不启用 `ctx_index`、`ctx_search` 或 recall/session 数据。2026-09-19 的真实隔离探针先以 source `wuji-synthetic-scope-20260919` 建索引；随后分别以项目 `E:/wuji-projects/wuji-legion-codex-2.0` 和 `E:/wuji-projects/wuji-scope-isolation-fixture` 搜索，两次均返回当前会话的同一 fixture。该行为不能证明可靠的项目隔离，因此相关入口保持禁用。

## 调用与验证

集成后的 CLI 只需提供两个命令：

```text
wuji context-mode-prepare --workspace ROOT --input FILE --operation jsonl-count
wuji context-mode-validate --contract CALL.json --result RESULT.json
```

prepare 会解析 workspace 和文件的真实路径，拒绝 symlink 越界，计算输入字节数与 SHA-256，并生成 `ctx_execute` 的 `code`/`timeout` 参数。validate 限制结果大小、重新读取并哈希原文件，并核对回执中的 schema、操作、输入字节与哈希。它不把“准备成功”误报为 MCP 已执行。

已完成的宿主探针另证明：775449-byte、10000 行的中文 JSONL（SHA-256 `8f1d7de6be0efffe2ab076c54f8523f32c59fcc1e790f86d14ff022629c89885`）能正确定位 104 个失败项（首项 0、末项 9991）；1000 ms 的 `setInterval` 探针被终止并返回 `isError: true`。这些是宿主行为证据，不放宽本适配器的固定 10000 ms 和无后台任务边界。

## 许可与发布

审读基线固定为 `mksglu/context-mode` v1.0.169、commit `6f0cc6841c687e754059f36714a11233fda1a02b`，上游声明 Elastic License 2.0。无极军团只保留本契约和独立实现，不复制上游源码、hooks 或安装器；若未来需要直接链接或再分发上游组件，必须先完成许可证审查、修改声明和托管服务限制评估。
