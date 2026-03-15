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
	tests := []struct {
		name          string
		setup         func(t *testing.T, dest string)
		opts          CopyOptions
		wantInstalled int
		wantSkipped   int
		wantWarned    int
		checkMeta     func(t *testing.T, dest string)
	}{
		{
			name: "fresh_install",
			opts: CopyOptions{Mode: ModeInstall, Name: "tool", Version: "v1.0.0"},
			wantInstalled: 1,
			checkMeta: func(t *testing.T, dest string) {
				if _, err := os.Stat(filepath.Join(dest, "mytool", "SKILL.md")); err != nil {
					t.Errorf("SKILL.md not found after install: %v", err)
				}
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
			},
		},
		{
			name: "skips_existing",
			setup: func(t *testing.T, dest string) {
				if _, err := CopySkills(newTestFS(), dest, CopyOptions{Mode: ModeInstall, Name: "tool", Version: "v1"}); err != nil {
					t.Fatalf("first install: %v", err)
				}
			},
			opts:        CopyOptions{Mode: ModeInstall, Name: "tool", Version: "v2"},
			wantSkipped: 1,
			checkMeta: func(t *testing.T, dest string) {
				meta, _ := ReadMeta(filepath.Join(dest, "mytool"))
				if meta.Version != "v1" {
					t.Errorf("expected version to remain v1, got %q", meta.Version)
				}
			},
		},
		{
			name: "warns_unmanaged",
			setup: func(t *testing.T, dest string) {
				if err := os.MkdirAll(filepath.Join(dest, "mytool"), 0o755); err != nil {
					t.Fatal(err)
				}
			},
			opts:       CopyOptions{Mode: ModeInstall, Name: "tool", Version: "v1"},
			wantWarned: 1,
		},
		{
			name: "force_unmanaged",
			setup: func(t *testing.T, dest string) {
				if err := os.MkdirAll(filepath.Join(dest, "mytool"), 0o755); err != nil {
					t.Fatal(err)
				}
			},
			opts:          CopyOptions{Mode: ModeInstall, Name: "tool", Version: "v1", Force: true},
			wantInstalled: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := newTestFS()
			dest := t.TempDir()
			if tt.setup != nil {
				tt.setup(t, dest)
			}
			result, err := CopySkills(src, dest, tt.opts)
			if err != nil {
				t.Fatalf("CopySkills: %v", err)
			}
			if len(result.Installed()) != tt.wantInstalled {
				t.Errorf("installed = %d, want %d; skills: %v", len(result.Installed()), tt.wantInstalled, result.Installed())
			}
			if len(result.Skipped()) != tt.wantSkipped {
				t.Errorf("skipped = %d, want %d; skills: %v", len(result.Skipped()), tt.wantSkipped, result.Skipped())
			}
			if len(result.Warned()) != tt.wantWarned {
				t.Errorf("warned = %d, want %d; skills: %v", len(result.Warned()), tt.wantWarned, result.Warned())
			}
			if tt.checkMeta != nil {
				tt.checkMeta(t, dest)
			}
		})
	}
}

func TestCopySkills_Update(t *testing.T) {
	tests := []struct {
		name          string
		installVer    string
		updateVer     string
		wantInstalled int
		wantAction    string
		wantSkipped   int
		wantVersion   string
	}{
		{
			name:          "higher_version",
			installVer:    "v1.0.0",
			updateVer:     "v2.0.0",
			wantInstalled: 1,
			wantAction:    "updated",
			wantVersion:   "v2.0.0",
		},
		{
			name:        "same_version_skipped",
			installVer:  "v1.0.0",
			updateVer:   "v1.0.0",
			wantSkipped: 1,
		},
		{
			name:        "lower_version_skipped",
			installVer:  "v2.0.0",
			updateVer:   "v1.0.0",
			wantSkipped: 1,
			wantVersion: "v2.0.0",
		},
		{
			name:          "higher_version_without_v_prefix",
			installVer:    "1.0.0",
			updateVer:     "2.0.0",
			wantInstalled: 1,
			wantAction:    "updated",
			wantVersion:   "2.0.0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := newTestFS()
			dest := t.TempDir()
			if _, err := CopySkills(src, dest, CopyOptions{Mode: ModeInstall, Name: "tool", Version: tt.installVer}); err != nil {
				t.Fatalf("install: %v", err)
			}
			result, err := CopySkills(src, dest, CopyOptions{Mode: ModeUpdate, Name: "tool", Version: tt.updateVer})
			if err != nil {
				t.Fatalf("update: %v", err)
			}
			if len(result.Installed()) != tt.wantInstalled {
				t.Errorf("installed = %d, want %d; skills: %v", len(result.Installed()), tt.wantInstalled, result.Installed())
			}
			if tt.wantAction != "" {
				gotAction := ""
				if len(result.Installed()) > 0 {
					gotAction = result.Installed()[0].Action
				}
				if gotAction != tt.wantAction {
					t.Errorf("action = %q, want %q", gotAction, tt.wantAction)
				}
			}
			if len(result.Skipped()) != tt.wantSkipped {
				t.Errorf("skipped = %d, want %d; skills: %v", len(result.Skipped()), tt.wantSkipped, result.Skipped())
			}
			if tt.wantVersion != "" {
				meta, _ := ReadMeta(filepath.Join(dest, "mytool"))
				if meta.Version != tt.wantVersion {
					t.Errorf("version = %q, want %q", meta.Version, tt.wantVersion)
				}
			}
		})
	}
}

func TestCopySkills_Reinstall(t *testing.T) {
	tests := []struct {
		name          string
		installVer    string
		reinstallVer  string
		force         bool
		wantInstalled int
		wantAction    string
		wantSkipped   int
		wantWarned    int
		wantVersion   string
		setup         func(t *testing.T, src fstest.MapFS, dest string)
	}{
		{
			name:          "same_version",
			installVer:    "v1.0.0",
			reinstallVer:  "v1.0.0",
			wantInstalled: 1,
			wantAction:    "reinstalled",
		},
		{
			name:          "higher_version",
			installVer:    "v1.0.0",
			reinstallVer:  "v2.0.0",
			wantInstalled: 1,
			wantAction:    "reinstalled",
			wantVersion:   "v2.0.0",
		},
		{
			name:         "downgrade_skipped",
			installVer:   "v2.0.0",
			reinstallVer: "v1.0.0",
			wantSkipped:  1,
		},
		{
			name:          "force_downgrade",
			installVer:    "v2.0.0",
			reinstallVer:  "v1.0.0",
			force:         true,
			wantInstalled: 1,
			wantAction:    "reinstalled",
			wantVersion:   "v1.0.0",
		},
		{
			name: "skips_unmanaged",
			setup: func(t *testing.T, src fstest.MapFS, dest string) {
				if err := os.MkdirAll(filepath.Join(dest, "mytool"), 0o755); err != nil {
					t.Fatal(err)
				}
			},
			reinstallVer: "v1.0.0",
			wantWarned:   1,
		},
		{
			name: "force_unmanaged",
			setup: func(t *testing.T, src fstest.MapFS, dest string) {
				if err := os.MkdirAll(filepath.Join(dest, "mytool"), 0o755); err != nil {
					t.Fatal(err)
				}
			},
			reinstallVer:  "v1.0.0",
			force:         true,
			wantInstalled: 1,
			wantAction:    "reinstalled",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := newTestFS()
			dest := t.TempDir()
			if tt.setup != nil {
				tt.setup(t, src, dest)
			} else if tt.installVer != "" {
				if _, err := CopySkills(src, dest, CopyOptions{Mode: ModeInstall, Name: "tool", Version: tt.installVer}); err != nil {
					t.Fatalf("install: %v", err)
				}
			}
			result, err := CopySkills(src, dest, CopyOptions{Mode: ModeReinstall, Name: "tool", Version: tt.reinstallVer, Force: tt.force})
			if err != nil {
				t.Fatalf("CopySkills reinstall: %v", err)
			}
			if len(result.Installed()) != tt.wantInstalled {
				t.Errorf("installed = %d, want %d; skills: %v", len(result.Installed()), tt.wantInstalled, result.Installed())
			}
			if tt.wantAction != "" {
				gotAction := ""
				if len(result.Installed()) > 0 {
					gotAction = result.Installed()[0].Action
				}
				if gotAction != tt.wantAction {
					t.Errorf("action = %q, want %q", gotAction, tt.wantAction)
				}
			}
			if len(result.Skipped()) != tt.wantSkipped {
				t.Errorf("skipped = %d, want %d; skills: %v", len(result.Skipped()), tt.wantSkipped, result.Skipped())
			}
			if len(result.Warned()) != tt.wantWarned {
				t.Errorf("warned = %d, want %d; skills: %v", len(result.Warned()), tt.wantWarned, result.Warned())
			}
			if tt.wantVersion != "" {
				meta, _ := ReadMeta(filepath.Join(dest, "mytool"))
				if meta.Version != tt.wantVersion {
					t.Errorf("version = %q, want %q", meta.Version, tt.wantVersion)
				}
			}
		})
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
