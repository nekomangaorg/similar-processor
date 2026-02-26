package calculate

import (
	"os"
	"testing"
)

func TestWriteLineToDebugFilePermissions(t *testing.T) {
	// Clean up before test
	_ = os.RemoveAll("debug")
	defer os.RemoveAll("debug")

	fileName := "test_security_debug"
	line := "test line"

	WriteLineToDebugFile(fileName, line)

	// Check directory permissions
	info, err := os.Stat("debug")
	if err != nil {
		t.Fatalf("Failed to stat debug directory: %v", err)
	}
	mode := info.Mode().Perm()

	// We want to ensure group and others have NO permissions (0700)
	// If any bit in 0077 is set, fail.
	if mode&0077 != 0 {
		t.Errorf("Debug directory has insecure permissions: %o (expected rwx------)", mode)
	}

	// Check file permissions
	info, err = os.Stat("debug/" + fileName + ".txt")
	if err != nil {
		t.Fatalf("Failed to stat debug file: %v", err)
	}
	mode = info.Mode().Perm()
	// We want 0600. So check 0077.
if mode != 0600 {
	t.Errorf("Debug file has incorrect permissions: %o (expected rw-------)", mode)
}
	}
}
