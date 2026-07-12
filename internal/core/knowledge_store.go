package core

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func KnowledgeWorkspaceScope(workspace string) (string, error) {
	workspace, err := normalizeWorkspacePath(workspace)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256([]byte(strings.ToLower(filepath.Clean(workspace))))
	return "workspace:" + hex.EncodeToString(hash[:]), nil
}

func normalizeKnowledgeScope(scope string) (string, error) {
	scope = strings.ToLower(strings.TrimSpace(scope))
	if scope == "global" || knowledgeWorkspaceScope.MatchString(scope) {
		return scope, nil
	}
	return "", fmt.Errorf("knowledge scope must be global or a canonical workspace identity")
}

func storeKnowledgeLocation(store, value string) (string, error) {
	location, err := normalizeKnowledgeLocation(value, false)
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(location, "https://") {
		return location, nil
	}
	data, err := readKnowledgeFileBounded(location, knowledgeMaxObjectBytes)
	if err != nil {
		return "", err
	}
	return storeKnowledgeObject(store, data)
}

func storeKnowledgeVerification(store, value string, now time.Time) (string, string, error) {
	path, err := normalizeKnowledgeLocation(value, true)
	if err != nil {
		return "", "", err
	}
	data, err := readKnowledgeFileBounded(path, knowledgeMaxObjectBytes)
	if err != nil {
		return "", "", err
	}
	var receipt knowledgeVerificationReceipt
	if err := json.Unmarshal(data, &receipt); err != nil {
		return "", "", fmt.Errorf("verification receipt is not valid JSON: %w", err)
	}
	verifiedAt, err := time.Parse(time.RFC3339, receipt.VerifiedAt)
	if err != nil {
		return "", "", fmt.Errorf("verification receipt has an invalid verified_at")
	}
	if receipt.SchemaVersion != 1 || receipt.Type != "wuji-verification-receipt" || !receipt.Passed || strings.TrimSpace(receipt.Verifier) == "" {
		return "", "", fmt.Errorf("verification receipt did not pass the trusted schema")
	}
	if verifiedAt.After(now.Add(5*time.Minute)) || now.Sub(verifiedAt) > knowledgeRetention {
		return "", "", fmt.Errorf("verification receipt is outside the accepted validity window")
	}
	ref, err := storeKnowledgeObject(store, data)
	if err != nil {
		return "", "", err
	}
	return ref, strings.TrimPrefix(ref, "object:sha256:"), nil
}

func storeKnowledgeObject(store string, data []byte) (string, error) {
	hash := sha256.Sum256(data)
	digest := hex.EncodeToString(hash[:])
	path := filepath.Join(filepath.Clean(store), "v1", "objects", digest[:2], digest)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := writeKnowledgeFile(path, data); err != nil {
			return "", err
		}
	} else if err != nil {
		return "", err
	}
	return "object:sha256:" + digest, nil
}

func knowledgeObjectPath(store, ref string) (string, error) {
	match := knowledgeObjectReference.FindStringSubmatch(strings.TrimSpace(ref))
	if len(match) != 2 {
		return "", fmt.Errorf("invalid content-addressed object reference")
	}
	return filepath.Join(filepath.Clean(store), "v1", "objects", match[1][:2], match[1]), nil
}

func hashKnowledgeObject(store, ref string) (string, error) {
	path, err := knowledgeObjectPath(store, ref)
	if err != nil {
		return "", err
	}
	data, err := readKnowledgeFileBounded(path, knowledgeMaxObjectBytes)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

func readKnowledgeFileBounded(path string, maxBytes int64) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if info.Size() > maxBytes {
		return nil, fmt.Errorf("file exceeds %d bytes: %s", maxBytes, path)
	}
	data, err := io.ReadAll(io.LimitReader(file, maxBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxBytes {
		return nil, fmt.Errorf("file exceeds %d bytes: %s", maxBytes, path)
	}
	return data, nil
}

func withKnowledgeStoreLock(store string, fn func() error) error {
	root := filepath.Join(filepath.Clean(store), "v1")
	if err := os.MkdirAll(root, 0o700); err != nil {
		return err
	}
	lockPath := filepath.Join(root, ".lock")
	deadline := time.Now().Add(knowledgeLockWait)
	for {
		file, err := os.OpenFile(lockPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if err == nil {
			_, _ = fmt.Fprintf(file, "%d\n", os.Getpid())
			_ = file.Close()
			defer os.Remove(lockPath)
			return fn()
		}
		if !os.IsExist(err) {
			return err
		}
		if info, statErr := os.Stat(lockPath); statErr == nil && time.Since(info.ModTime()) > knowledgeLockWait*4 {
			_ = os.Remove(lockPath)
			continue
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("knowledge store lock timed out")
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func repairKnowledgeStore(store string) error {
	if _, err := os.Stat(knowledgeTransactionPath(store)); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	if err := rebuildKnowledgeIndexes(store); err != nil {
		return err
	}
	if err := sweepKnowledgeObjects(store); err != nil {
		return err
	}
	return os.Remove(knowledgeTransactionPath(store))
}

func rebuildKnowledgeIndexes(store string) error {
	indexRoot := filepath.Join(filepath.Clean(store), "v1", "indexes")
	if err := os.RemoveAll(indexRoot); err != nil {
		return err
	}
	records, err := listKnowledgeRecords(store)
	if err != nil {
		return err
	}
	for _, record := range records {
		if err := validateKnowledgeRecord(store, record); err != nil {
			return err
		}
		if err := ensureKnowledgeIndexCapacity(store, record); err != nil {
			return err
		}
		for _, term := range knowledgeRecordTerms(record) {
			if err := writeKnowledgeReference(store, "terms", knowledgeScopedIndexValue(record.NormalizedScope, term), record); err != nil {
				return err
			}
		}
		for _, relation := range record.Relations {
			if err := writeKnowledgeReference(store, "relations", knowledgeScopedIndexValue(record.NormalizedScope, relation.Target), record); err != nil {
				return err
			}
			key := knowledgeRelationKey(relation.Predicate, relation.Target)
			if err := writeKnowledgeReference(store, "relations", knowledgeScopedIndexValue(record.NormalizedScope, key), record); err != nil {
				return err
			}
		}
	}
	return nil
}

func listKnowledgeRecords(store string) ([]KnowledgeRecord, error) {
	root := filepath.Join(filepath.Clean(store), "v1", "nodes")
	records := []KnowledgeRecord{}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if os.IsNotExist(walkErr) {
			return nil
		}
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			return nil
		}
		record, err := readKnowledgeRecord(path)
		if err != nil {
			return err
		}
		records = append(records, record)
		return nil
	})
	if os.IsNotExist(err) {
		return records, nil
	}
	return records, err
}

func enforceKnowledgeStoreCapacity(store string, now time.Time, maxNodes int, maxBytes int64) (KnowledgeCapacity, error) {
	records, err := listKnowledgeRecords(store)
	if err != nil {
		return KnowledgeCapacity{}, err
	}
	sort.Slice(records, func(i, j int) bool { return records[i].UpdatedAt < records[j].UpdatedAt })
	evicted := 0
	cutoff := now.Add(-knowledgeRetention)
	for _, record := range records {
		updated, parseErr := time.Parse(time.RFC3339Nano, record.UpdatedAt)
		if parseErr == nil && updated.Before(cutoff) {
			if err := os.Remove(knowledgeRecordPath(store, record.Kind, record.ID)); err != nil && !os.IsNotExist(err) {
				return KnowledgeCapacity{}, err
			}
			evicted++
		}
	}
	if evicted > 0 {
		if err := rebuildKnowledgeIndexes(store); err != nil {
			return KnowledgeCapacity{}, err
		}
		if err := sweepKnowledgeObjects(store); err != nil {
			return KnowledgeCapacity{}, err
		}
		records, err = listKnowledgeRecords(store)
		if err != nil {
			return KnowledgeCapacity{}, err
		}
		sort.Slice(records, func(i, j int) bool { return records[i].UpdatedAt < records[j].UpdatedAt })
	}
	for {
		capacity, err := knowledgeStoreCapacity(store)
		if err != nil {
			return KnowledgeCapacity{}, err
		}
		if len(records) <= maxNodes && capacity.StoreBytes <= maxBytes {
			capacity.EvictedRecords = evicted
			return capacity, nil
		}
		if len(records) == 0 {
			return KnowledgeCapacity{}, fmt.Errorf("knowledge store cannot satisfy its hard byte quota")
		}
		oldest := records[0]
		records = records[1:]
		if err := os.Remove(knowledgeRecordPath(store, oldest.Kind, oldest.ID)); err != nil && !os.IsNotExist(err) {
			return KnowledgeCapacity{}, err
		}
		evicted++
		if err := rebuildKnowledgeIndexes(store); err != nil {
			return KnowledgeCapacity{}, err
		}
		if err := sweepKnowledgeObjects(store); err != nil {
			return KnowledgeCapacity{}, err
		}
	}
}

func knowledgeStoreCapacity(store string) (KnowledgeCapacity, error) {
	capacity := KnowledgeCapacity{MaxNodes: knowledgeMaxNodes, MaxStoreBytes: knowledgeMaxStoreBytes, RetentionDays: int(knowledgeRetention / (24 * time.Hour))}
	root := filepath.Join(filepath.Clean(store), "v1")
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if os.IsNotExist(walkErr) {
			return nil
		}
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		capacity.StoreBytes += info.Size()
		if strings.Contains(filepath.ToSlash(path), "/nodes/") && strings.HasSuffix(entry.Name(), ".json") {
			capacity.NodeCount++
		}
		return nil
	})
	if os.IsNotExist(err) {
		return capacity, nil
	}
	return capacity, err
}

func sweepKnowledgeObjects(store string) error {
	records, err := listKnowledgeRecords(store)
	if err != nil {
		return err
	}
	referenced := map[string]bool{}
	for _, record := range records {
		for _, ref := range []string{record.Location, record.Verification} {
			if match := knowledgeObjectReference.FindStringSubmatch(ref); len(match) == 2 {
				referenced[match[1]] = true
			}
		}
	}
	root := filepath.Join(filepath.Clean(store), "v1", "objects")
	err = filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if os.IsNotExist(walkErr) {
			return nil
		}
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if !referenced[entry.Name()] {
			return os.Remove(path)
		}
		return nil
	})
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
