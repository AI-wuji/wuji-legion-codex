package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AI-wuji/wuji-legion-codex-2.0/internal/core"
)

const usage = `usage: wuji <route|response-policy|orchestrate|change-capsule|context-select|graph-sync|knowledge-record|knowledge-query|task-gate|task-record|requirement-record|decision-record|requirement-project|execution-record|execution-result|execution-project|acceptance-reconcile|staff-create|staff-update|staff-status|conversation-link|conversation-resolve|provenance-record|provenance-resolve|source-assess|source-impact|asset-select|graph-govern|audit-record|lineage-sync|security-gate|officer-select|dispatch|validate-receipt|verify|source-audit|evolve> [flags]

Commands:
  route           select a capability for a user request
  response-policy compile the final-writer rule overlay and session state
  orchestrate     prepare ordered native-worker contracts for General Staff scheduling
  change-capsule  create a bounded high-risk change contract
  context-select  select ranked code excerpts within a byte budget
  graph-sync      build or refresh the local workspace relation graph
	knowledge-record record a verified knowledge node; --feedback-id admits an eligible failure candidate only
  knowledge-query  query the event-triggered cross-project knowledge graph
  task-gate        check whether a task strategy may run under its circuit policy
  task-record      persist a task outcome for circuit enforcement
  requirement-record create a versioned requirement graph node
  decision-record   create a decision node bound to active requirements
  requirement-project create a sparse requirement/decision projection artifact
  execution-record create an execution node bound to exact requirements
  execution-result persist a deterministic execution result and evidence handles
  execution-project create a sparse execution graph projection artifact
  acceptance-reconcile bind an active requirement, succeeded execution, artifacts, and evidence
  staff-create     create a persistent task-level General Staff instance
  staff-update     incrementally update or replace a General Staff instance
  staff-status     read a persistent General Staff instance
  conversation-link link a requirement revision to opaque host message handles
  conversation-resolve resolve opaque message handles without chat replay
  provenance-record record one scope- and reader-bounded provenance edge
  provenance-resolve resolve provenance with a reader ACL recheck
  source-assess retain a minimal source/version decision for explicit reuse
  source-impact locate lineage nodes affected by a candidate source version
  asset-select select a compatible fusion asset and trusted invocation contract
  graph-govern validate bounded graph retention and perform eligible archival GC
  audit-record     append one bounded, idempotent audit event
  lineage-sync      persist compact fusion lineage and asset reachability metadata
  security-gate     evaluate a deterministic pre-side-effect security gate
  officer-select    select recommendation-only officers for a task signal
  dispatch        prepare one read-only native-host worker contract
  validate-receipt check receipt/route consistency (not native execution)
  verify          verify one or all capability manifests
  source-audit    report which retained sources are actually auto-selectable
  evolve          evaluate or apply a capability candidate`

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, usage)
		return 2
	}
	if args[0] == "help" || args[0] == "-h" || args[0] == "--help" {
		fmt.Fprintln(stdout, usage)
		return 0
	}

	root := discoverRoot()
	var output any
	var err error
	exitCode := 0
	switch args[0] {
	case "route":
		fs := newFlagSet("route", stderr)
		query := fs.String("query", "", "user request")
		rootFlag := fs.String("root", root, "Wuji 3.0 root")
		model := fs.String("model", "", "explicit user-selected model; empty uses the GPT hierarchy")
		contextArtifact := fs.String("context-artifact", "", "verified context artifact for bounded delegation")
		parentContextRequired := fs.Bool("parent-context-required", false, "block delegation when parent context must be replayed")
		selfContainedHandoff := fs.Bool("self-contained-handoff", false, "confirm the compact worker contract contains all required task context")
		responsePolicyActive := fs.Bool("response-policy-active", false, "carry an explicitly activated response policy into this turn")
		if code := parseFlags(fs, args[1:], stderr); code >= 0 {
			return code
		}
		if strings.TrimSpace(*query) == "" {
			return reportError(stderr, 2, errors.New("--query is required"))
		}
		items, err := core.LoadManifests(*rootFlag)
		if err != nil {
			return reportError(stderr, 1, err)
		}
		delegationContext := core.DelegationContext{ParentContextRequired: *parentContextRequired, SelfContained: *selfContainedHandoff}
		if strings.TrimSpace(*contextArtifact) != "" {
			delegationContext, err = core.LoadContextArtifact(*contextArtifact)
			if err != nil {
				return reportError(stderr, 2, err)
			}
			if *parentContextRequired {
				delegationContext.ParentContextRequired = true
			}
			if *selfContainedHandoff {
				delegationContext.SelfContained = true
			}
		}
		output = core.RouteWithContextModelAndResponseState(*query, items, delegationContext, *model, *responsePolicyActive)

	case "response-policy":
		fs := newFlagSet("response-policy", stderr)
		query := fs.String("query", "", "current user request")
		rootFlag := fs.String("root", root, "Wuji 3.0 root")
		active := fs.Bool("active", false, "policy was active in the preceding turn")
		if code := parseFlags(fs, args[1:], stderr); code >= 0 {
			return code
		}
		if strings.TrimSpace(*query) == "" {
			return reportError(stderr, 2, errors.New("--query is required"))
		}
		items, err := core.LoadManifests(*rootFlag)
		if err != nil {
			return reportError(stderr, 1, err)
		}
		var interaction *core.Manifest
		for index := range items {
			if items[index].ID == "interaction" {
				interaction = &items[index]
				break
			}
		}
		if interaction == nil {
			return reportError(stderr, 1, errors.New("interaction capability is unavailable"))
		}
		output, err = core.CompileResponsePolicy(*interaction, *query, *active)

	case "orchestrate":
		fs := newFlagSet("orchestrate", stderr)
		query := fs.String("query", "", "user request")
		rootFlag := fs.String("root", root, "Wuji 3.0 root")
		model := fs.String("model", "", "explicit user-selected model; empty uses the GPT hierarchy")
		workspace := fs.String("workspace", ".", "workspace passed to read-only workers")
		outputDir := fs.String("output-dir", ".wuji/dispatch", "directory for worker output evidence")
		contextArtifact := fs.String("context-artifact", "", "verified context artifact for bounded delegation")
		parentContextRequired := fs.Bool("parent-context-required", false, "compact parent context is unavailable to workers")
		selfContainedHandoff := fs.Bool("self-contained-handoff", false, "confirm the compact worker contract contains all required task context")
		codexPath := fs.String("codex", "", "Codex executable override")
		timeoutSeconds := fs.Int("timeout-seconds", 90, "per-worker timeout in seconds")
		maxParallel := fs.Int("max-parallel", 3, "maximum independent workers running together")
		dryRun := fs.Bool("dry-run", false, "plan an explicitly requested external CLI compatibility run")
		compatibilityExec := fs.Bool("compatibility-exec", false, "run external codex exec for local diagnostics only; never produces native worker evidence")
		if code := parseFlags(fs, args[1:], stderr); code >= 0 {
			return code
		}
		if strings.TrimSpace(*query) == "" {
			return reportError(stderr, 2, errors.New("--query is required"))
		}
		if *timeoutSeconds <= 0 || *maxParallel <= 0 {
			return reportError(stderr, 2, errors.New("--timeout-seconds and --max-parallel must be greater than zero"))
		}
		items, err := core.LoadManifests(*rootFlag)
		if err != nil {
			return reportError(stderr, 1, err)
		}

		delegationContext := core.DelegationContext{ParentContextRequired: *parentContextRequired, SelfContained: *selfContainedHandoff}
		if strings.TrimSpace(*contextArtifact) != "" {
			delegationContext, err = core.LoadContextArtifact(*contextArtifact)
			if err != nil {
				return reportError(stderr, 2, err)
			}
			if *parentContextRequired {
				delegationContext.ParentContextRequired = true
			}
			if *selfContainedHandoff {
				delegationContext.SelfContained = true
			}
		}
		initial := core.RouteWithContextAndModel(*query, items, delegationContext, *model)
		output, err = core.OrchestrateRoute(initial, core.OrchestrationOptions{
			Dispatch:    core.DispatchOptions{CodexPath: *codexPath, Workspace: *workspace, OutputDir: *outputDir, Timeout: time.Duration(*timeoutSeconds) * time.Second, DryRun: *dryRun, CompatibilityExec: *compatibilityExec, TrustedManifests: items},
			MaxParallel: *maxParallel,
		})
		if err != nil {
			return reportError(stderr, 1, err)
		}

	case "change-capsule":
		fs := newFlagSet("change-capsule", stderr)
		intent := fs.String("intent", "", "change intent")
		scopeOut := fs.String("scope-out", "", "explicitly excluded scope")
		acceptance := fs.String("acceptance", "", "semicolon-separated acceptance scenarios")
		verification := fs.String("verification", "", "semicolon-separated verification commands or evidence")
		rollback := fs.String("rollback", "", "rollback boundary")
		strict := fs.Bool("strict", false, "require scope, acceptance, verification, and rollback boundaries")
		if code := parseFlags(fs, args[1:], stderr); code >= 0 {
			return code
		}
		if strings.TrimSpace(*intent) == "" {
			return reportError(stderr, 2, errors.New("--intent is required"))
		}
		capsule := core.NewChangeCapsule(*intent, *scopeOut, *acceptance, *verification, *rollback)
		if *strict {
			validation := core.ValidateChangeCapsule(capsule, true)
			output = core.ChangeCapsuleValidationResult{Capsule: capsule, Validation: validation}
			if !validation.Valid {
				if err := writeJSON(stdout, output); err != nil {
					return reportError(stderr, 1, err)
				}
				return 1
			}
		} else {
			output = capsule
		}

	case "context-select":
		fs := newFlagSet("context-select", stderr)
		workspace := fs.String("workspace", ".", "workspace to search")
		query := fs.String("query", "", "retrieval query")
		budget := fs.Int("max-bytes", 12288, "maximum emitted context bytes")
		artifactDir := fs.String("artifact-dir", "", "content-addressed artifact directory (default: <workspace>/.wuji/context)")
		if code := parseFlags(fs, args[1:], stderr); code >= 0 {
			return code
		}
		result, err := core.SelectContext(*workspace, *query, *budget)
		if err != nil {
			return reportError(stderr, 2, err)
		}
		targetDir := *artifactDir
		if strings.TrimSpace(targetDir) == "" {
			targetDir = filepath.Join(result.Workspace, ".wuji", "context")
		}
		artifactPath, err := core.WriteContextArtifact(result, targetDir)
		if err != nil {
			return reportError(stderr, 1, err)
		}
		result.ArtifactPath = artifactPath
		output = result

	case "source-audit":
		fs := newFlagSet("source-audit", stderr)
		rootFlag := fs.String("root", root, "Wuji 3.0 root")
		if code := parseFlags(fs, args[1:], stderr); code >= 0 {
			return code
		}
		items, err := core.LoadManifests(*rootFlag)
		if err != nil {
			return reportError(stderr, 1, err)
		}
		output = core.AuditSources(items)

	case "graph-sync":
		fs := newFlagSet("graph-sync", stderr)
		workspace := fs.String("workspace", ".", "workspace to index")
		if code := parseFlags(fs, args[1:], stderr); code >= 0 {
			return code
		}
		result, err := core.SyncWorkspaceGraph(*workspace)
		if err != nil {
			return reportError(stderr, 1, err)
		}
		output = result

	case "knowledge-record":
		fs := newFlagSet("knowledge-record", stderr)
		store := fs.String("store", core.DefaultKnowledgeStore(), "knowledge store")
		kind := fs.String("kind", "", "knowledge node kind")
		key := fs.String("key", "", "stable knowledge key")
		scope := fs.String("scope", "", "canonical scope: global or workspace:<sha256>")
		workspace := fs.String("workspace", "", "workspace path used to derive a canonical scope")
		summary := fs.String("summary", "", "compact verified summary")
		rootCause := fs.String("root-cause", "", "verified root cause for a failure node")
		location := fs.String("location", "", "solution location or HTTPS URL")
		verification := fs.String("verification", "", "local verification evidence file")
		tags := fs.String("tags", "", "comma-separated tags")
		relations := fs.String("relations", "", "comma-separated predicate=target relations")
		feedbackID := fs.String("feedback-id", "", "eligible failed execution feedback candidate id")
		feedbackStore := fs.String("feedback-store", core.DefaultExecutionFeedbackStore(), "execution feedback store used with --feedback-id")
		if code := parseFlags(fs, args[1:], stderr); code >= 0 {
			return code
		}
		scopeValue, err := resolveKnowledgeScope(*scope, *workspace)
		if err != nil {
			return reportError(stderr, 2, err)
		}
		input := core.KnowledgeRecordInput{
			Kind: *kind, Key: *key, Scope: scopeValue, Summary: *summary, RootCause: *rootCause,
			Location: *location, Verification: *verification, Tags: splitCSV(*tags), Relations: parseKnowledgeRelations(*relations),
		}
		var record core.KnowledgeRecord
		if strings.TrimSpace(*feedbackID) != "" {
			record, err = core.RecordVerifiedFailureFeedbackKnowledge(*feedbackStore, *store, *feedbackID, input)
		} else {
			record, err = core.RecordKnowledge(*store, input)
		}
		if err != nil {
			return reportError(stderr, 2, err)
		}
		output = record

	case "knowledge-query":
		fs := newFlagSet("knowledge-query", stderr)
		store := fs.String("store", core.DefaultKnowledgeStore(), "knowledge store")
		trigger := fs.String("trigger", "", "event trigger: failure, explicit-reuse, capability-miss, or verification-trace")
		kind := fs.String("kind", "", "optional knowledge node kind")
		key := fs.String("key", "", "knowledge key")
		scope := fs.String("scope", "", "canonical scope: global or workspace:<sha256>")
		workspace := fs.String("workspace", "", "workspace path used to derive a canonical scope")
		tags := fs.String("tags", "", "comma-separated tags")
		relatedTo := fs.String("related-to", "", "relation target")
		relation := fs.String("relation", "", "optional relation predicate")
		limit := fs.Int("limit", 3, "maximum matches")
		if code := parseFlags(fs, args[1:], stderr); code >= 0 {
			return code
		}
		scopeValue := strings.TrimSpace(*scope)
		var err error
		if scopeValue == "" && strings.TrimSpace(*workspace) == "" && strings.TrimSpace(*trigger) == "" {
			scopeValue = "global"
		} else {
			scopeValue, err = resolveKnowledgeScope(*scope, *workspace)
		}
		if err != nil {
			return reportError(stderr, 2, err)
		}
		result, err := core.QueryKnowledge(*store, core.KnowledgeQuery{
			Trigger: *trigger, Kind: *kind, Key: *key, Scope: scopeValue, Tags: splitCSV(*tags), RelatedTo: *relatedTo, Relation: *relation, Limit: *limit,
		})
		if err != nil {
			return reportError(stderr, 2, err)
		}
		output = result

	case "task-gate", "task-record":
		fs := newFlagSet(args[0], stderr)
		store := fs.String("store", filepath.Join(root, ".wuji", "task-circuits"), "task circuit state store")
		taskID := fs.String("task", "", "stable task id")
		strategyID := fs.String("strategy", "", "stable strategy id")
		policyID := fs.String("policy", "", "stable circuit policy id")
		maxNoProgress := fs.Int("max-no-progress", 2, "maximum consecutive no-progress outcomes before blocking")
		attemptID := fs.String("attempt", "", "deterministic attempt signature")
		outcome := fs.String("outcome", "", "progress, success, no-progress, or failure (task-record only)")
		transientFailure := fs.Bool("transient-failure", false, "failure was caused by a transient external condition")
		evidenceSHA256 := fs.String("evidence-sha256", "", "optional local evidence SHA-256")
		if code := parseFlags(fs, args[1:], stderr); code >= 0 {
			return code
		}
		policy := core.TaskCircuitPolicy{ID: *policyID, MaxNoProgress: *maxNoProgress}
		input := core.TaskAttemptInput{TaskID: *taskID, StrategyID: *strategyID, AttemptID: *attemptID, Outcome: *outcome, TransientFailure: *transientFailure, EvidenceSHA256: *evidenceSHA256}
		var err error
		if args[0] == "task-gate" {
			output, err = core.CheckTaskCircuit(*store, policy, input)
		} else {
			output, err = core.RecordTaskAttempt(*store, policy, input)
		}
		if err != nil {
			return reportError(stderr, 2, err)
		}

	case "requirement-record":
		fs := newFlagSet("requirement-record", stderr)
		store := fs.String("store", filepath.Join(root, ".wuji", "requirement-graph"), "requirement graph store")
		id := fs.String("id", "", "stable requirement id")
		summary := fs.String("summary", "", "compact requirement summary")
		goal := fs.String("goal", "", "desired outcome")
		wants := fs.String("wants", "", "comma-separated desired properties")
		avoids := fs.String("avoids", "", "comma-separated exclusions")
		constraints := fs.String("constraints", "", "comma-separated hard constraints")
		preferences := fs.String("preferences", "", "comma-separated preferences")
		priority := fs.String("priority", "", "priority label")
		acceptance := fs.String("acceptance", "", "comma-separated acceptance criteria")
		decisions := fs.String("decisions", "", "comma-separated decisions")
		openQuestions := fs.String("open-questions", "", "comma-separated open questions")
		conflicts := fs.String("conflicts", "", "comma-separated conflicts")
		sourceMessages := fs.String("source-messages", "", "comma-separated source message references")
		dependsOn := fs.String("depends-on", "", "comma-separated requirement version references")
		sources := fs.String("sources", "", "comma-separated provenance references")
		if code := parseFlags(fs, args[1:], stderr); code >= 0 {
			return code
		}
		result, err := core.UpsertRequirement(*store, core.RequirementInput{ID: *id, Summary: *summary, Goal: *goal, Wants: splitCSV(*wants), Avoids: splitCSV(*avoids), Constraints: splitCSV(*constraints), Preferences: splitCSV(*preferences), Priority: *priority, Acceptance: splitCSV(*acceptance), Decisions: splitCSV(*decisions), OpenQuestions: splitCSV(*openQuestions), Conflicts: splitCSV(*conflicts), SourceMessages: splitCSV(*sourceMessages), DependsOn: splitCSV(*dependsOn), Sources: splitCSV(*sources)})
		if err != nil {
			return reportError(stderr, 2, err)
		}
		output = result

	case "decision-record":
		fs := newFlagSet("decision-record", stderr)
		store := fs.String("store", filepath.Join(root, ".wuji", "requirement-graph"), "requirement graph store")
		id := fs.String("id", "", "stable decision id")
		summary := fs.String("summary", "", "compact decision summary")
		goal := fs.String("goal", "", "desired outcome")
		wants := fs.String("wants", "", "comma-separated desired properties")
		avoids := fs.String("avoids", "", "comma-separated exclusions")
		constraints := fs.String("constraints", "", "comma-separated hard constraints")
		preferences := fs.String("preferences", "", "comma-separated preferences")
		priority := fs.String("priority", "", "priority label")
		acceptance := fs.String("acceptance", "", "comma-separated acceptance criteria")
		decisions := fs.String("decisions", "", "comma-separated decisions")
		openQuestions := fs.String("open-questions", "", "comma-separated open questions")
		conflicts := fs.String("conflicts", "", "comma-separated conflicts")
		sourceMessages := fs.String("source-messages", "", "comma-separated source message references")
		sources := fs.String("sources", "", "comma-separated provenance references")
		requirements := fs.String("requirements", "", "comma-separated active requirement ids or revisions")
		status := fs.String("status", "proposed", "proposed, accepted, or rejected")
		if code := parseFlags(fs, args[1:], stderr); code >= 0 {
			return code
		}
		result, err := core.RecordDecision(*store, core.DecisionInput{ID: *id, Summary: *summary, Goal: *goal, Wants: splitCSV(*wants), Avoids: splitCSV(*avoids), Constraints: splitCSV(*constraints), Preferences: splitCSV(*preferences), Priority: *priority, Acceptance: splitCSV(*acceptance), Decisions: splitCSV(*decisions), OpenQuestions: splitCSV(*openQuestions), Conflicts: splitCSV(*conflicts), SourceMessages: splitCSV(*sourceMessages), Sources: splitCSV(*sources), Requirements: splitCSV(*requirements), Status: *status})
		if err != nil {
			return reportError(stderr, 2, err)
		}
		output = result

	case "requirement-project":
		fs := newFlagSet("requirement-project", stderr)
		store := fs.String("store", filepath.Join(root, ".wuji", "requirement-graph"), "requirement graph store")
		id := fs.String("id", "", "requirement or decision id to project")
		maxBytes := fs.Int("max-bytes", 4096, "hard projection byte budget")
		artifactDir := fs.String("artifact-dir", "", "projection artifact directory (default: <store>/v1/projections)")
		if code := parseFlags(fs, args[1:], stderr); code >= 0 {
			return code
		}
		projection, err := core.ProjectRequirementGraph(*store, *id, *maxBytes)
		if err != nil {
			return reportError(stderr, 2, err)
		}
		targetDir := *artifactDir
		if strings.TrimSpace(targetDir) == "" {
			targetDir = filepath.Join(*store, "v1", "projections")
		}
		projection.ArtifactPath, err = core.WriteRequirementGraphProjection(projection, targetDir)
		if err != nil {
			return reportError(stderr, 1, err)
		}
		output = projection

	case "execution-record":
		fs := newFlagSet("execution-record", stderr)
		store := fs.String("store", filepath.Join(root, ".wuji", "execution-graph"), "execution graph store")
		requirementStore := fs.String("requirements-store", filepath.Join(root, ".wuji", "requirement-graph"), "requirement graph store")
		id := fs.String("id", "", "stable execution node id")
		authority := fs.String("authority", "executor", "execution-node authority")
		goal := fs.String("goal", "", "execution goal")
		avoids := fs.String("avoids", "", "comma-separated exclusions")
		requirements := fs.String("requirements", "", "comma-separated exact requirement revisions")
		dependsOn := fs.String("depends-on", "", "comma-separated execution node versions")
		inputs := fs.String("inputs", "", "comma-separated input handles")
		allowedContext := fs.String("allowed-context", "", "comma-separated context handles")
		outputs := fs.String("outputs", "", "comma-separated output handles")
		model := fs.String("model", "", "selected model")
		modelReason := fs.String("model-reason", "", "model selection reason")
		acceptance := fs.String("acceptance", "", "comma-separated acceptance criteria")
		verification := fs.String("verification", "", "comma-separated verification commands")
		evidence := fs.String("evidence-required", "", "comma-separated evidence handles")
		timeBudget := fs.Int("time-budget-seconds", 0, "time budget")
		costBudget := fs.Int64("cost-budget-microunits", 0, "cost budget")
		maxAttempts := fs.Int("max-attempts", 0, "maximum attempts")
		networkBoundary := fs.String("network-boundary", "none", "network boundary")
		writeBoundary := fs.String("write-boundary", "executor-owned", "write boundary")
		branchBoundary := fs.String("branch-boundary", "current", "branch boundary")
		taskInstance := fs.String("task-instance", "", "task-level General Staff instance id")
		graphVersion := fs.String("graph-version", "", "task graph version")
		attempt := fs.String("attempt", "", "execution attempt id")
		if code := parseFlags(fs, args[1:], stderr); code >= 0 {
			return code
		}
		result, err := core.RecordVersionedExecutionNode(*store, *requirementStore, core.VersionedExecutionNodeInput{ExecutionNodeInput: core.ExecutionNodeInput{ID: *id, Authority: *authority, Goal: *goal, Avoids: splitCSV(*avoids), RequirementRevisions: splitCSV(*requirements), DependsOn: splitCSV(*dependsOn), Inputs: splitCSV(*inputs), AllowedContext: splitCSV(*allowedContext), Outputs: splitCSV(*outputs), Model: *model, ModelReason: *modelReason, Acceptance: splitCSV(*acceptance), Verification: splitCSV(*verification), EvidenceRequired: splitCSV(*evidence), TimeBudgetSeconds: *timeBudget, CostBudgetMicrounits: *costBudget, MaxAttempts: *maxAttempts, NetworkBoundary: *networkBoundary, WriteBoundary: *writeBoundary, BranchBoundary: *branchBoundary}, TaskInstanceID: *taskInstance, GraphVersion: *graphVersion, AttemptID: *attempt})
		if err != nil {
			return reportError(stderr, 2, err)
		}
		output = result

	case "execution-result":
		fs := newFlagSet("execution-result", stderr)
		store := fs.String("store", filepath.Join(root, ".wuji", "execution-graph"), "execution graph store")
		requirementStore := fs.String("requirements-store", filepath.Join(root, ".wuji", "requirement-graph"), "requirement graph store")
		id := fs.String("id", "", "execution node id")
		status := fs.String("status", "", "planned, running, succeeded, failed, or invalidated")
		failure := fs.String("failure", "", "failure summary")
		recovery := fs.String("recovery", "", "recovery summary")
		artifacts := fs.String("artifacts", "", "comma-separated artifact handles")
		verification := fs.String("verification", "", "comma-separated verification handles")
		taskInstance := fs.String("task-instance", "", "task-level General Staff instance id")
		graphVersion := fs.String("graph-version", "", "task graph version")
		attempt := fs.String("attempt", "", "execution attempt id")
		if code := parseFlags(fs, args[1:], stderr); code >= 0 {
			return code
		}
		result, err := core.RecordVersionedExecutionResult(*store, *requirementStore, core.VersionedExecutionResultInput{ExecutionResultInput: core.ExecutionResultInput{ID: *id, Status: *status, Failure: *failure, Recovery: *recovery, ArtifactHandles: splitCSV(*artifacts), VerificationHandles: splitCSV(*verification)}, TaskInstanceID: *taskInstance, GraphVersion: *graphVersion, AttemptID: *attempt})
		if err != nil {
			return reportError(stderr, 2, err)
		}
		output = result

	case "execution-project":
		fs := newFlagSet("execution-project", stderr)
		store := fs.String("store", filepath.Join(root, ".wuji", "execution-graph"), "execution graph store")
		requirementStore := fs.String("requirements-store", filepath.Join(root, ".wuji", "requirement-graph"), "requirement graph store")
		id := fs.String("id", "", "execution node id or version")
		maxBytes := fs.Int("max-bytes", 4096, "hard projection byte budget")
		artifactDir := fs.String("artifact-dir", "", "projection artifact directory")
		if code := parseFlags(fs, args[1:], stderr); code >= 0 {
			return code
		}
		projection, err := core.ProjectExecutionGraph(*store, *requirementStore, *id, *maxBytes)
		if err != nil {
			return reportError(stderr, 2, err)
		}
		targetDir := *artifactDir
		if strings.TrimSpace(targetDir) == "" {
			targetDir = filepath.Join(*store, "v1", "projections")
		}
		projection.ArtifactPath, err = core.WriteExecutionGraphProjection(projection, targetDir)
		if err != nil {
			return reportError(stderr, 1, err)
		}
		output = projection

	case "acceptance-reconcile":
		fs := newFlagSet("acceptance-reconcile", stderr)
		store := fs.String("store", filepath.Join(root, ".wuji", "acceptance"), "acceptance ledger store")
		requirementStore := fs.String("requirements-store", filepath.Join(root, ".wuji", "requirement-graph"), "requirement graph store")
		executionStore := fs.String("execution-store", filepath.Join(root, ".wuji", "execution-graph"), "execution graph store")
		id := fs.String("id", "", "stable acceptance id")
		requirement := fs.String("requirement", "", "exact active requirement revision")
		execution := fs.String("execution", "", "exact succeeded execution version")
		artifacts := fs.String("artifacts", "", "comma-separated artifact handles")
		verification := fs.String("verification", "", "comma-separated verification handles")
		if code := parseFlags(fs, args[1:], stderr); code >= 0 {
			return code
		}
		output, err = core.ReconcileAcceptance(*store, *requirementStore, *executionStore, core.AcceptanceInput{ID: *id, RequirementRevision: *requirement, ExecutionVersion: *execution, ArtifactHandles: splitCSV(*artifacts), VerificationHandles: splitCSV(*verification)})
		if err != nil {
			return reportError(stderr, 2, err)
		}

	case "staff-create":
		fs := newFlagSet("staff-create", stderr)
		store := fs.String("store", core.DefaultGeneralStaffStore(), "General Staff state store")
		taskInstance := fs.String("task-instance", "", "task-level instance id")
		sessionKey := fs.String("session-key", "", "sticky task-state session key")
		requirementVersion := fs.String("requirement-version", "", "active requirement version")
		graphVersion := fs.String("graph-version", "", "task graph version")
		if code := parseFlags(fs, args[1:], stderr); code >= 0 {
			return code
		}
		output, err = core.CreateGeneralStaffState(*store, core.GeneralStaffSnapshot{TaskInstanceID: *taskInstance, SessionKey: *sessionKey, RequirementVersion: *requirementVersion, GraphVersion: *graphVersion})
		if err != nil {
			return reportError(stderr, 2, err)
		}

	case "staff-update":
		fs := newFlagSet("staff-update", stderr)
		store := fs.String("store", core.DefaultGeneralStaffStore(), "General Staff state store")
		taskInstance := fs.String("task-instance", "", "task-level instance id")
		sessionKey := fs.String("session-key", "", "sticky task-state session key")
		requirementVersion := fs.String("requirement-version", "", "active requirement version")
		graphVersion := fs.String("graph-version", "", "task graph version")
		veto := fs.Bool("veto", false, "replace the current instance and task graph")
		if code := parseFlags(fs, args[1:], stderr); code >= 0 {
			return code
		}
		output, err = core.UpdateGeneralStaffState(*store, core.GeneralStaffUpdate{Snapshot: core.GeneralStaffSnapshot{TaskInstanceID: *taskInstance, SessionKey: *sessionKey, RequirementVersion: *requirementVersion, GraphVersion: *graphVersion}, Veto: *veto})
		if err != nil {
			return reportError(stderr, 2, err)
		}

	case "staff-status":
		fs := newFlagSet("staff-status", stderr)
		store := fs.String("store", core.DefaultGeneralStaffStore(), "General Staff state store")
		if code := parseFlags(fs, args[1:], stderr); code >= 0 {
			return code
		}
		output, err = core.ReadGeneralStaffState(*store)
		if err != nil {
			return reportError(stderr, 2, err)
		}

	case "conversation-link":
		fs := newFlagSet("conversation-link", stderr)
		store := fs.String("store", filepath.Join(root, ".wuji", "conversation-evidence"), "opaque conversation evidence store")
		requirementStore := fs.String("requirements-store", filepath.Join(root, ".wuji", "requirement-graph"), "requirement graph store")
		revision := fs.String("revision", "", "exact requirement or decision revision")
		messages := fs.String("messages", "", "comma-separated opaque host message handles")
		if code := parseFlags(fs, args[1:], stderr); code >= 0 {
			return code
		}
		output, err = core.LinkConversationEvidence(*store, *requirementStore, *revision, splitCSV(*messages))
		if err != nil {
			return reportError(stderr, 2, err)
		}

	case "conversation-resolve":
		fs := newFlagSet("conversation-resolve", stderr)
		store := fs.String("store", filepath.Join(root, ".wuji", "conversation-evidence"), "opaque conversation evidence store")
		revision := fs.String("revision", "", "exact requirement or decision revision")
		message := fs.String("message", "", "opaque host message handle")
		if code := parseFlags(fs, args[1:], stderr); code >= 0 {
			return code
		}
		output, err = core.ResolveConversationEvidence(*store, core.ConversationEvidenceQuery{Revision: *revision, MessageHandle: *message})
		if err != nil {
			return reportError(stderr, 2, err)
		}

	case "provenance-record":
		fs := newFlagSet("provenance-record", stderr)
		store := fs.String("store", filepath.Join(root, ".wuji", "provenance"), "provenance store")
		id := fs.String("id", "", "immutable provenance id")
		scope := fs.String("scope", "global", "global or canonical workspace scope")
		subject := fs.String("subject", "", "scoped provenance subject handle")
		predicate := fs.String("predicate", "", "bounded provenance predicate")
		target := fs.String("target", "", "scoped provenance target handle")
		readers := fs.String("readers", "aji", "comma-separated permitted readers, or *")
		if code := parseFlags(fs, args[1:], stderr); code >= 0 {
			return code
		}
		output, err = core.RecordProvenance(*store, core.ProvenanceInput{ID: *id, Scope: *scope, Subject: *subject, Predicate: *predicate, Target: *target, Readers: splitCSV(*readers)})
		if err != nil {
			return reportError(stderr, 2, err)
		}

	case "provenance-resolve":
		fs := newFlagSet("provenance-resolve", stderr)
		store := fs.String("store", filepath.Join(root, ".wuji", "provenance"), "provenance store")
		scope := fs.String("scope", "global", "global or canonical workspace scope")
		subject := fs.String("subject", "", "scoped provenance subject handle")
		principal := fs.String("principal", "aji", "authenticated host principal for the ACL recheck")
		if code := parseFlags(fs, args[1:], stderr); code >= 0 {
			return code
		}
		output, err = core.ResolveProvenance(*store, core.ProvenanceQuery{Scope: *scope, Subject: *subject, Principal: *principal})
		if err != nil {
			return reportError(stderr, 2, err)
		}

	case "source-assess":
		fs := newFlagSet("source-assess", stderr)
		store := fs.String("store", filepath.Join(root, ".wuji", "source-assessments"), "minimal source assessment store")
		source := fs.String("source", "", "stable source id")
		version := fs.String("version", "", "source version")
		decision := fs.String("decision", "", "adopted, rejected, or deferred")
		reason := fs.String("reason", "", "bounded decision reason")
		evidence := fs.String("evidence", "", "comma-separated evidence handles")
		reanalyze := fs.String("reanalyze-when", "", "comma-separated reanalysis conditions")
		if code := parseFlags(fs, args[1:], stderr); code >= 0 {
			return code
		}
		output, err = core.AssessSource(*store, core.SourceAssessmentInput{SourceID: *source, Version: *version, Decision: *decision, Reason: *reason, EvidenceHandles: splitCSV(*evidence), ReanalyzeWhen: splitCSV(*reanalyze)})
		if err != nil {
			return reportError(stderr, 2, err)
		}

	case "source-impact":
		fs := newFlagSet("source-impact", stderr)
		rootFlag := fs.String("root", root, "Wuji root and lineage catalog owner")
		source := fs.String("source", "", "stable source id")
		version := fs.String("candidate-version", "", "candidate source version")
		if code := parseFlags(fs, args[1:], stderr); code >= 0 {
			return code
		}
		items, err := core.LoadManifests(*rootFlag)
		if err != nil {
			return reportError(stderr, 1, err)
		}
		catalog, err := core.SyncLineageCatalog(*rootFlag, items)
		if err != nil {
			return reportError(stderr, 1, err)
		}
		output, err = core.SourceImpact(catalog.Catalog, *source, *version)
		if err != nil {
			return reportError(stderr, 2, err)
		}

	case "asset-select":
		fs := newFlagSet("asset-select", stderr)
		rootFlag := fs.String("root", root, "Wuji root used to resolve trusted fusion assets")
		capability := fs.String("capability", "", "optional capability id")
		domain := fs.String("domain", "", "fusion adapter domain")
		assetID := fs.String("asset-id", "", "stable asset id or capability-qualified asset id")
		compatibility := fs.String("compatibility", "", "comma-separated required compatibility tags")
		if code := parseFlags(fs, args[1:], stderr); code >= 0 {
			return code
		}
		items, err := core.LoadManifests(*rootFlag)
		if err != nil {
			return reportError(stderr, 1, err)
		}
		output, err = core.SelectFusionAsset(items, core.AssetSelectionRequest{Capability: *capability, Domain: *domain, AssetID: *assetID, Compatibility: splitCSV(*compatibility)})
		if err != nil {
			return reportError(stderr, 2, err)
		}

	case "graph-govern":
		fs := newFlagSet("graph-govern", stderr)
		graph := fs.String("graph", "", "acceptance, conversation-evidence, provenance, or source-assessments")
		store := fs.String("store", "", "graph store; defaults under the selected root")
		if code := parseFlags(fs, args[1:], stderr); code >= 0 {
			return code
		}
		if strings.TrimSpace(*graph) == "" {
			output = core.GraphRetentionPolicies()
			break
		}
		selectedStore := *store
		if strings.TrimSpace(selectedStore) == "" {
			selectedStore = governedStore(root, *graph)
		}
		output, err = core.MaintainGraph(selectedStore, *graph, time.Now().UTC())
		if err != nil {
			return reportError(stderr, 2, err)
		}

	case "audit-record":
		fs := newFlagSet("audit-record", stderr)
		store := fs.String("store", filepath.Join(root, ".wuji", "audit"), "audit event store")
		eventID := fs.String("event-id", "", "idempotency key")
		eventType := fs.String("event-type", "", "event type")
		actor := fs.String("actor", "aji", "actor")
		authority := fs.String("authority", "aji-merge", "authority")
		target := fs.String("target", "", "target handle")
		inputRevision := fs.String("input-revision", "", "input revision")
		resultHandle := fs.String("result-handle", "", "result handle")
		evidence := fs.String("evidence", "", "comma-separated evidence handles")
		if code := parseFlags(fs, args[1:], stderr); code >= 0 {
			return code
		}
		event := core.AuditEvent{EventID: *eventID, EventType: *eventType, Actor: *actor, Authority: *authority, Target: *target, InputRevision: *inputRevision, ResultHandle: *resultHandle, EvidenceHandles: splitCSV(*evidence)}
		if err := core.AuditEventRecord(*store, event); err != nil {
			return reportError(stderr, 2, err)
		}
		output = event

	case "lineage-sync":
		fs := newFlagSet("lineage-sync", stderr)
		rootFlag := fs.String("root", root, "Wuji 3.0 root and lineage catalog owner")
		if code := parseFlags(fs, args[1:], stderr); code >= 0 {
			return code
		}
		items, err := core.LoadManifests(*rootFlag)
		if err != nil {
			return reportError(stderr, 1, err)
		}
		output, err = core.SyncLineageCatalog(*rootFlag, items)
		if err != nil {
			return reportError(stderr, 1, err)
		}

	case "security-gate":
		fs := newFlagSet("security-gate", stderr)
		kind := fs.String("kind", "", "controlled action kind")
		target := fs.String("target", "", "bounded action target")
		workspace := fs.String("workspace", "", "optional workspace boundary for file actions")
		explicitUserIntent := fs.Bool("explicit-user-intent", false, "the user explicitly approved this controlled action")
		auditStore := fs.String("audit-store", filepath.Join(root, ".wuji", "audit"), "append-only audit event store")
		if code := parseFlags(fs, args[1:], stderr); code >= 0 {
			return code
		}
		gate := core.EvaluateSecurityGate(core.SecurityAction{Kind: *kind, Target: *target, Workspace: *workspace, ExplicitUserIntent: *explicitUserIntent})
		output = gate
		if err := core.AuditEventRecord(*auditStore, core.AuditEvent{EventType: "security-gate", Actor: "aji", Authority: "deterministic-gate", Target: *kind + ":" + *target, ResultHandle: gate.Decision}); err != nil {
			return reportError(stderr, 1, err)
		}
		if !gate.Allowed {
			exitCode = 1
		}

	case "officer-select":
		fs := newFlagSet("officer-select", stderr)
		query := fs.String("query", "", "task signal to evaluate")
		if code := parseFlags(fs, args[1:], stderr); code >= 0 {
			return code
		}
		if strings.TrimSpace(*query) == "" {
			return reportError(stderr, 2, errors.New("--query is required"))
		}
		output = core.SelectOfficerRecommendations(*query)

	case "dispatch":
		fs := newFlagSet("dispatch", stderr)
		routePath := fs.String("route", "", "route JSON file")
		rootFlag := fs.String("root", root, "Wuji 3.0 root used to validate selected sources")
		workerID := fs.String("worker", "", "worker id from the route")
		workspace := fs.String("workspace", ".", "workspace passed to the read-only worker")
		outputDir := fs.String("output-dir", ".wuji/dispatch", "directory for worker output evidence")
		codexPath := fs.String("codex", "", "Codex executable override")
		timeoutSeconds := fs.Int("timeout-seconds", 90, "per-attempt timeout in seconds")
		dryRun := fs.Bool("dry-run", false, "plan an explicitly requested external CLI compatibility run")
		compatibilityExec := fs.Bool("compatibility-exec", false, "run external codex exec for local diagnostics only; never produces native worker evidence")
		if code := parseFlags(fs, args[1:], stderr); code >= 0 {
			return code
		}
		if strings.TrimSpace(*routePath) == "" || strings.TrimSpace(*workerID) == "" {
			return reportError(stderr, 2, errors.New("--route and --worker are required"))
		}
		if *timeoutSeconds <= 0 {
			return reportError(stderr, 2, errors.New("--timeout-seconds must be greater than zero"))
		}
		routeData, err := os.ReadFile(*routePath)
		if err != nil {
			return reportError(stderr, 1, err)
		}
		route, err := decodeRoute(routeData)
		if err != nil {
			return reportError(stderr, 2, err)
		}
		worker := findRouteWorker(route, *workerID)
		if worker == nil {
			return reportError(stderr, 2, fmt.Errorf("worker not found in route: %s", *workerID))
		}
		items, err := core.LoadManifests(*rootFlag)
		if err != nil {
			return reportError(stderr, 1, err)
		}
		result, err := core.DispatchWorker(*worker, core.DispatchOptions{
			CodexPath:         *codexPath,
			Workspace:         *workspace,
			OutputDir:         *outputDir,
			Timeout:           time.Duration(*timeoutSeconds) * time.Second,
			DryRun:            *dryRun,
			CompatibilityExec: *compatibilityExec,
			TrustedManifests:  items,
		})
		if err != nil {
			return reportError(stderr, 2, err)
		}
		output = result
		if strings.HasPrefix(result.Status, "compatibility-exec-failed-") {
			exitCode = 1
		}

	case "validate-receipt":
		fs := newFlagSet("validate-receipt", stderr)
		routePath := fs.String("route", "", "route JSON file")
		receiptPath := fs.String("receipt", "", "worker receipt JSON file")
		workerID := fs.String("worker", "", "worker id from the route")
		if code := parseFlags(fs, args[1:], stderr); code >= 0 {
			return code
		}
		if strings.TrimSpace(*routePath) == "" || strings.TrimSpace(*receiptPath) == "" || strings.TrimSpace(*workerID) == "" {
			return reportError(stderr, 2, errors.New("--route, --receipt and --worker are required"))
		}
		routeData, err := os.ReadFile(*routePath)
		if err != nil {
			return reportError(stderr, 1, err)
		}
		receiptData, err := os.ReadFile(*receiptPath)
		if err != nil {
			return reportError(stderr, 1, err)
		}
		route, err := decodeRoute(routeData)
		if err != nil {
			return reportError(stderr, 2, err)
		}
		var receipt core.WorkerExecutionReceipt
		if err := json.Unmarshal(receiptData, &receipt); err != nil {
			return reportError(stderr, 2, fmt.Errorf("decode receipt: %w", err))
		}
		worker := findRouteWorker(route, *workerID)
		if worker == nil {
			return reportError(stderr, 2, fmt.Errorf("worker not found in route: %s", *workerID))
		}
		if err := core.ValidateWorkerReceiptConsistency(*worker, receipt); err != nil {
			return reportError(stderr, 1, err)
		}
		output = map[string]any{
			"contract_consistent": true,
			"execution_verified":  false,
			"reason":              "a Desktop native-host attestation or real MCP handler result is required",
			"worker_id":           *workerID,
			"effective_model":     receipt.EffectiveModel,
			"savings_microunits":  receipt.SavingsMicrounits,
		}

	case "verify":
		fs := newFlagSet("verify", stderr)
		capability := fs.String("capability", "all", "capability id or all")
		rootFlag := fs.String("root", root, "Wuji 3.0 root")
		if code := parseFlags(fs, args[1:], stderr); code >= 0 {
			return code
		}
		items, err := core.LoadManifests(*rootFlag)
		if err != nil {
			return reportError(stderr, 1, err)
		}
		results := []core.VerifyResult{}
		failed := false
		for _, item := range items {
			if *capability == "all" || *capability == item.ID {
				result := core.Verify(*rootFlag, item)
				results = append(results, result)
				failed = failed || !result.Passed
			}
		}
		if len(results) == 0 {
			return reportError(stderr, 2, fmt.Errorf("capability not found: %s", *capability))
		}
		if err := writeJSON(stdout, results); err != nil {
			return reportError(stderr, 1, err)
		}
		if failed {
			return reportError(stderr, 1, errors.New("capability verification failed"))
		}
		return 0

	case "evolve":
		fs := newFlagSet("evolve", stderr)
		candidate := fs.String("candidate", "", "candidate manifest path")
		apply := fs.Bool("apply", false, "admit or replace a behavior-verified candidate")
		rootFlag := fs.String("root", root, "Wuji 3.0 root")
		if code := parseFlags(fs, args[1:], stderr); code >= 0 {
			return code
		}
		if strings.TrimSpace(*candidate) == "" {
			return reportError(stderr, 2, errors.New("--candidate is required"))
		}
		result, err := core.EvaluateCandidate(*rootFlag, *candidate, *apply)
		if err != nil {
			return reportError(stderr, 1, err)
		}
		output = result

	default:
		return reportError(stderr, 2, fmt.Errorf("unknown command: %s", args[0]))
	}

	if err := writeJSON(stdout, output); err != nil {
		return reportError(stderr, 1, err)
	}
	return exitCode
}

func resolveKnowledgeScope(scope, workspace string) (string, error) {
	if strings.TrimSpace(workspace) != "" {
		if strings.TrimSpace(scope) != "" {
			return "", errors.New("use --workspace or --scope, not both")
		}
		return core.KnowledgeWorkspaceScope(workspace)
	}
	if strings.TrimSpace(scope) == "" {
		return "", errors.New("--scope or --workspace is required")
	}
	return strings.ToLower(strings.TrimSpace(scope)), nil
}

func findRouteWorker(route core.RouteResult, workerID string) *core.WorkerTask {
	for index := range route.PreflightWorkers {
		if route.PreflightWorkers[index].ID == workerID {
			return &route.PreflightWorkers[index]
		}
	}
	for index := range route.Workers {
		if route.Workers[index].ID == workerID {
			return &route.Workers[index]
		}
	}
	for index := range route.OfficerWorkers {
		if route.OfficerWorkers[index].ID == workerID {
			return &route.OfficerWorkers[index]
		}
	}
	return nil
}

func decodeRoute(data []byte) (core.RouteResult, error) {
	var route core.RouteResult
	if err := json.Unmarshal(bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF}), &route); err != nil {
		return core.RouteResult{}, fmt.Errorf("decode route: %w", err)
	}
	return route, nil
}

func newFlagSet(name string, stderr io.Writer) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(stderr)
	return fs
}

func parseFlags(fs *flag.FlagSet, args []string, stderr io.Writer) int {
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if fs.NArg() > 0 {
		return reportError(stderr, 2, fmt.Errorf("unexpected arguments: %s", strings.Join(fs.Args(), " ")))
	}
	return -1
}

func writeJSON(writer io.Writer, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(writer, string(data))
	return err
}

func splitCSV(value string) []string {
	values := strings.Split(value, ",")
	result := make([]string, 0, len(values))
	for _, item := range values {
		if item = strings.TrimSpace(item); item != "" {
			result = append(result, item)
		}
	}
	return result
}

func parseKnowledgeRelations(value string) []core.KnowledgeRelation {
	result := []core.KnowledgeRelation{}
	for _, item := range splitCSV(value) {
		parts := strings.SplitN(item, "=", 2)
		if len(parts) == 2 {
			result = append(result, core.KnowledgeRelation{Predicate: strings.TrimSpace(parts[0]), Target: strings.TrimSpace(parts[1])})
		}
	}
	return result
}

func reportError(writer io.Writer, code int, err error) int {
	fmt.Fprintln(writer, "error:", err)
	return code
}

func discoverRoot() string {
	if value := os.Getenv("WUJI_ROOT"); value != "" {
		return value
	}
	current, err := os.Getwd()
	if err != nil {
		return "."
	}
	for {
		if _, err := os.Stat(filepath.Join(current, "SKILL.md")); err == nil {
			if _, err := os.Stat(filepath.Join(current, "capabilities")); err == nil {
				return current
			}
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return "."
}

func governedStore(root, graph string) string {
	switch graph {
	case "acceptance":
		return filepath.Join(root, ".wuji", "acceptance")
	case "conversation-evidence":
		return filepath.Join(root, ".wuji", "conversation-evidence")
	case "provenance":
		return filepath.Join(root, ".wuji", "provenance")
	case "source-assessments":
		return filepath.Join(root, ".wuji", "source-assessments")
	default:
		return ""
	}
}
