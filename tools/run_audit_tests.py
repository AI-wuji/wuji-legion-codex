"""Run only the new P0 helper tests and persist their actual result."""

import sys
import unittest

from preflight import ROOT, digest, stamp, write_json


if __name__ == "__main__":
    suite = unittest.defaultTestLoader.discover(str(ROOT / "tools"), pattern="test_*.py")
    result = unittest.TextTestRunner(verbosity=2).run(suite)
    report = {"schema_version": 1, "observed_at": stamp(), "kind": "P0-helper-unit-tests",
              "tests_run": result.testsRun, "passed": result.wasSuccessful(),
              "failures": [str(case) for case, _ in result.failures],
              "errors": [str(case) for case, _ in result.errors],
              "skipped": [str(case) for case, _ in result.skipped],
              "source_hashes": {path.name: digest(path) for path in sorted((ROOT / "tools").glob("*.py"))},
              "runtime_acceptance_scenarios_run": 0, "G0": "not_passed",
              "boundary": "Helper tests do not prove source reading, expert admission or native execution."}
    write_json(ROOT / "outputs/p0/audit-tests.json", report)
    sys.exit(0 if result.wasSuccessful() else 1)
