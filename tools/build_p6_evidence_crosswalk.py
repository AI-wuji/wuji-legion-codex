from __future__ import annotations

import json
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
ACCEPTANCE = ROOT / "docs/acceptance-map.json"
OUTPUT = ROOT / "outputs/p6/evidence-crosswalk.json"


def main() -> None:
    acceptance = json.loads(ACCEPTANCE.read_text(encoding="utf-8"))
    tests = acceptance["tests"]
    from p6_acceptance import evaluate
    receipt = json.loads((ROOT / "outputs/p6/test-execution.json").read_text(encoding="utf-8"))
    execution = evaluate(ROOT, acceptance, receipt)
    matrix = [
        {"id": row["id"], "gate": row["gate"], "execution_class": row["execution_class"],
         "authoritative_status": row["status"], "classification": row["status"],
         "required_for_g6": row["required_for_g6"], "evidence_refs": row["evidence_refs"]}
        for row in execution["matrix"]
    ]
    report = {
        "schema_version": 1,
        "release": "p6-evidence-crosswalk-2",
        "status": "supporting_evidence_separated_from_authoritative_acceptance",
        "authority": "Derived view of the separate current execution ledger; frozen specs unchanged. Component/partial/native/professional evidence remain distinct.",
        "execution_ledger": "outputs/p6/acceptance-execution.json",
        "execution_evidence_valid": execution["evidence_valid"],
        "summary": {
            "acceptance_total": len(matrix),
            "authoritative_not_run": sum(item["classification"] == "not_run" for item in matrix),
            "authoritative_passed": execution["summary"]["passed_for_g6"],
            "status_counts": execution["summary"]["status_counts"],
            "required_for_g6": execution["summary"]["required_for_g6"],
            "supporting_evidence_only_records": 14,
            "blocked_external_scopes": 1,
            "internal_incomplete_scopes": 3,
            "deferred_p7_scopes": 1,
        },
        "authoritative_matrix": matrix,
        "supporting_evidence_only": [
            {
                "id": "E-REVIEWED-VIDEO-BINDING",
                "kind": "actual_bound_local_video_dispatch_and_independent_decode",
                "path": "outputs/p4/video-framework-workflow-evidence.json",
                "boundary": "Current local process/tool/schema/implementation binding plus real eight-frame video and drift refusals; not all MCP/API tool discovery or full T85/professional acceptance.",
            },
            {
                "id": "E-OFFICE-DOCUMENT-BOUNDED",
                "kind": "actual_bounded_word_and_xlsx_document_workflow",
                "path": "outputs/p4/office-document-workflow-evidence.json",
                "boundary": "XLSX formula/cached-value and OfficeCLI readback passed; DOCX structure/readback passed but the approved renderer was unavailable, so no Word completion or professional-effectiveness claim is made.",
            },
            {
                "id": "E-LEGION-TASK",
                "kind": "actual_bounded_isolated_utf8_task",
                "path": "outputs/p2/legion-first-task-receipt.json",
                "current_recheck": "outputs/p2/legion-first-task-current-recheck.json",
                "boundary": "Real file artifact, independent bytes/hash/UTF-8 checks and one invocation on replay; not model repair or full professional acceptance.",
            },
            {
                "id": "E-WUJI4-FRAMEWORK",
                "kind": "actual_project_local_framework_smoke",
                "path": "outputs/p2/framework-doctor-current.json",
                "boundary": "Project-local status/doctor/on-demand route plus one existing Rust/SQLite bounded file task; no resident service, second database, model dispatch, installation or professional effectiveness claim.",
            },
            {
                "id": "E-WUJI4-CAPABILITIES",
                "kind": "existing_tool_contract_view",
                "path": "outputs/p2/framework-capabilities-current.json",
                "boundary": "Read-only view of existing bounded adapters and their contracts; does not execute tools, grant permissions, prove professional effectiveness or add generic dispatch.",
            },
            {
                "id": "E-LEGION-EXPERIENCE",
                "kind": "local_exact_experience_followup_and_invalidation",
                "path": "outputs/p5/legion-experience-task-receipt.json",
                "boundary": "Local reviewed candidate used in current manifest guard/regression and refused after source change; not autonomous model evolution or public/global admission.",
            },
            {
                "id": "E-LEGION-EXPERIENCE-WORKFLOW",
                "kind": "actual_cli_related_task_and_scope_regression",
                "path": "outputs/p5/experience-workflow-evidence.json",
                "boundary": "Current log-bound T21 task adoption and T22 project/review-permit checks; T22 remains partial without a complete per-user identity/ACL layer. No autonomous model or global admission.",
            },
            {
                "id": "E-G2-U",
                "kind": "bounded_native_acceptance",
                "path": "outputs/p2/native-real-execution-report-2026-10-03.json",
                "boundary": "Supports the bounded u task lifecycle and independent acceptance only; not the 95-scenario matrix or host capability claims.",
            },
            {
                "id": "E-G3-COLD",
                "kind": "cold_catalog_quality",
                "path": "outputs/p3/p3-quality-report.json",
                "boundary": "Supports catalog structure and cold-state checks only; not expert effectiveness or runtime admission.",
            },
            {
                "id": "E-G4-LOCAL",
                "kind": "deterministic_local_adapter",
                "path": "outputs/p4/software-repair-u-evidence.json",
                "boundary": "Supports local software-repair file validation only; not ComfyUI, OfficeCLI, video, or audio professional artifacts.",
            },
            {
                "id": "E-G4-COMFYUI-BOUNDED",
                "kind": "bounded_real_comfyui_probe",
                "path": "outputs/p4/comfyui-probe-evidence.json",
                "boundary": "Supports one real local LoadImage -> ImageScale -> SaveImage chain and artifact validation only; not professional image quality or full G4 effectiveness.",
            },
            {
                "id": "E-G4-FFMPEG-BOUNDED",
                "kind": "bounded_real_ffmpeg_probe",
                "path": "outputs/p4/ffmpeg-probe-evidence.json",
                "boundary": "Supports one real local PNG -> H.264 MP4 render and decode validation only; not professional video editing, audio, or full G4 effectiveness.",
            },
            {
                "id": "E-G5-GOVERNANCE",
                "kind": "governance_core_tests",
                "path": "outputs/p5/p5-governance-report.json",
                "boundary": "Supports deterministic delta/CAS/impact/revocation mechanics only; not model or professional effectiveness.",
            },
            {
                "id": "E-RUST-MECHANISM",
                "kind": "mechanism_tests",
                "path": "outputs/p6/rust-full-tests.log",
                "boundary": "Supports Rust core mechanism regressions only; does not substitute for task-relevant native or professional acceptance.",
            },
            {
                "id": "E-PYTHON-AUDIT",
                "kind": "audit_helper_tests",
                "path": "outputs/p6/python-tool-tests.log",
                "boundary": "Supports audit/helper behavior only; does not substitute for the 95 acceptance scenarios.",
            },
        ],
        "internal_incomplete": [
            {
                "id": "B-G4-COMFYUI-HOLDOUT",
                "scope": "Professional ComfyUI artifact and quality holdout",
                "reason": "A bounded probe passed, but professional effectiveness and holdout evidence are not recorded.",
            },
            {
                "id": "B-G4-MEDIA",
                "scope": "Professional video editing, audio and REAPER real artifacts",
                "reason": "A bounded ffmpeg render passed, but professional editing/audio and holdout evidence are not recorded.",
            },
            {
                "id": "B-G6-HOLDOUT",
                "scope": "Professional holdout and effectiveness evidence",
                "reason": "No valid professional holdout execution is recorded.",
            },
        ],
        "blocked_external": [
            {
                "id": "B-G6-HOST",
                "scope": "Effective backend model, effort, quota, and fee attestations",
                "reason": "Host capability facts remain unknown and are not inferred from configuration or self-report.",
            },
        ],
        "deferred_p7": [
            {
                "id": "B-G7-INSTALL",
                "scope": "Codex installation, representative task, and recovery rehearsal",
                "reason": "P7 is not authorized and was not performed.",
            },
        ],
    }
    OUTPUT.write_text(json.dumps(report, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(json.dumps({"path": str(OUTPUT), "summary": report["summary"]}))


if __name__ == "__main__":
    main()
