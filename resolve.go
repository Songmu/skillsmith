package skillsmith

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// agentScopePaths maps (agent, scope) to installation path patterns.
// A leading "~" is expanded to the user's home directory.
var agentScopePaths = map[string]map[string]string{
	"codex": {
		"user": "~/.codex/skills",
		"repo": ".agents/skills",
	},
	"claude": {
		"user": "~/.claude/skills",
		"repo": ".claude/skills",
	},
	"agents": {
		"user": "~/.agents/skills",
		"repo": ".agents/skills",
	},
}

// ResolveInstallDir returns the skill installation directory for the given
// agent and scope. An empty agent defaults to "claude"; an empty scope
// defaults to "user".
func ResolveInstallDir(agent, scope string) (string, error) {
	if agent == "" {
		agent = "claude"
	}
	if scope == "" {
		scope = "user"
	}

	scopes, ok := agentScopePaths[agent]
	if !ok {
		return "", fmt.Errorf("unknown agent %q (supported: codex, claude, agents)", agent)
	}

	p, ok := scopes[scope]
	if !ok {
		return "", fmt.Errorf("unknown scope %q for agent %q (supported: user, repo)", scope, agent)
	}

	return expandHome(p)
}

// expandHome replaces a leading "~" with the user's home directory.
func expandHome(p string) (string, error) {
	if !strings.HasPrefix(p, "~") {
		return p, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving home directory: %w", err)
	}
	return filepath.Join(home, p[1:]), nil
}
