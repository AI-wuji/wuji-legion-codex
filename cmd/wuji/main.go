package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/AI-wuji/wuji-legion-codex-2.0/internal/core"
)

const usage = `usage: wuji <route|context-select|verify|evolve> [flags]

Commands:
  route           select a capability for a user request
  context-select  select ranked code excerpts within a byte budget
  verify          verify one or all capability manifests
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
	switch args[0] {
	case "route":
		fs := newFlagSet("route", stderr)
		query := fs.String("query", "", "user request")
		rootFlag := fs.String("root", root, "Wuji 2.0 root")
		contextArtifact := fs.String("context-artifact", "", "verified context artifact for bounded delegation")
		parentContextRequired := fs.Bool("parent-context-required", false, "keep execution on Aji because parent context must be replayed")
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
		delegationContext := core.DelegationContext{ParentContextRequired: *parentContextRequired}
		if strings.TrimSpace(*contextArtifact) != "" {
			delegationContext, err = core.LoadContextArtifact(*contextArtifact)
			if err != nil {
				return reportError(stderr, 2, err)
			}
			delegationContext.ParentContextRequired = *parentContextRequired
		}
		output = core.RouteWithContext(*query, items, delegationContext)

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
	return 0
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
