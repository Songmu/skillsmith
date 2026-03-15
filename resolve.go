package skillsmith

import (
	"fmt"
	"os"
	"path/filepath"
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
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolving home directory: %w", err)
		}
		return filepath.Join(home, ".agents", "skills"), nil
	case "repo":
		return filepath.Join(".agents", "skills"), nil
	default:
		return "", fmt.Errorf("unknown scope %q (supported: user, repo)", scope)
	}
}
