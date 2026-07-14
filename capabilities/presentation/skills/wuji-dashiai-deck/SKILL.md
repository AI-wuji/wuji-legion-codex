---
name: wuji-dashiai-deck
description: Produce an editable browser-native HTML deck through the installed DashiAI runtime.
---

# Wuji DashiAI Deck Adapter

Use this adapter only for a browser-editable, template-driven HTML deck. The
local runtime is `~/.codex/skills/dashiai-ppt-skill/project`; do not load the
full upstream Skill body or unrelated templates.

1. Query a small candidate set with `npm --prefix <runtime> run --silent
   layout:query -- --theme <theme> --role <role> --limit 3 --seed <seed>`.
2. Build a concrete `goal.json` with explicit layouts and props. Validate it
   with `validate:goal-spec`, render with `render:goal`, then run both
   `validate:swiss` and `validate:goal-copy` against the result.
3. Deliver the editable `index.html`. Use the runtime preview server only when
   an interactive local editing session is needed; do not substitute a static
   server. Export PPTX or PDF only when the user asks.
4. Stop after two correction passes. If validation still fails, return the
   failing command and a compact cause instead of retrying without new input.
