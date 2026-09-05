---
name: feishu-lark
description: Use the official lark-cli to read, search, create, or update Feishu/Lark documents, wiki spaces, tasks, sheets, Base records, Drive files, calendar items, and messages. Use for Feishu/Lark URLs, document or wiki tokens, 飞书文档, 飞书知识库, 飞书任务, or requests that explicitly require an operation in Feishu/Lark.
---

# Feishu / Lark

Use the official `lark-cli` as the only execution surface. Keep existing cloud content read-only unless the user explicitly requests a targeted write.

## Select The Official Domain Skill

Before running a command, read `official/lark-shared/SKILL.md`, then read only the matching domain Skill:

- Documents and `/docx/` links: `official/lark-doc/SKILL.md`
- Wiki spaces and wiki nodes: `official/lark-wiki/SKILL.md`
- Tasks and task lists: `official/lark-task/SKILL.md`
- Sheets, Base, Drive, calendar, mail, IM, or other supported domains: read the matching `official/lark-*/SKILL.md`
- Unsupported operations: read `official/lark-openapi-explorer/SKILL.md` and inspect official API docs before using `lark-cli api`; never guess paths or parameters

Follow the selected official Skill's required references and command schema. Do not load unrelated Lark Skills.

## Execution Rules

1. Verify `lark-cli` exists. Use `lark-cli auth status --json --verify` when authenticated access is needed.
2. Default to `--as user` for user-owned documents, wiki content, tasks, Drive, and calendar data.
3. For reads, fetch the smallest useful scope or fragment and preserve the source URL/token in the result.
4. For creates, edits, sends, moves, permission changes, or deletes, require an explicit target and action. Use `--dry-run` when supported.
5. Treat exit code `10` with `confirmation_required` as a hard user-confirmation boundary. Show the action and risk; append `--yes` only after explicit approval.
6. Judge success from exit code `0` and JSON `ok: true`. A route, mount, command preview, or auth prompt is not execution evidence.

## Authentication Boundary

Do not create an app, begin login, or request scopes unless the user asks to configure/authenticate or the requested operation requires it.

When authentication is required, request the minimum domain/scope and follow the official split flow: start with `--no-wait --json`, preserve the verification URL exactly, generate its QR code with `lark-cli auth qrcode`, return control to the user, and finish `--device-code` only after the user confirms authorization. Never store verification URLs, device codes, app secrets, or access tokens in rules, repositories, or chat summaries.

## Safety

- Never print app secrets or access tokens.
- Keep local file arguments relative to the current working directory.
- Do not silently retry writes or reinterpret a failed write as success.
- Do not use third-party Feishu MCP servers when the official CLI covers the operation.
- If configuration, identity, or scope is missing, report the exact blocker and minimum next action.
