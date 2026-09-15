# Context Mode 兼容切片契约

本文件记录对 [context-mode](https://github.com/mksglu/context-mode) 的最小兼容边界。它是可选宿主适配参考，不是无极军团的第二路由器、第二知识库或必需依赖。

## 允许借鉴的能力

| 切片 | 无极军团中的归属 | 验收要求 |
| --- | --- | --- |
| 摘要式批量执行 | `context` 能力的可选执行器 | 输入、输出字节、运行时长、子进程和工作区均有硬上限；超时必须返回可区分的失败状态。 |
| FTS/BM25 局部检索 | 现有工作区图/经验图的候选实现 | 只返回有界句柄和摘要；不得把原始会话、凭据或整库内容注入热上下文。 |
| 会话连续性 | 需求快照、任务图版本和反馈账本 | 恢复必须绑定 workspace、任务图版本和内容哈希；旧版本或跨项目记录必须拒绝。 |
| Codex 生命周期适配 | 宿主边界层 | hook 只能转换事件和施加门禁，不能宣布模型已调用、任务已完成或能力已晋级。 |

## 明确不引入

- 不复制 context-mode 的 ELv2 源码、构建产物或整套 hooks。
- 不让外部 SQLite/FTS 数据库成为无极军团的状态权威；唯一状态仍由现有受限存储和证据目录管理。
- 不接受“98% 节省”等上游宣传数字作为验收结果，必须用本仓库的真实探针重新测量。
- 不把工具输出压缩结果直接升级为 `behavior-verified` 或 `primary`；仍需独立验证文件、SHA-256、对照和晋级回执。

## 兼容接口草案

适配器只需提供以下三个有界操作，返回内容寻址句柄，不返回无限正文：

```text
context.summarize(input, limits) -> {handle, bytes, sha256, outcome}
context.search(query, scope, limits) -> {handles[], coverage, bytes, outcome}
context.resume(workspace, graph_version, task_id, limits) -> {handles[], stale, outcome}
```

`limits` 至少包含最大输入/输出字节、最大执行时间、最大候选数和允许的工作区根。任何缺失或超限字段都应拒绝调用。适配器的真实行为验证必须写入 `WUJI_PROBE_EVIDENCE_DIR`，并由独立验证器计算哈希。

## 许可与发布

上游仓库当前声明 Elastic License 2.0。无极军团只保留本契约和独立实现；若未来需要直接链接或再分发上游组件，必须先完成许可证审查、修改声明和托管服务限制评估。
