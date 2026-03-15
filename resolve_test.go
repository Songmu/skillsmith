package skillsmith

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstallDirForScope_User(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home directory:", err)
	}

	tests := []struct {
		scope string
		want  string
	}{
		{"", filepath.Join(home, ".agents", "skills")},
		{"user", filepath.Join(home, ".agents", "skills")},
	}
	for _, tt := range tests {
		got, err := InstallDirForScope(tt.scope)
		if err != nil {
			t.Errorf("InstallDirForScope(%q) error: %v", tt.scope, err)
			continue
		}
		if got != tt.want {
			t.Errorf("InstallDirForScope(%q) = %q, want %q", tt.scope, got, tt.want)
		}
	}
}

func TestInstallDirForScope_Repo(t *testing.T) {
	got, err := InstallDirForScope("repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(".agents", "skills")
	if got != want {
		t.Errorf("InstallDirForScope(\"repo\") = %q, want %q", got, want)
	}
}

func TestInstallDirForScope_UnknownScope(t *testing.T) {
	_, err := InstallDirForScope("unknown")
	if err == nil {
		t.Error("expected error for unknown scope, got nil")
	}
}
