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
CATALOG = ROOT / "catalog/p3"
DELEGATION_MANIFEST = CATALOG / "delegation-manifest.json"
EXPERT_CATALOG = CATALOG / "experts.json"


def _bounded_text(value, field, limit=512):
    if not isinstance(value, str) or not value.strip() or len(value.encode("utf-8")) > limit:
        raise ValueError(f"{field} must be a nonempty bounded string")
    if any(ord(character) < 32 for character in value):
        raise ValueError(f"{field} must not contain control characters")
    return value


def _bounded_list(value, field, limit=32):
    if not isinstance(value, list) or len(value) > limit or any(not isinstance(item, str) for item in value):
        raise ValueError(f"{field} must be a bounded string list")
    for item in value:
        _bounded_text(item, field, 256)
    return value


def _load_delegation_manifest():
    manifest = strict_json(DELEGATION_MANIFEST.read_text(encoding="utf-8"))
    if manifest.get("schema_version") != 1 or manifest.get("kind") != "bounded_staff_delegation_manifest":
        raise ValueError("Unsupported staff delegation manifest")
    rules = manifest.get("rules")
    if not isinstance(rules, dict) or rules.get("all_expert_fallback") is not False:
        raise ValueError("Staff delegation must disable all-expert fallback")
    recipes = manifest.get("recipes")
    if not isinstance(recipes, list) or not recipes:
        raise ValueError("Staff delegation requires bounded leader recipes")
    if manifest.get("recipe_count") != len(recipes):
        raise ValueError("Leader recipe count does not match the manifest")
    family_ids = manifest.get("leader_family_ids")
    if (not isinstance(family_ids, list) or manifest.get("leader_family_count") != len(family_ids)
            or len(set(family_ids)) != len(family_ids)):
        raise ValueError("Leader family index is malformed")
    recipe_ids = set()
    for recipe in recipes:
        if not isinstance(recipe, dict) or not isinstance(recipe.get("recipe_id"), str):
            raise ValueError("Every leader recipe requires a unique recipe_id")
        if recipe["recipe_id"] in recipe_ids:
            raise ValueError("Leader recipe ids must be unique")
        recipe_ids.add(recipe["recipe_id"])
        if not isinstance(recipe.get("domains"), list) or not recipe["domains"]:
            raise ValueError("Every leader recipe requires bounded domains")
        if not isinstance(recipe.get("typed_intents"), list) or not recipe["typed_intents"]:
            raise ValueError("Every leader recipe requires bounded typed intents")
        leader = recipe.get("leader")
        if (not isinstance(leader, dict) or not leader.get("id")
                or leader.get("family_id") not in family_ids):
            raise ValueError("Every leader recipe requires a leader identity")
        if not isinstance(recipe.get("expert_refs"), list) or not recipe["expert_refs"]:
            raise ValueError("Every leader recipe requires expert references")
        references = recipe["expert_refs"]
        role_ids = [_reference_role_id(reference) for reference in references]
        if None in role_ids or len(role_ids) != len(set(role_ids)):
            raise ValueError("Leader expert references must be valid and unique")
        if not any(_selection_rule(reference).get("required") is True for reference in references):
            raise ValueError("Every leader recipe requires a bounded primary responsibility")
        for reference in references:
            _selection_rule(reference)
        if not isinstance(recipe.get("acceptance_defaults"), list) or not recipe["acceptance_defaults"]:
            raise ValueError("Every leader recipe requires acceptance defaults")
    coverage = manifest.get("expert_coverage")
    if not isinstance(coverage, dict) or coverage.get("unbound_role_ids") != []:
        raise ValueError("Complete leader catalog must declare no unbound expert roles")
    return manifest


def _load_expert_metadata():
    catalog = strict_json(EXPERT_CATALOG.read_text(encoding="utf-8"))
    roles = catalog.get("roles")
    if catalog.get("role_count") != 57 or not isinstance(roles, list):
        raise ValueError("Cold expert catalog is malformed")
    metadata = {}
    for role in roles:
        if not isinstance(role, dict) or not isinstance(role.get("id"), str) or role["id"] in metadata:
            raise ValueError("Cold expert catalog contains an invalid or duplicate role")
        metadata[role["id"]] = role
    if len(metadata) != catalog["role_count"]:
        raise ValueError("Cold expert role count does not match the catalog")
    return catalog, metadata


def _reference_role_id(reference):
    if isinstance(reference, str):
        return reference
    if isinstance(reference, dict) and isinstance(reference.get("role_id"), str):
        return reference["role_id"]
    return None


def _validate_work_item(value, field_prefix):
    if not isinstance(value, dict):
        raise ValueError(f"{field_prefix} must be an object")
    task_id = _bounded_text(value.get("task_id", value.get("subtask_id")), f"{field_prefix}.task_id", 128)
    if "task_id" in value and "subtask_id" in value and value["task_id"] != value["subtask_id"]:
        raise ValueError("Conflicting task_id and subtask_id in one work item")
    goal = _bounded_text(value.get("goal"), f"{field_prefix}.goal", 4096)
    domain = _bounded_text(value.get("domain"), f"{field_prefix}.domain", 128)
    typed_intents = _bounded_list(value.get("typed_intents", []), f"{field_prefix}.typed_intents", 16)
    languages = _bounded_list(value.get("languages", []), f"{field_prefix}.languages", 8)
    inputs = _bounded_list(value.get("inputs", []), f"{field_prefix}.inputs", 16)
    deliverables = _bounded_list(value.get("deliverables", []), f"{field_prefix}.deliverables", 16)
    if not deliverables:
        raise ValueError(f"{field_prefix}.deliverables must not be empty")
    acceptance_ids = _bounded_list(value.get("acceptance_ids", []), f"{field_prefix}.acceptance_ids", 32)
    write_roots = _bounded_list(value.get("write_roots", []), f"{field_prefix}.write_roots", 16)
    constraints = value.get("constraints", [])
    _bounded_list(constraints, f"{field_prefix}.constraints", 32)
    depends_on = _bounded_list(value.get("depends_on", []), f"{field_prefix}.depends_on", 16)
    return {"task_id": task_id, "subtask_id": task_id, "goal": goal, "domain": domain,
            "typed_intents": typed_intents, "languages": languages, "inputs": inputs, "deliverables": deliverables,
            "acceptance_ids": acceptance_ids, "write_roots": write_roots,
            "constraints": constraints, "depends_on": depends_on}


def _validate_delegation_request(request):
    if not isinstance(request, dict):
        raise ValueError("Delegation request must be an object")
    task_id = _bounded_text(request.get("task_id"), "task_id", 128)
    task_class = request.get("task_class", "complex")
    if task_class not in ("simple", "complex"):
        raise ValueError("task_class must be simple or complex")
    parallel_cap = request.get("max_parallel_instances", 2)
    if type(parallel_cap) is not int or not 1 <= parallel_cap <= 3:
        raise ValueError("max_parallel_instances must be a construction cap in 1..3, not a native quota")
    raw_subtasks = request.get("subtasks")
    if raw_subtasks is None:
        item = _validate_work_item(request, "request")
        item["subtask_id"] = task_id
        if item["depends_on"]:
            raise ValueError("Single-task delegation cannot reference undeclared dependencies")
        return {**item, "task_id": task_id, "task_class": task_class, "composite": False,
                "max_parallel_instances": parallel_cap,
                "subtasks": [item]}
    if task_class != "complex":
        raise ValueError("Composite delegation requires task_class=complex")
    if not isinstance(raw_subtasks, list) or not 1 <= len(raw_subtasks) <= 16:
        raise ValueError("subtasks must contain 1..16 bounded task objects")
    subtasks = []
    seen = set()
    for index, raw in enumerate(raw_subtasks):
        item = _validate_work_item(raw, f"subtasks[{index}]")
        requested_id = _bounded_text(raw.get("subtask_id", item["task_id"]),
                                     f"subtasks[{index}].subtask_id", 128)
        if requested_id in seen:
            raise ValueError("subtask_id values must be unique")
        seen.add(requested_id)
        item["subtask_id"] = requested_id
        subtasks.append(item)
    known = {item["subtask_id"] for item in subtasks}
    for item in subtasks:
        if item["subtask_id"] in item["depends_on"] or any(dependency not in known for dependency in item["depends_on"]):
            raise ValueError("subtask dependencies must reference other subtasks and cannot self-reference")
    goal = _bounded_text(request.get("goal", f"复合任务 {task_id}"), "goal", 4096)
    domain = _bounded_text(request.get("domain", "composite"), "domain", 128)
    typed_intents = _bounded_list(request.get("typed_intents", []), "typed_intents", 16)
    inputs = _bounded_list(request.get("inputs", []), "inputs", 16)
    deliverables = _bounded_list(request.get("deliverables", ["subtask-receipts"]), "deliverables", 16)
    acceptance_ids = _bounded_list(request.get("acceptance_ids", []), "acceptance_ids", 32)
    write_roots = _bounded_list(request.get("write_roots", []), "write_roots", 16)
    constraints = request.get("constraints", [])
    _bounded_list(constraints, "constraints", 32)
    for item in subtasks:
        item["constraints"] = _bounded_list(list(dict.fromkeys(constraints + item["constraints"])),
                                            "inherited constraints", 32)
        item["inputs"] = _bounded_list(list(dict.fromkeys(inputs + item["inputs"])), "inherited inputs", 32)
        if "write_roots" in request and any(not any(_root_contains(root, child) for root in write_roots)
                               for child in item["write_roots"]):
            raise ValueError("Subtask write roots cannot exceed the declared parent scope")
    return {"task_id": task_id, "subtask_id": task_id, "goal": goal, "domain": domain,
            "typed_intents": typed_intents, "inputs": inputs, "deliverables": deliverables,
            "acceptance_ids": acceptance_ids, "write_roots": write_roots,
            "constraints": constraints, "task_class": task_class, "composite": True,
            "max_parallel_instances": parallel_cap,
            "subtasks": subtasks}


def _recipe_matches(manifest, task):
    domain = task["domain"]
    intents = set(task["typed_intents"])
    scored = []
    for recipe in manifest["recipes"]:
        recipe_domains = set(recipe.get("domains", []))
        recipe_intents = set(recipe.get("typed_intents", []))
        score = (3 if domain in recipe_domains else 0) + 2 * len(intents.intersection(recipe_intents))
        if score:
            scored.append((score, recipe))
    best_score = max((score for score, _ in scored), default=0)
    return [recipe for score, recipe in scored if score == best_score]


def _selection_rule(reference):
    if not isinstance(reference, dict) or not _reference_role_id(reference):
        raise ValueError("Expert references require explicit selection conditions")
    selection = reference.get("selection")
    if not isinstance(selection, dict) or not selection:
        raise ValueError("Expert selection conditions must be a nonempty object")
    if set(selection) == {"required"} and selection["required"] is True:
        return selection
    if not set(selection).issubset({"any_intents", "any_languages"}):
        raise ValueError("Expert selection permits only required or exact intent/language conditions")
    conditions = []
    for field, values in selection.items():
        conditions.extend(_bounded_list(values, f"selection.{field}", 32))
    if not conditions:
        raise ValueError("Optional expert requires at least one explicit selection condition")
    return selection


def _role_needed(reference, task):
    selection = _selection_rule(reference)
    if selection.get("required") is True:
        return True, "required_by_leader_sop"
    matched_intents = sorted(set(selection.get("any_intents", [])).intersection(task["typed_intents"]))
    matched_languages = sorted(set(selection.get("any_languages", [])).intersection(task["languages"]))
    return bool(matched_intents or matched_languages), {
        "matched_intents": matched_intents, "matched_languages": matched_languages}


def _select_experts(recipe, task, expert_metadata):
    selected = []
    cold = []
    missing = []
    references = recipe.get("expert_refs")
    if not isinstance(references, list) or not references:
        raise ValueError("Leader recipe must contain at least one expert reference")
    for reference in references:
        role_id = _reference_role_id(reference)
        if not role_id:
            missing.append(str(role_id))
            continue
        role = expert_metadata.get(role_id)
        if role is None:
            missing.append(role_id)
            continue
        needed, reason = _role_needed(reference, task)
        assignment = reference.get("assignment") if isinstance(reference, dict) else None
        if not assignment:
            assignment = f"按目录职责 {role.get('design_target') or role.get('baseline_id')} 处理任务中的对应专业边界"
        if not needed:
            cold.append({"role_id": role_id, "reason": reason, "state": "cold_candidate"})
            continue
        selected.append({
            "role_id": role_id,
            "assignment": assignment,
            "catalog_ref": "catalog/p3/experts.json",
            "baseline_id": role.get("baseline_id"),
            "kind": role.get("kind"),
            "design_target": role.get("design_target"),
            "five_elements": role.get("five_elements"),
            "anti_trigger": role.get("anti_trigger"),
            "owner_rule": role.get("owner_rule"),
            "composition_dependencies": role.get("composition_dependencies", []),
            "scope_rule": role.get("scope_rule"),
            "cancellation": role.get("cancellation"),
            "runtime_admission": False,
            "active_release": role.get("active_release") is True,
            "effectiveness": role.get("effectiveness"),
            "selection_basis": reason,
        })
    return selected, cold, missing


def _canonical_root(value):
    return str((ROOT / value).resolve()).replace("\\", "/").rstrip("/").casefold()


def _root_contains(parent, child):
    parent = _canonical_root(parent)
    child = _canonical_root(child)
    return child == parent or child.startswith(f"{parent}/")


def _write_roots_conflict(left, right):
    for first in left:
        first = _canonical_root(first)
        for second in right:
            second = _canonical_root(second)
            if first == second or first.startswith(f"{second}/") or second.startswith(f"{first}/"):
                return True
    return False


def _parallel_schedule(subtasks, branches, parallel_cap):
    pending = {item["subtask_id"]: item for item in subtasks}
    visited = set()
    blocked = {}
    gap_ids = {branch["subtask_id"] for branch in branches if branch["state"] == "selection_gap"}
    groups = []
    while pending:
        ready = [item for item in pending.values() if set(item["depends_on"]).issubset(visited)]
        if not ready:
            raise ValueError("subtask dependency graph contains a cycle")
        group = []
        for item in sorted(ready, key=lambda value: value["subtask_id"]):
            subtask_id = item["subtask_id"]
            blockers = set().union(*(set(blocked.get(dependency, [])) for dependency in item["depends_on"]))
            if subtask_id in gap_ids:
                blockers.add(subtask_id)
            if blockers:
                blocked[subtask_id] = sorted(blockers)
                visited.add(subtask_id)
                del pending[subtask_id]
                continue
            if len(group) >= parallel_cap:
                break
            if not any(_write_roots_conflict(item["write_roots"], other["write_roots"]) for other in group):
                group.append(item)
        group_ids = [item["subtask_id"] for item in group]
        if group_ids:
            groups.append(group_ids)
        visited.update(group_ids)
        for subtask_id in group_ids:
            del pending[subtask_id]
    return {"groups": groups, "parallel_groups": [group for group in groups if len(group) > 1],
            "parallel_eligible": any(len(group) > 1 for group in groups),
            "blocked_branches": blocked,
            "construction_only": True, "native_execution_observed": False,
            "max_construction_parallelism": max((len(group) for group in groups), default=0)}


def _staff_subtask(manifest, catalog, expert_metadata, task):
    matches = _recipe_matches(manifest, task)
    common = {"subtask_id": task["subtask_id"], "task_id": task["task_id"], "goal": task["goal"],
              "depends_on": task["depends_on"], "staff_graph_required": True,
              "return_path": ["expert", "leader", "staff", "aji"]}
    if len(matches) != 1:
        return {**common, "staff_decision": "selection_gap", "state": "selection_gap",
                "selection_basis": {"domain": task["domain"], "typed_intents": task["typed_intents"],
                                     "candidate_recipe_count": len(matches)}, "selected_expert_refs": [],
                "cold_expert_refs": [], "runtime_admission": False, "execution_state": "not_dispatchable",
                "model_generation_submitted": False, "execution_authority": False, "P7": False,
                "not_done": ["select one exact leader recipe for this subtask", "expert dispatch",
                             "professional validation"]}
    recipe = matches[0]
    selected, cold, missing = _select_experts(recipe, task, expert_metadata)
    if missing or not selected:
        return {**common, "staff_decision": "selection_gap", "state": "selection_gap",
                "leader_recipe": recipe["recipe_id"], "missing_expert_refs": missing,
                "selected_expert_refs": selected, "cold_expert_refs": cold, "runtime_admission": False,
                "execution_state": "not_dispatchable", "model_generation_submitted": False,
                "execution_authority": False, "P7": False,
                "not_done": ["resolve missing catalog role or necessary expert condition", "expert dispatch",
                             "professional validation"]}
    return {**common, "staff_decision": "recipe_selected", "state": "prepared_not_executed",
            "leader_recipe": {**recipe["leader"], "recipe_id": recipe["recipe_id"],
                              "acceptance_defaults": recipe["acceptance_defaults"]},
            "selected_expert_refs": selected, "cold_expert_refs": cold,
            "selection_basis": {"domain": task["domain"], "typed_intents": task["typed_intents"],
                                 "matching": manifest["rules"]["matching"],
                                 "catalog_role_count": catalog["role_count"],
                                 "leader_recipe_count": len(manifest["recipes"]),
                                 "leader_family_count": manifest.get("leader_family_count"),
                                 "cold_catalog": True, "conditional_expert_selection": True},
            "required_context": {"inputs": task["inputs"], "deliverables": task["deliverables"],
                                  "acceptance_ids": task["acceptance_ids"] or recipe["acceptance_defaults"],
                                  "write_roots": task["write_roots"], "constraints": task["constraints"]},
            "handoff_contract": {
                "aji_to_staff": "structured_task_contract",
                "staff_to_leader": "selected_recipe_and_scope",
                "leader_to_experts": "assignment_input_output_acceptance",
                "expert_to_leader": "localized_artifact_or_gap",
                "leader_to_staff": "completion_or_revision_proposal",
                "staff_to_aji": "validated_receipt_or_blocker"},
            "runtime_admission": False, "actual_agents_started": 0,
            "formal_experts_activated": 0, "model_generation_submitted": False,
            "execution_authority": False, "professional_effectiveness": "not_claimed", "P7": False,
            "not_done": ["native model execution", "expert runtime admission", "professional validation"]}


def delegate(request):
    request = _validate_delegation_request(request)
    active = load_active()
    if request["task_class"] == "simple":
        entry = route("chat")
        return {**entry, "operation": "aji-direct", "task_id": request["task_id"],
                "goal": request["goal"], "staff_decision": "direct_to_aji",
                "execution_state": "direct_answer", "return_path": ["aji"],
                "not_done": ["professional delegation", "model execution"]}

    manifest = _load_delegation_manifest()
    catalog, expert_metadata = _load_expert_metadata()
    if manifest["expert_coverage"]["source_role_count"] != catalog["role_count"]:
        raise ValueError("Leader catalog source role count does not match the expert catalog")
    workflow_count = sum(role.get("kind") == "workflow" for role in expert_metadata.values())
    if manifest.get("workflow_role_count") != workflow_count:
        raise ValueError("Leader catalog workflow role count does not match the expert catalog")
    referenced_role_ids = set()
    for recipe in manifest["recipes"]:
        for reference in recipe["expert_refs"]:
            role_id = reference if isinstance(reference, str) else reference.get("role_id")
            if isinstance(role_id, str):
                referenced_role_ids.add(role_id)
    if referenced_role_ids != set(expert_metadata):
        missing = sorted(set(expert_metadata) - referenced_role_ids)
        extra = sorted(referenced_role_ids - set(expert_metadata))
        raise ValueError(f"Leader catalog coverage drift: missing={missing}, extra={extra}")
    branches = [_staff_subtask(manifest, catalog, expert_metadata, task) for task in request["subtasks"]]
    preparation = _parallel_schedule(request["subtasks"], branches, request["max_parallel_instances"])
    common = {"communicator": "aji", "operation": "staff-delegation", "task_id": request["task_id"],
              "goal": request["goal"], "active_plan_version": active["version"],
              "active_plan_sha256": active["sha256"], "staff_graph_required": True,
              "return_path": ["expert", "leader", "staff", "aji"], "runtime_admission": False,
              "actual_agents_started": 0, "formal_experts_activated": 0,
              "model_generation_submitted": False, "execution_authority": False,
              "professional_effectiveness": "not_claimed", "P7": False,
              "shared_budget": {"max_parallel_instances": request["max_parallel_instances"],
                                "applies_to": "all_leader_and_expert_instances",
                                "effective_host_quota": "unknown", "native_admission": False},
              "parallel_preparation": preparation}
    if not request["composite"]:
        return {**branches[0], **common}
    has_gap = any(branch["state"] == "selection_gap" for branch in branches)
    for branch in branches:
        branch["dependency_blockers"] = preparation["blocked_branches"].get(branch["subtask_id"], [])
    return {**common, "state": "selection_gap" if has_gap else "prepared_not_executed",
            "staff_decision": "selection_gap" if has_gap else "recipes_selected",
            "branches": branches,
            "required_context": {field: request[field] for field in
                                 ("inputs", "deliverables", "acceptance_ids", "write_roots", "constraints")},
            "not_done": ["native model execution", "expert runtime admission", "professional validation"]}


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
    delegation = commands.add_parser("delegate")
    delegation.add_argument("request_json")
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
        elif arguments.command == "delegate":
            request_path = Path(arguments.request_json).resolve(strict=True)
            if not request_path.is_file() or request_path.stat().st_size > 32 * 1024:
                raise ValueError("Delegation request must be a bounded JSON file")
            result = delegate(strict_json(request_path.read_text(encoding="utf-8")))
        elif arguments.command == "summary":
            result = call("execution-summary", isolated_workspace(arguments.workspace), arguments.task)
        else:
            result = run_copy(arguments.workspace, arguments.task, arguments.source, arguments.output, arguments.confirmation)
        print(json.dumps(result, ensure_ascii=False))
    except (ValueError, OSError, subprocess.SubprocessError) as error:
        print(json.dumps({"success": False, "error": str(error), "P7": False}, ensure_ascii=False), file=sys.stderr)
        raise SystemExit(1)


if __name__ == "__main__":
    sys.stdout.reconfigure(encoding="utf-8")
    sys.stderr.reconfigure(encoding="utf-8")
    main()
