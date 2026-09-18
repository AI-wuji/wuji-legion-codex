package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/AI-wuji/wuji-legion-codex-2.0/internal/core"
)

const expertBridgeMaxJSONBytes int64 = 1 << 20

// runExpertBridge is intentionally registration-free: main.go owns command registration.
func runExpertBridge(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "expert-bridge requires select, prepare, dispatch, or verify")
		return 2
	}
	switch args[0] {
	case "select":
		fs := flag.NewFlagSet("expert-bridge select", flag.ContinueOnError)
		fs.SetOutput(stderr)
		root := fs.String("root", ".", "repository root")
		query := fs.String("query", "", "task query")
		if fs.Parse(args[1:]) != nil || *query == "" {
			return 2
		}
		result, err := core.SelectExpert(*root, *query)
		return expertBridgeOutput(result, err, stdout, stderr)
	case "prepare":
		fs := flag.NewFlagSet("expert-bridge prepare", flag.ContinueOnError)
		fs.SetOutput(stderr)
		root := fs.String("root", ".", "repository root")
		workspace := fs.String("workspace", ".", "workspace")
		query := fs.String("query", "", "task query")
		workerPath := fs.String("worker", "", "worker JSON")
		task := fs.String("task-instance", "", "task instance")
		graph := fs.String("graph-version", "", "graph version")
		node := fs.String("execution-node", "", "execution node")
		attempt := fs.String("attempt", "", "attempt")
		if fs.Parse(args[1:]) != nil || *query == "" || *workerPath == "" {
			return 2
		}
		var worker core.WorkerTask
		if err := readExpertJSON(*workerPath, &worker); err != nil {
			fmt.Fprintln(stderr, "error:", err)
			return 2
		}
		result, err := core.PrepareExpertHandoff(*root, *workspace, *query, worker, *task, *graph, *node, *attempt)
		return expertBridgeOutput(result, err, stdout, stderr)
	case "dispatch":
		fs := flag.NewFlagSet("expert-bridge dispatch", flag.ContinueOnError)
		fs.SetOutput(stderr)
		contractPath := fs.String("contract", "", "contract JSON")
		output := fs.String("output-dir", ".wuji/dispatch", "dispatch output")
		dry := fs.Bool("dry-run", false, "prepare only")
		if fs.Parse(args[1:]) != nil || *contractPath == "" {
			return 2
		}
		var contract core.ExpertHandoffContract
		if err := readExpertJSON(*contractPath, &contract); err != nil {
			fmt.Fprintln(stderr, "error:", err)
			return 2
		}
		result, err := core.DispatchExpertHandoff(contract, core.DispatchOptions{OutputDir: *output, DryRun: *dry, Timeout: 90 * time.Second})
		return expertBridgeOutput(result, err, stdout, stderr)
	case "verify":
		fs := flag.NewFlagSet("expert-bridge verify", flag.ContinueOnError)
		fs.SetOutput(stderr)
		contractPath := fs.String("contract", "", "contract JSON")
		receiptPath := fs.String("receipt", "", "receipt JSON")
		store := fs.String("execution-store", core.DefaultExecutionGraphStore(), "execution store")
		requirements := fs.String("requirement-store", core.DefaultRequirementGraphStore(), "requirement store")
		if fs.Parse(args[1:]) != nil || *contractPath == "" || *receiptPath == "" {
			return 2
		}
		var contract core.ExpertHandoffContract
		var receipt core.ExpertExecutionReceipt
		if err := readExpertJSON(*contractPath, &contract); err != nil {
			fmt.Fprintln(stderr, "error:", err)
			return 2
		}
		if err := readExpertJSON(*receiptPath, &receipt); err != nil {
			fmt.Fprintln(stderr, "error:", err)
			return 2
		}
		result, err := core.VerifyAndRecordExpertReceipt(contract, receipt, *store, *requirements)
		return expertBridgeOutput(result, err, stdout, stderr)
	default:
		fmt.Fprintln(stderr, "unknown expert-bridge action:", args[0])
		return 2
	}
}

func readExpertJSON(path string, target any) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return err
	}
	if info.Size() > expertBridgeMaxJSONBytes {
		return fmt.Errorf("expert bridge JSON exceeds %d bytes", expertBridgeMaxJSONBytes)
	}
	data, err := io.ReadAll(io.LimitReader(file, expertBridgeMaxJSONBytes+1))
	if err != nil {
		return err
	}
	if int64(len(data)) > expertBridgeMaxJSONBytes {
		return fmt.Errorf("expert bridge JSON exceeds %d bytes", expertBridgeMaxJSONBytes)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("multiple JSON values are not allowed")
		}
		return err
	}
	return nil
}
func expertBridgeOutput(value any, err error, stdout, stderr io.Writer) int {
	if err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 1
	}
	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 1
	}
	return 0
}
