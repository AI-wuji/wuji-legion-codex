import copy
import unittest

from native_host_session import ObservationTimeout
from native_task_driver import MODEL, RETURN_KEYS, run_task


class FakeSession:
    def __init__(self, events=None, timeout=False):
        self.events = list(events or [])
        self.timeout = timeout
        self.calls = []
        self.event_calls = []
        self.closed = []
        self.started = 0
        self.results = {
            0: {"ok": True},
            1: {
                "thread": {"id": "thread-1", "modelProvider": "fixture"},
                "model": MODEL,
            },
            2: {"turn": {"id": "turn-1", "status": "inProgress", "items": []}},
            3: {"status": "unsubscribed"},
        }

    def start(self):
        self.started += 1

    def request(self, request_id, method, params, timeout=30):
        self.calls.append(("request", request_id, method, copy.deepcopy(params)))
        return copy.deepcopy(self.results[request_id])

    def notify(self, method, params):
        self.calls.append(("notify", method, copy.deepcopy(params)))

    def next_event(self, thread_id, timeout=30):
        self.event_calls.append((thread_id, timeout))
        if self.timeout:
            raise ObservationTimeout("fixture timeout")
        if not self.events:
            raise ObservationTimeout("fixture timeout")
        return copy.deepcopy(self.events.pop(0))

    def close(self, timeout=5):
        self.closed.append(timeout)
        return {"owned_process_exited": True, "exit_code": 0}


def started(turn_id="turn-1"):
    return {
        "method": "turn/started",
        "params": {
            "threadId": "thread-1",
            "turn": {"id": turn_id, "status": "inProgress", "items": []},
        },
    }


def item(text, turn_id="turn-1"):
    return {
        "method": "item/completed",
        "params": {
            "threadId": "thread-1",
            "turnId": turn_id,
            "item": {"type": "agentMessage", "text": text},
        },
    }


def completed(turn_id="turn-1", text=None):
    event = {
        "method": "turn/completed",
        "params": {
            "threadId": "thread-1",
            "turn": {"id": turn_id, "status": "completed", "items": []},
        },
    }
    if text is not None:
        event["params"]["turn"]["items"] = [{"type": "agentMessage", "text": text}]
    return event


def closed():
    return {"method": "thread/closed", "params": {"threadId": "thread-1"}}


class NativeTaskDriverTests(unittest.TestCase):
    def test_success_preserves_bindings_and_returns_only_declared_keys(self):
        session = FakeSession([started(), item("first"), item("second"), completed(), closed()])
        thread_params = {"model": MODEL, "sandbox": {"type": "read-only"}}
        turn_params = {"model": MODEL, "effort": "xhigh", "input": {"nested": ["value"]}}
        original_thread = copy.deepcopy(thread_params)
        original_turn = copy.deepcopy(turn_params)

        result = run_task(session, thread_params, turn_params, timeout=1, close_timeout=2)

        self.assertEqual(RETURN_KEYS, set(result))
        self.assertEqual("thread-1", result["thread_id"])
        self.assertEqual("turn-1", result["turn_id"])
        self.assertEqual("first\nsecond", result["generated_text"])
        self.assertEqual("completed", result["status"])
        self.assertTrue(result["thread_closed_observed"])
        self.assertEqual({"owned_process_exited": True, "exit_code": 0}, result["owned_process_exit"])
        self.assertEqual(original_thread, thread_params)
        self.assertEqual(original_turn, turn_params)
        self.assertEqual(1, session.started)
        self.assertEqual(2, session.closed[0])
        self.assertEqual(5, len(session.event_calls))
        self.assertEqual("thread-1", session.event_calls[0][0])
        calls = session.calls
        self.assertEqual(("request", 0, "initialize", {
            "clientInfo": {"name": "wuji4-native-task-driver", "version": "0.1.0"},
            "capabilities": {"experimentalApi": True},
        }), calls[0])
        self.assertEqual(("notify", "initialized", {}), calls[1])
        self.assertEqual(True, calls[2][3]["ephemeral"])
        self.assertEqual(MODEL, calls[2][3]["model"])
        self.assertEqual("thread-1", calls[3][3]["threadId"])
        self.assertEqual(MODEL, calls[3][3]["model"])
        self.assertEqual("xhigh", calls[3][3]["effort"])
        self.assertEqual(("request", 3, "thread/unsubscribe", {"threadId": "thread-1"}), calls[4])
        self.assertEqual([1, 2, "turn/started", "turn/completed", "thread/closed"],
                         [observation.get("id", observation.get("method")) for observation in result["observations"]])
        self.assertEqual("reconstructed", result["observations"][0]["transport_provenance"]["envelope"])
        self.assertNotIn("first", repr(result["observations"]))

    def test_foreign_turn_notifications_are_not_claimed(self):
        session = FakeSession([
            started("foreign"),
            item("foreign-text", "foreign"),
            started(),
            item("owned-text"),
            completed("foreign"),
            completed(),
            closed(),
        ])

        result = run_task(session, {}, {"effort": "medium"}, timeout=1, close_timeout=1)

        self.assertEqual("owned-text", result["generated_text"])
        methods = [observation["method"] for observation in result["observations"] if "method" in observation]
        self.assertEqual(["turn/started", "turn/completed", "thread/closed"], methods[0:])
        self.assertNotIn("foreign", repr(result["observations"]))

    def test_timeout_is_unknown_without_generation_retry(self):
        session = FakeSession(timeout=True)

        result = run_task(session, {}, {"effort": "high"}, timeout=1, close_timeout=1)

        self.assertEqual("unknown", result["status"])
        self.assertEqual(1, len([call for call in session.calls if call[0] == "request" and call[2] == "turn/start"]))
        self.assertEqual({"owned_process_exited": True, "exit_code": 0}, result["owned_process_exit"])
        self.assertFalse(result["thread_closed_observed"])

    def test_turn_completion_items_are_preserved_and_close_has_own_budget(self):
        session = FakeSession([started(), completed(text="completed-text"), closed()])
        result = run_task(session, {}, {"effort": "high"}, timeout=0.1, close_timeout=2)
        self.assertEqual("completed-text", result["generated_text"])
        self.assertEqual("completed", result["status"])
        self.assertTrue(result["thread_closed_observed"])

    def test_same_agent_message_id_is_not_collected_twice(self):
        event = completed()
        event["params"]["turn"]["items"] = [{"type": "agentMessage", "id": "message-1", "text": "same"}]
        item_event = item("same")
        item_event["params"]["item"]["id"] = "message-1"
        session = FakeSession([started(), item_event, event, closed()])
        result = run_task(session, {}, {"effort": "high"}, timeout=1, close_timeout=1)
        self.assertEqual("same", result["generated_text"])


if __name__ == "__main__":
    unittest.main()
