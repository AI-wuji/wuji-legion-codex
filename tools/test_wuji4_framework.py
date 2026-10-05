import io
import json
import sys
import tempfile
import types
import unittest
from contextlib import redirect_stdout
from pathlib import Path
from unittest import mock

import wuji4 as framework


class Wuji4FrameworkTests(unittest.TestCase):
    def test_office_failed_state_without_status_exits_nonzero_and_retains_receipt(self):
        failed = {"receipt": {"state": "failed_evidence_retained", "passed": False,
                              "errors": [{"stage": "create", "message": "failure retained"}]}, "P7": False}
        output = io.StringIO()
        arguments = ["wuji4.py", "run-office", "unused-workspace", "confirm-isolated-officecli-task"]
        with mock.patch.object(sys, "argv", arguments), mock.patch.object(framework, "_workflow_result", return_value=failed), redirect_stdout(output):
            with self.assertRaises(SystemExit) as caught:
                framework.main()
        self.assertEqual(caught.exception.code, 1)
        self.assertEqual(json.loads(output.getvalue()), failed)

    def test_confirmation_is_checked_before_professional_execution(self):
        for arguments in (["run-office", "unused-workspace", "not-confirmed"],
                          ["run-video", "unused-workspace", "source.png", "not-confirmed"],
                          ["run-office-doc", "unused-workspace", "{}", "not-confirmed"]):
            with self.subTest(arguments=arguments), mock.patch.object(sys, "argv", ["wuji4.py", *arguments]):
                with mock.patch.object(framework, "_workflow_result") as execute, mock.patch("sys.stderr", io.StringIO()):
                    with self.assertRaises(SystemExit) as caught:
                        framework.main()
                self.assertEqual(caught.exception.code, 1)
                execute.assert_not_called()

    def test_office_document_dispatch_binds_request_and_confirmation(self):
        result = {"receipt": {"passed": True, "result_state": "completed_bounded_task"}, "P7": False}
        arguments = ["wuji4.py", "run-office-doc", "workspace", '{"schema_version":1}',
                     "confirm-isolated-office-document-task", "--receipt", "receipt.json"]
        output = io.StringIO()
        with mock.patch.object(sys, "argv", arguments), mock.patch.object(
                framework, "_office_document_workflow_result", return_value=result) as execute, redirect_stdout(output):
            framework.main()
        execute.assert_called_once_with("workspace", '{"schema_version":1}', "receipt.json")
        self.assertEqual(json.loads(output.getvalue()), result)

    def test_video_result_requires_completion_and_returns_failure_nonzero(self):
        for passed in (True, False):
            result = {"receipt": {"passed": passed,
                                 "result_state": "completed_bounded_task" if passed else "failed_evidence_retained"},
                      "P7": False}
            arguments = ["wuji4.py", "run-video", "unused-workspace", "source.png", "confirm-isolated-video-task"]
            output = io.StringIO()
            with self.subTest(passed=passed), mock.patch.object(sys, "argv", arguments), mock.patch.object(
                    framework, "_workflow_result", return_value=result) as execute, redirect_stdout(output):
                if passed:
                    framework.main()
                else:
                    with self.assertRaises(SystemExit) as caught:
                        framework.main()
                    self.assertEqual(caught.exception.code, 1)
            self.assertEqual(json.loads(output.getvalue()), result)
            execute.assert_called_once_with("isolated-png-to-video", "adapters.p4.ffmpeg_workflow",
                                            "unused-workspace", None, "source.png")

    def test_status_reports_aji_v17_uninstalled_policy_and_rust_sqlite_authority(self):
        result = framework.status()

        self.assertEqual(result["communicator"], "aji")
        self.assertEqual(result["active_plan"]["version"], "1.7")
        self.assertFalse(result["entry"]["installed"])
        self.assertFalse(result["entry"]["automatic_discovery_enabled"])
        self.assertFalse(result["entry"]["P7"])
        self.assertEqual(result["authority"], "existing Rust/SQLite workspace only; no second task database")

    def test_doctor_checks_project_pack_and_existing_isolated_rust_cli_without_config_write(self):
        before_content = framework.CONFIG.read_bytes()
        before_mtime = framework.CONFIG.stat().st_mtime_ns

        result = framework.doctor()

        self.assertTrue(result["passed"], result)
        self.assertTrue(result["checks"]["active_baseline"])
        self.assertTrue(result["checks"]["manifest_policy"])
        self.assertTrue(result["checks"]["pack_files"])
        self.assertTrue(result["checks"]["rust_cli"])
        self.assertTrue(framework.run_legion_task.EXECUTABLE.is_file())
        self.assertIn("entry", result["details"]["pack_files"])
        self.assertIn("engineering", result["details"]["pack_files"])
        self.assertEqual(before_content, framework.CONFIG.read_bytes())
        self.assertEqual(before_mtime, framework.CONFIG.stat().st_mtime_ns)
        self.assertFalse(result["P7"])

    def test_route_loads_only_needed_chat_or_professional_context(self):
        chat = framework.run_legion_task.route("chat")
        professional = framework.run_legion_task.route("engineering")

        self.assertTrue(chat["direct_answer"])
        self.assertFalse(professional["direct_answer"])
        self.assertEqual(len(chat["context"]), 1)
        self.assertEqual(len(professional["context"]), 2)
        self.assertEqual(professional["context"][0], chat["context"][0])
        self.assertTrue(professional["context"][1]["path"].endswith("references/engineering.md"))
        self.assertFalse(professional["execution_authority"])

    def test_capabilities_reads_existing_tool_contracts_without_generic_dispatch(self):
        result = framework.capabilities()

        self.assertEqual(result["communicator"], "aji")
        self.assertEqual(result["active_plan"]["version"], "1.7")
        self.assertTrue(result["tools"])
        self.assertIn("officecli", {tool["id"] for tool in result["tools"]})
        self.assertIn("no generic dispatch fallback", result["selection_rule"])
        self.assertFalse(result["P7"])

    def test_unknown_route_is_rejected_without_all_expert_fallback(self):
        with self.assertRaisesRegex(ValueError, "no all-expert fallback"):
            framework.run_legion_task.route("all-experts")

    def test_run_copy_requires_confirmation_and_delegates_existing_workspace(self):
        with tempfile.TemporaryDirectory(dir=framework.ROOT / ".dev" / "core-test-workspaces") as temporary:
            workspace = Path(temporary)
            (workspace / "source.txt").write_text("framework delegation test", encoding="utf-8")
            with self.assertRaisesRegex(ValueError, "explicit scoped confirmation"):
                framework.run_legion_task.run_copy(workspace, "framework-copy", "source.txt", "result.txt", "not-confirmed")
            self.assertFalse((workspace / ".wuji4").exists())

            result = framework.run_legion_task.run_copy(
                workspace, "framework-copy", "source.txt", "result.txt", "confirm-isolated-legion-file-copy"
            )
            self.assertTrue(result["completed"])
            self.assertTrue((workspace / ".wuji4" / "state.sqlite").is_file())
            self.assertEqual((workspace / "source.txt").read_bytes(), (workspace / "result.txt").read_bytes())

            arguments = [
                "wuji4.py",
                "run-copy",
                str(workspace),
                "framework-copy-replay",
                "source.txt",
                "result.txt",
                "confirm-isolated-legion-file-copy",
            ]
            with mock.patch.object(framework.run_legion_task, "run_copy", return_value={"completed": True}) as delegated:
                output = io.StringIO()
                with mock.patch.object(sys, "argv", arguments), redirect_stdout(output):
                    framework.main()

            delegated.assert_called_once_with(
                str(workspace), "framework-copy-replay", "source.txt", "result.txt", "confirm-isolated-legion-file-copy"
            )
            self.assertEqual(json.loads(output.getvalue()), {"completed": True})

    def test_professional_workflow_wrapper_keeps_receipt_and_input_inside_workspace(self):
        with tempfile.TemporaryDirectory(dir=framework.ROOT / ".dev" / "core-test-workspaces") as temporary:
            workspace = Path(temporary)
            input_path = workspace / "input.png"
            input_path.write_bytes(b"declared input")
            module = types.ModuleType("test_professional_workflow")
            calls = []

            def run(*, workspace, output=None, input_path=None):
                calls.append((workspace, output, input_path))
                return {"status": "bounded_real_probe_passed"}

            module.run = run
            with mock.patch.dict(sys.modules, {module.__name__: module}):
                result = framework._workflow_result("test-operation", module.__name__, str(workspace),
                                                    "receipt.json", "input.png")

            self.assertEqual(result["operation"], "test-operation")
            self.assertEqual(result["receipt"]["status"], "bounded_real_probe_passed")
            self.assertEqual(len(calls), 1)
            self.assertEqual(calls[0][0], workspace)
            self.assertEqual(calls[0][1], workspace / "receipt.json")
            self.assertEqual(calls[0][2], input_path)

    def test_professional_workflow_wrapper_rejects_nested_receipt_and_missing_input(self):
        with tempfile.TemporaryDirectory(dir=framework.ROOT / ".dev" / "core-test-workspaces") as temporary:
            workspace = Path(temporary)
            module = types.ModuleType("test_professional_workflow_rejection")
            module.run = lambda **kwargs: {"status": "unexpected"}
            with mock.patch.dict(sys.modules, {module.__name__: module}):
                with self.assertRaisesRegex(ValueError, "receipt"):
                    framework._workflow_result("test-operation", module.__name__, str(workspace), "nested/receipt.json")
                with self.assertRaisesRegex(ValueError, "input"):
                    framework._workflow_result("test-operation", module.__name__, str(workspace), input_relative="missing.png")


if __name__ == "__main__":
    unittest.main()
