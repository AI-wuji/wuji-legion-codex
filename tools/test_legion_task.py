import json
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path
from unittest import mock

import execution_baseline
import run_legion_task as legion
import test_core_cli


class LegionTaskTests(unittest.TestCase):
    def setUp(self):
        self.fixture = test_core_cli.CoreCliTests("test_complete_copy_repeat_and_current_file")
        self.fixture.setUp()
        self.root = self.fixture.workspace

    def test_simple_and_professional_route_load_only_needed_context_without_agents(self):
        simple = legion.route("chat")
        self.assertTrue(simple["direct_answer"])
        self.assertFalse(simple["staff_graph_required"])
        self.assertEqual(len(simple["context"]), 1)
        professional = legion.route("engineering")
        self.assertEqual(len(professional["context"]), 2)
        self.assertIn("engineering.md", professional["context"][1]["path"])
        self.assertEqual(professional["actual_agents_started"], 0)
        self.assertFalse(professional["execution_authority"])
        for kind in ("research", "documents", "media", "evolution"):
            related = legion.route(kind)
            self.assertEqual(len(related["context"]), 2)
            self.assertIn(f"{kind}.md", related["context"][1]["path"])
            self.assertFalse(related["model_generation_submitted"])
        with self.assertRaises(ValueError):
            legion.route("all-experts")

    def test_core_skill_is_the_only_legion_entry_without_a_legacy_runtime(self):
        core_skill = (legion.PACK / "skills/wuji-legion-4-0/SKILL.md").read_text(encoding="utf-8")
        self.assertIn("name: wuji-legion-4-0", core_skill)
        self.assertIn("唯一的无极军团入口", core_skill)
        self.assertIn("不要求口令", core_skill)
        self.assertIn("懒人行动", core_skill)
        self.assertNotIn("wuji-legion-codex-3-0", core_skill)
        self.assertNotIn("并列共存", core_skill)

    def test_actual_thin_entry_writes_validates_summarizes_and_replays_without_second_invocation(self):
        (self.root / "source.txt").write_bytes("真实隔离军团任务，不是模型生成".encode("utf-8"))
        arguments = (self.root, "legion-task", "source.txt", "output.txt", "confirm-isolated-legion-file-copy")
        first = legion.run_copy(*arguments)
        repeated = legion.run_copy(*arguments)
        self.assertEqual(first["summary"]["result_state"], "completed_bounded_task")
        self.assertTrue(first["completed"])
        self.assertEqual(repeated["summary"]["recorded_invocations"], 1)
        self.assertEqual(first["summary"]["unclosed_slots"], 0)
        self.assertFalse(first["model_generation_submitted"])
        self.assertFalse(first["P7"])
        self.assertEqual((self.root / "output.txt").read_bytes(), (self.root / "source.txt").read_bytes())
        (self.root / "output.txt").write_text("changed after validation", encoding="utf-8")
        summary = legion.call("execution-summary", self.root, "legion-task")
        self.assertEqual(summary["result_state"], "incomplete")
        self.assertTrue(summary["unmet_nodes"])
        with self.assertRaisesRegex(ValueError, "ValidationStale"):
            legion.run_copy(*arguments)
        self.assertEqual((self.root / "output.txt").read_text(encoding="utf-8"), "changed after validation")

    def test_actual_summary_recheck_does_not_keep_success_text_after_output_changes(self):
        (self.root / "source.txt").write_text("Real current-summary race input", encoding="utf-8")
        actual_call = legion.call

        def mutate_after_real_validation(*arguments):
            result = actual_call(*arguments)
            if arguments[0] == "run-local":
                (self.root / "output.txt").write_text("changed after validation", encoding="utf-8")
            return result

        with mock.patch.object(legion, "call", side_effect=mutate_after_real_validation):
            stale = legion.run_copy(self.root, "legion-summary-race", "source.txt", "output.txt", "confirm-isolated-legion-file-copy")
        self.assertFalse(stale["completed"])
        self.assertEqual(stale["summary"]["result_state"], "incomplete")
        self.assertEqual(stale["summary"]["recorded_invocations"], 1)
        self.assertIn("未通过完成检查", stale["user_summary"])
        self.assertEqual((self.root / "output.txt").read_text(encoding="utf-8"), "changed after validation")

    def test_wrong_permission_outside_workspace_and_unassigned_parent_paths_are_rejected(self):
        (self.root / "source.txt").write_text("safe", encoding="utf-8")
        with self.assertRaises(ValueError):
            legion.run_copy(self.root, "wrong", "source.txt", "output.txt", "proposal-is-not-permission")
        with self.assertRaises(ValueError):
            legion.isolated_workspace(execution_baseline.ROOT)
        with self.assertRaises(ValueError):
            legion.isolated_workspace(execution_baseline.ROOT / ".dev/runtime-target/debug")
        with self.assertRaises(ValueError):
            legion.run_copy(self.root, "escape", "source.txt", "../escape.txt", "confirm-isolated-legion-file-copy")
        self.assertFalse((self.root / "output.txt").exists())

    def test_actual_command_returns_nonzero_for_unknown_route_without_configuration_change(self):
        result = subprocess.run([sys.executable, str(execution_baseline.ROOT / "tools/run_legion_task.py"), "route", "unknown"],
                                capture_output=True, encoding="utf-8", timeout=30,
                                creationflags=getattr(subprocess, "CREATE_NO_WINDOW", 0))
        self.assertNotEqual(result.returncode, 0)
        self.assertFalse(json.loads(result.stderr)["success"])

    def test_followup_manifest_rejects_duplicate_authority_or_activation_before_loading(self):
        manifest = json.loads((legion.PACK / "task-entry.json").read_text(encoding="utf-8"))
        with tempfile.TemporaryDirectory(dir=self.root) as temporary:
            pack = Path(temporary)
            path = pack / "task-entry.json"
            with mock.patch.object(legion, "PACK", pack):
                for field in ("installed", "automatic_discovery_enabled", "P7"):
                    changed = dict(manifest, **{field: True})
                    path.write_text(json.dumps(changed), encoding="utf-8")
                    with self.assertRaisesRegex(ValueError, "isolated uninstalled"):
                        legion.route("engineering")
                    value = json.dumps(manifest)
                    path.write_text(value[:-1] + "," + json.dumps(field) + ":false}", encoding="utf-8")
                    with self.assertRaisesRegex(ValueError, "Duplicate"):
                        legion.route("engineering")

    def test_malformed_routes_do_not_fall_back_to_all_experts_or_load_external_content(self):
        manifest = json.loads((legion.PACK / "task-entry.json").read_text(encoding="utf-8"))
        with tempfile.TemporaryDirectory(dir=self.root) as temporary:
            pack = Path(temporary)
            entry = pack / manifest["entry"]
            entry.parent.mkdir(parents=True)
            entry.write_text("Isolated entry fixture", encoding="utf-8")
            with mock.patch.object(legion, "PACK", pack):
                for routes in ([], {}, {**manifest["routes"], "engineering": None},
                               {**manifest["routes"], "engineering": "../outside.md"},
                               {**manifest["routes"], "chat": "references/media.md"}):
                    (pack / "task-entry.json").write_text(json.dumps(dict(manifest, routes=routes)), encoding="utf-8")
                    with self.assertRaises(ValueError):
                        legion.route("engineering")


if __name__ == "__main__":
    unittest.main()
