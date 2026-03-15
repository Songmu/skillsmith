package skillsmith

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func newTestFS() fstest.MapFS {
	return fstest.MapFS{
		"mytool/SKILL.md": {
			Data: []byte(`---
name: mytool
description: A test skill
---
body
`),
		},
		"mytool/README.md": {
			Data: []byte("# mytool\n"),
		},
	}
}

func TestCopySkills_Install(t *testing.T) {
	src := newTestFS()
	dest := t.TempDir()

	result, err := CopySkills(src, dest, CopyOptions{
		Mode:    ModeInstall,
		Name:    "tool",
		Version: "v1.0.0",
	})
	if err != nil {
		t.Fatalf("CopySkills: %v", err)
	}

	installed := result.Installed()
	if len(installed) != 1 || installed[0].Dir != "mytool" {
		t.Errorf("expected 1 installed skill, got: %v", installed)
	}

	// SKILL.md should be on disk.
	if _, err := os.Stat(filepath.Join(dest, "mytool", "SKILL.md")); err != nil {
		t.Errorf("SKILL.md not found after install: %v", err)
	}

	// .skillsmith.json should be on disk.
	if !IsManaged(filepath.Join(dest, "mytool")) {
		t.Error(".skillsmith.json not found after install")
	}

	meta, err := ReadMeta(filepath.Join(dest, "mytool"))
	if err != nil {
		t.Fatalf("ReadMeta: %v", err)
	}
	if meta.InstalledBy != "tool" {
		t.Errorf("InstalledBy = %q, want %q", meta.InstalledBy, "tool")
	}
	if meta.Version != "v1.0.0" {
		t.Errorf("Version = %q, want %q", meta.Version, "v1.0.0")
	}
}

func TestCopySkills_Install_SkipsExisting(t *testing.T) {
	src := newTestFS()
	dest := t.TempDir()

	// First install.
	if _, err := CopySkills(src, dest, CopyOptions{Mode: ModeInstall, Name: "tool", Version: "v1"}); err != nil {
		t.Fatalf("first install: %v", err)
	}

	// Second install should skip.
	result, err := CopySkills(src, dest, CopyOptions{Mode: ModeInstall, Name: "tool", Version: "v2"})
	if err != nil {
		t.Fatalf("second install: %v", err)
	}

	if len(result.Skipped()) != 1 {
		t.Errorf("expected 1 skipped skill on second install, got: %v", result.Skipped())
	}

	// Version should still be v1 (not overwritten).
	meta, _ := ReadMeta(filepath.Join(dest, "mytool"))
	if meta.Version != "v1" {
		t.Errorf("expected version to remain v1, got %q", meta.Version)
	}
}

func TestCopySkills_Install_WarnsUnmanaged(t *testing.T) {
	src := newTestFS()
	dest := t.TempDir()

	// Create an unmanaged skill directory (no .skillsmith.json).
	if err := os.MkdirAll(filepath.Join(dest, "mytool"), 0o755); err != nil {
		t.Fatal(err)
	}

	result, err := CopySkills(src, dest, CopyOptions{Mode: ModeInstall, Name: "tool", Version: "v1"})
	if err != nil {
		t.Fatalf("CopySkills: %v", err)
	}

	if len(result.Warned()) != 1 {
		t.Errorf("expected 1 warned skill for unmanaged, got: %v", result.Warned())
	}
}

func TestCopySkills_Install_ForceUnmanaged(t *testing.T) {
	src := newTestFS()
	dest := t.TempDir()

	if err := os.MkdirAll(filepath.Join(dest, "mytool"), 0o755); err != nil {
		t.Fatal(err)
	}

	result, err := CopySkills(src, dest, CopyOptions{Mode: ModeInstall, Name: "tool", Version: "v1", Force: true})
	if err != nil {
		t.Fatalf("CopySkills: %v", err)
	}

	if len(result.Installed()) != 1 {
		t.Errorf("expected 1 installed skill with --force, got: %v", result.Installed())
	}
}

func TestCopySkills_Update_SameVersion(t *testing.T) {
	src := newTestFS()
	dest := t.TempDir()

	if _, err := CopySkills(src, dest, CopyOptions{Mode: ModeInstall, Name: "tool", Version: "v1"}); err != nil {
		t.Fatalf("install: %v", err)
	}

	result, err := CopySkills(src, dest, CopyOptions{Mode: ModeUpdate, Name: "tool", Version: "v1"})
	if err != nil {
		t.Fatalf("update: %v", err)
	}

	if len(result.Skipped()) != 1 {
		t.Errorf("expected 1 skipped skill (same version), got: %v", result.Skipped())
	}
}

func TestCopySkills_Update_DifferentVersion(t *testing.T) {
	src := newTestFS()
	dest := t.TempDir()

	if _, err := CopySkills(src, dest, CopyOptions{Mode: ModeInstall, Name: "tool", Version: "v1"}); err != nil {
		t.Fatalf("install: %v", err)
	}

	result, err := CopySkills(src, dest, CopyOptions{Mode: ModeUpdate, Name: "tool", Version: "v2"})
	if err != nil {
		t.Fatalf("update: %v", err)
	}

	if len(result.Installed()) != 1 || result.Installed()[0].Action != "updated" {
		t.Errorf("expected 1 updated skill, got: %v", result.Installed())
	}

	meta, _ := ReadMeta(filepath.Join(dest, "mytool"))
	if meta.Version != "v2" {
		t.Errorf("expected version v2 after update, got %q", meta.Version)
	}
}

func TestCopySkills_Reinstall(t *testing.T) {
	src := newTestFS()
	dest := t.TempDir()

	if _, err := CopySkills(src, dest, CopyOptions{Mode: ModeInstall, Name: "tool", Version: "v1"}); err != nil {
		t.Fatalf("install: %v", err)
	}

	result, err := CopySkills(src, dest, CopyOptions{Mode: ModeReinstall, Name: "tool", Version: "v1"})
	if err != nil {
		t.Fatalf("reinstall: %v", err)
	}

	if len(result.Installed()) != 1 || result.Installed()[0].Action != "reinstalled" {
		t.Errorf("expected 1 reinstalled skill, got: %v", result.Installed())
	}
}

func TestCopySkills_DryRun(t *testing.T) {
	src := newTestFS()
	dest := t.TempDir()

	result, err := CopySkills(src, dest, CopyOptions{Mode: ModeInstall, Name: "tool", Version: "v1", DryRun: true})
	if err != nil {
		t.Fatalf("CopySkills dry-run: %v", err)
	}

	if len(result.Installed()) != 1 {
		t.Errorf("expected 1 dry-run installed skill, got: %v", result.Installed())
	}

	// Nothing should have been written to disk.
	if _, err := os.Stat(filepath.Join(dest, "mytool")); err == nil {
		t.Error("expected no files written in dry-run mode, but directory exists")
	}
}

func TestCopySkills_InvalidSKILLmd(t *testing.T) {
	src := fstest.MapFS{
		"bad-skill/SKILL.md": {Data: []byte("no frontmatter")},
	}
	dest := t.TempDir()

	result, err := CopySkills(src, dest, CopyOptions{Mode: ModeInstall})
	if err != nil {
		t.Fatalf("CopySkills: %v", err)
	}

	// The bad skill should produce a warning.
	if len(result.Warned()) == 0 {
		t.Error("expected at least one warning for bad SKILL.md, got none")
	}
}
