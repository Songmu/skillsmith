package skillsmith

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveInstallDir_Defaults(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home directory:", err)
	}

	got, err := ResolveInstallDir("", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(home, ".claude", "skills")
	if got != want {
		t.Errorf("ResolveInstallDir(\"\",\"\") = %q, want %q", got, want)
	}
}

func TestResolveInstallDir_AllMappings(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home directory:", err)
	}

	tests := []struct {
		agent, scope string
		want         string
	}{
		{"codex", "user", filepath.Join(home, ".codex", "skills")},
		{"codex", "repo", filepath.Join(".agents", "skills")},
		{"claude", "user", filepath.Join(home, ".claude", "skills")},
		{"claude", "repo", filepath.Join(".claude", "skills")},
		{"agents", "user", filepath.Join(home, ".agents", "skills")},
		{"agents", "repo", filepath.Join(".agents", "skills")},
	}
	for _, tt := range tests {
		got, err := ResolveInstallDir(tt.agent, tt.scope)
		if err != nil {
			t.Errorf("ResolveInstallDir(%q,%q) error: %v", tt.agent, tt.scope, err)
			continue
		}
		if got != tt.want {
			t.Errorf("ResolveInstallDir(%q,%q) = %q, want %q", tt.agent, tt.scope, got, tt.want)
		}
	}
}

func TestResolveInstallDir_UnknownAgent(t *testing.T) {
	_, err := ResolveInstallDir("unknown", "user")
	if err == nil {
		t.Error("expected error for unknown agent, got nil")
	}
}

func TestResolveInstallDir_UnknownScope(t *testing.T) {
	_, err := ResolveInstallDir("claude", "unknown")
	if err == nil {
		t.Error("expected error for unknown scope, got nil")
	}
}
