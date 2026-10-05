import json
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch

import host_probe as h


class HostProbeTests(unittest.TestCase):
    def test_secrets_urls_and_instructions_not_emitted(self):
        with tempfile.TemporaryDirectory() as folder:
            config = Path(folder) / "config.toml"
            config.write_text('''model = "gpt-6.1-sol"
model_provider = "custom"
model_reasoning_effort = "medium"
developer_instructions = "private content must not be emitted"
[agents]
max_concurrent_threads_per_session = 3
[model_providers.custom]
base_url = "https://private.example/secret-path?secret=true"
credential_fixture = "redacted-test-value"
wire_api = "responses"
''', encoding="utf-8")
            with patch.object(h.shutil, "which", return_value=None):
                value = h.probe(config)
            rendered = json.dumps(value)
            self.assertNotIn("redacted-test-value", rendered)
            self.assertNotIn("private.example", rendered)
            self.assertNotIn("private content must not be emitted", rendered)
            self.assertEqual(value["configured_concurrency_cap"], 3)
            self.assertEqual(value["effective_concurrency_cap"], "unknown")
            self.assertEqual(value["effective_model"], "unknown")

    def test_missing_config_is_unknown_not_auth_failure(self):
        with tempfile.TemporaryDirectory() as folder:
            with patch.object(h.shutil, "which", return_value=None):
                value = h.probe(Path(folder) / "missing.toml")
            self.assertEqual(value["config"]["status"], "missing")
            self.assertFalse(value["third_party_request_verified"])
            self.assertEqual(value["requested_policy"]["model"], "gpt-6.1-sol")


if __name__ == "__main__":
    unittest.main()
