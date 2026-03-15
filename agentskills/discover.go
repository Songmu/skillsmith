package agentskills

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
// the returned slice; each such failure is collected into the returned error as
// a [SkillError] (joined via [errors.Join]) so callers can use [errors.As] to
// retrieve per-skill directory information.
func Discover(fsys fs.FS) ([]*Skill, error) {
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, fmt.Errorf("discover: reading root directory: %w", err)
	}

	var skills []*Skill
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
			// Other I/O errors are recorded as non-fatal per-skill errors.
			errs = append(errs, &SkillError{Dir: dirName, Err: fmt.Errorf("open error: %w", err)})
			continue
		}

		skill, err := Parse(f)
		f.Close() //nolint:errcheck
		if err != nil {
			errs = append(errs, &SkillError{Dir: dirName, Err: fmt.Errorf("parse error: %w", err)})
			continue
		}

		result := Validate(skill, dirName)
		if !result.OK() {
			for _, e := range result.Errors {
				errs = append(errs, &SkillError{Dir: dirName, Err: fmt.Errorf("validation error: %s", e)})
			}
			continue
		}

		skill.Dir = dirName
		skills = append(skills, skill)
	}

	return skills, errors.Join(errs...)
}
