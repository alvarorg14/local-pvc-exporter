package diskusage

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMeasure(t *testing.T) {
	root := t.TempDir()

	if err := os.WriteFile(filepath.Join(root, "a.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	sub := filepath.Join(root, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "b.txt"), []byte("world!"), 0o644); err != nil {
		t.Fatal(err)
	}

	w := New(30 * time.Second)
	result, err := w.Measure(context.Background(), root)
	if err != nil {
		t.Fatalf("Measure() error: %v", err)
	}

	if result.UsedBytes <= 0 {
		t.Errorf("expected positive UsedBytes, got %d", result.UsedBytes)
	}
	if result.InodesUsed < 3 {
		t.Errorf("expected at least 3 inodes (root dir, sub dir, 2 files), got %d", result.InodesUsed)
	}
	if result.InodesTotal <= 0 {
		t.Errorf("expected positive InodesTotal, got %d", result.InodesTotal)
	}
	if result.InodesFree <= 0 {
		t.Errorf("expected positive InodesFree, got %d", result.InodesFree)
	}
}

func TestMeasureNotDirectory(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "file.txt")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	w := New(time.Second)
	_, err := w.Measure(context.Background(), file)
	if err == nil {
		t.Fatal("expected error for non-directory root")
	}
}
