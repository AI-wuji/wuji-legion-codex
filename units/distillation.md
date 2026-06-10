# Distillation Mirror

Mirror source: `kernel-source.json`

## Goal

Distillation does not collect more named systems.

It breaks sources into atoms, then decides:

- `resident`
- `mount-on-demand`
- `replace`
- `retire`
- `reject`

Evolution is the admission judge, not another runtime. Before any cleanup,
merge, deletion, plugin enablement, or skill promotion, `evolution-profile`
must decide whether the candidate fills a real gap, replaces a weaker atom, or
only adds a new shell.

## Source pools

- `old-wuji`
- `opensquilla`
- `headroom-style`
- `reasonix-style`
- `hermes-style`
- `ecc-style`
- `context7-mcp-style`
- `research-mode-style`
- `superpowers-style-approved-atoms`
- `prior-art-rca-tooling-style`
- `fault-localization-style`
- `multi-agent-rca-style`
- `patch-debt-cure-style`
- `terminal-verification-style`
- `refactor-recipe-style`
- `agent-eval-style`
- `github-trending-20260608-style`
- `omx-style`
- `baoyu-skills-style`
- `khazix-skills-style`
- `humanizer-zh-style`
- `neat-freak-style`
- `competitive-teardown-style`
- `hv-analysis-style`
- `ppt-keynote-style`
- `ian-xiaohei-style`
- `prompt-cache-block-control-style`
- `llmlingua-style`
- `memgpt-style`
- `external-agent-shells`
- `external-write-surfaces`

## Distilled atom kernel

Mirror marker: `distilled_atom_kernel`.

Keep exactly the existing 21 kernel atoms unless the user approves an explicit redesign. New source pools may only fill gaps, replace weaker behavior, or be rejected:

- `assumption-ledger`
- `claim-fact-check`
- `reversible-evidence-handle`
- `content-type-compression-router`
- `version-doc-mcp`
- `guarded-realtime-source-search`
- `research-evidence-pack`
- `skill-stocktake-daily-library`
- `verified-learning-loop`
- `disciplined-debug-loop`
- `prior-art-solution-search`
- `root-cause-radar`
- `parallel-hypothesis-fanout`
- `patch-debt-root-cure`
- `terminal-real-run-verification`
- `html-native-design-canvas`
- `brand-asset-protocol`
- `anti-ai-slop-visual-rules`
- `design-direction-triad`
- `html-deck-to-editable-pptx`
- `motion-stage-sprite-engine`

## GitHub 热榜蒸馏边界

- `last30days`, `taste-skill`, `open-notebook`, `tolaria`, `turbovec-style`, `goose`, `pg_durable-style`, `opencv-style`, and `openai-plugins` are source-pool lessons only.
- Useful parts must land in existing owners and atoms: intelligence, visual, data, execution-base, quality, guard-office, or evolution.
- `GhostTrack`, `ChinaTextbook`, `AiToEarn`, `project-nomad`, plugin runtimes, and external agent shells are reject/guard signals unless a later task gives a narrow, reviewed, lawful use case.

## 2026-06-10 Source Pool Merge

No new default atom is added. The new projects are source pools only:

- `OMX`: keep hook, HUD, sidecar lifecycle, and plugin bridge lessons; reject the external runtime shell.
- `Huashu Design 2.0`, `ppt-keynote`, `baoyu-article-illustrator`, `ian-xiaohei`: strengthen visual atoms, not a second design system.
- `khazix-writer`, `wechat-toutiao-article-writer`, `humanizer-zh`: strengthen content-profile and quality-inspection writing checks.
- `hv-analysis`, `competitive-teardown`: strengthen `research-evidence-pack` with historical and competitive analysis cards.
- `baoyu-url-to-markdown`, `baoyu-youtube-transcript`: strengthen guarded search and evidence packs after guard-office screening.
- `baoyu-post-to-x/wechat/weibo`, `baoyu-electron-extract`: reject as default runtimes; keep only draft formatting, explicit authorization, and safety checklists.
- `neat-freak`: strengthens context/document cleanup by preferring deletion, consolidation, and evidence handles over more notes.

Context-cache note: prompt caching is useful only when the stable prefix is small. A 200k cached/blue-hit line is treated as long-context bloat until `runtime-context-audit` proves otherwise.

## Admission test

An atom is kept only if it improves the fused kernel without creating split brain:

- smaller stable prefix
- better routing precision
- lighter context assembly
- fewer retries and less rework
- equal or stronger evidence and discipline
- a recipe, counterexample, and verification path exist for structural changes
- an eval set or real-run evidence exists before promotion to resident behavior

## Rejection test

Reject atoms that mainly:

- create a second entry
- create a parallel routing shell
- dump whole skill bodies into context
- save tokens by removing necessary evidence
- duplicate an existing stronger atom
- turn a source pool into an active commander
- preserve a temporary patch chain instead of replacing the weak atom

## Prior-art first

When solving a problem, search existing tools, papers, docs, issues, and open-source implementations before inventing a new mechanism.

The output must be distilled into atoms, evidence handles, and local actions. Whole external systems must not become new commanders or route owners.

## Evolution gates

`进化主帅` uses these gates before later execution steps:

- `source-pool-not-shell`: external systems stay source pools unless a task needs a
  narrow local atom.
- `refactor-recipe-gate`: structural cleanup needs target, match surface,
  transform, counterexample, rollback path, and verification command.
- `eval-set-before-upgrade`: recurring preferences or prompt changes need a
  small comparison set before becoming resident rules.
- `retire-after-replace`: a weaker skill, rule, route, or script is retired only
  after its useful atom has a named owner.
- `no-sensitive-learning`: evolution may keep hashes, counts, and strategy
  signals, but never raw user secrets, accounts, sessions, addresses, or private
  task text.

## Record of truth

Decision truth lives in:

- `fusion-matrix.json`
- `residual-entrypoints.json`
- `acceptance-checklists.json`
