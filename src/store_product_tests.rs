#[test]
fn registered_input_is_exact_scoped_and_read_bounded() {
    let root = workspace("input-register");
    fs::write(root.join("source.txt"),b"source").unwrap();
    let mut store = Store::open(&root).unwrap();
    let reference = store.register_local_input("user/source",Path::new("source.txt")).unwrap();
    assert_eq!(store.register_local_input("user/source",Path::new("source.txt")).unwrap(),reference);
    assert_eq!(count(&store,"input_files"),1);
    let mut blueprint = plan(&store,"task",&[("write","result.txt")],&[],1);
    blueprint["payload"]["nodes"][0]["inputs"] = json!([reference]);
    rehash(&mut blueprint);
    assert_eq!(store.plan(&blueprint).unwrap_err().kind,ErrorKind::PathDenied);
    blueprint["payload"]["nodes"][0]["read_roots"] = json!(["source.txt"]);
    rehash(&mut blueprint);
    store.plan(&blueprint).unwrap();
    let claim = store.claim_local("task","write").unwrap();
    store.write_local_file(&claim,Path::new("result.txt"),b"real derived output").unwrap();
    store.verify_local_file(&claim).unwrap();
    store.close_local_handler(&claim).unwrap();
    assert_eq!(store.connection.query_row("SELECT status FROM tasks",[],|row| row.get::<_,String>(0)).unwrap(),"succeeded");
}

#[test]
fn input_change_persistently_invalidates_full_downstream_without_release() {
    let root = workspace("input-invalidate");
    fs::write(root.join("source.txt"),b"original").unwrap();
    let mut store = Store::open(&root).unwrap();
    let reference = store.register_local_input("user/source",Path::new("source.txt")).unwrap();
    let mut blueprint = plan(&store,"task",&[("first","first.txt"),("second","second.txt"),("third","third.txt")],&[("first","second"),("second","third")],2);
    blueprint["payload"]["nodes"][0]["inputs"] = json!([reference]);
    blueprint["payload"]["nodes"][0]["read_roots"] = json!(["source.txt"]);
    rehash(&mut blueprint);
    store.plan(&blueprint).unwrap();
    let first = store.claim_local("task","first").unwrap();
    store.write_local_file(&first,Path::new("first.txt"),b"first").unwrap();
    store.verify_local_file(&first).unwrap();
    let second = store.claim_local("task","second").unwrap();
    fs::write(root.join("source.txt"),b"changed").unwrap();
    let result = store.refresh_local_evidence().unwrap();
    assert_eq!(result["affected_nodes"],3);
    assert_eq!(count(&store,"acceptance_links"),0);
    assert_eq!(store.status().unwrap()["open_slots"],2);
    assert_eq!(store.write_local_file(&second,Path::new("second.txt"),b"late").unwrap_err().kind,ErrorKind::StaleReceipt);
    drop(store);
    let mut reopened = Store::open_existing(&root).unwrap();
    assert_eq!(reopened.connection.query_row("SELECT count(*) FROM nodes WHERE state='blocked'",[],|row| row.get::<_,i64>(0)).unwrap(),3);
    let next = reopened.register_local_input("user/source",Path::new("source.txt")).unwrap();
    assert_eq!(next.revision,2);
    assert_eq!(reopened.claim_local("task","third").unwrap_err().kind,ErrorKind::ValidationStale);
    assert!(root.join("first.txt").exists());
}

fn revision(blueprint: &Value, changed: &[(&str,&str)], affected: &[&str], next: i64) -> Value {
    let mut revised = blueprint.clone();
    revised["metadata"]["revision"] = json!(next);
    revised["payload"]["version"] = json!(next.to_string());
    for node in revised["payload"]["nodes"].as_array_mut().unwrap() {
        let id = node["id"].as_str().unwrap().to_owned();
        if affected.contains(&id.as_str()) { node["revision"] = json!(node["revision"].as_u64().unwrap()+1); }
        if let Some((_,path)) = changed.iter().find(|(name,_)| *name == id) { node["write_roots"] = json!([path]); }
    }
    rehash(&mut revised);
    revised
}

#[test]
fn graph_cas_retains_unrelated_adoption_and_rejects_old_result() {
    let root = workspace("graph-cas");
    let mut store = Store::open(&root).unwrap();
    let blueprint = plan(&store,"task",&[("first","first.txt"),("second","second.txt"),("independent","other.txt")],&[("first","second")],2);
    store.plan(&blueprint).unwrap();
    let other = store.claim_local("task","independent").unwrap();
    store.write_local_file(&other,Path::new("other.txt"),b"preserved").unwrap();
    store.verify_local_file(&other).unwrap();
    store.close_local_handler(&other).unwrap();
    let first = store.claim_local("task","first").unwrap();
    let revised = revision(&blueprint,&[("first","first-v2.txt")],&["first","second"],2);
    let result = store.revise_local_plan(&revised,1,"user-change-1").unwrap();
    assert_eq!(result["affected_nodes"],json!(["first","second"]));
    assert_eq!(store.revise_local_plan(&revised,1,"user-change-1").unwrap(),result);
    assert_eq!(store.revise_local_plan(&revised,1,"different-change").unwrap_err().kind,ErrorKind::RevisionConflict);
    assert_eq!(count(&store,"task_plans"),2);
    assert_eq!(count(&store,"acceptance_links"),3);
    assert_eq!(store.status().unwrap()["open_slots"],1);
    assert_eq!(store.write_local_file(&first,Path::new("first-v2.txt"),b"late").unwrap_err().kind,ErrorKind::StaleReceipt);
    store.close_local_handler(&first).unwrap();
    let fresh = store.claim_local("task","first").unwrap();
    store.write_local_file(&fresh,Path::new("first-v2.txt"),b"new").unwrap();
    store.verify_local_file(&fresh).unwrap();
    store.close_local_handler(&fresh).unwrap();
    assert_eq!(fs::read(root.join("other.txt")).unwrap(),b"preserved");
}

#[test]
fn graph_revisions_cannot_reset_budget_or_repeat_no_progress() {
    let root = workspace("revision-budget");
    let mut store = Store::open(&root).unwrap();
    let blueprint = plan(&store,"task",&[("write","one.txt")],&[],1);
    store.plan(&blueprint).unwrap();
    let mut reset = revision(&blueprint,&[("write","two.txt")],&["write"],2);
    reset["payload"]["budget"]["max_extra_retries"] = json!(999);
    rehash(&mut reset);
    assert_eq!(store.revise_local_plan(&reset,1,"reset").unwrap_err().kind,ErrorKind::BudgetExhausted);
    let unchanged = revision(&blueprint,&[],&[],2);
    assert_eq!(store.revise_local_plan(&unchanged,1,"no-progress").unwrap_err().kind,ErrorKind::RevisionConflict);
    let second = revision(&blueprint,&[("write","two.txt")],&["write"],2);
    store.revise_local_plan(&second,1,"two").unwrap();
    let third = revision(&second,&[("write","three.txt")],&["write"],3);
    store.revise_local_plan(&third,2,"three").unwrap();
    let fourth = revision(&third,&[("write","four.txt")],&["write"],4);
    assert_eq!(store.revise_local_plan(&fourth,3,"four").unwrap_err().kind,ErrorKind::BudgetExhausted);
    assert_eq!(count(&store,"task_plans"),3);
}

#[test]
fn old_attempt_keeps_immutable_path_reservation_after_revision() {
    let root = workspace("immutable-lock");
    let mut store = Store::open(&root).unwrap();
    let blueprint = plan(&store,"task",&[("write","old.txt")],&[],2);
    store.plan(&blueprint).unwrap();
    let claim = store.claim_local("task","write").unwrap();
    let revised = revision(&blueprint,&[("write","new.txt")],&["write"],2);
    store.revise_local_plan(&revised,1,"move-target").unwrap();
    let another = plan(&store,"other-task",&[("write","old.txt")],&[],2);
    store.plan(&another).unwrap();
    assert_eq!(store.claim_local("other-task","write").unwrap_err().kind,ErrorKind::OwnerConflict);
    store.close_local_handler(&claim).unwrap();
    store.claim_local("other-task","write").unwrap();
}

#[test]
fn checkpoint_is_durable_but_cannot_resolve_unknown_or_extend_lease() {
    let root = workspace("checkpoint");
    let mut store = Store::open(&root).unwrap();
    let blueprint = plan(&store,"task",&[("write","result.txt")],&[],1);
    store.plan(&blueprint).unwrap();
    let checkpoint = store.checkpoint_local("task").unwrap();
    let id = checkpoint["checkpoint_id"].as_str().unwrap();
    assert_eq!(store.inspect_checkpoint(id).unwrap()["reusable_as_context"],true);
    let claim = store.claim_local("task","write").unwrap();
    assert_eq!(store.inspect_checkpoint(id).unwrap()["reusable_as_context"],false);
    store.begin_local_dispatch(&claim,&store.workspace.output(Path::new("result.txt")).unwrap(),b"possible").unwrap();
    fs::write(root.join("result.txt"),b"possible").unwrap();
    store.recover().unwrap();
    let query = store.inspect_local_attempt(claim.attempt_id()).unwrap();
    assert_eq!(query["file_observation"]["matches_intent"],true);
    assert_eq!(query["file_observation"]["ownership_proven_by_content"],false);
    assert_eq!(query["invocation_state"],"unknown");
    assert_eq!(store.status().unwrap()["open_slots"],1);
    assert_eq!(count(&store,"artifacts"),0);
    assert_eq!(store.close_local_handler(&claim).unwrap_err().kind,ErrorKind::UnknownSubmission);
    fs::write(Path::new(checkpoint["path"].as_str().unwrap()),b"tampered").unwrap();
    assert_eq!(store.inspect_checkpoint(id).unwrap_err().kind,ErrorKind::HashMismatch);
}

#[test]
fn prior_schema_id_is_preserved_without_silent_migration() {
    let root = workspace("old-schema");
    let store = Store::open(&root).unwrap();
    store.connection.pragma_update(None,"user_version",1).unwrap();
    drop(store);
    let database = root.join(".wuji4/state.sqlite");
    let original = fs::read(&database).unwrap();
    assert_eq!(Store::open_existing(&root).err().unwrap().kind,ErrorKind::MigrationUnsupported);
    assert_eq!(fs::read(database).unwrap(),original);
}

#[test]
fn invalidated_unknown_attempt_cannot_be_revived_by_recovery() {
    let root = workspace("unknown-no-revival");
    fs::write(root.join("source.txt"),b"original").unwrap();
    let mut store = Store::open(&root).unwrap();
    let reference = store.register_local_input("source",Path::new("source.txt")).unwrap();
    let mut blueprint = plan(&store,"task",&[("write","result.txt")],&[],1);
    blueprint["payload"]["nodes"][0]["inputs"] = json!([reference]);
    blueprint["payload"]["nodes"][0]["read_roots"] = json!(["source.txt"]);
    rehash(&mut blueprint);
    store.plan(&blueprint).unwrap();
    let claim = store.claim_local("task","write").unwrap();
    store.begin_local_dispatch(&claim,&store.workspace.output(Path::new("result.txt")).unwrap(),b"possible").unwrap();
    fs::write(root.join("source.txt"),b"changed").unwrap();
    store.recover().unwrap();
    assert_eq!(store.inspect_local_attempt(claim.attempt_id()).unwrap()["attempt_state"],"superseded");
    assert_eq!(store.status().unwrap()["unknown_invocations"],1);
    assert_eq!(store.status().unwrap()["open_slots"],1);
    assert_eq!(store.write_local_file(&claim,Path::new("result.txt"),b"late").unwrap_err().kind,ErrorKind::StaleReceipt);
}

#[test]
fn closing_last_observed_local_slot_finalizes_requested_cancellation() {
    let root = workspace("cancel-finalize");
    let mut store = Store::open(&root).unwrap();
    store.plan(&plan(&store,"task",&[("write","result.txt")],&[],1)).unwrap();
    let claim = store.claim_local("task","write").unwrap();
    store.write_local_file(&claim,Path::new("result.txt"),b"kept").unwrap();
    store.connection.execute("UPDATE tasks SET status='cancel_requested'",[]).unwrap();
    store.connection.execute("UPDATE nodes SET state='cancel_requested'",[]).unwrap();
    store.close_local_handler(&claim).unwrap();
    assert_eq!(store.connection.query_row("SELECT status FROM tasks",[],|row| row.get::<_,String>(0)).unwrap(),"cancelled");
    assert_eq!(store.status().unwrap()["open_slots"],0);
    assert_eq!(fs::read(root.join("result.txt")).unwrap(),b"kept");
}

#[test]
fn declared_input_copy_requires_exact_source_and_has_separate_validation() {
    let root = workspace("input-copy");
    fs::write(root.join("source.txt"),"真实输入".as_bytes()).unwrap();
    fs::write(root.join("unassigned.txt"),b"not adopted").unwrap();
    let mut store = Store::open(&root).unwrap();
    let reference = store.register_local_input("source",Path::new("source.txt")).unwrap();
    let mut blueprint = plan(&store,"task",&[("write","result.txt")],&[],1);
    blueprint["payload"]["nodes"][0]["inputs"] = json!([reference]);
    blueprint["payload"]["nodes"][0]["read_roots"] = json!(["source.txt"]);
    rehash(&mut blueprint);
    store.plan(&blueprint).unwrap();
    assert_eq!(store.copy_local_input("task","write",Path::new("unassigned.txt"),Path::new("result.txt")).unwrap_err().kind,ErrorKind::Reference);
    assert_eq!(count(&store,"attempts"),0);
    let result = store.copy_local_input("task","write",Path::new("source.txt"),Path::new("result.txt")).unwrap();
    assert_eq!(result["professional_or_native_verified"],false);
    assert_eq!(result["status"]["open_slots"],0);
    assert_eq!(fs::read(root.join("result.txt")).unwrap(),"真实输入".as_bytes());
    assert_eq!(store.copy_local_input("task","write",Path::new("source.txt"),Path::new("result.txt")).unwrap(),result);
    assert_eq!(count(&store,"attempts"),1);
    assert_eq!(count(&store,"validations"),1);
}

#[test]
fn repeated_write_replays_without_redispatch_and_changed_payload_conflicts() {
    let root = workspace("write-replay");
    let mut store = Store::open(&root).unwrap();
    store.plan(&plan(&store,"task",&[("write","result.txt")],&[],1)).unwrap();
    let claim = store.claim_local("task","write").unwrap();
    let produced = store.write_local_file(&claim,Path::new("result.txt"),b"same").unwrap();
    assert_eq!(store.write_local_file(&claim,Path::new("result.txt"),b"same").unwrap(),produced);
    assert_eq!(store.write_local_file(&claim,Path::new("result.txt"),b"different").unwrap_err().kind,ErrorKind::EventConflict);
    store.verify_local_file(&claim).unwrap();
    store.close_local_handler(&claim).unwrap();
    assert_eq!(store.write_local_file(&claim,Path::new("result.txt"),b"same").unwrap(),produced);
    assert_eq!(count(&store,"artifacts"),1);
    assert_eq!(count(&store,"attempts"),1);
    assert_eq!(fs::read(root.join("result.txt")).unwrap(),b"same");
}

struct OwnedProcess(std::process::Child);

impl Drop for OwnedProcess {
    fn drop(&mut self) {
        if self.0.try_wait().is_ok_and(|status| status.is_none()) {
            let _ = self.0.kill();
            let _ = self.0.wait();
        }
    }
}

fn fixture(root: &Path, mode: &str) -> OwnedProcess {
    OwnedProcess(std::process::Command::new(std::env::current_exe().unwrap())
        .args(["--exact","store::tests::process_fixture","--nocapture"])
        .env("WUJI4_PRODUCT_FIXTURE_ROOT",root).env("WUJI4_PRODUCT_FIXTURE_MODE",mode)
        .stdout(std::process::Stdio::piped()).stderr(std::process::Stdio::piped()).spawn().unwrap())
}

fn await_file(path: &Path) {
    let deadline = std::time::Instant::now()+Duration::from_secs(15);
    while !path.exists() {
        assert!(std::time::Instant::now()<deadline,"fixture handshake timeout: {}",path.display());
        std::thread::sleep(Duration::from_millis(5));
    }
}

#[test]
fn process_fixture() {
    let Ok(root) = std::env::var("WUJI4_PRODUCT_FIXTURE_ROOT") else { return; };
    let root = PathBuf::from(root);
    let expected = Path::new(env!("CARGO_MANIFEST_DIR")).join(".dev/core-test-workspaces");
    assert!(root.starts_with(expected));
    let mode = std::env::var("WUJI4_PRODUCT_FIXTURE_MODE").unwrap();
    let mut store = Store::open_existing(&root).unwrap();
    if mode.starts_with("write-") {
        let node = mode.strip_prefix("write-").unwrap();
        let claim = store.claim_local("task",node).unwrap();
        fs::write(root.join(format!("{mode}.ready")),b"reserved-local-slot").unwrap();
        await_file(&root.join("release-write"));
        store.write_local_file(&claim,Path::new(&format!("{node}.txt")),node.as_bytes()).unwrap();
        store.verify_local_file(&claim).unwrap();
        store.close_local_handler(&claim).unwrap();
        return;
    }
    if mode.starts_with("claim-") {
        fs::write(root.join(format!("{mode}.ready")),b"ready").unwrap();
        await_file(&root.join("release-claim"));
        match store.claim_local("task","write") {
            Ok(_) => println!("CLAIM_OK"),
            Err(error) => { assert_eq!(error.kind,ErrorKind::RevisionConflict); println!("CLAIM_REJECTED"); }
        }
        return;
    }
    if mode == "uncommitted" {
        let transaction = store.connection.transaction_with_behavior(TransactionBehavior::Immediate).unwrap();
        transaction.execute("UPDATE tasks SET graph_revision=99,status='blocked'",[]).unwrap();
        fs::write(root.join("crash.ready"),b"transaction-uncommitted").unwrap();
        await_file(&root.join("never-release-transaction"));
        transaction.commit().unwrap();
        return;
    }
    let claim = store.claim_local("task","write").unwrap();
    let target = store.workspace.output(Path::new("result.txt")).unwrap();
    store.begin_local_dispatch(&claim,&target,b"process-owned-output").unwrap();
    if mode == "after-file-sync" {
        let mut file = OpenOptions::new().write(true).create_new(true).open(target).unwrap();
        file.write_all(b"process-owned-output").unwrap();
        file.sync_all().unwrap();
        drop(file);
    }
    fs::write(root.join("crash.ready"),b"dispatch-committed").unwrap();
    await_file(&root.join("never-release-handler"));
}

#[test]
fn two_os_processes_have_one_claim_winner() {
    let root = workspace("process-claim");
    let mut store = Store::open(&root).unwrap();
    store.plan(&plan(&store,"task",&[("write","result.txt")],&[],2)).unwrap();
    drop(store);
    let mut first = fixture(&root,"claim-one");
    let mut second = fixture(&root,"claim-two");
    await_file(&root.join("claim-one.ready"));
    await_file(&root.join("claim-two.ready"));
    fs::write(root.join("release-claim"),b"both-may-start").unwrap();
    assert!(first.0.wait().unwrap().success());
    assert!(second.0.wait().unwrap().success());
    let reopened = Store::open_existing(&root).unwrap();
    assert_eq!(count(&reopened,"attempts"),1);
    assert_eq!(reopened.status().unwrap()["open_slots"],1);
}

#[test]
fn independent_file_handlers_run_in_two_os_processes() {
    let root = workspace("process-two-files");
    let mut store = Store::open(&root).unwrap();
    store.plan(&plan(&store,"task",&[("first","first.txt"),("second","second.txt")],&[],2)).unwrap();
    drop(store);
    let mut first = fixture(&root,"write-first");
    let mut second = fixture(&root,"write-second");
    await_file(&root.join("write-first.ready"));
    await_file(&root.join("write-second.ready"));
    assert_eq!(Store::open_existing(&root).unwrap().status().unwrap()["open_slots"],2);
    fs::write(root.join("release-write"),b"both-local-handlers-may-write").unwrap();
    assert!(first.0.wait().unwrap().success());
    assert!(second.0.wait().unwrap().success());
    let reopened = Store::open_existing(&root).unwrap();
    assert_eq!(fs::read(root.join("first.txt")).unwrap(),b"first");
    assert_eq!(fs::read(root.join("second.txt")).unwrap(),b"second");
    assert_eq!(count(&reopened,"validations"),2);
    assert_eq!(count(&reopened,"acceptance_links"),6);
    assert_eq!(reopened.status().unwrap()["open_slots"],0);
    assert_eq!(reopened.connection.query_row("SELECT status FROM tasks",[],|row| row.get::<_,String>(0)).unwrap(),"succeeded");
}

fn killed_dispatch(mode: &str) {
    let root = workspace(mode);
    let mut store = Store::open(&root).unwrap();
    store.plan(&plan(&store,"task",&[("write","result.txt")],&[],1)).unwrap();
    drop(store);
    let mut child = fixture(&root,mode);
    await_file(&root.join("crash.ready"));
    child.0.kill().unwrap();
    child.0.wait().unwrap();
    let mut reopened = Store::open_existing(&root).unwrap();
    reopened.recover().unwrap();
    assert_eq!(reopened.status().unwrap()["unknown_invocations"],1);
    assert_eq!(reopened.status().unwrap()["open_slots"],1);
    assert_eq!(count(&reopened,"artifacts"),0);
    assert_eq!(count(&reopened,"acceptance_links"),0);
    let attempt: String = reopened.connection.query_row("SELECT id FROM attempts",[],|row| row.get(0)).unwrap();
    let observation = reopened.inspect_local_attempt(&attempt).unwrap();
    assert_eq!(observation["invocation_state"],"unknown");
    assert_eq!(observation["file_observation"]["readable"],mode=="after-file-sync");
    if mode == "after-file-sync" { assert_eq!(observation["file_observation"]["matches_intent"],true); }
    for _ in 0..3 {
        reopened.recover().unwrap();
        let queried = reopened.inspect_local_attempt(&attempt).unwrap();
        assert_eq!(queried["invocation_state"],"unknown");
        assert_eq!(queried["redispatch_permitted"],false);
        assert_eq!(queried["slots_released"],0);
        assert_eq!(queried["native_verified"],false);
        assert!(reopened.claim_local("task","write").is_err());
        assert_eq!(count(&reopened,"invocations"),1);
        assert_eq!(count(&reopened,"artifacts"),0);
        assert_eq!(count(&reopened,"acceptance_links"),0);
    }
}

#[test]
fn os_kill_after_dispatch_requires_query_and_holds_slot() { killed_dispatch("after-dispatch"); }

#[test]
fn os_kill_after_file_sync_cannot_fabricate_producer_receipt() { killed_dispatch("after-file-sync"); }

#[test]
fn os_kill_in_transaction_rolls_back_graph_change() {
    let root = workspace("process-rollback");
    let mut store = Store::open(&root).unwrap();
    store.plan(&plan(&store,"task",&[("write","result.txt")],&[],1)).unwrap();
    drop(store);
    let mut child = fixture(&root,"uncommitted");
    await_file(&root.join("crash.ready"));
    child.0.kill().unwrap();
    child.0.wait().unwrap();
    let reopened = Store::open_existing(&root).unwrap();
    let revision: i64 = reopened.connection.query_row("SELECT graph_revision FROM tasks",[],|row| row.get(0)).unwrap();
    assert_eq!(revision,1);
    assert_eq!(count(&reopened,"attempts"),0);
    assert_eq!(reopened.status().unwrap()["open_slots"],0);
}

#[test]
fn t11_full_local_revision_revalidates_only_affected_chain_and_preserves_history() {
    let root = workspace("t11-full-revision");
    let mut store = Store::open(&root).unwrap();
    let blueprint = plan(&store,"task",&[("first","first.txt"),("second","second.txt"),("independent","other.txt")],&[("first","second")],2);
    store.plan(&blueprint).unwrap();
    let mut prior_claims = Vec::new();
    for (node,path,content) in [("first","first.txt",b"first".as_slice()),("second","second.txt",b"second".as_slice()),("independent","other.txt",b"preserved".as_slice())] {
        let claim = store.claim_local("task",node).unwrap();
        store.write_local_file(&claim,Path::new(path),content).unwrap();
        store.verify_local_file(&claim).unwrap();
        store.close_local_handler(&claim).unwrap();
        prior_claims.push(claim);
    }
    assert_eq!(store.task_status("task").unwrap()["state"],"succeeded");
    let revised = revision(&blueprint,&[("first","first-v2.txt"),("second","second-v2.txt")],&["first","second"],2);
    let result = store.revise_local_plan(&revised,1,"changed-user-requirement").unwrap();
    assert_eq!(result["affected_nodes"],json!(["first","second"]));
    assert_eq!(count(&store,"acceptance_links"),3);
    assert_eq!(count(&store,"validations"),3);
    assert_eq!(store.revise_local_plan(&revised,1,"changed-user-requirement").unwrap(),result);
    assert_eq!(store.revise_local_plan(&revised,1,"stale-other-event").unwrap_err().kind,ErrorKind::RevisionConflict);
    assert!(store.verify_local_file(&prior_claims[0]).is_err());
    assert_eq!(store.claim_local("task","second").unwrap_err().kind,ErrorKind::DependencyNotAccepted);
    for (node,path,content) in [("first","first-v2.txt",b"new-first".as_slice()),("second","second-v2.txt",b"new-second".as_slice())] {
        let claim = store.claim_local("task",node).unwrap();
        store.write_local_file(&claim,Path::new(path),content).unwrap();
        let checked = store.verify_local_file(&claim).unwrap();
        assert_eq!(checked["professional_or_native_verified"],false);
        store.close_local_handler(&claim).unwrap();
    }
    assert_eq!(count(&store,"acceptance_links"),9);
    assert_eq!(count(&store,"validations"),5);
    assert_eq!(count(&store,"task_plans"),2);
    assert_eq!(fs::read(root.join("first.txt")).unwrap(),b"first");
    assert_eq!(fs::read(root.join("second.txt")).unwrap(),b"second");
    assert_eq!(fs::read(root.join("other.txt")).unwrap(),b"preserved");
    drop(store);
    let reopened = Store::open_existing(&root).unwrap();
    assert_eq!(reopened.task_status("task").unwrap()["state"],"succeeded");
    assert_eq!(reopened.status().unwrap()["open_slots"],0);
    assert_eq!(count(&reopened,"validations"),5);
}

#[test]
fn t12_full_local_repeated_events_never_create_second_dispatch_or_change_artifact() {
    let root = workspace("t12-event-replay");
    let mut store = Store::open(&root).unwrap();
    let blueprint = plan(&store,"task",&[("write","result.txt")],&[],1);
    let planned = store.plan(&blueprint).unwrap();
    let claim = store.claim_local("task","write").unwrap();
    let produced = store.write_local_file(&claim,Path::new("result.txt"),b"one actual dispatch").unwrap();
    let verified = store.verify_local_file(&claim).unwrap();
    store.close_local_handler(&claim).unwrap();
    for _ in 0..20 {
        assert_eq!(store.plan(&blueprint).unwrap(),planned);
        assert_eq!(store.write_local_file(&claim,Path::new("result.txt"),b"one actual dispatch").unwrap(),produced);
        assert_eq!(store.verify_local_file(&claim).unwrap(),verified);
        store.close_local_handler(&claim).unwrap();
    }
    assert_eq!(store.write_local_file(&claim,Path::new("result.txt"),b"conflicting replay").unwrap_err().kind,ErrorKind::EventConflict);
    assert_eq!(count(&store,"tasks"),1);
    assert_eq!(count(&store,"attempts"),1);
    assert_eq!(count(&store,"invocations"),1);
    assert_eq!(count(&store,"validations"),1);
    assert_eq!(count(&store,"artifacts"),1);
    assert_eq!(fs::read(root.join("result.txt")).unwrap(),b"one actual dispatch");
    assert_eq!(store.status().unwrap()["open_slots"],0);
    assert_eq!(produced["host_class"],"test_local");
}

#[test]
fn t18_changed_adopted_file_needs_new_revision_and_independent_current_validation() {
    let root = workspace("t18-revalidation");
    let mut store = Store::open(&root).unwrap();
    let blueprint = plan(&store,"task",&[("write","result.txt")],&[],1);
    store.plan(&blueprint).unwrap();
    let old = store.claim_local("task","write").unwrap();
    store.write_local_file(&old,Path::new("result.txt"),b"original validated output").unwrap();
    let prior = store.verify_local_file(&old).unwrap();
    store.close_local_handler(&old).unwrap();
    fs::write(root.join("result.txt"),b"modified outside producer").unwrap();
    assert!(store.verify_local_file(&old).is_err());
    assert_eq!(count(&store,"acceptance_links"),0);
    assert_eq!(count(&store,"validations"),1);
    assert_eq!(store.task_status("task").unwrap()["state"],"blocked");
    let revised = revision(&blueprint,&[("write","current.txt")],&["write"],2);
    store.revise_local_plan(&revised,1,"revalidate-changed-output").unwrap();
    assert!(store.write_local_file(&old,Path::new("current.txt"),b"old receipt cannot authorize new file").is_err());
    let current = store.claim_local("task","write").unwrap();
    store.write_local_file(&current,Path::new("current.txt"),b"new independently validated output").unwrap();
    assert_eq!(count(&store,"acceptance_links"),0);
    let accepted = store.verify_local_file(&current).unwrap();
    assert_ne!(prior,accepted);
    store.close_local_handler(&current).unwrap();
    assert_eq!(count(&store,"acceptance_links"),3);
    assert_eq!(count(&store,"validations"),2);
    assert_eq!(fs::read(root.join("result.txt")).unwrap(),b"modified outside producer");
    assert_eq!(store.task_status("task").unwrap()["graph_revision"],2);
    assert_eq!(store.task_status("task").unwrap()["state"],"succeeded");
}
