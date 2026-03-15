package skillsmith

import (
	"os"
	"path/filepath"
	"strings"
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
	if !filepath.IsAbs(got) {
		t.Errorf("InstallDirForScope(\"repo\") = %q, want absolute path", got)
	}
	if !strings.HasSuffix(got, filepath.Join(".agents", "skills")) {
		t.Errorf("InstallDirForScope(\"repo\") = %q, want path ending with .agents/skills", got)
	}

	root, err := findRepoRoot()
	if err != nil {
		t.Fatalf("findRepoRoot error: %v", err)
	}
	want := filepath.Join(root, ".agents", "skills")
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

func TestFindRepoRoot_FromRepoDir(t *testing.T) {
	root, err := findRepoRoot()
	if err != nil {
		t.Fatalf("findRepoRoot() error: %v", err)
	}
	if !filepath.IsAbs(root) {
		t.Errorf("findRepoRoot() = %q, want absolute path", root)
	}
	fi, err := os.Lstat(filepath.Join(root, ".git"))
	if err != nil {
		t.Fatalf("no .git at reported root %q: %v", root, err)
	}
	if !fi.IsDir() && !fi.Mode().IsRegular() {
		t.Errorf(".git at %q is neither a directory nor a regular file", root)
	}
}

func TestFindRepoRoot_FromSubdir(t *testing.T) {
	// Change to a subdirectory of the repo and verify we still find the root.
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(orig); err != nil {
			t.Errorf("restoring working directory: %v", err)
		}
	})

	subdir := filepath.Join(orig, "agentskill")
	if err := os.Chdir(subdir); err != nil {
		t.Skipf("cannot chdir to agentskill subdir: %v", err)
	}

	root, err := findRepoRoot()
	if err != nil {
		t.Fatalf("findRepoRoot() from subdir error: %v", err)
	}

	// The root found from subdir should match the root found from the original dir.
	if err := os.Chdir(orig); err != nil {
		t.Fatal(err)
	}
	expected, err := findRepoRoot()
	if err != nil {
		t.Fatal(err)
	}
	if root != expected {
		t.Errorf("findRepoRoot() from subdir = %q, want %q", root, expected)
	}
}
