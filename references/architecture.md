# Wuji Legion 2.0 Architecture

## Decision

Wuji 2.0 is a system-level Codex Skill, not another agent host. Codex executes; the Go CLI supplies deterministic routing, package verification, sparse context selection, and evolution gates.

## One Brain

Aji and the former General Staff are one strongest reasoning surface. Aji owns requirement normalization, dependency-aware fan-out, final merge, write authority, and completion judgment. Nuwa is removed because staffing belongs to routing and capability ownership, while evolution belongs to the Evolution Commander.

## Experts

Use three layers, none resident in the hot prompt:

1. A small domain owner contract names the capability and acceptance boundary.
2. A complete cold package retains upstream Skills, scripts, assets, templates, UI, and entrypoints.
3. A bounded task worker receives only its branch contract, selected context handles, and the mounted package.

Large domains expose one scenario-oriented suite, or a very small number when output formats have genuinely different runtimes. Presentation exposes exactly two: `wuji-web-deck` and `wuji-editable-deck`. HTML-PPT, Slidev, stage fluid, PPT Master, Huashu, Humanize PPT, and Baoyu are internal templates, components, scripts, or shared planning layers. Users never choose upstream projects.

Parallelize independent branches. Keep dependency edges sequential. Workers never merge each other and never become a second command chain.

## Evolution Commander

The Evolution Commander is capability governance, not a content persona. It must inventory the upstream package, retain execution assets, map overlaps, run the same fixture against upstream and integrated routes, compare artifacts, and then admit, replace, or reject. A name, rule list, or passing registry assertion is not capability evidence.

## Context

The default path uses a small stable prefix plus task-local retrieval. `context-select` emits ranked excerpts within a hard byte budget. RTK is the preferred optional command-output filter. Codebase Memory MCP is a read-only cold index for large repositories. Context Mode is useful as a tool-output sandbox but remains optional because its license and host hooks make it unsuitable as copied core code. Heavy temporal GraphRAG remains off the hot path. Web scouting uses GPT/Luna, never Agnes.

## Model Policy

- Aji planning, architecture, merge, and high-risk judgment: `gpt-5.6-sol`, highest available reasoning.
- Bounded implementation and verification workers: Terra when explicitly delegated.
- Broad web scouting uses the default GPT route. Independent mechanical source branches may use Luna; Aji performs the final analysis.
- Other mechanical extraction uses Luna only when explicitly delegated and no project context replay is required.

Model switching is never used merely to save tokens when it would require replaying project context. Each worker receives a compact task handoff instead of the parent transcript.
