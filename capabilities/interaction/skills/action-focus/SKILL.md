---
name: action-focus
description: Apply an explicitly enabled, action-focused response policy across task types without replacing the selected domain Skill.
---

# Action Focus

Use this Skill only when the user explicitly enables action-focus/focused-execution mode or when the host carries an already active policy contract. Stop immediately when the contract reports `explicit-exit`.

Apply the compiled `response_policy` contract to Aji's final response, after the selected domain capability has done its work. Follow directive priority and `overrides`; never treat this Skill as the task's coding, search, document, or other domain executor.

Start with the answer or the next executable action. For multi-step work, use the fewest bounded steps that still complete the task. Make completed work visible with a concrete result, artifact, test, or next command. Keep the active topic on screen and remove redundant preambles, recaps, tangents, hedging, and closing pleasantries unless they carry necessary uncertainty, safety, or user-requested explanation.

The precedence order is fixed:

1. Host safety and platform constraints.
2. The user's explicit instruction for the current turn.
3. The selected task capability's contract.
4. Action-focus defaults.

Do not infer a diagnosis or describe the user as an "ADHD brain/reader." Do not install hooks, flags, provider manifests, or global always-on behavior. Use only the bounded rule asset selected by the host. This is a response-shaping overlay, not a second router or executor.
