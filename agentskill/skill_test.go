package agentskill

import (
	"strings"
	"testing"
)

const validSKILL = `---
name: mytool-cli
description: A helpful CLI skill
license: MIT
compatibility:
  - claude
  - codex
allowed_tools:
  - Bash
---

# mytool-cli

This skill teaches AI agents how to use mytool-cli effectively.
`

func TestParse_Valid(t *testing.T) {
	s, err := Parse(strings.NewReader(validSKILL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Name != "mytool-cli" {
		t.Errorf("Name = %q, want %q", s.Name, "mytool-cli")
	}
	if s.Description != "A helpful CLI skill" {
		t.Errorf("Description = %q, want %q", s.Description, "A helpful CLI skill")
	}
	if s.License != "MIT" {
		t.Errorf("License = %q, want %q", s.License, "MIT")
	}
	if len(s.Compatibility) != 2 || s.Compatibility[0] != "claude" {
		t.Errorf("Compatibility = %v, want [claude codex]", s.Compatibility)
	}
	if len(s.AllowedTools) != 1 || s.AllowedTools[0] != "Bash" {
		t.Errorf("AllowedTools = %v, want [Bash]", s.AllowedTools)
	}
	if !strings.Contains(s.Body, "mytool-cli") {
		t.Errorf("Body does not contain expected content, got: %q", s.Body)
	}
}

func TestParse_MissingOpeningDelimiter(t *testing.T) {
	_, err := Parse(strings.NewReader("name: foo\ndescription: bar\n"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestParse_MissingClosingDelimiter(t *testing.T) {
	_, err := Parse(strings.NewReader("---\nname: foo\ndescription: bar\n"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestParse_InvalidYAML(t *testing.T) {
	_, err := Parse(strings.NewReader("---\n: invalid: yaml: [\n---\n"))
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}

func TestParse_EmptyBody(t *testing.T) {
	input := "---\nname: foo\ndescription: bar\n---\n"
	s, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Body != "" {
		t.Errorf("Body = %q, want empty string", s.Body)
	}
}

func TestParse_WithMetadata(t *testing.T) {
	input := `---
name: foo
description: bar
metadata:
  key: value
  num: 42
---
body
`
	s, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Metadata == nil {
		t.Fatal("Metadata is nil, want map")
	}
	if s.Metadata["key"] != "value" {
		t.Errorf("Metadata[key] = %v, want \"value\"", s.Metadata["key"])
	}
}
