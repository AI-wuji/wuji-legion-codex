package core

import (
	"crypto/sha256"
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
		if existingProof.Probe == nil || proof.Probe == nil || existingProof.Probe.Signature != proof.Probe.Signature {
			result.Decision = "hold"
			result.Comparison = fmt.Sprintf("fixture %s produced different behavior signatures", candidate.Probe.Fixture)
			result.RequiredActions = []string{"make candidate and current artifact behavior signatures equivalent", "compare a deliberately versioned fixture before replacement"}
			return result, nil
		}
		existingComparison := findProbeArtifact(existingProof.Probe.Evidence, existing.Probe.ComparisonEvidence)
		candidateComparison := findProbeArtifact(proof.Probe.Evidence, candidate.Probe.ComparisonEvidence)
		if existing.Probe.ComparisonEvidence != candidate.Probe.ComparisonEvidence || existingComparison == nil || candidateComparison == nil || existingComparison.SHA256 != candidateComparison.SHA256 {
			result.Decision = "hold"
			result.Comparison = fmt.Sprintf("fixture %s produced different verified assertion evidence", candidate.Probe.Fixture)
			result.RequiredActions = []string{"inspect the verified assertion reports", "accept a changed fixture contract explicitly before replacement"}
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
	installed := candidate
	if overlap {
		archiveDir := filepath.Join(root, "retired", candidate.ID, time.Now().UTC().Format("20060102T150405.000000000Z"))
		archive := filepath.Join(archiveDir, "manifest.json")
		if err := os.MkdirAll(filepath.Dir(archive), 0o755); err != nil {
			return result, err
		}
		if err := copyFile(target, archive); err != nil {
			return result, err
		}
		if existing.PromotionReceipt != "" {
			existingReceipt := filepath.Join(root, "capabilities", existing.ID, filepath.FromSlash(existing.PromotionReceipt))
			if err := copyFile(existingReceipt, filepath.Join(archiveDir, "promotion-receipt.json")); err != nil {
				return result, err
			}
		}
		archiveRelative, err := filepath.Rel(root, archive)
		if err != nil {
			return result, err
		}
		receiptRelative, err := persistPromotionReceipt(root, existing, existingProofFrom(result), candidate, proof, filepath.ToSlash(archiveRelative))
		if err != nil {
			return result, err
		}
		installed.Status = "primary"
		installed.PromotionReceipt = receiptRelative
	} else {
		installed.Status = "behavior-verified"
		installed.PromotionReceipt = ""
	}
	pretty, err := json.MarshalIndent(installed, "", "  ")
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

func existingProofFrom(result EvolutionResult) VerifyResult {
	if result.ExistingProof == nil {
		return VerifyResult{}
	}
	return *result.ExistingProof
}

func persistPromotionReceipt(root string, baseline Manifest, baselineProof VerifyResult, candidate Manifest, candidateProof VerifyResult, baselineManifest string) (string, error) {
	baselineSummary, err := promotionProof(baseline, baselineProof)
	if err != nil {
		return "", err
	}
	candidateSummary, err := promotionProof(candidate, candidateProof)
	if err != nil {
		return "", err
	}
	receipt := PromotionReceipt{
		SchemaVersion:    1,
		Capability:       candidate.ID,
		Decision:         "replace",
		Fixture:          candidate.Probe.Fixture,
		BaselineManifest: baselineManifest,
		Baseline:         baselineSummary,
		Candidate:        candidateSummary,
	}
	data, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		return "", err
	}
	data = append(data, '\n')
	digest := fmt.Sprintf("%x", sha256.Sum256(data))
	relative := filepath.ToSlash(filepath.Join("releases", digest+".json"))
	target := filepath.Join(root, "capabilities", candidate.ID, filepath.FromSlash(relative))
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return "", err
	}
	if err := atomicWriteFile(target, data, 0o644); err != nil {
		return "", err
	}
	return relative, nil
}

func promotionProof(manifest Manifest, proof VerifyResult) (PromotionProof, error) {
	if manifest.Probe == nil || proof.Probe == nil || rank(proof.Effective) < rank("behavior-verified") {
		return PromotionProof{}, fmt.Errorf("capability %s has no verified behavior proof", manifest.ID)
	}
	comparison := findProbeArtifact(proof.Probe.Evidence, manifest.Probe.ComparisonEvidence)
	if comparison == nil {
		return PromotionProof{}, fmt.Errorf("capability %s is missing comparison evidence %s", manifest.ID, manifest.Probe.ComparisonEvidence)
	}
	contractHash, err := manifestContractSHA256(manifest)
	if err != nil {
		return PromotionProof{}, err
	}
	return PromotionProof{
		ContractSHA256:  contractHash,
		EffectiveStatus: proof.Effective,
		Signature:       proof.Probe.Signature,
		ComparisonEvidence: PromotionArtifact{
			ID:     comparison.ID,
			SHA256: comparison.SHA256,
			Size:   comparison.Size,
		},
	}, nil
}

func findProbeArtifact(items []ProbeArtifact, id string) *ProbeArtifact {
	for index := range items {
		if items[index].ID == id {
			return &items[index]
		}
	}
	return nil
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
