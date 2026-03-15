package agentskill

import (
	"bufio"
	"bytes"
	"errors"
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
	// License is the SPDX license identifier from frontmatter.
	License string
	// Compatibility lists agent compatibility strings from frontmatter.
	Compatibility []string
	// Metadata holds arbitrary client-extension metadata from frontmatter.
	Metadata map[string]any
	// AllowedTools lists permitted tool names from frontmatter.
	AllowedTools []string
	// Body is the Markdown body after the closing frontmatter delimiter.
	Body string
	// Dir is the directory name of the skill (set by Discover).
	Dir string
}

// frontmatter holds the raw YAML fields parsed from the SKILL.md header.
type frontmatter struct {
	Name          string         `yaml:"name"`
	Description   string         `yaml:"description"`
	License       string         `yaml:"license"`
	Compatibility []string       `yaml:"compatibility"`
	Metadata      map[string]any `yaml:"metadata"`
	AllowedTools  []string       `yaml:"allowed_tools"`
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

	s := &Skill{
		Name:          fm.Name,
		Description:   fm.Description,
		License:       fm.License,
		Compatibility: fm.Compatibility,
		Metadata:      fm.Metadata,
		AllowedTools:  fm.AllowedTools,
		Body:          body,
	}
	return s, nil
}

// splitFrontmatter splits the raw content of a SKILL.md into the YAML bytes
// (without delimiters) and the remaining body text.
// It returns an error when no valid frontmatter block is found.
func splitFrontmatter(data []byte) (yamlBytes []byte, body string, err error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if scanErr := scanner.Err(); scanErr != nil {
		return nil, "", scanErr
	}

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

	return []byte(yamlContent), strings.TrimLeft(bodyContent, "\n"), nil
}
