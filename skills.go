package skillsmith

import (
	"embed"
	"io/fs"
)

//go:embed skills/**
var skillsFS embed.FS

// DemoFS returns the embedded demo skills filesystem with the "skills/"
// prefix stripped so callers receive a top-level directory of skill dirs.
func DemoFS() fs.FS {
	sub, err := fs.Sub(skillsFS, "skills")
	if err != nil {
		// This can only happen if the embed path is wrong, which is a
		// programming error caught at build time.
		panic(err)
	}
	return sub
}
