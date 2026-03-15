package agentskills

import (
	"strings"
	"testing"
)

func TestValidate(t *testing.T) {
	longName := strings.Repeat("a", 65)

	tests := []struct {
		name         string
		skill        *Skill
		dirName      string
		wantOK       bool
		wantWarnings int
		wantErrors   int
	}{
		{
			name:    "valid",
			skill:   &Skill{Name: "foo", Description: "some description"},
			dirName: "foo",
			wantOK:  true,
		},
		{
			name:       "empty_description",
			skill:      &Skill{Name: "foo", Description: ""},
			dirName:    "foo",
			wantOK:     false,
			wantErrors: 1,
		},
		{
			name:         "name_mismatch",
			skill:        &Skill{Name: "other-name", Description: "some description"},
			dirName:      "foo",
			wantOK:       true,
			wantWarnings: 1,
		},
		{
			name:         "name_too_long",
			skill:        &Skill{Name: longName, Description: "some description"},
			dirName:      longName,
			wantOK:       true,
			wantWarnings: 1,
		},
		{
			name:    "no_name_no_dir",
			skill:   &Skill{Name: "", Description: "desc"},
			dirName: "",
			wantOK:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Validate(tt.skill, tt.dirName)
			if result.OK() != tt.wantOK {
				t.Errorf("OK() = %v, want %v; errors: %v", result.OK(), tt.wantOK, result.Errors)
			}
			if tt.wantWarnings > 0 && len(result.Warnings) == 0 {
				t.Errorf("expected at least %d warning(s), got none", tt.wantWarnings)
			}
			if tt.wantErrors > 0 && len(result.Errors) == 0 {
				t.Errorf("expected at least %d error(s), got none", tt.wantErrors)
			}
		})
	}
}
