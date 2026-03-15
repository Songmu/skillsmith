package agentskills

import (
	"errors"
	"testing"
	"testing/fstest"
)

func TestDiscover_NormalCases(t *testing.T) {
	tests := []struct {
		name      string
		fsys      fstest.MapFS
		wantCount int
		wantDir   string
		wantErr   bool
	}{
		{
			name: "valid_discovers_all",
			fsys: fstest.MapFS{
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
			},
			wantCount: 2,
		},
		{
			name: "skips_non_dirs",
			fsys: fstest.MapFS{
				"README.md": {Data: []byte("hello")},
				"my-skill/SKILL.md": {
					Data: []byte(`---
name: my-skill
description: A skill
---
`),
				},
			},
			wantCount: 1,
			wantDir:   "my-skill",
		},
		{
			name: "skips_no_skill_md",
			fsys: fstest.MapFS{
				"my-skill/README.md": {Data: []byte("hello")},
			},
			wantCount: 0,
		},
		{
			name: "warning_does_not_skip",
			fsys: fstest.MapFS{
				"my-skill/SKILL.md": {
					Data: []byte(`---
name: different-name
description: A skill
---
`),
				},
			},
			wantCount: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			skills, err := Discover(tt.fsys)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if len(skills) != tt.wantCount {
				t.Errorf("expected %d skill(s), got %d", tt.wantCount, len(skills))
			}
			if tt.wantDir != "" {
				gotDir := ""
				if len(skills) > 0 {
					gotDir = skills[0].Dir
				}
				if gotDir != tt.wantDir {
					t.Errorf("Dir = %q, want %q", gotDir, tt.wantDir)
				}
			}
		})
	}
}

func TestDiscover_ErrorCases(t *testing.T) {
	tests := []struct {
		name           string
		fsys           fstest.MapFS
		wantSkillCount int
		wantDirs       []string
	}{
		{
			name: "parse_error",
			fsys: fstest.MapFS{
				"bad-skill/SKILL.md": {Data: []byte("no frontmatter here")},
				"good-skill/SKILL.md": {
					Data: []byte(`---
name: good-skill
description: A good skill
---
`),
				},
			},
			wantSkillCount: 1,
			wantDirs:       []string{"good-skill"},
		},
		{
			name: "validation_error",
			fsys: fstest.MapFS{
				"no-desc/SKILL.md": {
					Data: []byte(`---
name: no-desc
description: ""
---
`),
				},
			},
			wantSkillCount: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			skills, err := Discover(tt.fsys)
			if err == nil {
				t.Error("expected at least one error, got none")
			}
			var se *SkillError
			if !errors.As(err, &se) {
				t.Errorf("expected *SkillError in error chain, got: %T %v", err, err)
			}
			if len(skills) != tt.wantSkillCount {
				t.Errorf("expected %d skill(s), got %d", tt.wantSkillCount, len(skills))
			}
			if len(tt.wantDirs) > 0 {
				if len(skills) != len(tt.wantDirs) {
					t.Fatalf("expected skills from dirs %v, but got %d skills", tt.wantDirs, len(skills))
				}
				gotDirs := make(map[string]struct{}, len(skills))
				for _, s := range skills {
					gotDirs[s.Dir] = struct{}{}
				}
				for _, wantDir := range tt.wantDirs {
					if _, ok := gotDirs[wantDir]; !ok {
						t.Errorf("expected discovered skills to include dir %q, got dirs %v", wantDir, gotDirs)
					}
				}
			}
		})
	}
}
