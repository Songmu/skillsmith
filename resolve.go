package skillsmith

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// FindRepoRoot traverses parent directories from the current working directory
// to find the repository root (the directory containing a .git entry).
//
// It uses os.Lstat to avoid following symlinks: .git must be a directory
// (normal repo) or a regular file (git worktree). A symlink named .git is
// ignored for security.
//
// Returns an error if the filesystem root is reached without finding .git.
func FindRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getting current directory: %w", err)
	}

	for {
		fi, err := os.Lstat(filepath.Join(dir, ".git"))
		if err == nil {
			mode := fi.Mode()
			if mode.IsDir() || mode.IsRegular() {
				return dir, nil
			}
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("checking .git in %s: %w", dir, err)
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("not in a git repository")
		}
		dir = parent
	}
}

// InstallDirForScope returns the skill installation directory for the given
// scope. An empty scope defaults to "user".
//
//   - user (default): ~/.agents/skills  (absolute, under the user's home directory)
//   - repo:           <repo-root>/.agents/skills  (absolute, under the repository root
//     found by traversing parent directories for .git)
func InstallDirForScope(scope string) (string, error) {
	switch scope {
	case "", "user":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolving home directory: %w", err)
		}
		return filepath.Join(home, ".agents", "skills"), nil
	case "repo":
		root, err := FindRepoRoot()
		if err != nil {
			return "", err
		}
		return filepath.Join(root, ".agents", "skills"), nil
	default:
		return "", fmt.Errorf("unknown scope %q (supported: user, repo)", scope)
	}
}
