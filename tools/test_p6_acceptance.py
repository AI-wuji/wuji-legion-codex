from __future__ import annotations

import json
from pathlib import Path
import tempfile
import unittest

from p6_acceptance import ROOT, SEMANTIC_CASES, SEMANTIC_PYTHON_CASES, SUPPORTING_CASES, SUPPORTING_PYTHON_CASES, digest, evaluate, semantic_cases, source_snapshot


class P6AcceptanceTests(unittest.TestCase):
    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory(prefix="p6-ledger-test-", dir=ROOT / ".dev")
        self.root = Path(self.temporary.name).resolve()
        self.root.relative_to((ROOT / ".dev").resolve())
        self.addCleanup(self.temporary.cleanup)
        (self.root / "docs").mkdir()
        (self.root / "outputs/p6").mkdir(parents=True)
        self.specs = {"tests": [self.spec(identifier) for identifier in (*SEMANTIC_CASES, *SEMANTIC_PYTHON_CASES)] +
                      [self.spec("T20"), self.spec("T95", kind="actual_professional_artifact"),
                       self.spec("T43", gate="G7"), self.spec("T44", gate="G7"), self.spec("T74", gate="G7")]}
        (self.root / "docs/acceptance-map.json").write_text(json.dumps(self.specs), encoding="utf-8")
        names = [name for identifier in SEMANTIC_CASES for name in semantic_cases(identifier)] + SUPPORTING_CASES["T20"]
        rust = self.root / "outputs/p6/rust.log"
        rust.write_text("\n".join(f"test {name} ... ok" for name in names) + "\n", encoding="utf-8")
        python = self.root / "outputs/p6/python.log"
        python_names = ["unit_test.Fake.test_one", *[name for cases in SEMANTIC_PYTHON_CASES.values() for name in cases]]
        python.write_text("\n".join(f"{name.rsplit('.', 1)[-1]} ({name}) ... ok" for name in python_names) + "\n", encoding="utf-8")
        snapshot = source_snapshot(self.root)
        self.receipt = {
            "acceptance_spec_sha256": digest(self.root / "docs/acceptance-map.json"),
            "source_hashes": snapshot, "source_hashes_after": dict(snapshot),
            "rust": {"exit_code": 0, "cases": dict.fromkeys(names, "passed"),
                     "log_path": "outputs/p6/rust.log", "log_sha256": digest(rust)},
            "python": {"passed": True, "exit_code": 0, "cases": dict.fromkeys(python_names, "passed"),
                       "log_path": "outputs/p6/python.log", "log_sha256": digest(python)},
        }

    @staticmethod
    def spec(identifier, gate="G6", kind="deterministic_and_task_relevant_evidence"):
        return {"id": identifier, "name": identifier, "gate": gate, "execution_class": kind,
                "normal_spec": "normal", "negative_spec": "negative"}

    def report(self):
        return evaluate(self.root, self.specs, self.receipt)

    def test_complete_semantics_only_and_partial_never_promoted(self):
        report = self.report()
        self.assertTrue(report["evidence_valid"])
        self.assertEqual(report["summary"]["passed_for_g6"], len(SEMANTIC_CASES) + len(SEMANTIC_PYTHON_CASES))
        self.assertEqual(report["summary"]["status_counts"]["partial"], 1)
        self.assertEqual(report["summary"]["status_counts"]["blocked_external"], 1)
        self.assertFalse(report["summary"]["g6_scenarios_passed"])

    def test_p7_is_deferred_not_deleted_or_treated_as_g6(self):
        report = self.report()
        self.assertEqual(report["summary"]["total"], len(SEMANTIC_CASES) + len(SEMANTIC_PYTHON_CASES) + 5)
        self.assertEqual(report["summary"]["required_for_g6"], len(SEMANTIC_CASES) + len(SEMANTIC_PYTHON_CASES) + 2)
        self.assertEqual(report["summary"]["status_counts"]["deferred_p7"], 3)

    def test_changed_source_invalidates_previously_passing_evidence(self):
        (self.root / "Cargo.toml").write_text("changed", encoding="utf-8")
        report = self.report()
        self.assertFalse(report["evidence_valid"])
        self.assertEqual(report["summary"]["passed_for_g6"], 0)

    def test_failed_rust_command_is_not_saved_as_pass(self):
        self.receipt["rust"]["exit_code"] = 1
        self.assertFalse(self.report()["evidence_valid"])

    def test_changed_or_missing_logs_invalidate_evidence(self):
        (self.root / "outputs/p6/rust.log").write_text("forged", encoding="utf-8")
        self.assertFalse(self.report()["evidence_valid"])

    def test_forged_case_map_cannot_replace_actual_log(self):
        self.receipt["rust"]["cases"]["not-actually-run"] = "passed"
        report = self.report()
        self.assertFalse(report["checks"]["rust_case_records_match_log"])
        self.assertEqual(report["summary"]["passed_for_g6"], 0)

    def test_sources_changing_during_execution_fail_closed(self):
        self.receipt["source_hashes_after"]["changed"] = "0" * 64
        self.assertFalse(self.report()["evidence_valid"])

    def test_python_case_map_cannot_replace_the_actual_verbose_log(self):
        self.receipt["python"]["cases"]["not.actually.executed"] = "passed"
        self.assertFalse(self.report()["checks"]["python_case_records_match_log"])
        self.assertEqual(self.report()["summary"]["passed_for_g6"], 0)

    def test_python_failure_exit_cannot_be_overridden_by_success_flag(self):
        self.receipt["python"]["exit_code"] = 1
        self.assertFalse(self.report()["checks"]["python_suite_passed"])

    def test_t53_requires_real_process_crash_and_dispatch_guard_not_only_component_success(self):
        missing = semantic_cases("T53")[1]
        del self.receipt["rust"]["cases"][missing]
        path = self.root / self.receipt["rust"]["log_path"]
        path.write_text(path.read_text(encoding="utf-8").replace(f"test {missing} ... ok\n", ""), encoding="utf-8")
        self.receipt["rust"]["log_sha256"] = digest(path)
        report = self.report()
        self.assertTrue(report["evidence_valid"])
        row = next(row for row in report["matrix"] if row["id"] == "T53")
        self.assertEqual(row["status"], "failed_or_stale")

    def test_received_delta_acceptance_requires_all_owner_and_dependency_cases(self):
        for identifier in ("T77", "T84"):
            missing = semantic_cases(identifier)[1]
            del self.receipt["rust"]["cases"][missing]
            path = self.root / self.receipt["rust"]["log_path"]
            path.write_text(path.read_text(encoding="utf-8").replace(f"test {missing} ... ok\n", ""), encoding="utf-8")
            self.receipt["rust"]["log_sha256"] = digest(path)
            row = next(row for row in self.report()["matrix"] if row["id"] == identifier)
            self.assertEqual(row["status"], "failed_or_stale")


    def test_python_semantic_workflows_require_actual_records_not_supporting_rust(self):
        missing = SEMANTIC_PYTHON_CASES["T83"][0]
        del self.receipt["python"]["cases"][missing]
        path = self.root / self.receipt["python"]["log_path"]
        path.write_text("\n".join(line for line in path.read_text(encoding="utf-8").splitlines() if missing not in line) + "\n", encoding="utf-8")
        self.receipt["python"]["log_sha256"] = digest(path)
        report = self.report()
        self.assertTrue(report["evidence_valid"])
        row = next(row for row in report["matrix"] if row["id"] == "T83")
        self.assertEqual(row["status"], "failed_or_stale")

    def test_t21_requires_actual_related_task_workflow_not_only_persistence_components(self):
        missing = SEMANTIC_PYTHON_CASES["T21"][0]
        del self.receipt["python"]["cases"][missing]
        path = self.root / self.receipt["python"]["log_path"]
        path.write_text("\n".join(line for line in path.read_text(encoding="utf-8").splitlines() if missing not in line) + "\n", encoding="utf-8")
        self.receipt["python"]["log_sha256"] = digest(path)
        row = next(row for row in self.report()["matrix"] if row["id"] == "T21")
        self.assertEqual(row["status"], "failed_or_stale")

    def test_t22_project_scope_evidence_does_not_claim_complete_user_acl(self):
        self.specs["tests"].append(self.spec("T22"))
        path = self.root / "docs/acceptance-map.json"
        path.write_text(json.dumps(self.specs), encoding="utf-8")
        self.receipt["acceptance_spec_sha256"] = digest(path)
        self.receipt["source_hashes"] = source_snapshot(self.root)
        self.receipt["source_hashes_after"] = dict(self.receipt["source_hashes"])
        python_log = self.root / self.receipt["python"]["log_path"]
        for case in SUPPORTING_PYTHON_CASES["T22"]:
            self.receipt["python"]["cases"][case] = "passed"
            python_log.write_text(python_log.read_text(encoding="utf-8") + f"{case.rsplit('.', 1)[-1]} ({case}) ... ok\n", encoding="utf-8")
        self.receipt["python"]["log_sha256"] = digest(python_log)
        row = next(row for row in self.report()["matrix"] if row["id"] == "T22")
        self.assertEqual(row["status"], "partial")
        self.assertEqual(row["missing_scope"], "per_user_identity_and_acl")

    def test_local_revalidation_requires_complete_workflow_and_negative_cases(self):
        missing = semantic_cases("T18")[1]
        del self.receipt["rust"]["cases"][missing]
        path = self.root / self.receipt["rust"]["log_path"]
        path.write_text(path.read_text(encoding="utf-8").replace(f"test {missing} ... ok\n", ""), encoding="utf-8")
        self.receipt["rust"]["log_sha256"] = digest(path)
        row = next(row for row in self.report()["matrix"] if row["id"] == "T18")
        self.assertEqual(row["status"], "failed_or_stale")

    def test_missing_professional_implementation_is_not_blanket_external_blocker(self):
        self.specs["tests"].append(self.spec("T28", kind="actual_professional_artifact"))
        path = self.root / "docs/acceptance-map.json"
        path.write_text(json.dumps(self.specs), encoding="utf-8")
        self.receipt["acceptance_spec_sha256"] = digest(path)
        self.receipt["source_hashes"] = source_snapshot(self.root)
        self.receipt["source_hashes_after"] = dict(self.receipt["source_hashes"])
        row = next(row for row in self.report()["matrix"] if row["id"] == "T28")
        self.assertEqual(row["status"], "not_run")
        self.assertEqual(row["missing_scope"], "internal_workflow_and_matched_artifact_evidence")

    def test_t85_hash_checks_remain_partial_instead_of_complete_scenario(self):
        for identifier in ("T85",):
            self.specs["tests"].append(self.spec(identifier))
        path = self.root / "docs/acceptance-map.json"
        path.write_text(json.dumps(self.specs), encoding="utf-8")
        self.receipt["acceptance_spec_sha256"] = digest(path)
        self.receipt["source_hashes"] = source_snapshot(self.root)
        self.receipt["source_hashes_after"] = dict(self.receipt["source_hashes"])
        log_path = self.root / self.receipt["python"]["log_path"]
        with log_path.open("a", encoding="utf-8") as handle:
            for identifier in ("T85",):
                for case in SUPPORTING_PYTHON_CASES[identifier]:
                    self.receipt["python"]["cases"][case] = "passed"
                    handle.write(f"{case.rsplit('.', 1)[-1]} ({case}) ... ok\n")
        self.receipt["python"]["log_sha256"] = digest(log_path)
        report = self.report()
        self.assertTrue(report["evidence_valid"])
        for identifier in ("T85",):
            row = next(row for row in report["matrix"] if row["id"] == identifier)
            self.assertEqual(row["status"], "partial")
            self.assertTrue(row["missing_scope"])

if __name__ == "__main__":
    unittest.main(verbosity=2)
