# Capability Contract

Each capability package must declare:

- semantic triggers and one primary route;
- semantic triggers for one user-facing scenario Skill; upstream projects remain internal atoms;
- complete upstream source roots;
- required scripts, references, templates, and assets;
- callable entrypoints;
- a behavior probe that creates or inspects a real artifact;
- acceptance evidence and a fallback;
- replacement and retirement rules.

Lifecycle:

1. `known`: discovered only.
2. `doctrine-only`: useful ideas retained, no executable claim.
3. `assets-retained`: complete source assets are mounted, but no proven call.
4. `callable`: the declared entrypoint runs.
5. `behavior-verified`: representative fixture passes artifact checks.
6. `primary`: verified winner against the prior route.

Never promote directly from a prose summary. `behavior-verified` requires a probe command that exits successfully and leaves verifiable evidence.

When multiple sources overlap, build one normalized catalog. Group duplicate template/style ids as variants, choose one preferred implementation, and retain provenance instead of copying every upstream tree into the runtime.
