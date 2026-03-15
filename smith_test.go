package skillsmith

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"testing/fstest"
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
	s := &Smith{
		FS:        fsys,
		Version:   "v1.0.0",
		Name:      "testtool",
		OutWriter: out,
		ErrWriter: errW,
	}
	return s, out, errW
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
