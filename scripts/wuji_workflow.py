#!/usr/bin/env python3
"""Create and verify lightweight Wuji Legion workflow artifacts.

This script intentionally keeps the workflow small. It is for complex LEGION_TASK
runs that need an audit trail, not for ordinary one-shot tasks.
"""

from __future__ import annotations

import argparse
import json
import re
from datetime import datetime, timezone
from pathlib import Path


REQUIRED_FILES = ("contract.md", "state.json", "final-report.md")
REQUIRED_DIRS = ("packets", "results")


def now_iso() -> str:
    return datetime.now(timezone.utc).replace(microsecond=0).isoformat()


def slugify(value: str) -> str:
    slug = re.sub(r"[^a-z0-9]+", "-", value.lower()).strip("-")
    return slug[:64].strip("-") or "workflow"


def write_new(path: Path, content: str) -> None:
    if not path.exists():
        path.write_text(content, encoding="utf-8")


def load_state(run_dir: Path) -> dict:
    path = run_dir / "state.json"
    if not path.is_file():
        raise SystemExit(f"Missing state.json: {path}")
    return json.loads(path.read_text(encoding="utf-8"))


def save_state(run_dir: Path, state: dict) -> None:
    state["updated_at"] = now_iso()
    (run_dir / "state.json").write_text(
        json.dumps(state, ensure_ascii=False, indent=2) + "\n",
        encoding="utf-8",
    )


def command_new(args: argparse.Namespace) -> int:
    slug = slugify(args.slug or args.title)
    run_dir = Path(args.root) / slug
    (run_dir / "packets").mkdir(parents=True, exist_ok=True)
    (run_dir / "results").mkdir(parents=True, exist_ok=True)

    state = {
        "title": args.title,
        "slug": slug,
        "status": "planned",
        "created_at": now_iso(),
        "updated_at": now_iso(),
        "approval": {"required": False, "granted": None, "notes": ""},
        "packets": [],
        "verification": {"status": "not_started", "checks": []},
    }
    write_new(
        run_dir / "contract.md",
        f"""# Wuji Workflow: {args.title}

## Goal

## Success Criteria

## Constraints

## Risks And Approval Gates

## Packet Policy

- Each packet must be bounded and non-overlapping.
- Each packet must include do, do_not, expected_output, and verification.
- Reference files remain read-only unless the user explicitly authorizes edits.

## Final Verification
""",
    )
    write_new(
        run_dir / "final-report.md",
        f"""# Final Report: {args.title}

## Outcome

## Accepted

## Rejected

## Conflicts

## Verification Evidence

## Remaining Risks
""",
    )
    if not (run_dir / "state.json").exists():
        save_state(run_dir, state)
    print(run_dir)
    return 0


def command_packet(args: argparse.Namespace) -> int:
    run_dir = Path(args.workflow_dir)
    packet_id = re.sub(r"[^a-zA-Z0-9_.-]+", "-", args.packet_id).strip("-")
    if not packet_id:
        raise SystemExit("packet_id cannot be empty")
    packet_path = run_dir / "packets" / f"{packet_id}.md"
    write_new(
        packet_path,
        f"""# Packet {packet_id}

## Objective
{args.objective}

## Context

## Files Or Sources

## Do

## Do Not

## Expected Output

## Verification

## Status
pending
""",
    )
    state = load_state(run_dir)
    packets = state.setdefault("packets", [])
    if not any(p.get("id") == packet_id for p in packets):
        packets.append({"id": packet_id, "objective": args.objective, "status": "pending"})
    save_state(run_dir, state)
    print(packet_path)
    return 0


def command_result(args: argparse.Namespace) -> int:
    run_dir = Path(args.workflow_dir)
    result_id = re.sub(r"[^a-zA-Z0-9_.-]+", "-", args.result_id).strip("-")
    if not result_id:
        raise SystemExit("result_id cannot be empty")
    result_path = run_dir / "results" / f"{result_id}.md"
    write_new(
        result_path,
        f"""# Result {result_id}

## Accepted

## Rejected

## Decisions

## Verification

## Risks
""",
    )
    print(result_path)
    return 0


def command_collect(args: argparse.Namespace) -> int:
    run_dir = Path(args.workflow_dir)
    result_files = sorted((run_dir / "results").glob("*.md"))
    lines = [f"# Integration Checklist: {run_dir.name}", ""]
    for path in result_files:
        lines.extend([f"## {path.stem}", ""])
        text = path.read_text(encoding="utf-8")
        snippets = [
            line.strip()
            for line in text.splitlines()
            if line.strip().startswith(("#", "-", "*"))
            or any(marker in line.lower() for marker in ("accepted", "rejected", "decision", "risk", "verification"))
        ][:40]
        lines.extend(snippets or ["Inspect this result manually."])
        lines.append("")
    lines.extend(["## Final Decisions", "", "Accepted:", "", "Rejected:", "", "Remaining risks:", ""])
    output = run_dir / "integration-checklist.md"
    output.write_text("\n".join(lines), encoding="utf-8")
    print(output)
    return 0


def command_verify(args: argparse.Namespace) -> int:
    run_dir = Path(args.workflow_dir)
    failures: list[str] = []
    if not run_dir.is_dir():
        failures.append(f"Missing workflow directory: {run_dir}")
    for name in REQUIRED_FILES:
        path = run_dir / name
        if not path.is_file():
            failures.append(f"Missing file: {path}")
        elif not path.read_text(encoding="utf-8").strip():
            failures.append(f"Empty file: {path}")
    for name in REQUIRED_DIRS:
        path = run_dir / name
        if not path.is_dir():
            failures.append(f"Missing directory: {path}")
    try:
        state = load_state(run_dir)
    except Exception as exc:  # noqa: BLE001 - produce a user-readable verifier error.
        failures.append(str(exc))
        state = {}
    for key in ("title", "slug", "status", "approval", "packets", "verification"):
        if key not in state:
            failures.append(f"Missing state key: {key}")

    packet_files = sorted((run_dir / "packets").glob("*.md")) if (run_dir / "packets").is_dir() else []
    result_files = sorted((run_dir / "results").glob("*.md")) if (run_dir / "results").is_dir() else []
    if args.stage == "final":
        if not packet_files:
            failures.append("Final verification requires at least one packet file.")
        if not result_files:
            failures.append("Final verification requires at least one result file.")
        final_report = run_dir / "final-report.md"
        if final_report.is_file() and "Verification Evidence" not in final_report.read_text(encoding="utf-8"):
            failures.append("final-report.md must include Verification Evidence.")
    if failures:
        print("Wuji workflow verification failed:")
        for failure in failures:
            print(f"- {failure}")
        return 1
    print(f"Wuji workflow verification passed ({args.stage}): {run_dir}")
    return 0


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(description=__doc__)
    sub = parser.add_subparsers(dest="command", required=True)

    new = sub.add_parser("new", help="Create a workflow artifact directory.")
    new.add_argument("title")
    new.add_argument("--root", default="outputs/workflows")
    new.add_argument("--slug")
    new.set_defaults(func=command_new)

    packet = sub.add_parser("packet", help="Create a packet file.")
    packet.add_argument("workflow_dir")
    packet.add_argument("packet_id")
    packet.add_argument("objective")
    packet.set_defaults(func=command_packet)

    result = sub.add_parser("result", help="Create a result file.")
    result.add_argument("workflow_dir")
    result.add_argument("result_id")
    result.set_defaults(func=command_result)

    collect = sub.add_parser("collect", help="Collect result files into an integration checklist.")
    collect.add_argument("workflow_dir")
    collect.set_defaults(func=command_collect)

    verify = sub.add_parser("verify", help="Verify workflow artifact completeness.")
    verify.add_argument("workflow_dir")
    verify.add_argument("--stage", choices=("scaffold", "final"), default="scaffold")
    verify.set_defaults(func=command_verify)
    return parser


def main() -> int:
    parser = build_parser()
    args = parser.parse_args()
    return args.func(args)


if __name__ == "__main__":
    raise SystemExit(main())
