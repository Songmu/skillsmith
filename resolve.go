package skillsmith

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// InstallDirForScope returns the skill installation directory for the given
// scope. An empty scope defaults to "user".
//
//   - user (default): ~/.agents/skills  (absolute, under the user's home directory)
//   - repo:           .agents/skills    (relative to the current working directory;
//     the caller should ensure it is invoked from the repository root)
func InstallDirForScope(scope string) (string, error) {
	switch scope {
	case "", "user":
		return expandHome("~/.agents/skills")
	case "repo":
		return filepath.FromSlash(".agents/skills"), nil
	default:
		return "", fmt.Errorf("unknown scope %q (supported: user, repo)", scope)
	}
}

// expandHome replaces a leading "~" with the user's home directory and
// normalizes path separators for the current OS.
func expandHome(p string) (string, error) {
	if !strings.HasPrefix(p, "~") {
		return filepath.FromSlash(p), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving home directory: %w", err)
	}
	return filepath.Join(home, p[1:]), nil
}
