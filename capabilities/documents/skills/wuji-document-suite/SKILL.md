---
name: wuji-document-suite
description: Unified office-document Skill for Word, PDF, and spreadsheet creation, editing, rendering, redlining, and verification. Use whenever the primary deliverable is DOCX, PDF, XLSX, CSV, or a related office artifact.
---

# Wuji Document Suite

Route by output format while keeping one document workflow: preserve source fidelity, edit the real target, render the final artifact, inspect layout, and run format-specific validation. Use the complete official Word, PDF, or spreadsheet package only for the selected format.

## OfficeCLI atom

For an explicit Word/Excel structural dump or a one-shot HTML rendering request, mount `officecli-stateless-adapter` automatically. It is a narrow executor, not a second document route: keep this Skill as the primary workflow, retain the existing Word/spreadsheet fallback, and never mount it for PPT/PPTX work.

The adapter accepts only existing `.docx` and `.xlsx` files. Existing user files are read-only by default; writing requires an explicit request naming the target file. Invoke only its `ViewHtml` or `DumpJson` wrapper operations. The wrapper resolves its pinned local executable directly, forces `OFFICECLI_NO_AUTO_RESIDENT=1`, rejects overwrite by default, and forbids `open`, `close`, `watch`, `mcp`, plugin setup, and PATH discovery. Verify the generated HTML or JSON before reporting success.
