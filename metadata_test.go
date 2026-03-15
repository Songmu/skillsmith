package skillsmith

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWriteReadMeta(t *testing.T) {
	dir := t.TempDir()

	meta := &SkillMeta{
		InstalledBy: "mytool",
		Version:     "v1.2.3",
		InstalledAt: time.Date(2026, 3, 13, 16, 30, 0, 0, time.UTC),
	}

	if err := WriteMeta(dir, meta); err != nil {
		t.Fatalf("WriteMeta: %v", err)
	}

	got, err := ReadMeta(dir)
	if err != nil {
		t.Fatalf("ReadMeta: %v", err)
	}

	if got.InstalledBy != meta.InstalledBy {
		t.Errorf("InstalledBy = %q, want %q", got.InstalledBy, meta.InstalledBy)
	}
	if got.Version != meta.Version {
		t.Errorf("Version = %q, want %q", got.Version, meta.Version)
	}
	if !got.InstalledAt.Equal(meta.InstalledAt) {
		t.Errorf("InstalledAt = %v, want %v", got.InstalledAt, meta.InstalledAt)
	}
}

func TestWriteMeta_JSONFormat(t *testing.T) {
	dir := t.TempDir()

	meta := &SkillMeta{
		InstalledBy: "tool",
		Version:     "v0.1.0",
		InstalledAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	if err := WriteMeta(dir, meta); err != nil {
		t.Fatalf("WriteMeta: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".skillsmith.json"))
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}

	// Verify camelCase keys.
	for _, key := range []string{"installedBy", "version", "installedAt"} {
		if _, ok := raw[key]; !ok {
			t.Errorf("JSON key %q not found in output", key)
		}
	}
}

func TestIsManaged(t *testing.T) {
	dir := t.TempDir()

	if IsManaged(dir) {
		t.Error("expected IsManaged=false for empty directory, got true")
	}

	meta := &SkillMeta{InstalledBy: "tool", Version: "v1", InstalledAt: time.Now()}
	if err := WriteMeta(dir, meta); err != nil {
		t.Fatalf("WriteMeta: %v", err)
	}

	if !IsManaged(dir) {
		t.Error("expected IsManaged=true after WriteMeta, got false")
	}
}

func TestReadMeta_Missing(t *testing.T) {
	dir := t.TempDir()
	_, err := ReadMeta(dir)
	if err == nil {
		t.Error("expected error reading missing meta, got nil")
	}
}
