package calculate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopyFile(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "test-copyfile")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	srcPath := filepath.Join(tmpDir, "src.txt")
	dstPath := filepath.Join(tmpDir, "dst.txt")

	content := []byte("hello world")
	if err := os.WriteFile(srcPath, content, 0644); err != nil {
		t.Fatal(err)
	}

	if err := copyFile(srcPath, dstPath); err != nil {
		t.Fatal(err)
	}

	dstContent, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatal(err)
	}

	if string(dstContent) != string(content) {
		t.Errorf("expected %q, got %q", content, dstContent)
	}
}
