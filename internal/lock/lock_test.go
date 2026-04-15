package lock

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAcquireLock_EdgeCases(t *testing.T) {
	dir := t.TempDir()

	// 1. First Lock
	l1, err := Acquire(dir)
	if err != nil {
		t.Fatalf("failed to acquire initial lock: %v", err)
	}

	// 2. Edge Case: Double Lock (Simulating second instance)
	l2, err := Acquire(dir)
	if err == nil {
		_ = l2.Release()
		t.Fatal("expected error when acquiring already locked directory, got nil")
	}

	// 3. Edge Case: Release and Re-acquire
	err = l1.Release()
	if err != nil {
		t.Fatalf("failed to release lock: %v", err)
	}

	l3, err := Acquire(dir)
	if err != nil {
		t.Fatalf("failed to re-acquire lock after release: %v", err)
	}
	_ = l3.Release()

	// 4. Edge Case: Missing Directory Perms
	badDir := filepath.Join(dir, "readonly")
	_ = os.Mkdir(badDir, 0400) // Read-only
	_, err = Acquire(badDir)
	if err == nil {
		t.Fatal("expected error acquiring lock in read-only dir")
	}
}
