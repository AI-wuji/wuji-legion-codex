# Presentation External Candidate Review

Status: reviewed and distilled on 2026-07-18. Stars are GitHub REST snapshots collected on that date. External projects remain evidence or cold sources unless the local callable slice and behavior probe say otherwise.

## Search Evidence

Official documentation pages returned HTTP 200:

- https://gitbrent.github.io/PptxGenJS/
- https://revealjs.com/
- https://sli.dev/
- https://marp.app/
- https://python-pptx.readthedocs.io/en/latest/

The bounded source branch used the required official -> GitHub -> community order. The native preflight agent returned a read-only result, but its host did not expose a native agent id, payload hash, or result handle. Its claims are therefore corroborating research, not execution evidence.

## GitHub Candidates

| Project | Stars | License | Source and verification evidence | 2.1 decision |
| --- | ---: | --- | --- | --- |
| [hugohe3/ppt-master](https://github.com/hugohe3/ppt-master) | 39,669 | MIT | Local pinned snapshot `fa3161cf7bb1cb858be0cfe694f8f45d3e87d4eb`; `skills/ppt-master/SKILL.md`; `scripts/pptx_intake.py`, `pptx_to_svg.py`, `svg_to_pptx.py`, `pptx_template_import.py`, and `native_enhance_pptx.py`; MIT `LICENSE`. | Retain as internal atoms. Distill intake, template fidelity, SVG/DrawingML conversion, native enhancement, notes/timing/transition boundaries, and visual regression rules. Do not expose its workflow as a user route. |
| [hakimel/reveal.js](https://github.com/hakimel/reveal.js) | 71,961 | MIT | `dist/reveal.js`, `dist/reveal.mjs`, `scripts/test.js`, and HTML/React tests; MIT `LICENSE`. | Reject as a second web runtime. Distill the 16:9 stage, keyboard/deep-link navigation, presenter behavior, and export checks into the existing web-deck contract. |
| [gitbrent/PptxGenJS](https://github.com/gitbrent/PptxGenJS) | 5,857 | MIT | `src/pptxgen.ts`, `writeFile()` export, `TESTING.md`, browser worker/Vite/Node tests, and MIT `LICENSE`. Official docs returned HTTP 200. | Retain the generator contract only: native slide/text/shape/image/table/chart output and deterministic file export. The local artifact-tool generator is the callable implementation; do not add a parallel dependency. |
| [scanny/python-pptx](https://github.com/scanny/python-pptx) | 3,453 | MIT | `src/pptx/api.py`, `src/pptx/presentation.py`, `tests/test_api.py`, `tests/test_presentation.py`, and MIT `LICENSE`. | Reject for this fusion: its semantic editing model is useful evidence, but adding a Python runtime is not the smallest compatible JS/TS slice. |
| [marp-team/marp-cli](https://github.com/marp-team/marp-cli) | 3,701 | MIT | `bin/marp-cli.js`, `src/cli.ts`, `package.json` `--pptx` path, `test`, `test:coverage`, and MIT `LICENSE`. | Keep as cold secondary prior art for Markdown -> PPTX. Do not call it fused until an isolated local PPTX editability and render probe exists. |
| [slidevjs/slidev](https://github.com/slidevjs/slidev) | 47,703 | MIT | Local retained Slidev source is pinned as `slidev-runtime`; `package.json` has `build`, `dev`, and `export`; local lock pins `@slidev/cli` 52.17.0. Official docs returned HTTP 200. | Preserve the existing web-deck slice. No second Slidev route or duplicate runtime. |
| [nexu-io/open-design](https://github.com/nexu-io/open-design) | about 79,000 | Apache-2.0 | Repository-level design workbench with PPTX export, but much larger than the presentation boundary and not a local callable slice. | Reject wholesale. Keep only general design-system evidence; do not import its application shell, templates, or media. |
| [op7418/guizang-ppt-skill](https://github.com/op7418/guizang-ppt-skill) | 21,672 | AGPL-3.0 | Pinned local audit `014c572454065e905477a7432ae331dfc0fe6070`; `SKILL.md`; `scripts/validate-swiss-deck.mjs`; complete assets/templates/references. The current primary output is HTML, and its README states PPTX is not the main path. | Do not mount as the editable engine. Distill Swiss/layout hierarchy, image-slot discipline, and validation into the unified web and editable contracts. Keep the upstream package optional and cold because its delivery boundary and AGPL license do not match the primary route. |
| [chuspeeism/dashi-ppt-skill](https://github.com/chuspeeism/dashi-ppt-skill) | 3,780 | AGPL-3.0 | HEAD `fdbb145517ea0e289000aef9b7906bcb3e0cd19a`; local `dashi-ppt` audit; `SKILL.md`; `scripts/render_goal_deck.sh`, `project/scripts/validate-goal-spec.mjs`, `validate-swiss-deck.mjs`, `layout-query.mjs`, `inspect-layout.mjs`, and export/preview scripts; version 0.4.4. | Distill template-first layout registry, props/fillPlan, visible-copy overwrite, content-locked rejection, capacity budgets, unique-layout checks, and local-preview-before-export. Do not import the full runtime or expose a Dashi route. |
| [JimLiu/baoyu-skills](https://github.com/JimLiu/baoyu-skills) (`baoyu-slide-deck`) | not recorded in this snapshot | MIT | Pinned commit `6b7a2e417500561a5ecdd0b168332f4142584617`; `SKILL.md`; `scripts/merge-to-pdf.ts`, `scripts/merge-to-pptx.ts`; analysis/design/layout/style/density/mood/typography references. | Keep as a secondary style/narrative source. Distill density, mood, typography, image slots, and real-copy checks. Image-first output remains image-only unless independently reconstructed and verified as native PPTX. |
| [GordenSun/GordenPPTSkill](https://github.com/GordenSun/GordenPPTSkill) | 2,723 | Other / NOASSERTION | HEAD `7d4a61cbd8363b92e4a30784f8ca460a0a2b0a33`; `SKILL.md`; `scripts/build_pptx.py`, `render_slides.py`, `compute_capacity.py`, `build_manifest.py`, `apply_update.py`, `check_update.py`; `NOTICE`; README says personal/research use only. | Distill non-destructive template updates, explicit `edits.json` slot addressing, `detail.json` capacity/type scale/page roles, no ellipsis truncation, chart-data synchronization, and final render QA. Do not copy its scripts/templates or call it an admitted reusable package. |
| [GordenSun/GordenSuperPPTSkills](https://github.com/GordenSun/GordenSuperPPTSkills) | 1,512 | not identified | HEAD `8c05583dab8334182b71738e8dfbbec5c56a1951`; `GordenImagePPTGen`, `GordenImage2PPTX`, and `GordenSuperPPTSkill` entrypoints; manifests, RUN_ROOT, coordinate contracts, and visual comparison QA. | Retain only the image-only versus reconstructed-editable boundary, manifest discipline, task isolation, and visual comparison rules. Do not import its image-generation or reconstruction runtime. |
| [CRui5in/paper-ppt-agent](https://github.com/CRui5in/paper-ppt-agent) | 1,052 | AGPL-3.0 | HEAD `d9967905c8a52e97cb555632054a530dda8c6450`; FastAPI/Python/Node app; `pyproject.toml` includes `python-pptx`, `svglib`, `resvg`, `playwright`; `pytest.ini` and `tests`; README documents Strategist -> Executor -> Critic, static/visual QA, template import, version snapshots, and editor. | Do not import the application runtime. Distill staged critic gates, static plus visual QA, template import evidence, and version snapshots into the local completion evidence chain. |
| [khuynh22/paper-deck](https://github.com/khuynh22/paper-deck) | 11 | MIT | HEAD `30c35fea888641296ba33fb1876686077fee3d7f`; Next.js/Vitest app with tests, but it is an AI/ML paper aggregation reader, not a PPT generator. | Reject for PPT fusion; no relevant deck generation or editable artifact contract. |
| `gaoshou` exact repository/Skill | not found | n/a | Bounded GitHub repository search and `gaoshou ppt` search returned no verifiable exact PPT Skill/repository. Results were name collisions or unrelated projects. | Do not infer a capability from the name. Re-audit only when the user supplies an exact URL or owner/repository. |

## Existing Wuji And Earlier Sources

These are the original or previously retained presentation sources. They remain internal atoms, cold evidence packages, or catalog assets behind the two unified scenario Skills; they are not forgotten or replaced by the newer candidates.

| Existing source | Evidence / license boundary | Retained role in 2.1 |
| --- | --- | --- |
| `html-ppt-skill` / [lewislulu/html-ppt-skill](https://github.com/lewislulu/html-ppt-skill) | Pinned local audit `f3a8435d3901697d5ac5e64d356c933637e43107`; MIT; themes, templates, animation effects, runtime, and render script. GitHub snapshot was 7,210 stars. | Web-deck visual atoms, template catalog, browser render probe; never a second user-facing route. |
| `dashi-ppt-skill` / [chuspeeism/dashi-ppt-skill](https://github.com/chuspeeism/dashi-ppt-skill) | Local user Skill and audited upstream are distinct; AGPL upstream is not copied into the repository. | Layout/fillPlan/capacity doctrine only; local adapter is narrow and separately behavior-verified where its probe passes. |
| `ppt-master` / [hugohe3/ppt-master](https://github.com/hugohe3/ppt-master) | Pinned MIT source with intake, SVG/DrawingML, template import, native enhancement, and render scripts; 39,669 stars in the snapshot. | Editable-deck conversion, fidelity, enhancement, and regression atoms. |
| `huashu-design` / [alchaincyf/huashu-design](https://github.com/alchaincyf/huashu-design) | Pinned source in `sources.lock.json`; complete package retained cold. | HTML-to-PPTX design reasoning, export boundary, and verification atoms. |
| `humanize-ppt` / [LearnPrompt/humanize-ppt](https://github.com/LearnPrompt/humanize-ppt) | Pinned source and tests in `sources.lock.json`; retained complete and cold. | Narrative planning, audience intent, media slots, and slide-plan QA. |
| `frontend-slides` / [zarazhangrui/frontend-slides](https://github.com/zarazhangrui/frontend-slides) | Pinned source with MIT license and large template pack. | Web layout/style presets, template catalog inputs, and animation patterns. |
| `open-design` / [nexu-io/open-design](https://github.com/nexu-io/open-design) | Pinned Apache-2.0 audit; full design workbench is outside the presentation boundary. | General design-system evidence only; no application shell or second runtime. |
| `slidev-runtime` and `stage-fluid` | Local approved assets with pinned tree hashes in `sources.lock.json`. | Internal Markdown/Vue runtime and stage component; web-deck only. |
| OpenAI Presentations / Codex Grid | Installed primary runtime with native artifact-tool setup, templates, renderer, and slide tests. | Default editable PPTX composition and primary behavior probe. |
| `elite-powerpoint-designer`, `slide-studio`, `pptx-generation` | Existing local Skills; optional source entries with their own instructions/references. | Internal design, editing, and generation doctrine; no parallel user route. |
| `baoyu-slide-deck` | Existing pinned secondary source; image-first deck semantics are explicitly kept separate from editable PPTX. | Style, density, image-slot, and narrative atoms only. |
| `academic-pptx-skill`, `claude-office-skills`, `ppt-agent-skills` | Existing local audited material; retained only where its entrypoint and license permit. | Academic structure, OOXML/html2pptx, and agent QA evidence; selected by scenario, never all mounted together. |
| `reveal.js`, `Slidev`, `Marp`, `PptxGenJS`, `python-pptx` | High-signal upstream prior art; MIT projects except where stated; no duplicate runtime admitted. Known snapshot stars: reveal.js 71,961; Slidev 47,703; PptxGenJS 5,857; Marp CLI 3,701; python-pptx 3,453. | 16:9/navigation/export contract, native object model, deterministic export, and pagination/editability acceptance rules. |

## Candidate Distillation Matrix

| Capability atom | Sources | Unified behavior |
| --- | --- | --- |
| Narrative staging and critic loop | Humanize, Baoyu, Paper PPT Agent | Plan page role and density first; run static, render, and critic checks before completion. |
| Template fidelity and slots | PPT Master, Dashi, GordenPPT, Codex Grid | Inspect unknown templates, bind edits to explicit slots, preserve source, and export a new artifact. |
| Layout governance | Guizang, Dashi, HTML-PPT, Frontend Slides | Use a layout registry, unique page roles, aspect-ratio image slots, and no ungoverned visible placeholders. |
| Native editability | PptxGenJS, python-pptx, Huashu, Claude Office | Keep text/shapes/charts/tables native; a rendered image is never proof of editability. |
| Image-first reconstruction | Baoyu, GordenSuperPPTSkills | Track image-only and editable deliverables separately; reconstruction needs a manifest, coordinate contract, and comparison evidence. |
| Browser deck behavior | reveal.js, Slidev, Marp, HTML-PPT | One 16:9 stage, stable navigation, presenter/export checks, nonblank pixels, and static fallback. |
| Capacity and pagination | Dashi, GordenPPT, PptxGenJS community evidence | Measure text, use capacity as a warning, paginate before export, and never truncate with ellipses. |
| Verification evidence | Paper PPT Agent, PPT Master, GordenPPT, local probe | Preserve source hashes, rendered artifacts, independent checks, and final artifact evidence. |

## Community Evidence

The Stack Overflow API query for `pptxgenjs` returned 20 questions on 2026-07-18. The highest-signal boundary cases were [loading an existing PPTX](https://stackoverflow.com/questions/62191726/how-can-i-load-an-existing-pptx-file-with-pptxgenjs), [text height and slide pagination](https://stackoverflow.com/questions/77760607/how-to-accurately-calculate-text-height-for-slide-pagination-in-pptxgenjs), [large-file browser blocking](https://stackoverflow.com/questions/69156413/writing-large-file-causes-ui-to-freeze-for-a-short-time-in-react-js), and [HTML text conversion](https://stackoverflow.com/questions/65567697/converting-html-text-string-to-pptx-slide-using-pptxgenjs).

These reports become local acceptance rules: a generator is not a template editor; long text needs measured pagination or overflow checks; browser decks need nonblank and console-error checks; HTML-to-PPTX conversion must still produce editable native objects.

## Distilled Contract

The retained rules are implemented by the two existing scenario Skills:

- Editable PPTX defaults to the native artifact-tool route, keeps semantic text/shapes/charts/tables, and uses retained PPT Master atoms only for intake, template/conversion, and native enhancement boundaries.
- Existing PPTX enhancement is append-oriented: notes, timings, transitions, and relationships must not cause slide regeneration or loss of source fidelity.
- Every editable deck is rendered and inspected, with text overflow, object editability, and template fidelity checked before delivery.
- Web decks use one 16:9 stage with keyboard/hash navigation, presenter behavior, responsive browser rendering, and export-ready editable source. Markdown/Vue remains an internal Slidev route, not a user-facing upstream choice.
- PptxGenJS/Marp-style generator semantics are accepted only when the resulting local PPTX passes the same native-object and render contract.
- Guizang/Dashi-style web layouts are accepted only through the local layout registry and browser probe; image-first and template sources remain distinct evidence types.
- Gorden-style edits are accepted only through explicit slot maps and non-destructive output; capacity warnings never authorize truncation.

## Evidence Boundary

The presentation capability remains `behavior-verified` after the local probe. This file records provenance and distilled policy; it does not promote cold external projects, create a second workflow, or claim an upstream package was copied. The local probe proves the unified scenario contract, not behavior of every upstream candidate listed above.
