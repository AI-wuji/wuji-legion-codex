use rusqlite::Connection;
use std::fs;
use std::path::{Path, PathBuf};
use std::time::{SystemTime, UNIX_EPOCH};
use wuji4::error::ErrorKind;
use wuji4::store::Store;
use wuji4::strict_json;

fn old_workspace() -> PathBuf {
    let root = Path::new(env!("CARGO_MANIFEST_DIR")).join(".dev/migration-acceptance")
        .join(format!("{}-{}",std::process::id(),SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_nanos()));
    fs::create_dir_all(root.join(".wuji4")).unwrap();
    let database = Connection::open(root.join(".wuji4/state.sqlite")).unwrap();
    database.execute_batch(include_str!("../src/state_schema.sql")).unwrap();
    let canonical = fs::canonicalize(&root).unwrap();
    let root_hash = strict_json::sha256(canonical.to_string_lossy().as_bytes());
    let scope = format!("project:{}",&root_hash[..24]);
    database.execute("INSERT INTO workspace_meta(singleton,scope,root_hash,last_clock_ms) VALUES(1,?1,?2,0)",
        rusqlite::params![scope,root_hash]).unwrap();
    root
}

#[test]
fn schema2_observation_does_not_mutate_and_explicit_init_migrates_atomically() {
    let root = old_workspace();
    let path = root.join(".wuji4/state.sqlite");
    let original = fs::read(&path).unwrap();
    assert_eq!(Store::open_existing(&root).err().unwrap().kind,ErrorKind::MigrationUnsupported);
    assert_eq!(fs::read(&path).unwrap(),original);
    let store = Store::open(&root).unwrap();
    assert_eq!(store.status().unwrap()["schema_version"],6);
    drop(store);
    let database = Connection::open(&path).unwrap();
    assert_eq!(database.pragma_query_value(None,"user_version",|row|row.get::<_,i64>(0)).unwrap(),6);
    let tables: i64 = database.query_row("SELECT count(*) FROM sqlite_master WHERE type='table' AND name IN ('knowledge_records','experience_records','resource_events')",[],|row|row.get(0)).unwrap();
    assert_eq!(tables,3);
}

#[test]
fn failed_known_migration_rolls_back_without_partial_tables_or_version_change() {
    let root = old_workspace();
    let path = root.join(".wuji4/state.sqlite");
    let database = Connection::open(&path).unwrap();
    database.execute("CREATE TABLE experience_records(existing TEXT)",[]).unwrap();
    drop(database);
    assert_eq!(Store::open(&root).err().unwrap().kind,ErrorKind::Storage);
    let database = Connection::open(&path).unwrap();
    assert_eq!(database.pragma_query_value(None,"user_version",|row|row.get::<_,i64>(0)).unwrap(),2);
    let partial: i64 = database.query_row("SELECT count(*) FROM sqlite_master WHERE name IN ('knowledge_records','resource_events')",[],|row|row.get(0)).unwrap();
    assert_eq!(partial,0);
}

#[test]
fn schema3_observation_does_not_migrate_and_transfer_failure_rolls_back() {
    let root = old_workspace();
    let path = root.join(".wuji4/state.sqlite");
    let database = Connection::open(&path).unwrap();
    database.execute_batch(include_str!("../src/resource_schema.sql")).unwrap();
    database.execute("INSERT INTO resource_events VALUES('preserve','hash','{}')",[]).unwrap();
    drop(database);
    let original = fs::read(&path).unwrap();
    assert_eq!(Store::open_existing(&root).err().unwrap().kind,ErrorKind::MigrationUnsupported);
    assert_eq!(fs::read(&path).unwrap(),original);
    let database = Connection::open(&path).unwrap();
    database.execute("CREATE TABLE received_resources(existing TEXT)",[]).unwrap();
    drop(database);
    assert_eq!(Store::open(&root).err().unwrap().kind,ErrorKind::Storage);
    let database = Connection::open(&path).unwrap();
    assert_eq!(database.pragma_query_value(None,"user_version",|row|row.get::<_,i64>(0)).unwrap(),3);
    assert_eq!(database.query_row("SELECT count(*) FROM sqlite_master WHERE name='resource_transfers'",[],|row|row.get::<_,i64>(0)).unwrap(),0);
    database.execute("DROP TABLE received_resources",[]).unwrap();
    drop(database);
    let store = Store::open(&root).unwrap();
    assert_eq!(store.status().unwrap()["schema_version"],6);
    let database = Connection::open(&path).unwrap();
    assert_eq!(database.query_row("SELECT payload_hash FROM resource_events WHERE event_key='preserve'",[],|row|row.get::<_,String>(0)).unwrap(),"hash");
}

#[test]
fn schema4_observation_rejects_and_failed_task_catalog_migration_rolls_back() {
    let root = old_workspace();
    let path = root.join(".wuji4/state.sqlite");
    let database = Connection::open(&path).unwrap();
    database.execute_batch(include_str!("../src/resource_schema.sql")).unwrap();
    database.execute_batch(include_str!("../src/transfer_schema.sql")).unwrap();
    drop(database);
    let original = fs::read(&path).unwrap();
    assert_eq!(Store::open_existing(&root).err().unwrap().kind,ErrorKind::MigrationUnsupported);
    assert_eq!(fs::read(&path).unwrap(),original);
    let database = Connection::open(&path).unwrap();
    database.execute("CREATE TABLE task_catalog_locks(existing TEXT)",[]).unwrap();
    drop(database);
    assert_eq!(Store::open(&root).err().unwrap().kind,ErrorKind::Storage);
    let database = Connection::open(&path).unwrap();
    assert_eq!(database.pragma_query_value(None,"user_version",|row|row.get::<_,i64>(0)).unwrap(),4);
    let mut columns = database.prepare("PRAGMA table_info(tasks)").unwrap();
    assert!(!columns.query_map([],|row|row.get::<_,String>(1)).unwrap().any(|name|name.unwrap()=="catalog_binding_hash"));
    drop(columns);
    database.execute("DROP TABLE task_catalog_locks",[]).unwrap();
    drop(database);
    assert_eq!(Store::open(&root).unwrap().status().unwrap()["schema_version"],6);
}

#[test]
fn schema5_observation_does_not_migrate_and_received_version_failure_rolls_back() {
    let root=old_workspace();
    let path=root.join(".wuji4/state.sqlite");
    let database=Connection::open(&path).unwrap();
    database.execute_batch(include_str!("../src/resource_schema.sql")).unwrap();
    database.execute_batch(include_str!("../src/transfer_schema.sql")).unwrap();
    database.execute_batch(include_str!("../src/task_catalog_schema.sql")).unwrap();
    database.execute("INSERT INTO resource_events VALUES('preserve-received-upgrade','prior','{}')",[]).unwrap();
    drop(database);
    let original=fs::read(&path).unwrap();
    assert_eq!(Store::open_existing(&root).err().unwrap().kind,ErrorKind::MigrationUnsupported);
    assert_eq!(fs::read(&path).unwrap(),original);
    let database=Connection::open(&path).unwrap();
    database.execute("CREATE TABLE received_resource_versions(existing TEXT)",[]).unwrap();
    drop(database);
    assert_eq!(Store::open(&root).err().unwrap().kind,ErrorKind::Storage);
    let database=Connection::open(&path).unwrap();
    assert_eq!(database.pragma_query_value(None,"user_version",|row|row.get::<_,i64>(0)).unwrap(),5);
    assert_eq!(database.query_row("SELECT payload_hash FROM resource_events WHERE event_key='preserve-received-upgrade'",[],|row|row.get::<_,String>(0)).unwrap(),"prior");
    database.execute("DROP TABLE received_resource_versions",[]).unwrap();
    drop(database);
    assert_eq!(Store::open(&root).unwrap().status().unwrap()["schema_version"],6);
}
