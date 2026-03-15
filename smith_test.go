package skillsmith

import (
	"bytes"
	"context"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/Songmu/skillsmith/agentskills"
)

var testSkillFS = fstest.MapFS{
	"demo-skill/SKILL.md": {
		Data: []byte(`---
name: demo-skill
description: A demonstration skill
license: MIT
---
# demo-skill

Teaches the agent how to use demo.
`),
	},
}

func newTestSmith(t *testing.T, fsys fstest.MapFS) (*Smith, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()

	out := &bytes.Buffer{}
	errW := &bytes.Buffer{}
	s, err := New("testtool", "v1.0.0", fsys)
	if err != nil {
		t.Fatalf("New(%q, %q) failed: %v", "testtool", "v1.0.0", err)
	}
	s.OutWriter = out
	s.ErrWriter = errW
	return s, out, errW
}

func TestNew_VersionValidation(t *testing.T) {
	tests := []struct {
		name        string
		version     string
		wantErr     bool
		wantVersion string
	}{
		{
			name:    "invalid_version",
			version: "not-a-version",
			wantErr: true,
		},
		{
			name:        "valid_with_v_prefix",
			version:     "v1.2.3",
			wantVersion: "v1.2.3",
		},
		{
			name:        "valid_without_v_prefix",
			version:     "1.2.3",
			wantVersion: "1.2.3",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := New("tool", tt.version, testSkillFS)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("New: %v", err)
			}
			if s.Version() != tt.wantVersion {
				t.Errorf("Version() = %q, want %q", s.Version(), tt.wantVersion)
			}
		})
	}
}

func TestNew_AutoDetect(t *testing.T) {
	tests := []struct {
		name          string
		fsys          fstest.MapFS
		wantSkillDir  string
		// wantSkillsDir indicates whether a "skills/" directory should remain at the root
		// after auto-detection. If false, "skills/" is expected to be stripped/absent.
		wantSkillsDir bool
		wantReadme    bool // true means README.md should be present at root
	}{
		{
			name:          "single_dir_strips_skills_prefix",
			fsys:          wrappedSkillFS,
			wantSkillDir:  "demo-skill",
			wantSkillsDir: false,
		},
		{
			name:         "pre_stripped_uses_as_is",
			fsys:         testSkillFS,
			wantSkillDir: "demo-skill",
		},
		{
			name:         "mixed_root_uses_as_is",
			fsys:         mixedRootFS,
			wantSkillDir: "demo-skill",
			wantReadme:   true,
		},
		{
			name:          "skills_dir_with_file_at_root_strips_skills",
			fsys:          wrappedWithFileFS,
			wantSkillDir:  "demo-skill",
			wantSkillsDir: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := New("test", "1.0.0", tt.fsys)
			if err != nil {
				t.Fatalf("New: %v", err)
			}
			skills, discoverErr := agentskills.Discover(s.fs)
			if discoverErr != nil {
				t.Fatalf("Discover: %v", discoverErr)
			}
			found := false
			for _, sk := range skills {
				if sk.Dir == tt.wantSkillDir {
					found = true
				}
			}
			if !found {
				t.Errorf("expected %q in s.fs after auto-detect, got: %v", tt.wantSkillDir, skills)
			}
			if !tt.wantSkillsDir {
				if _, statErr := fs.Stat(s.fs, "skills"); statErr == nil {
					t.Error("'skills' dir should have been stripped from root, but it still exists")
				}
			}
			if tt.wantReadme {
				if _, statErr := fs.Stat(s.fs, "README.md"); statErr != nil {
					t.Errorf("expected README.md at root of s.fs, got: %v", statErr)
				}
			}
		})
	}
}

func TestSmith_Run_Dispatch(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		wantErr        bool
		wantErrOutput  string
	}{
		{
			name:    "unknown_subcommand",
			args:    []string{"unknown"},
			wantErr: true,
		},
		{
			name:    "no_args",
			args:    []string{},
			wantErr: false,
		},
		{
			name:          "help_flag",
			args:          []string{"--help"},
			wantErr:       false,
			wantErrOutput: "install",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, _, errW := newTestSmith(t, testSkillFS)
			err := s.Run(context.Background(), tt.args)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.wantErrOutput != "" && !strings.Contains(errW.String(), tt.wantErrOutput) {
				t.Errorf("help output does not contain %q, got: %q", tt.wantErrOutput, errW.String())
			}
		})
	}
}

func TestSmith_List(t *testing.T) {
	s, out, _ := newTestSmith(t, testSkillFS)
	err := s.Run(context.Background(), []string{"list"})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if !strings.Contains(out.String(), "demo-skill") {
		t.Errorf("list output missing 'demo-skill', got: %q", out.String())
	}
}

func TestSmith_Install_DryRun(t *testing.T) {
	s, out, _ := newTestSmith(t, testSkillFS)
	dir := t.TempDir()
	err := s.Run(context.Background(), []string{"install", "--prefix", dir, "--dry-run"})
	if err != nil {
		t.Fatalf("install --dry-run: %v", err)
	}
	if !strings.Contains(out.String(), "dry-run") {
		t.Errorf("dry-run output missing indicator, got: %q", out.String())
	}
}

func TestSmith_Install_ThenStatus(t *testing.T) {
	s, out, _ := newTestSmith(t, testSkillFS)
	dir := t.TempDir()

	// Install.
	if err := s.Run(context.Background(), []string{"install", "--prefix", dir}); err != nil {
		t.Fatalf("install: %v", err)
	}

	// Status.
	out.Reset()
	if err := s.Run(context.Background(), []string{"status", "--prefix", dir}); err != nil {
		t.Fatalf("status: %v", err)
	}
	if !strings.Contains(out.String(), "demo-skill") {
		t.Errorf("status output missing 'demo-skill', got: %q", out.String())
	}
	if !strings.Contains(out.String(), "up to date") {
		t.Errorf("status output should show 'up to date', got: %q", out.String())
	}
}

func TestSmith_Uninstall(t *testing.T) {
	s, out, _ := newTestSmith(t, testSkillFS)
	dir := t.TempDir()

	if err := s.Run(context.Background(), []string{"install", "--prefix", dir}); err != nil {
		t.Fatalf("install: %v", err)
	}

	out.Reset()
	if err := s.Run(context.Background(), []string{"uninstall", "--prefix", dir}); err != nil {
		t.Fatalf("uninstall: %v", err)
	}
	if !strings.Contains(out.String(), "uninstalled") {
		t.Errorf("uninstall output missing 'uninstalled', got: %q", out.String())
	}
}

func TestSmith_Update_NoChange(t *testing.T) {
	s, out, _ := newTestSmith(t, testSkillFS)
	dir := t.TempDir()

	if err := s.Run(context.Background(), []string{"install", "--prefix", dir}); err != nil {
		t.Fatalf("install: %v", err)
	}

	out.Reset()
	if err := s.Run(context.Background(), []string{"update", "--prefix", dir}); err != nil {
		t.Fatalf("update: %v", err)
	}
	if !strings.Contains(out.String(), "skipped") {
		t.Errorf("update with same version should skip, got: %q", out.String())
	}
}

func TestSmith_Reinstall(t *testing.T) {
	s, out, _ := newTestSmith(t, testSkillFS)
	dir := t.TempDir()

	if err := s.Run(context.Background(), []string{"install", "--prefix", dir}); err != nil {
		t.Fatalf("install: %v", err)
	}

	out.Reset()
	if err := s.Run(context.Background(), []string{"reinstall", "--prefix", dir}); err != nil {
		t.Fatalf("reinstall: %v", err)
	}
	if !strings.Contains(out.String(), "reinstalled") {
		t.Errorf("reinstall output missing 'reinstalled', got: %q", out.String())
	}
}

// wrappedSkillFS wraps testSkillFS under a "skills/" directory to simulate
// what //go:embed skills/** produces.
var wrappedSkillFS = fstest.MapFS{
	"skills/demo-skill/SKILL.md": {
		Data: []byte(`---
name: demo-skill
description: A demonstration skill
license: MIT
---
# demo-skill

Teaches the agent how to use demo.
`),
	},
}

// mixedRootFS has a file at root alongside a non-"skills" directory.
// The file is ignored for detection; the dir is not named "skills", so the FS
// is used as-is.
var mixedRootFS = fstest.MapFS{
	"README.md": {Data: []byte("readme")},
	"demo-skill/SKILL.md": {
		Data: []byte(`---
name: demo-skill
description: A demonstration skill
license: MIT
---
# demo-skill

Teaches the agent how to use demo.
`),
	},
}

// wrappedWithFileFS wraps skills under "skills/" and also has a README.md at root,
// simulating an embed FS where files coexist with the skills container dir.
var wrappedWithFileFS = fstest.MapFS{
	"README.md": {Data: []byte("readme")},
	"skills/demo-skill/SKILL.md": {
		Data: []byte(`---
name: demo-skill
description: A demonstration skill
license: MIT
---
# demo-skill

Teaches the agent how to use demo.
`),
	},
}
