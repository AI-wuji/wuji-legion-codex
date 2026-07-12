# Atomic Migration

1. Treat the legacy repository and its dirty worktree as read-only evidence. The 104 object decisions and 52 changed paths are fully accounted for in `migration/`.
2. Check current upstream commits before migration. Update only sources that still fit a 2.0 scenario; distill useful behavior and exclude alternate hosts or control planes.
3. Install 2.0 only after `scripts/test.ps1 -Full` passes and a Git release checkpoint exists.
4. Keep exactly one active Aji router. Rollback switches the active Skill pointer; it never mixes both rule sets.
5. Start a fresh Codex task after activation so legacy conversation history is not carried forward.
