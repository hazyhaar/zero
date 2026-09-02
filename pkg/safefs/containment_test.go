package safefs

import (
	"math"
	"os"
	"path/filepath"
	"testing"
)

func TestInBoundsNonWrapping(t *testing.T) {
	if !InBoundsNonWrapping(10, 20, 100) {
		t.Error("expected true for valid bounds")
	}
	if InBoundsNonWrapping(90, 20, 100) {
		t.Error("expected false for out of bounds")
	}
	// Test du wrap around avec UINT64_MAX
	if InBoundsNonWrapping(math.MaxUint64-5, 10, 100) {
		t.Error("expected false on wrap around attempt")
	}
}

func TestInRootBounds(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub", "file.txt")
	outside := filepath.Join(dir, "..", "outside.txt")

	if _, err := InRootBounds(dir, sub); err != nil {
		t.Errorf("expected inside root: %v", err)
	}
	if _, err := InRootBounds(dir, outside); err == nil {
		t.Error("expected error for escaping path")
	}
}

func TestWriteFileAtomic(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "test.txt")
	content := []byte("hello atomic world")

	if err := WriteFileAtomic(target, content, 0o644); err != nil {
		t.Fatalf("WriteFileAtomic failed: %v", err)
	}

	read, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if string(read) != string(content) {
		t.Fatalf("got %q; want %q", read, content)
	}
}
