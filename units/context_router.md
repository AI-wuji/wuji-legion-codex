# 上下文路由与装配 / Routing Mirror

Mirror source: `kernel-source.json`

## Purpose

本文件只镜像当前有效的三层主路由与上下文装配原则，不再承载第二套独立 doctrine。

## Three Layers

1. `task-routing`
   - 判定状态机
   - 选择 owner
   - 注入独立监督
   - 定义收口标准

2. `capability-mount`
   - 最小缺口优先
   - owner 确定后再挂载
   - 只挂需要的 atoms / plugins / MCP

3. `deterministic-execution`
   - Go `wuji-cli`
   - repeatable gates
   - audit / bench / context-pack

## Context Assembly

### resident

- 入口规则
- 状态机
- 三层主路由
- owner 归口
- 独立监督边界

### mount-on-demand

- distilled OpenSquilla atoms
- fused Wuji `distilled_atoms`
- approved MCP / plugins
- specialized skill fragments
- summarized large artifacts

## Atom namespace rule

- `source_lineage_atoms` are source-level capability hints only.
- `distilled_atoms` are fused Wuji atoms after gap-fill, replacement, and audit.
- These namespaces must not be mixed.

### forbidden-as-resident

- full skill bodies
- full logs
- full test output
- whole external pages
- one-off discussion noise
- sensitive data

## Optimization Rule

上下文优化必须同时满足：

- 稳定小前缀
- 轻量装配
- 证据保留
- 验证不降级

没有通过 `optimization-audit` 的装配策略，不得进入主链。
