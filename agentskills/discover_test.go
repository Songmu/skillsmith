package agentskills

import (
	"errors"
	"testing"
	"testing/fstest"
)

func TestDiscover_Valid(t *testing.T) {
	fsys := fstest.MapFS{
		"mytool-cli/SKILL.md": {
			Data: []byte(`---
name: mytool-cli
description: A CLI skill
license: MIT
---
body
`),
		},
		"other-tool/SKILL.md": {
			Data: []byte(`---
name: other-tool
description: Another skill
---
`),
		},
	}

	skills, err := Discover(fsys)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(skills) != 2 {
		t.Errorf("expected 2 skills, got %d", len(skills))
	}
}

func TestDiscover_SkipsNonDirs(t *testing.T) {
	fsys := fstest.MapFS{
		"README.md": {Data: []byte("hello")},
		"my-skill/SKILL.md": {
			Data: []byte(`---
name: my-skill
description: A skill
---
`),
		},
	}

	skills, err := Discover(fsys)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(skills) != 1 {
		t.Errorf("expected 1 skill, got %d", len(skills))
	}
	if skills[0].Dir != "my-skill" {
		t.Errorf("Dir = %q, want %q", skills[0].Dir, "my-skill")
	}
}

func TestDiscover_SkipsNoSKILLmd(t *testing.T) {
	fsys := fstest.MapFS{
		"my-skill/README.md": {Data: []byte("hello")},
	}

	skills, err := Discover(fsys)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(skills) != 0 {
		t.Errorf("expected 0 skills, got %d", len(skills))
	}
}

func TestDiscover_ParseError(t *testing.T) {
	fsys := fstest.MapFS{
		"bad-skill/SKILL.md": {
			Data: []byte("no frontmatter here"),
		},
		"good-skill/SKILL.md": {
			Data: []byte(`---
name: good-skill
description: A good skill
---
`),
		},
	}

	skills, err := Discover(fsys)
	if err == nil {
		t.Error("expected at least one error for bad-skill, got none")
	}
	// The error should be extractable as a SkillError.
	var se *SkillError
	if !errors.As(err, &se) {
		t.Errorf("expected *SkillError in error chain, got: %T %v", err, err)
	}
	if len(skills) != 1 || skills[0].Dir != "good-skill" {
		t.Errorf("expected 1 valid skill (good-skill), got %v", skills)
	}
}

func TestDiscover_ValidationError(t *testing.T) {
	fsys := fstest.MapFS{
		"no-desc/SKILL.md": {
			Data: []byte(`---
name: no-desc
description: ""
---
`),
		},
	}

	skills, err := Discover(fsys)
	if err == nil {
		t.Error("expected error for missing description, got none")
	}
	var se *SkillError
	if !errors.As(err, &se) {
		t.Errorf("expected *SkillError in error chain, got: %T %v", err, err)
	}
	if len(skills) != 0 {
		t.Errorf("expected 0 skills, got %d", len(skills))
	}
}

func TestDiscover_WarningDoesNotSkip(t *testing.T) {
	fsys := fstest.MapFS{
		"my-skill/SKILL.md": {
			Data: []byte(`---
name: different-name
description: A skill
---
`),
		},
	}

	skills, err := Discover(fsys)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(skills) != 1 {
		t.Errorf("expected 1 skill (warning should not skip), got %d", len(skills))
	}
}
