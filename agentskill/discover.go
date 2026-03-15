package agentskill

import (
	"errors"
	"fmt"
	"io/fs"
	"path"
)

// Discover enumerates skill directories inside fsys, parses each SKILL.md,
// and returns the valid skills together with any non-fatal errors encountered.
//
// The top-level directory in fsys is expected to contain skill subdirectories,
// for example:
//
//	skills/
//	  mytool-cli/
//	    SKILL.md
//
// Skills whose SKILL.md cannot be parsed or fails validation are omitted from
// the returned slice; a descriptive error is appended to the returned error
// slice instead.
func Discover(fsys fs.FS) ([]Skill, []error) {
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, []error{fmt.Errorf("discover: reading root directory: %w", err)}
	}

	var skills []Skill
	var errs []error

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dirName := entry.Name()
		skillPath := path.Join(dirName, "SKILL.md")

		f, err := fsys.Open(skillPath)
		if err != nil {
			// No SKILL.md in this directory — silently skip.
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			// Other I/O errors are recorded as non-fatal discovery errors.
			errs = append(errs, fmt.Errorf("discover: %s: open error: %w", skillPath, err))
			continue
		}

		skill, err := Parse(f)
		f.Close() //nolint:errcheck
		if err != nil {
			errs = append(errs, fmt.Errorf("discover: %s: parse error: %w", skillPath, err))
			continue
		}

		result := Validate(skill, dirName)
		if !result.OK() {
			for _, e := range result.Errors {
				errs = append(errs, fmt.Errorf("discover: %s: validation error: %s", skillPath, e))
			}
			continue
		}

		skill.Dir = dirName
		skills = append(skills, *skill)
	}

	return skills, errs
}
