CREATE TABLE knowledge_records (
    id TEXT NOT NULL,
    revision INTEGER NOT NULL CHECK(revision>=1),
    scope TEXT NOT NULL,
    content_hash TEXT NOT NULL,
    envelope_json TEXT NOT NULL,
    state TEXT NOT NULL CHECK(state IN ('candidate','active_local','superseded','retired')),
    PRIMARY KEY(id,revision)
);
CREATE UNIQUE INDEX active_local_knowledge ON knowledge_records(id) WHERE state='active_local';
CREATE TABLE experience_records (
    id TEXT NOT NULL,
    revision INTEGER NOT NULL CHECK(revision>=1),
    scope TEXT NOT NULL,
    content_hash TEXT NOT NULL,
    envelope_json TEXT NOT NULL,
    state TEXT NOT NULL CHECK(state IN ('candidate','active_local','superseded','retired')),
    PRIMARY KEY(id,revision)
);
CREATE UNIQUE INDEX active_local_experience ON experience_records(id) WHERE state='active_local';
CREATE TABLE resource_events (
    event_key TEXT PRIMARY KEY,
    payload_hash TEXT NOT NULL,
    result_json TEXT NOT NULL
);
PRAGMA user_version = 3;
