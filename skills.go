package skillsmith

import (
	"embed"
	"io/fs"
)

//go:embed skills/**
var skillsFS embed.FS

// DemoFS returns the embedded demo skills filesystem.
// The "skills/" prefix is stripped automatically by Smith.skillsFS().
func DemoFS() fs.FS {
	return skillsFS
}
