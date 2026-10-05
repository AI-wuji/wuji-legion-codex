import argparse
import hashlib
import json
import subprocess
import sys
import uuid
from pathlib import Path

from execution_baseline import ROOT, load_active
from p6_package_validation import strict_json


EXECUTABLE = ROOT / ".dev/runtime-target/debug/wuji4.exe"
PACK = ROOT / "legion"


def route(kind):
    active = load_active()
    manifest = strict_json((PACK / "task-entry.json").read_text(encoding="utf-8"))
    if (manifest.get("schema_version") != 1 or manifest.get("installed") is not False
            or manifest.get("automatic_discovery_enabled") is not False or manifest.get("P7") is not False):
        raise ValueError("Only the isolated uninstalled task entry is supported")
    routes = manifest.get("routes")
    if not isinstance(routes, dict) or set(routes) != {"chat", "rewrite", "engineering", "research", "documents", "media", "evolution"}:
        raise ValueError("The complete bounded task route set is required")
    if routes["chat"] is not None or routes["rewrite"] is not None:
        raise ValueError("Simple tasks must not load professional references")
    if kind not in routes:
        raise ValueError("Unknown task kind; no all-expert fallback")
    paths = [manifest.get("entry")]
    if kind not in ("chat", "rewrite"):
        paths.append(routes[kind])
    context = []
    for relative in paths:
        if not isinstance(relative, str) or not relative or Path(relative).is_absolute():
            raise ValueError("Task guidance requires a nonempty relative pack path")
        path = (PACK / relative).resolve()
        if not path.is_relative_to(PACK.resolve()) or not path.is_file():
            raise ValueError("Task guidance must exist within the independent pack")
        data = path.read_bytes()
        if len(data) > 32 * 1024:
            raise ValueError("Task guidance exceeds the bounded context target")
        context.append({"path": path.relative_to(ROOT).as_posix(), "sha256": hashlib.sha256(data).hexdigest(),
                        "content": data.decode("utf-8"), "bytes": len(data)})
    simple = kind in ("chat", "rewrite")
    return {"communicator": "aji", "kind": kind, "state": "guidance_loaded_not_executed",
            "active_plan_version": active["version"], "active_plan_sha256": active["sha256"],
            "direct_answer": simple, "staff_graph_required": False, "context": context,
            "actual_agents_started": 0, "model_generation_submitted": False, "execution_authority": False,
            "P7": False, "not_done": ["model execution", "professional validation", "installation"]}


def isolated_workspace(value):
    path = Path(value).resolve(strict=True)
    development = (ROOT / ".dev").resolve(strict=True)
    if not path.is_dir() or path == development or not path.is_relative_to(development):
        raise ValueError("Task execution requires an existing separate project .dev workspace")
    parents = [development / name for name in ("core-test-workspaces", "legion-task-workspaces")]
    if not any(path != parent.resolve() and path.is_relative_to(parent.resolve()) for parent in parents):
        raise ValueError("Only dedicated task workspaces are allowed; toolchain and runtime directories are protected")
    return path


def call(*arguments):
    if not EXECUTABLE.is_file():
        raise ValueError("Build the existing isolated Rust CLI first; no auto-install")
    completed = subprocess.run([str(EXECUTABLE), *map(str, arguments)], cwd=ROOT, capture_output=True,
                               encoding="utf-8", timeout=30, creationflags=getattr(subprocess, "CREATE_NO_WINDOW", 0))
    if completed.returncode:
        error = strict_json(completed.stderr)
        raise ValueError(f"{error['error']}: {error['detail']}")
    return strict_json(completed.stdout)


def run_copy(workspace, task, source, output, confirmation):
    if confirmation != "confirm-isolated-legion-file-copy":
        raise ValueError("An explicit scoped confirmation is required; a skill does not authorize actions")
    workspace = isolated_workspace(workspace)
    if not task or len(task.encode("utf-8")) > 128 or any(ord(character) < 32 for character in task):
        raise ValueError("Task identity must be nonempty and bounded")
    guidance = route("engineering")
    state = call("init", workspace)
    reference = call("register-input", workspace, "user/legion-copy-source", source)
    role = state["admitted_local_program_ref"]
    plan = {"schema_version": 1, "contract_type": "WorkflowPlan", "payload": {
        "workflow_id": task, "version": "1", "catalog_version": role["release"],
        "nodes": [{"id": "copy", "role_ref": role, "owner": "local-executor", "revision": 1,
                   "inputs": [reference], "read_roots": [source], "write_roots": [output],
                   "acceptance_ids": ["local.file-readable", "local.hash-current", "local.utf8"],
                   "status": "planned", "execution_form": "program"}], "edges": [],
        "budget": {"max_candidates": 1, "max_query_objects": 8, "max_hops": 1,
                   "target_contract_bytes": 16384, "target_evidence_bytes": 16384, "hard_bytes_limit": None,
                   "measured_tokens": None, "effective_token_limit": None, "wall_ms": 60000,
                   "max_parallel_workers": 1, "max_extra_retries": 0, "max_point_revisions": 2, "max_no_progress": 1}},
        "metadata": {"id": task, "type": "WorkflowPlan", "scope": state["scope"], "schema_version": 1,
                     "revision": 1, "owner": "aji-local", "authority": "review_proposal", "status": "proposal",
                     "source_refs": [], "evidence_refs": [], "created_at_utc_ms": 1, "updated_at_utc_ms": 1,
                     "valid_until_utc_ms": None, "classification": "project_private", "parent_refs": []}}
    directory = ROOT / ".dev/legion-invocations" / uuid.uuid4().hex
    directory.mkdir(parents=True)
    proposal = directory / "workflow.json"
    proposal.write_text(json.dumps(plan, ensure_ascii=False), encoding="utf-8")
    plan["metadata"]["content_hash"] = call("object-hash", proposal)["sha256"]
    proposal.write_text(json.dumps(plan, ensure_ascii=False), encoding="utf-8")
    call("plan", workspace, proposal)
    executed = call("run-local", workspace, task, "copy", source, output)
    summary = call("execution-summary", workspace, task)
    completed = summary["result_state"] == "completed_bounded_task"
    return {"communicator": "aji", "operation": "isolated-legion-utf8-file-copy", "context_refs": [
        {key: item[key] for key in ("path", "sha256", "bytes")} for item in guidance["context"]],
        "active_plan_version": guidance["active_plan_version"], "actual_execution": executed,
        "summary": summary, "completed": completed, "model_generation_submitted": False,
        "user_summary": ("完成声明范围内的本地文件任务，并核验当前内容/hash/UTF-8；未执行模型修复、专业质量评审或安装4.0。"
                         if completed else "当前本地文件任务未通过完成检查；保留未满足项和历史回执，不自动重做、修复产物或宣称已完成。"),
        "task_state_authority": "existing Rust/SQLite workspace only; no second task database",
        "P7": False, "shutdown": False}


def main():
    parser = argparse.ArgumentParser(description="Lightweight isolated task entry; existing CLI owns task state")
    commands = parser.add_subparsers(dest="command", required=True)
    routing = commands.add_parser("route")
    routing.add_argument("kind")
    copying = commands.add_parser("run-copy")
    copying.add_argument("workspace")
    copying.add_argument("task")
    copying.add_argument("source")
    copying.add_argument("output")
    copying.add_argument("confirmation")
    summary = commands.add_parser("summary")
    summary.add_argument("workspace")
    summary.add_argument("task")
    arguments = parser.parse_args()
    try:
        if arguments.command == "route":
            result = route(arguments.kind)
        elif arguments.command == "summary":
            result = call("execution-summary", isolated_workspace(arguments.workspace), arguments.task)
        else:
            result = run_copy(arguments.workspace, arguments.task, arguments.source, arguments.output, arguments.confirmation)
        print(json.dumps(result, ensure_ascii=False))
    except (ValueError, OSError, subprocess.SubprocessError) as error:
        print(json.dumps({"success": False, "error": str(error), "P7": False}, ensure_ascii=False), file=sys.stderr)
        raise SystemExit(1)


if __name__ == "__main__":
    main()
