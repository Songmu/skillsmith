package skillsmith

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const metaFilename = ".skillsmith.json"

// SkillMeta holds metadata written alongside an installed skill.
type SkillMeta struct {
	InstalledBy string    `json:"installedBy"`
	Version     string    `json:"version"`
	InstalledAt time.Time `json:"installedAt"`
}

// ReadMeta reads .skillsmith.json from the given skill directory.
// It returns an error when the file cannot be read or decoded.
func ReadMeta(dir string) (*SkillMeta, error) {
	data, err := os.ReadFile(filepath.Join(dir, metaFilename))
	if err != nil {
		return nil, err
	}
	var m SkillMeta
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// WriteMeta writes meta as .skillsmith.json into dir.
func WriteMeta(dir string, meta *SkillMeta) error {
	data, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, metaFilename), data, 0o644)
}

// IsManaged reports whether a .skillsmith.json file exists in dir.
func IsManaged(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, metaFilename))
	return err == nil
}
