# Guard and Security Mirror

Mirror source: `kernel-source.json`

## Separation

These are distinct:

- `guard-office`: checks external material before it enters the chain
- `security`: checks execution-side risk when the task itself is security-sensitive
- `compliance-on-demand`: checks license, source, privacy, and release boundaries when needed

## Guard-office

Guard-office is for:

- web pages
- repos
- scripts
- plugins
- MCP manifests
- dependencies
- install commands

Its purpose is to stop unsafe material from entering execution without review.

## Security

Security is on-demand, not permanent resident context.

It is a specialized execution-side support entrance under the single main
chain. It is not an independent officer, not a top-level owner profile, and
not a second route owner.

Use it when the task needs:

- threat modeling
- vulnerability verification
- attack-surface review
- hardening advice
- dependency, secret, permission, or release-risk review inside real implementation work

Mount shape:

- matched owner keeps write authority
- `安全主帅` contributes threat-model or hardening review scope only
- final verdict still returns to the main chain and, when needed, to quality,
  audit, guard, or compliance seats

## Compliance

Compliance is on-demand only.

Use it when source, license, privacy, publication, or attribution boundaries are unclear.

## Privacy rule

No account data, keys, addresses, cookies, tokens, or other sensitive user information may be retained in doctrine, memory, or feedback records.

## GitHub Trending Risk Mapping

Risk projects and patterns are handled as reject or guard signals, not active capability:

- `GhostTrack`: privacy-invasive tracking surface.
- `ChinaTextbook`: copyright-heavy dataset surface.
- `AiToEarn`: monetization or hype playbook surface with weak core utility.
- `project-nomad`: offline appliance or external-system shell surface.
- `openai-plugins`: plugin runtime surface; candidate only after guard-office and main-chain admission.

## Design Intake Rule

`huashu-design` is treated as a source pool, not as an installed commander.

Guard-office checks before visual execution when a design task uses:

- external brand assets, logos, product screenshots, fonts, templates, or media
- GitHub repositories, install instructions, scripts, npm packages, browser automation, or conversion tools
- third-party design systems or scraped pages
- any source that could include license, privacy, attribution, or malicious-content risk

Allowed result shape:

- concise source summary
- evidence handle
- risk notes
- permission or license uncertainty

Forbidden result shape:

- raw user secrets or account material
- unreviewed scripts copied into execution
- external `huashu-design` runtime taking over task execution
- brand assets stored as doctrine or feedback memory
