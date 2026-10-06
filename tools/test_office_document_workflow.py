import json
from pathlib import Path
import tempfile
import unittest
from unittest import mock

from adapters.p4 import office_document_workflow as workflow
from adapters.p4.office_document_workflow import run, validate_request


ROOT = Path(__file__).resolve().parents[1]


class OfficeDocumentWorkflowTests(unittest.TestCase):
    def test_renderer_uses_unique_installed_bundle_not_retired_version(self):
        with tempfile.TemporaryDirectory(dir=ROOT / ".dev/core-test-workspaces") as temporary:
            root = Path(temporary)
            renderer = root / "current/skills/documents/render_docx.py"
            renderer.parent.mkdir(parents=True)
            renderer.write_text("pass", encoding="utf-8")
            with mock.patch.object(workflow, "DOCX_RENDERER_ROOT", root):
                self.assertEqual(workflow._docx_renderer(), renderer)

    def test_absent_and_ambiguous_renderer_do_not_fall_back_or_claim_libreoffice_missing(self):
        with tempfile.TemporaryDirectory(dir=ROOT / ".dev/core-test-workspaces") as temporary:
            root = Path(temporary)
            with mock.patch.object(workflow, "DOCX_RENDERER_ROOT", root):
                with self.assertRaises(workflow.DocxRenderUnavailable) as absent:
                    workflow._docx_renderer()
                self.assertEqual(absent.exception.reason, "renderer_not_found")
                for version in ("first", "second"):
                    renderer = root / version / "skills/documents/render_docx.py"
                    renderer.parent.mkdir(parents=True)
                    renderer.write_text("pass", encoding="utf-8")
                with self.assertRaises(workflow.DocxRenderUnavailable) as ambiguous:
                    workflow._docx_renderer()
                self.assertEqual(ambiguous.exception.reason, "renderer_selection_required")

    def test_missing_libreoffice_preserves_actual_render_command_and_specific_reason(self):
        with tempfile.TemporaryDirectory(dir=ROOT / ".dev/core-test-workspaces") as temporary:
            workspace = Path(temporary)
            receipt = {}
            completed = workflow.subprocess.CompletedProcess([], 1, b"", b"FileNotFoundError: LibreOffice soffice.exe was not found on PATH")
            with mock.patch.object(workflow, "_docx_renderer", return_value=workspace / "renderer.py"), mock.patch.object(
                    workflow.subprocess, "run", return_value=completed):
                with self.assertRaises(workflow.DocxRenderUnavailable) as unavailable:
                    workflow._run_render(workspace / "existing.docx", workspace, receipt)
            self.assertEqual(unavailable.exception.reason, "libreoffice_not_found")
            self.assertEqual(receipt["commands"][0]["exit_code"], 1)
            self.assertIn("soffice.exe", receipt["commands"][0]["stderr"])

    def test_word_and_sheet_requests_preserve_unicode_and_xml_sensitive_text(self):
        word = validate_request({
            "schema_version": 1,
            "kind": "word",
            "title": "\u65e0\u6781\u519b\u56e2 4.0 & <QA>",
            "paragraphs": ["\u963f\u6781\u662f\u552f\u4e00\u6c9f\u901a\u5165\u53e3\u3002"],
        })
        self.assertEqual(word["title"], "\u65e0\u6781\u519b\u56e2 4.0 & <QA>")
        sheet = validate_request({
            "schema_version": 1,
            "kind": "sheet",
            "title": "\u6210\u672c\u8868",
            "rows": [{"label": "\u7814\u7a76 & <A>", "quantity": 2, "unit_price": "0.00"}],
        })
        self.assertEqual(sheet["rows"][0]["label"], "\u7814\u7a76 & <A>")

    def test_request_schema_rejects_undeclared_permissions_and_unsafe_values(self):
        base = {"schema_version": 1, "kind": "sheet", "title": "Cost", "rows": [{"label": "A", "quantity": 1, "unit_price": "1.00"}]}
        invalid = [
            {**base, "formula": "=NOW()"},
            {**base, "rows": [{"label": "A", "quantity": True, "unit_price": "1.00"}]},
            {**base, "rows": [{"label": "A", "quantity": 1, "unit_price": 1.0}]},
            {**base, "rows": [{"label": "A", "quantity": 1, "unit_price": "1"}]},
            {**base, "rows": [{"label": "A\x00", "quantity": 1, "unit_price": "1.00"}]},
            {**base, "rows": [{"label": "A", "quantity": 1, "unit_price": "1000000.01"}]},
        ]
        for request in invalid:
            with self.subTest(request=request):
                with self.assertRaises(ValueError):
                    validate_request(request)

    def test_actual_sheet_creates_formula_bound_workbook_and_rejects_replay_overwrite(self):
        with tempfile.TemporaryDirectory(prefix="office-document-", dir=ROOT / ".dev/core-test-workspaces") as temporary:
            workspace = Path(temporary)
            request = {"schema_version": 1, "kind": "sheet", "title": "Project Cost Summary", "rows": [
                {"label": "Research design & <A>", "quantity": 2, "unit_price": "120.50"},
                {"label": "Isolation test", "quantity": 3, "unit_price": "80.00"},
            ]}
            result = run(workspace, request)
            self.assertTrue(result["passed"], result)
            self.assertEqual(result["build"]["grand_total"], "481.00")
            self.assertTrue(result["checks"]["officecli_observations_completed"])
            self.assertTrue(result["configuration_protection"]["unchanged"])
            replay = run(workspace, request, Path("replay-receipt.json"))
            self.assertFalse(replay["passed"])
            self.assertIn("existing artifact", replay["failure"]["message"])

    def test_actual_word_preserves_structure_and_records_render_block_without_false_success(self):
        with tempfile.TemporaryDirectory(prefix="office-document-word-", dir=ROOT / ".dev/core-test-workspaces") as temporary:
            workspace = Path(temporary)
            request = {"schema_version": 1, "kind": "word", "title": "Wuji Legion 4.0", "paragraphs": ["Aji is the only entry."]}
            with mock.patch.object(workflow, "_run_render", side_effect=workflow.DocxRenderUnavailable(
                    "libreoffice_not_found", "LibreOffice unavailable in isolated failure fixture")):
                result = run(workspace, request)
            self.assertFalse(result["passed"])
            self.assertTrue(result["checks"]["exact_text"])
            self.assertTrue(result["checks"]["officecli_observations_completed"])
            self.assertEqual(result["failure"]["type"], "render_unavailable")
            self.assertEqual(result["render"]["status"], "blocked_external")
            self.assertEqual(result["render"]["reason"], "libreoffice_not_found")
            self.assertTrue(Path(result["artifacts"]["office_file"]["path"]).is_file())
            self.assertTrue(result["configuration_protection"]["unchanged"])


if __name__ == "__main__":
    unittest.main()
