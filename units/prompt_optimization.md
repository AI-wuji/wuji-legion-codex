# Prompt Optimization Mirror

Mirror source: `kernel-source.json`

## Role

Prompt optimization is an offline governance loop, not a default runtime layer.

## Offline Chain

`feedback-log -> feedback-dataset -> candidate -> audit -> eval -> distill -> absorb / defer / reject`

## Rules

- Keep only distilled preference signals, not raw user text.
- Do not promote a candidate into resident runtime behavior without eval or real-run evidence.
- Do not wrap runtime token noise as a prompt upgrade.

## 200k Cached Diagnosis

If backend usage shows around 200k cached or blue-hit tokens, treat it as likely long-context bloat until `runtime-context-audit` proves otherwise.

Likely causes:

- oversized stable prefix
- replayed logs, pages, transcripts, or readmes
- too many mounted officers or skills
- too much old thread history

Fix direction:

`runtime-context-audit -> replace replay with summaries/handles -> shrink stable prefix -> keep one owner and only needed mounts`
