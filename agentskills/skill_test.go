package agentskills

import (
	"strings"
	"testing"
)

const validSKILL = `---
name: mytool-cli
description: A helpful CLI skill
license: MIT
compatibility: Requires Claude Code or Codex
allowed-tools: Bash Read
---

# mytool-cli

This skill teaches AI agents how to use mytool-cli effectively.
`

func TestParse_Errors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "missing_opening_delimiter",
			input: "name: foo\ndescription: bar\n",
		},
		{
			name:  "missing_closing_delimiter",
			input: "---\nname: foo\ndescription: bar\n",
		},
		{
			name:  "invalid_yaml",
			input: "---\n: invalid: yaml: [\n---\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(strings.NewReader(tt.input))
			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestParse_Success(t *testing.T) {
	withMetadata := `---
name: foo
description: bar
metadata:
  key: value
  num: 42
---
body
`
	tests := []struct {
		name       string
		input      string
		checkSkill func(t *testing.T, s *Skill)
	}{
		{
			name:  "valid_full",
			input: validSKILL,
			checkSkill: func(t *testing.T, s *Skill) {
				if s.Name != "mytool-cli" {
					t.Errorf("Name = %q, want %q", s.Name, "mytool-cli")
				}
				if s.Description != "A helpful CLI skill" {
					t.Errorf("Description = %q, want %q", s.Description, "A helpful CLI skill")
				}
				if s.License != "MIT" {
					t.Errorf("License = %q, want %q", s.License, "MIT")
				}
				if s.Compatibility != "Requires Claude Code or Codex" {
					t.Errorf("Compatibility = %q, want %q", s.Compatibility, "Requires Claude Code or Codex")
				}
				if s.AllowedTools != "Bash Read" {
					t.Errorf("AllowedTools = %q, want %q", s.AllowedTools, "Bash Read")
				}
				if !strings.Contains(s.Body, "mytool-cli") {
					t.Errorf("Body does not contain expected content, got: %q", s.Body)
				}
			},
		},
		{
			name:  "empty_body",
			input: "---\nname: foo\ndescription: bar\n---\n",
			checkSkill: func(t *testing.T, s *Skill) {
				if s.Body != "" {
					t.Errorf("Body = %q, want empty string", s.Body)
				}
			},
		},
		{
			name:  "with_metadata",
			input: withMetadata,
			checkSkill: func(t *testing.T, s *Skill) {
				if s.Metadata == nil {
					t.Fatal("Metadata is nil, want map")
				}
				if s.Metadata["key"] != "value" {
					t.Errorf("Metadata[key] = %v, want \"value\"", s.Metadata["key"])
				}
			},
		},
		{
			name: "legacy_list_fields",
			input: `---
name: foo
description: bar
compatibility:
  - claude
  - codex
allowed_tools:
  - Bash
  - Read
---
`,
			checkSkill: func(t *testing.T, s *Skill) {
				if s.Compatibility != "claude, codex" {
					t.Errorf("Compatibility = %q, want %q", s.Compatibility, "claude, codex")
				}
				if s.AllowedTools != "Bash Read" {
					t.Errorf("AllowedTools = %q, want %q", s.AllowedTools, "Bash Read")
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			tt.checkSkill(t, s)
		})
	}
}
