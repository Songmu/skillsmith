package skillsmith

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallDirForScope(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home directory:", err)
	}

	tests := []struct {
		name    string
		scope   string
		want    string
		wantErr bool
	}{
		{name: "empty defaults to user", scope: "", want: filepath.Join(home, ".agents", "skills")},
		{name: "explicit user", scope: "user", want: filepath.Join(home, ".agents", "skills")},
		{name: "unknown scope", scope: "unknown", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := InstallDirForScope(tt.scope)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}

	t.Run("repo scope returns absolute path under repo root", func(t *testing.T) {
		// Create a fake repo and chdir into a nested directory under it so the test
		// does not depend on the real working directory containing a .git entry.
		fakeRoot := setupFakeRepo(t, false)

		nestedDir := filepath.Join(fakeRoot, "sub", "dir")
		if err := os.MkdirAll(nestedDir, 0o755); err != nil {
			t.Fatalf("failed to create nested directory: %v", err)
		}

		origWD, err := os.Getwd()
		if err != nil {
			t.Fatalf("failed to get working directory: %v", err)
		}

		if err := os.Chdir(nestedDir); err != nil {
			t.Fatalf("failed to chdir to nested directory: %v", err)
		}

		t.Cleanup(func() {
			if err := os.Chdir(origWD); err != nil {
				t.Fatalf("failed to restore working directory: %v", err)
			}
		})

		got, err := InstallDirForScope("repo")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := filepath.Join(fakeRoot, ".agents", "skills")
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

// setupFakeRepo creates a temporary directory containing a .git entry
// and returns the path. If gitFile is true, .git is created as a regular
// file (simulating a worktree); otherwise it is created as a directory.
func setupFakeRepo(t *testing.T, gitFile bool) string {
	t.Helper()
	root := t.TempDir()
	// Resolve symlinks so tests match os.Getwd() on macOS where
	// /var is a symlink to /private/var.
	root, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	gitPath := filepath.Join(root, ".git")
	if gitFile {
		if err := os.WriteFile(gitPath, []byte("gitdir: /some/other/path\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	} else {
		if err := os.Mkdir(gitPath, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func TestFindRepoRoot(t *testing.T) {
	tests := []struct {
		name    string
		gitFile bool // true = .git file (worktree), false = .git directory
		subdir  string
	}{
		{name: "root with .git dir", gitFile: false},
		{name: "root with .git file (worktree)", gitFile: true},
		{name: "subdir with .git dir", gitFile: false, subdir: "a/b/c"},
		{name: "subdir with .git file (worktree)", gitFile: true, subdir: "deep/nested"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := setupFakeRepo(t, tt.gitFile)

			startDir := root
			if tt.subdir != "" {
				startDir = filepath.Join(root, tt.subdir)
				if err := os.MkdirAll(startDir, 0o755); err != nil {
					t.Fatal(err)
				}
			}

			got, err := findRepoRoot(startDir)
			if err != nil {
				t.Fatalf("findRepoRoot(%q) error: %v", startDir, err)
			}
			if got != root {
				t.Errorf("findRepoRoot(%q) = %q, want %q", startDir, got, root)
			}
		})
	}

	t.Run("not in a git repository", func(t *testing.T) {
		dir := t.TempDir() // no .git
		_, err := findRepoRoot(dir)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "not in a git repository") {
			t.Errorf("error = %q, want it to contain %q", err.Error(), "not in a git repository")
		}
	})

	t.Run("symlink .git is rejected", func(t *testing.T) {
		root := t.TempDir()
		target := t.TempDir()
		if err := os.Symlink(target, filepath.Join(root, ".git")); err != nil {
			t.Skip("cannot create symlink:", err)
		}
		_, err := findRepoRoot(root)
		if err == nil {
			t.Fatal("expected error for symlink .git, got nil")
		}
		if !strings.Contains(err.Error(), "symlink") {
			t.Errorf("error = %q, want it to mention symlink", err.Error())
		}
	})
}
