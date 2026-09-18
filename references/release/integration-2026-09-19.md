# 2026-09-19 集成记录

本轮保留最小可调用切片，不把上游仓库整包复制进主线，也不把 smoke 或适配器单测称为完整融合。

## 已落地

| 切片 | 当前证据 | 可宣称范围 |
| --- | --- | --- |
| 用户记忆 | Go core、CLI 子命令和行为测试 | 用户明确确认后可保存、分作用域召回、TTL 清理、版本冲突保护、疑似密钥拒绝和撤销；正 TTL 按整秒设下限，避免亚秒值序列化成 `0` 后被误解为永不过期；记录是用户事实，不是已核验世界事实。 |
| 专家桥 | 确定性选择、交接契约、内容哈希、独立证据一致性检查和陈旧 attempt 拒绝测试 | 可准备交接并检查回执一致性；verify 固定不确认宿主执行、不允许图变更。没有原生模型/session 宿主绑定证据，不能宣称六类专业工作流已经完整执行。 |
| context-mode 兼容切片 | 生成参数经真实 MCP 执行并由 Go 独立复算；Node 回归 | 固定 JSONL 过滤切片通过一万行中文 fixture；indexed fixture 跨无关显式项目路径可见，因此禁用索引记忆复用。 |
| Graphify pilot | 有界适配器及其测试 | 已修正过滤节点后可能遗留的 dangling edge；仅为冷、可选、无 hook 的 code-only 后端候选；本机缺少可用的固定 Python 运行时，上游提取尚未行为验证。 |

## 明确未完成

- 原生 Codex 宿主未提供本仓库可用的精确模型和 `session_key` 绑定入口；CLI 契约不是原生执行回执。
- 专家目录的六个专业流程没有因此自动变为完成或 `behavior-verified`。
- context-mode 的 indexed memory 不进入记忆协议，不启用自动捕获、恢复注入或全局 hook。
- Graphify 没有接入主路由或主工作区图，也没有安装 hook、watcher、MCP 或 LLM provider。
- 本轮未修改账号认证。

## 操作边界

统一入口与示例见[冷集成操作指南](../integrations/operations.md)。外部/本地记忆的数据与授权边界见[外部知识库与长期记忆协议](../integrations/external-memory-protocol.md)。

## 本地验证

- `go test ./...`：通过（`cmd/wuji` 与 `internal/core`）。
- `scripts/build.ps1`：通过，生成 `bin/wuji.exe`。
- CLI smoke：`--help` 显示三个集成入口；`expert-bridge select` 返回内容寻址目录选择；`user-memory remember -> recall -> revoke` 通过；`context-mode-prepare` 返回带 `language: javascript`、固定超时和 `prepared_only: true` 的无状态调用契约。该 smoke 不替代真实 MCP 执行。
- Graphify adapter probe：通过；真实固定上游运行返回 `status: unavailable` / `runtime-not-found`，因此保持 cold/unavailable。
- Fast audit 仍受仓库已知 size gates 阻断，不能宣称整库 audit 全绿。

## 实际适配器验收

修正后的 `context-mode-prepare` 生成参数已直接交给本机 `ctx_execute` 执行，再由 `context-mode-validate` 独立复算。10,000 行中文 JSONL 共 775,450 字节（含末尾换行），正确返回 104 个失败、前 100 个行号和截断标记。输入未进入调用正文；准备契约文件为 3,143 字节，包含代码回显的宿主响应文件为 2,940 字节，纯结果文件为 724 字节。这些是文件字节数，不是模型 token 或计费节省比例。

本地证据位于 `.wuji-probe-evidence/integration-20260919/`，经独立文件哈希校验：

| 文件 | SHA-256 |
| --- | --- |
| `context-large-call.json` | `5c8e94afed2f909e1e0cdcfcf716e1a211bc1d9b197f715d8257880632976d5b` |
| `context-large-host.json` | `4c65c058be8288c3e8d338db0514309df4f919c0258127bf4780a74644822790` |
| `context-large-result.json` | `9c7f0f916b86ca2d8446ca42baecabcc33e169c95dcbb9595f59611344e56cdd` |

该合成 fixture 只证明固定过滤切片；没有晋级任何 capability 为 primary。新增 Node 回归还覆盖零匹配、缺失字段与 null、数值格式、HTML 字符截断预算和尾随 JSON 拒绝。用户记忆另经独立 CLI 进程完成保存、召回、拒绝无版本覆盖及撤销。

缺失的 native agent、token、cache 或成本遥测不得补写或推断。
