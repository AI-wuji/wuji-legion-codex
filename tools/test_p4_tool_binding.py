import copy
import hashlib
import json
from pathlib import Path
import tempfile
import unittest
from unittest import mock

from adapters.p4 import ffmpeg_workflow as video


ROOT = Path(__file__).resolve().parents[1]


def digest(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def load_json(relative: str) -> dict:
    return json.loads((ROOT / relative).read_text(encoding="utf-8"))


class P4ToolBindingTests(unittest.TestCase):
    def test_video_binding_includes_exact_version_schema_and_identity(self):
        binding = video.reviewed_binding()
        self.assertEqual(binding["tool_id"], "video-render")
        self.assertEqual(binding["version"], "ffmpeg-7.1-observed")
        self.assertEqual(binding["identity"]["server"], "local-process")
        self.assertEqual(binding["identity"]["tool"], "video-render/png-to-mp4")
        self.assertEqual(binding["interface"], video.VIDEO_INTERFACE)

    def test_video_binding_rejects_same_name_alias_and_directory_fallback(self):
        original = json.loads(video.BINDING_MANIFEST.read_text(encoding="utf-8"))
        for mutation in ("same-name", "missing-target"):
            with self.subTest(mutation=mutation), tempfile.TemporaryDirectory(
                    dir=ROOT / ".dev" / "core-test-workspaces") as temporary:
                altered = Path(temporary) / "tool-manifest.json"
                manifest = copy.deepcopy(original)
                if mutation == "same-name":
                    alias = copy.deepcopy(next(tool for tool in manifest["tools"] if tool["id"] == "video-render"))
                    alias["id"] = "video-render-alias"
                    manifest["tools"].append(alias)
                else:
                    manifest["tools"] = [tool for tool in manifest["tools"] if tool["id"] != "video-render"]
                altered.write_text(json.dumps(manifest), encoding="utf-8")
                with mock.patch.object(video, "BINDING_MANIFEST", altered):
                    with self.assertRaises(ValueError):
                        video.reviewed_binding()

    def test_video_dispatch_rejects_binding_drift_before_process_start(self):
        with tempfile.TemporaryDirectory(dir=ROOT / ".dev" / "core-test-workspaces") as temporary:
            workspace = Path(temporary)
            source = workspace / "source.png"
            source.write_bytes((ROOT / "outputs/p4/comfyui-probe-output.png").read_bytes())
            binding = workspace / "tool-manifest.json"
            binding.write_bytes(video.BINDING_MANIFEST.read_bytes())
            original_append = video.media.append_process_command
            process_calls = []

            def mutate_before_dispatch(receipt, phase, argv, cwd, timeout, parse_json=False, before_dispatch=None):
                if phase == "ffmpeg-version":
                    binding.write_text(binding.read_text(encoding="utf-8") + "\n", encoding="utf-8")
                return original_append(receipt, phase, argv, cwd, timeout, parse_json, before_dispatch)

            with mock.patch.object(video, "BINDING_MANIFEST", binding), mock.patch.object(
                    video.media, "append_process_command", side_effect=mutate_before_dispatch), mock.patch.object(
                    video.media.subprocess, "run", side_effect=lambda *args, **kwargs: process_calls.append(args) or None):
                result = video.run(workspace, input_path=source)

            self.assertFalse(result["passed"])
            self.assertIn("drifted before dispatch", result["failure"]["message"])
            self.assertEqual(process_calls, [])
            self.assertEqual(result["commands"][0]["phase"], "ffmpeg-version")
            self.assertEqual(result["commands"][0]["dispatch_guard"], "rejected")
            self.assertFalse(result["commands"][0]["owned_process_exit_observed"])

    def test_framework_video_evidence_binds_current_sources_artifact_and_real_decode(self):
        evidence = load_json("outputs/p4/video-framework-workflow-evidence.json")
        self.assertTrue(evidence["passed"])
        self.assertEqual(evidence["result_state"], "completed_bounded_task")
        self.assertEqual(evidence["framework_execution"]["exit_code"], 0)
        self.assertEqual(evidence["framework_execution"]["source_sha256"], digest(ROOT / "tools/wuji4.py"))
        for relative, expected in evidence["implementation_hashes"].items():
            self.assertEqual(expected, digest(ROOT / relative))
        artifact = evidence["retained_artifact"]
        self.assertEqual(artifact["sha256"], digest(ROOT / artifact["path"]))
        self.assertEqual(artifact["sha256"], evidence["artifacts"]["mp4"]["sha256"])
        self.assertTrue(all(evidence["checks"].values()))
        binding = evidence["reviewed_binding"]
        self.assertEqual(binding["manifest_sha256"], digest(ROOT / "catalog/p4/tool-manifests.json"))
        self.assertEqual(binding["identity"]["server"], "local-process")
        self.assertEqual(binding["identity"]["tool"], "video-render/png-to-mp4")
        self.assertTrue(evidence["checks"]["reviewed_binding_current"])
        decode = next(command for command in evidence["commands"] if command["phase"] == "ffmpeg-independent-decode")
        self.assertEqual(decode["exit_code"], 0)
        self.assertTrue(decode["owned_process_exit_observed"])
        self.assertEqual(evidence["validation"]["progress"]["frame"], "8")
        self.assertEqual(evidence["validation"]["progress"]["out_time_us"], "1000000")
        self.assertFalse(evidence["comfyui_submitted"])
        self.assertFalse(evidence["audio_device_opened"])
        self.assertFalse(evidence["P7"])

    def test_comfyui_historical_probe_is_not_reported_as_current_health(self):
        tool = next(tool for tool in load_json("catalog/p4/tool-manifests.json")["tools"] if tool["id"] == "comfyui")
        self.assertEqual(tool["status"], "blocked")
        self.assertFalse(tool["runtime_admission"])
        self.assertEqual(tool["current_health_evidence"], "outputs/p4/media-framework-failure-evidence.json")
        self.assertNotEqual(tool["bounded_evidence"], tool["current_health_evidence"])

    def test_office_document_binding_has_bounded_word_and_sheet_contract(self):
        manifest = load_json("catalog/p4/tool-manifests.json")
        tool = next(tool for tool in manifest["tools"] if tool["id"] == "officecli-documents")
        self.assertEqual(tool["status"], "available-local-bounded")
        self.assertTrue(tool["runtime_admission"])
        self.assertEqual(tool["adapter"], "adapters/p4/office_document_workflow.py")
        self.assertIn("word", tool["action_contract"]["input"])
        self.assertIn("sheet", tool["action_contract"]["input"])
        self.assertFalse(tool["action_contract"]["permissions"].find("global") == -1)

    def test_office_document_evidence_preserves_xlsx_success_and_word_render_block(self):
        evidence = load_json("outputs/p4/office-document-workflow-evidence.json")
        self.assertFalse(evidence["passed"])
        self.assertEqual(evidence["status"], "partial_evidence_retained")
        self.assertTrue(evidence["formats"]["sheet"]["passed"])
        self.assertEqual(evidence["formats"]["sheet"]["build"]["grand_total"], "481.00")
        self.assertFalse(evidence["formats"]["word"]["passed"])
        self.assertEqual(evidence["formats"]["word"]["failure"]["type"], "render_unavailable")
        self.assertTrue(evidence["summary"]["word_structure_and_officecli_readback"])
        self.assertFalse(evidence["summary"]["word_required_render"])
        self.assertTrue(evidence["summary"]["word_render_blocked_external"])
        self.assertTrue(evidence["summary"]["configuration_unchanged"])
        for relative, expected in evidence["implementation_hashes"].items():
            self.assertEqual(expected, digest(ROOT / relative))
        for result in evidence["formats"].values():
            for artifact in (result["receipt"], result["framework_output"], result["artifact"]):
                path = ROOT / artifact["path"]
                self.assertTrue(path.is_file())
                self.assertEqual(artifact["sha256"], digest(path))

    def test_manifest_binds_current_tool_versions_contracts_and_evidence(self):
        manifest = load_json("catalog/p4/tool-manifests.json")
        tools = manifest["tools"]
        self.assertEqual(len({tool["id"] for tool in tools}), len(tools))
        by_id = {tool["id"]: tool for tool in tools}
        for tool in tools:
            self.assertTrue(tool["owner"])
            self.assertIn(tool["status"], {"available-local", "available-local-bounded", "blocked"})
            if tool["status"] == "blocked":
                self.assertTrue(tool.get("reason"))
            else:
                self.assertTrue(tool["version"])
                self.assertIn("action_contract", tool)
                self.assertIsInstance(tool["action_contract"], dict)
                self.assertIn("failure", tool["action_contract"])
            self.assertIsInstance(tool["runtime_admission"], bool)

        comfy = load_json("outputs/p4/comfyui-probe-evidence.json")
        self.assertTrue(by_id["comfyui"]["version"].startswith(comfy["version_observed"]))
        self.assertEqual(by_id["comfyui"]["endpoint"], comfy["endpoint"])
        self.assertEqual([node["class_type"] for node in comfy["workflow"]],
                         ["LoadImage", "ImageScale", "SaveImage"])
        comfy_artifact = ROOT / comfy["artifact"]["path"]
        self.assertEqual(comfy["artifact"]["sha256"], digest(comfy_artifact))

        video = load_json("outputs/p4/ffmpeg-probe-evidence.json")
        ffmpeg = Path(video["binary_path"])
        self.assertIn("7.1", by_id["video-render"]["version"])
        self.assertTrue(ffmpeg.is_file())
        self.assertEqual(video["version_observed"], "ffmpeg version 7.1")
        video_artifact = ROOT / video["artifact"]["path"]
        self.assertEqual(video["artifact"]["sha256"], digest(video_artifact))
        self.assertEqual(video["validation"]["decode_to_null_exit_code"], 0)

        office = load_json("outputs/p4/officecli-probe-evidence.json")
        office_binary = Path(office["binary_path"])
        self.assertEqual(by_id["officecli"]["version"], "1.0.152-exact-original-release-observed")
        self.assertTrue(office["passed"])
        self.assertEqual(office["binary_sha256"], digest(office_binary))
        self.assertEqual(office["adapter_sha256"], digest(ROOT / "adapters/p4/officecli_adapter.py"))
        office_artifact = ROOT / office["artifact"]["path"]
        self.assertEqual(office["artifact"]["sha256"], digest(office_artifact))

        framework_office = load_json("outputs/p4/officecli-framework-workflow-evidence.json")
        self.assertEqual(framework_office["state"], "completed_bounded_task")
        self.assertTrue(framework_office["passed"])
        self.assertEqual(framework_office["artifact"]["sha256"], digest(Path(framework_office["artifact"]["path"])))
        self.assertEqual(framework_office["protected_configurations_before"],
                         framework_office["protected_configurations_after"])

    def test_changed_binding_evidence_is_not_accepted_as_current(self):
        manifest = load_json("catalog/p4/tool-manifests.json")
        evidence = load_json("outputs/p4/officecli-probe-evidence.json")
        office_tool = next(tool for tool in manifest["tools"] if tool["id"] == "officecli")
        original_version = office_tool["version"]
        self.assertTrue(original_version.startswith(evidence["version"]))
        office_tool["version"] = "unverified-version"
        self.assertFalse(office_tool["version"].startswith(evidence["version"]))
        office_tool["version"] = original_version
        original_adapter_hash = evidence["adapter_sha256"]
        evidence["adapter_sha256"] = "tampered"
        self.assertNotEqual(evidence["adapter_sha256"], digest(ROOT / "adapters/p4/officecli_adapter.py"))
        evidence["adapter_sha256"] = original_adapter_hash

        media_failure = load_json("outputs/p4/media-framework-failure-evidence.json")
        self.assertEqual(media_failure["status"], "failed_real_error")
        self.assertEqual(media_failure["failure"]["code"], "comfyui_unavailable")
        self.assertTrue(media_failure["configuration_protection"]["unchanged"])
        self.assertFalse(media_failure["P7"])


if __name__ == "__main__":
    unittest.main()
