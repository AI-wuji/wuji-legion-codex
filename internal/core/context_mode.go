package core

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	ContextModeMaxInputBytes  = 1024 * 1024
	ContextModeMaxResultBytes = 4096
	ContextModeTimeoutMS      = 10000
)

type ContextModeRequest struct {
	Operation  string `json:"operation"`
	Field      string `json:"field,omitempty"`
	Value      string `json:"value,omitempty"`
	MaxMatches int    `json:"max_matches,omitempty"`
}

type ContextModeToolCall struct {
	SchemaVersion int                    `json:"schema_version"`
	PreparedOnly  bool                   `json:"prepared_only"`
	Tool          string                 `json:"tool"`
	Arguments     map[string]interface{} `json:"arguments"`
	Workspace     string                 `json:"workspace"`
	InputPath     string                 `json:"input_path"`
	InputBytes    int                    `json:"input_bytes"`
	InputSHA256   string                 `json:"input_sha256"`
	Operation     string                 `json:"operation"`
	Request       ContextModeRequest     `json:"request"`
	ResultCap     int                    `json:"result_cap_bytes"`
	Limitations   []string               `json:"limitations"`
}

type ContextModeResult struct {
	SchemaVersion int           `json:"schema_version"`
	InputBytes    int           `json:"input_bytes"`
	InputSHA256   string        `json:"input_sha256"`
	Operation     string        `json:"operation"`
	LineCount     int           `json:"line_count"`
	InvalidLines  int           `json:"invalid_lines"`
	MatchCount    int           `json:"match_count"`
	Truncated     bool          `json:"truncated"`
	Matches       []interface{} `json:"matches,omitempty"`
}

// PrepareContextModeExecute prepares, but does not execute, one bounded ctx_execute call.
func PrepareContextModeExecute(workspace, inputPath string, request ContextModeRequest) (ContextModeToolCall, error) {
	workspace, err := canonicalDirectory(workspace)
	if err != nil {
		return ContextModeToolCall{}, err
	}
	inputPath, content, err := readCanonicalWorkspaceFile(workspace, inputPath)
	if err != nil {
		return ContextModeToolCall{}, err
	}
	if len(content) > ContextModeMaxInputBytes {
		return ContextModeToolCall{}, fmt.Errorf("context-mode input exceeds %d bytes", ContextModeMaxInputBytes)
	}
	request.Operation = strings.TrimSpace(request.Operation)
	if request.MaxMatches == 0 {
		request.MaxMatches = 20
	}
	if request.MaxMatches < 1 || request.MaxMatches > 100 {
		return ContextModeToolCall{}, fmt.Errorf("max-matches must be between 1 and 100")
	}
	switch request.Operation {
	case "jsonl-count":
		if request.Field != "" || request.Value != "" {
			return ContextModeToolCall{}, fmt.Errorf("jsonl-count does not accept field or value")
		}
	case "jsonl-filter-eq":
		if !validContextModeField(request.Field) {
			return ContextModeToolCall{}, fmt.Errorf("filter field is invalid")
		}
	case "jsonl-extract":
		if !validContextModeField(request.Field) || request.Value != "" {
			return ContextModeToolCall{}, fmt.Errorf("extract requires a valid field and no value")
		}
	default:
		return ContextModeToolCall{}, fmt.Errorf("unsupported context-mode operation %q", request.Operation)
	}
	digest := sha256.Sum256(content)
	hash := hex.EncodeToString(digest[:])
	code, err := contextModeJavaScript(inputPath, content, hash, request)
	if err != nil {
		return ContextModeToolCall{}, err
	}
	relative, _ := filepath.Rel(workspace, inputPath)
	return ContextModeToolCall{
		SchemaVersion: 1, PreparedOnly: true, Tool: "ctx_execute",
		Arguments: map[string]interface{}{"language": "javascript", "code": code, "timeout": ContextModeTimeoutMS},
		Workspace: workspace, InputPath: filepath.ToSlash(relative), InputBytes: len(content), InputSHA256: hash,
		Operation: request.Operation, ResultCap: ContextModeMaxResultBytes,
		Request:     request,
		Limitations: []string{"host must invoke ctx_execute", "no index/search/recall", "no shell/network/background/intent", "prepared call is not execution evidence"},
	}, nil
}

func ValidateContextModeResult(call ContextModeToolCall, raw []byte) (ContextModeResult, error) {
	if len(raw) == 0 || len(raw) > call.ResultCap || call.ResultCap != ContextModeMaxResultBytes {
		return ContextModeResult{}, fmt.Errorf("context-mode result size is outside the prepared contract")
	}
	var result ContextModeResult
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.UseNumber()
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&result); err != nil {
		return ContextModeResult{}, fmt.Errorf("decode context-mode result: %w", err)
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		return ContextModeResult{}, fmt.Errorf("context-mode result contains trailing content")
	}
	if result.SchemaVersion != 1 || result.InputBytes != call.InputBytes || result.InputSHA256 != call.InputSHA256 || result.Operation != call.Operation {
		return ContextModeResult{}, fmt.Errorf("context-mode result does not match prepared input")
	}
	_, content, err := readCanonicalWorkspaceFile(call.Workspace, call.InputPath)
	if err != nil {
		return ContextModeResult{}, fmt.Errorf("re-read prepared input: %w", err)
	}
	digest := sha256.Sum256(content)
	if len(content) != call.InputBytes || hex.EncodeToString(digest[:]) != call.InputSHA256 {
		return ContextModeResult{}, fmt.Errorf("prepared context-mode input changed")
	}
	if result.LineCount < 0 || result.InvalidLines < 0 || result.MatchCount < 0 || result.InvalidLines > result.LineCount {
		return ContextModeResult{}, fmt.Errorf("context-mode result contains invalid counts")
	}
	expectedLines, expectedInvalid, expectedMatches, expectedValues, err := summarizeContextModeInput(content, call.Request)
	if err != nil {
		return ContextModeResult{}, err
	}
	if result.LineCount != expectedLines || result.InvalidLines != expectedInvalid || result.MatchCount != expectedMatches {
		return ContextModeResult{}, fmt.Errorf("context-mode result summary does not match independent validation")
	}
	if call.Operation != "jsonl-count" {
		for len(expectedValues) > 0 && contextModeRenderedBytes(call, expectedLines, expectedInvalid, expectedMatches, expectedValues) > ContextModeMaxResultBytes {
			expectedValues = expectedValues[:len(expectedValues)-1]
		}
	}
	expectedTruncated := call.Operation != "jsonl-count" && expectedMatches > len(expectedValues)
	if !sameContextModeMatches(result.Matches, expectedValues) || result.Truncated != expectedTruncated {
		return ContextModeResult{}, fmt.Errorf("context-mode result matches do not match independent validation")
	}
	return result, nil
}

func contextModeRenderedBytes(call ContextModeToolCall, lines, invalid, matches int, values []interface{}) int {
	result := map[string]interface{}{
		"schema_version": 1, "input_bytes": call.InputBytes, "input_sha256": call.InputSHA256,
		"operation": call.Operation, "line_count": lines, "invalid_lines": invalid,
		"match_count": matches, "truncated": matches > len(values), "matches": values,
	}
	var encoded bytes.Buffer
	encoder := json.NewEncoder(&encoded)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(result)
	return encoded.Len() - 1
}

func summarizeContextModeInput(content []byte, request ContextModeRequest) (lines, invalid, matches int, values []interface{}, err error) {
	scanner := bufio.NewScanner(bytes.NewReader(content))
	scanner.Buffer(make([]byte, 64*1024), ContextModeMaxInputBytes)
	for scanner.Scan() {
		lines++
		var row interface{}
		decoder := json.NewDecoder(strings.NewReader(scanner.Text()))
		decoder.UseNumber()
		if decodeErr := decoder.Decode(&row); decodeErr != nil || decoder.Decode(&struct{}{}) != io.EOF {
			invalid++
			continue
		}
		object, ok := row.(map[string]interface{})
		if !ok || object == nil {
			invalid++
			continue
		}
		switch request.Operation {
		case "jsonl-count":
			matches++
		case "jsonl-filter-eq":
			value, exists := object[request.Field]
			if scalar, ok := contextModeJSString(value); exists && ok && scalar == request.Value {
				matches++
				if len(values) < request.MaxMatches {
					values = append(values, json.Number(fmt.Sprint(lines)))
				}
			}
		case "jsonl-extract":
			if value, exists := object[request.Field]; exists {
				matches++
				if rendered, supported := contextModeJSString(value); supported && len(values) < request.MaxMatches {
					values = append(values, truncateRunes(rendered, 128))
				}
			}
		}
	}
	if scanErr := scanner.Err(); scanErr != nil {
		return 0, 0, 0, nil, fmt.Errorf("independently scan context-mode input: %w", scanErr)
	}
	return lines, invalid, matches, values, nil
}

func contextModeJSString(value interface{}) (string, bool) {
	switch typed := value.(type) {
	case string:
		return typed, true
	case json.Number:
		number, err := strconv.ParseFloat(typed.String(), 64)
		if err != nil {
			return "", false
		}
		if number == 0 {
			return "0", true
		}
		encoded, err := json.Marshal(number)
		return string(encoded), err == nil
	case bool:
		if typed {
			return "true", true
		}
		return "false", true
	case nil:
		return "null", true
	default:
		return "", false
	}
}

func truncateRunes(value string, limit int) string {
	runes := []rune(value)
	if len(runes) > limit {
		runes = runes[:limit]
	}
	return string(runes)
}

func sameContextModeMatches(actual, expected []interface{}) bool {
	if len(actual) == 0 && len(expected) == 0 {
		return true
	}
	left, leftErr := json.Marshal(actual)
	right, rightErr := json.Marshal(expected)
	return leftErr == nil && rightErr == nil && bytes.Equal(left, right)
}

func canonicalDirectory(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", fmt.Errorf("resolve workspace: %w", err)
	}
	info, err := os.Stat(resolved)
	if err != nil || !info.IsDir() {
		return "", fmt.Errorf("workspace must be an existing directory")
	}
	return filepath.Clean(resolved), nil
}

func readCanonicalWorkspaceFile(workspace, path string) (string, []byte, error) {
	if !filepath.IsAbs(path) {
		path = filepath.Join(workspace, path)
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", nil, err
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", nil, fmt.Errorf("resolve input: %w", err)
	}
	rel, err := filepath.Rel(workspace, resolved)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", nil, fmt.Errorf("context-mode input escapes canonical workspace")
	}
	info, err := os.Stat(resolved)
	if err != nil || !info.Mode().IsRegular() {
		return "", nil, fmt.Errorf("context-mode input must be a regular file")
	}
	if info.Size() > ContextModeMaxInputBytes {
		return "", nil, fmt.Errorf("context-mode input exceeds %d bytes", ContextModeMaxInputBytes)
	}
	content, err := os.ReadFile(resolved)
	return resolved, content, err
}

func validContextModeField(field string) bool {
	if field == "" || utf8.RuneCountInString(field) > 64 {
		return false
	}
	for _, r := range field {
		if !(r == '_' || r == '-' || unicode.IsLetter(r) || unicode.IsDigit(r)) {
			return false
		}
	}
	return true
}

func contextModeJavaScript(path string, content []byte, hash string, request ContextModeRequest) (string, error) {
	config, err := json.Marshal(request)
	if err != nil {
		return "", err
	}
	// The program has fixed behavior and reads only the already-canonicalized selected file.
	return fmt.Sprintf(`const fs = require("fs"), crypto = require("crypto");
const inputPath = %q, expectedBytes = %d, expectedSHA = %q;
const stat = fs.statSync(inputPath); if (!stat.isFile() || stat.size > %d || stat.size !== expectedBytes) throw new Error("input size changed");
const raw = fs.readFileSync(inputPath); if (raw.length !== expectedBytes || crypto.createHash("sha256").update(raw).digest("hex") !== expectedSHA) throw new Error("input hash changed");
const input = raw.toString("utf8");
const cfg = %s;
const lines = input.split(/\r?\n/); if (lines.length && lines[lines.length-1] === "") lines.pop();
const scalarString = value => value === null ? "null" : (["string","number","boolean"].includes(typeof value) ? String(value) : null);
let invalid = 0, count = 0, matches = [];
for (let i = 0; i < lines.length; i++) { let row; try { row = JSON.parse(lines[i]); } catch (_) { invalid++; continue; }
  if (row === null || Array.isArray(row) || typeof row !== "object") { invalid++; continue; }
  if (cfg.operation === "jsonl-count") count++;
  else if (cfg.operation === "jsonl-filter-eq" && scalarString(row[cfg.field]) === cfg.value) { count++; if (matches.length < cfg.max_matches) matches.push(i + 1); }
  else if (cfg.operation === "jsonl-extract" && Object.prototype.hasOwnProperty.call(row, cfg.field)) { count++; const value = scalarString(row[cfg.field]); if (value !== null && matches.length < cfg.max_matches) matches.push(Array.from(value).slice(0, 128).join("")); }
}
const out = {schema_version:1,input_bytes:expectedBytes,input_sha256:expectedSHA,operation:cfg.operation,line_count:lines.length,invalid_lines:invalid,match_count:count,truncated:cfg.operation === "jsonl-count" ? false : count>matches.length};
if (cfg.operation !== "jsonl-count") out.matches = matches;
while (Buffer.byteLength(JSON.stringify(out), "utf8") > %d && out.matches.length) { out.matches.pop(); out.truncated = true; }
const rendered = JSON.stringify(out); if (Buffer.byteLength(rendered, "utf8") > %d) throw new Error("result cap exceeded"); console.log(rendered);`, filepath.ToSlash(path), len(content), hash, ContextModeMaxInputBytes, config, ContextModeMaxResultBytes, ContextModeMaxResultBytes), nil
}
