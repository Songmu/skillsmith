package agentskills

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/goccy/go-yaml"
)

// Skill represents a parsed agentskill from a SKILL.md file.
type Skill struct {
	// Name is the skill name from frontmatter.
	Name string
	// Description is the skill description from frontmatter.
	Description string
	// License is the license name or bundled license reference from frontmatter.
	License string
	// Compatibility describes environment requirements from frontmatter.
	Compatibility string
	// Metadata holds arbitrary client-extension metadata from frontmatter.
	Metadata map[string]any
	// AllowedTools is the space-separated tool pattern string from frontmatter.
	AllowedTools string
	// Body is the Markdown body after the closing frontmatter delimiter.
	Body string
	// Dir is the directory name of the skill (set by Discover).
	Dir string
}

// frontmatter holds the raw YAML fields parsed from the SKILL.md header.
type frontmatter struct {
	Name               string         `yaml:"name"`
	Description        string         `yaml:"description"`
	License            string         `yaml:"license"`
	Compatibility      any            `yaml:"compatibility"`
	Metadata           map[string]any `yaml:"metadata"`
	AllowedTools       any            `yaml:"allowed-tools"`
	LegacyAllowedTools any            `yaml:"allowed_tools"`
}

// Parse reads a SKILL.md file from r and returns a Skill.
// It returns an error when the YAML frontmatter is unparseable or missing.
func Parse(r io.Reader) (*Skill, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	// Extract YAML frontmatter between the first pair of "---" delimiters.
	yamlBytes, body, err := splitFrontmatter(data)
	if err != nil {
		return nil, err
	}

	var fm frontmatter
	if err := yaml.Unmarshal(yamlBytes, &fm); err != nil {
		return nil, err
	}

	compatibility, err := normalizeStringField(fm.Compatibility, ", ")
	if err != nil {
		return nil, fmt.Errorf("compatibility: %w", err)
	}
	allowedToolsValue := fm.AllowedTools
	if allowedToolsValue == nil {
		allowedToolsValue = fm.LegacyAllowedTools
	}
	allowedTools, err := normalizeStringField(allowedToolsValue, " ")
	if err != nil {
		return nil, fmt.Errorf("allowed-tools: %w", err)
	}

	s := &Skill{
		Name:          fm.Name,
		Description:   fm.Description,
		License:       fm.License,
		Compatibility: compatibility,
		Metadata:      fm.Metadata,
		AllowedTools:  allowedTools,
		Body:          body,
	}
	return s, nil
}

func normalizeStringField(value any, separator string) (string, error) {
	switch value := value.(type) {
	case nil:
		return "", nil
	case string:
		return value, nil
	case []any:
		values := make([]string, len(value))
		for i, item := range value {
			text, ok := item.(string)
			if !ok {
				return "", fmt.Errorf("list item %d must be a string", i)
			}
			values[i] = text
		}
		return strings.Join(values, separator), nil
	default:
		return "", fmt.Errorf("must be a string")
	}
}

// splitFrontmatter splits the raw content of a SKILL.md into the YAML bytes
// (without delimiters) and the remaining body text.
// It returns an error when no valid frontmatter block is found.
func splitFrontmatter(data []byte) (yamlBytes []byte, body string, err error) {
	lines := strings.Split(string(data), "\n")

	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return nil, "", errors.New("SKILL.md: missing frontmatter opening delimiter")
	}

	// Find the closing "---".
	closingIdx := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			closingIdx = i
			break
		}
	}
	if closingIdx < 0 {
		return nil, "", errors.New("SKILL.md: missing frontmatter closing delimiter")
	}

	yamlContent := strings.Join(lines[1:closingIdx], "\n")
	bodyContent := strings.Join(lines[closingIdx+1:], "\n")

	return []byte(yamlContent), strings.TrimLeft(bodyContent, "\r\n"), nil
}
