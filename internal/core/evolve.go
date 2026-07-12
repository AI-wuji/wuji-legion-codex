package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func EvaluateCandidate(root, candidatePath string, apply bool) (EvolutionResult, error) {
	data, err := os.ReadFile(candidatePath)
	if err != nil {
		return EvolutionResult{}, err
	}
	candidate, err := decodeManifest(data)
	if err != nil {
		return EvolutionResult{}, err
	}
	proof := Verify(root, candidate)
	result := EvolutionResult{Candidate: candidate.ID, CandidateProof: proof, Decision: "reject"}
	items, err := LoadManifests(root)
	if err != nil {
		return result, err
	}
	existing, overlap := FindManifest(items, candidate.ID)
	if overlap {
		result.ExistingStatus = existing.Status
	}
	if !proof.Passed {
		result.RequiredActions = []string{"retain as known/doctrine-only only", "repair package evidence before admission"}
		return result, nil
	}
	if rank(proof.Effective) < rank("behavior-verified") {
		result.Decision = "hold"
		result.RequiredActions = []string{"add a real artifact behavior probe", "do not describe this candidate as fused"}
		return result, nil
	}

	if overlap {
		existingProof := Verify(root, existing)
		result.ExistingProof = &existingProof
		if !existingProof.Passed {
			result.Decision = "hold"
			result.RequiredActions = []string{"repair the current capability baseline", "rerun current and candidate verification before replacement"}
			return result, nil
		}
		if existing.Probe == nil || candidate.Probe == nil || existing.Probe.Fixture == "" || existing.Probe.Fixture != candidate.Probe.Fixture {
			result.Decision = "hold"
			result.RequiredActions = []string{"declare the same representative fixture on current and candidate probes", "compare only equivalent behavior evidence"}
			return result, nil
		}
		if rank(proof.Effective) < rank(existingProof.Effective) {
			result.Decision = "reject"
			result.Comparison = fmt.Sprintf("fixture %s: candidate %s is below existing %s", candidate.Probe.Fixture, proof.Effective, existingProof.Effective)
			result.RequiredActions = []string{"retain the current capability", "improve the candidate without regressing required behavior"}
			return result, nil
		}
		result.Decision = "replace"
		result.Comparison = fmt.Sprintf("fixture %s passed on both routes: candidate %s is no worse than existing %s", candidate.Probe.Fixture, proof.Effective, existingProof.Effective)
	} else {
		result.Decision = "admit"
	}

	if !apply {
		return result, nil
	}
	targetDir := filepath.Join(root, "capabilities", candidate.ID)
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return result, err
	}
	target := filepath.Join(targetDir, "manifest.json")
	if overlap {
		archive := filepath.Join(root, "retired", candidate.ID, time.Now().UTC().Format("20060102T150405.000000000Z"), "manifest.json")
		if err := os.MkdirAll(filepath.Dir(archive), 0o755); err != nil {
			return result, err
		}
		if err := copyFile(target, archive); err != nil {
			return result, err
		}
	}
	pretty, err := json.MarshalIndent(candidate, "", "  ")
	if err != nil {
		return result, err
	}
	pretty = append(pretty, '\n')
	if err := atomicWriteFile(target, pretty, 0o644); err != nil {
		return result, err
	}
	result.Applied = true
	return result, nil
}

func copyFile(source, target string) error {
	data, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	if err := atomicWriteFile(target, data, 0o644); err != nil {
		return fmt.Errorf("archive %s: %w", target, err)
	}
	return nil
}

func atomicWriteFile(target string, data []byte, mode os.FileMode) error {
	temp, err := os.CreateTemp(filepath.Dir(target), ".manifest-*.tmp")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)
	if err := temp.Chmod(mode); err != nil {
		temp.Close()
		return err
	}
	if _, err := temp.Write(data); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Sync(); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tempPath, target); err != nil {
		return fmt.Errorf("replace %s atomically: %w", target, err)
	}
	return nil
}
