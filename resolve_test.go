package skillsmith

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultInstallDir(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home directory:", err)
	}

	got, err := DefaultInstallDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(home, ".agents", "skills")
	if got != want {
		t.Errorf("DefaultInstallDir() = %q, want %q", got, want)
	}
}
