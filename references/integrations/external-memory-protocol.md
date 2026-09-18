# 外部知识库与长期记忆协议

飞书、夸克网盘和本地记忆都采用“按需检索、最小投影”模式，不做默认全量导入。

## 触发

- 用户明确要求“去飞书查”“查夸克网盘”“从我的知识库找”时，才启用对应连接器。
- 用户明确要求“记住”“保存为长期偏好/项目记忆”时，才允许写入记忆载体；普通检索结果不自动变成记忆。
- 未明确指定外部来源时，优先使用当前项目和已验证本地能力；不主动扫描飞书或网盘。

## 统一返回

连接器应返回内容句柄，而不是把整篇文档塞进 Prompt：

```json
{
  "source": "feishu|quark|local",
  "scope": "user|project|session",
  "handle": "content://sha256/...",
  "title": "...",
  "updated_at": "...",
  "content_sha256": "...",
  "acl": "read|write",
  "excerpt": "最小相关片段"
}
```

阿极和专家只接收与当前任务相关的摘要/片段及句柄；需要全文时必须由任务契约明确要求，并受字节、时间和文件数预算限制。

## 记忆层级

| 层级 | 默认范围 | 允许内容 |
| --- | --- | --- |
| 会话记忆 | 当前会话 | 未提交的对话状态和临时决定 |
| 用户记忆 | 跨会话 | 用户明确保存的偏好、长期约束和工作方式 |
| 项目记忆 | 指定项目 | 需求、决策、验证经验和项目知识 |
| 外部知识库 | 飞书/夸克授权范围 | 原文或索引内容，必须保留来源、权限、更新时间和哈希 |

跨项目读取默认拒绝，除非用户明确指定项目或授权共享范围。外部文档是证据来源，不自动成为事实或全局记忆；删除、撤销授权、版本变化或哈希不一致时，句柄必须失效。

## 当前状态

- 飞书：已有官方 `feishu-lark` 能力 manifest，CLI 探针通过；用户已在先前任务中完成授权，本轮没有复核令牌当前有效性，不要求重复授权。
- 夸克网盘：已安装夸克官网提供的官方 `quarkclouddrive` Skill 1.0.9，CLI smoke 探针通过；用户已在先前任务中完成授权，本轮没有复核令牌当前有效性。具体读写仍按用户任务范围执行。
- 本地用户记忆：`wuji user-memory remember|recall|revoke` 已提供显式持久化入口。默认以解析后的工作区路径哈希隔离；只有显式 `--shared-scope` 才跨项目共享。写入必须带 key、value 和用户确认来源 `--provenance`，并拒绝常见密钥形态；支持 TTL、版本冲突检查、容量上限和撤销。更新已有 key 时必须提供当前 `--expected-version`；TTL 为零表示不自动过期。

本地记忆中的内容是“用户确认过的偏好、约束或项目约定”，不是已经核验的世界事实。调用方不得把一次 recall 当成事实验证，也不得从普通对话、外部检索或 context-mode 索引中自动写入。当前 CLI 的 confirmation boundary 是显式调用加非空 provenance；宿主仍须在调用前确认用户确实要求保存。

最小示例：

```powershell
./bin/wuji.exe user-memory remember --workspace . --key "language" --value "默认使用中文" --provenance "用户于 2026-09-19 明确要求记住"
./bin/wuji.exe user-memory recall --workspace . --query "language"
./bin/wuji.exe user-memory revoke --workspace . --key "language"
```

默认存储位于 `.wuji/memory/user-memory/v1/records.json`。它包含用户提供的值和 provenance，必须按用户数据保护；仓库不得提交该文件。详细边界与其他冷集成见[操作指南](operations.md)。

跨项目共享须同时指定同一个 `--store` 目录和同一个 `--shared-scope`；只使用相同 scope 名但各项目各自的默认存储，并不会互相读取。
