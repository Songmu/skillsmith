package skillsmith

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DefaultInstallDir returns the default skill installation directory
// (~/.agents/skills), which is the cross-client standard for agent skills.
func DefaultInstallDir() (string, error) {
	return expandHome("~/.agents/skills")
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
