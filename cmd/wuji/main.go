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

const usage = `usage: wuji <route|orchestrate|change-capsule|context-select|graph-sync|knowledge-record|knowledge-query|dispatch|validate-receipt|verify|source-audit|evolve> [flags]

Commands:
  route           select a capability for a user request
  orchestrate     prepare ordered native-worker contracts for Aji/host merge
  change-capsule  create a bounded high-risk change contract
  context-select  select ranked code excerpts within a byte budget
  graph-sync      build or refresh the local workspace relation graph
  knowledge-record record a verified cross-project knowledge node
  knowledge-query  query the event-triggered cross-project knowledge graph
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
	exitCode := 0
	switch args[0] {
	case "route":
		fs := newFlagSet("route", stderr)
		query := fs.String("query", "", "user request")
		rootFlag := fs.String("root", root, "Wuji 2.0 root")
		contextArtifact := fs.String("context-artifact", "", "verified context artifact for bounded delegation")
		parentContextRequired := fs.Bool("parent-context-required", false, "keep execution on Aji because parent context must be replayed")
		selfContainedHandoff := fs.Bool("self-contained-handoff", false, "confirm the compact worker contract contains all required task context")
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
			delegationContext.ParentContextRequired = *parentContextRequired
			delegationContext.SelfContained = *selfContainedHandoff
		}
		output = core.RouteWithContext(*query, items, delegationContext)

	case "orchestrate":
		fs := newFlagSet("orchestrate", stderr)
		query := fs.String("query", "", "user request")
		rootFlag := fs.String("root", root, "Wuji 2.0 root")
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
			delegationContext.ParentContextRequired = *parentContextRequired
			delegationContext.SelfContained = *selfContainedHandoff
		}
		initial := core.RouteWithContext(*query, items, delegationContext)
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
		rootFlag := fs.String("root", root, "Wuji 2.0 root")
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
		if code := parseFlags(fs, args[1:], stderr); code >= 0 {
			return code
		}
		scopeValue, err := resolveKnowledgeScope(*scope, *workspace)
		if err != nil {
			return reportError(stderr, 2, err)
		}
		record, err := core.RecordKnowledge(*store, core.KnowledgeRecordInput{
			Kind: *kind, Key: *key, Scope: scopeValue, Summary: *summary, RootCause: *rootCause,
			Location: *location, Verification: *verification, Tags: splitCSV(*tags), Relations: parseKnowledgeRelations(*relations),
		})
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

	case "dispatch":
		fs := newFlagSet("dispatch", stderr)
		routePath := fs.String("route", "", "route JSON file")
		rootFlag := fs.String("root", root, "Wuji 2.0 root used to validate selected sources")
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
		rootFlag := fs.String("root", root, "Wuji 2.0 root")
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
		rootFlag := fs.String("root", root, "Wuji 2.0 root")
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
