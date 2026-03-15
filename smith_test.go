package skillsmith

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"testing/fstest"
	"io/fs"
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

func newTestSmith(fsys fstest.MapFS) (*Smith, *bytes.Buffer, *bytes.Buffer) {
	out := &bytes.Buffer{}
	errW := &bytes.Buffer{}
	s, err := New("testtool", "v1.0.0", fsys)
	if err != nil {
		panic(err)
	}
	s.OutWriter = out
	s.ErrWriter = errW
	return s, out, errW
}

func TestNew_InvalidVersion(t *testing.T) {
	_, err := New("tool", "not-a-version", testSkillFS)
	if err == nil {
		t.Error("expected error for invalid version, got nil")
	}
}

func TestNew_ValidVersion_WithV(t *testing.T) {
	s, err := New("tool", "v1.2.3", testSkillFS)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if s.Version() != "1.2.3" {
		t.Errorf("expected version %q stored without 'v', got %q", "1.2.3", s.Version())
	}
}

func TestNew_ValidVersion_WithoutV(t *testing.T) {
	s, err := New("tool", "1.2.3", testSkillFS)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if s.Version() != "1.2.3" {
		t.Errorf("expected version %q stored without 'v', got %q", "1.2.3", s.Version())
	}
}

func TestSmith_Run_UnknownSubcommand(t *testing.T) {
	s, _, _ := newTestSmith(testSkillFS)
	err := s.Run(context.Background(), []string{"unknown"})
	if err == nil {
		t.Error("expected error for unknown subcommand, got nil")
	}
}

func TestSmith_Run_NoArgs(t *testing.T) {
	s, _, _ := newTestSmith(testSkillFS)
	err := s.Run(context.Background(), []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSmith_Run_Help(t *testing.T) {
	s, _, errW := newTestSmith(testSkillFS)
	err := s.Run(context.Background(), []string{"--help"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !strings.Contains(errW.String(), "install") {
		t.Errorf("help output does not mention 'install', got: %q", errW.String())
	}
}

func TestSmith_List(t *testing.T) {
	s, out, _ := newTestSmith(testSkillFS)
	err := s.Run(context.Background(), []string{"list"})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if !strings.Contains(out.String(), "demo-skill") {
		t.Errorf("list output missing 'demo-skill', got: %q", out.String())
	}
}

func TestSmith_Install_DryRun(t *testing.T) {
	s, out, _ := newTestSmith(testSkillFS)
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
	s, out, _ := newTestSmith(testSkillFS)
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
	s, out, _ := newTestSmith(testSkillFS)
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
	s, out, _ := newTestSmith(testSkillFS)
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
	s, out, _ := newTestSmith(testSkillFS)
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

// TestNew_AutoDetect_SingleDir verifies that New strips the "skills/" prefix when
// it is the only directory at the root of skillFS.
func TestNew_AutoDetect_SingleDir(t *testing.T) {
	s, err := New("test", "1.0.0", wrappedSkillFS)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, err := fs.Stat(s.fs, "demo-skill/SKILL.md"); err != nil {
		t.Fatalf("expected 'demo-skill' at root of skill FS after auto-detect strip, stat error: %v", err)
	}
}

// TestNew_AutoDetect_PreStripped verifies that New uses the FS as-is when skills
// are already at the root (no "skills/" prefix to strip).
func TestNew_AutoDetect_PreStripped(t *testing.T) {
	s, err := New("test", "1.0.0", testSkillFS)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, err := fs.Stat(s.fs, "demo-skill/SKILL.md"); err != nil {
		t.Fatalf("expected 'demo-skill' at root of pre-stripped skill FS, stat error: %v", err)
	}
}

// TestNew_AutoDetect_MixedRoot verifies that New uses the FS as-is when the only
// directory at root is not named "skills".
func TestNew_AutoDetect_MixedRoot(t *testing.T) {
	s, err := New("test", "1.0.0", mixedRootFS)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, err := fs.Stat(s.fs, "demo-skill/SKILL.md"); err != nil {
		t.Fatalf("expected 'demo-skill' at root of mixed-root skill FS, stat error: %v", err)
	}
}

// TestNew_AutoDetect_SkillsDirWithFileAtRoot verifies that New strips the "skills/"
// prefix even when files are present alongside it at root.
func TestNew_AutoDetect_SkillsDirWithFileAtRoot(t *testing.T) {
	s, err := New("test", "1.0.0", wrappedWithFileFS)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, err := fs.Stat(s.fs, "demo-skill/SKILL.md"); err != nil {
		t.Fatalf("expected 'demo-skill' at root of skill FS after auto-detect strip (with file at root), stat error: %v", err)
	}
}


