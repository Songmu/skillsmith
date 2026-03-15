package skillsmith

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/Songmu/skillsmith/agentskill"
)

// CopyMode controls the install/update/reinstall behavior.
type CopyMode int

const (
	// ModeInstall installs new skills; skips existing managed skills.
	ModeInstall CopyMode = iota
	// ModeUpdate re-installs managed skills whose version has changed.
	ModeUpdate
	// ModeReinstall overwrites all managed skills regardless of version.
	ModeReinstall
)

// CopyOptions configures the behavior of CopySkills.
type CopyOptions struct {
	// Mode determines whether this is an install, update, or reinstall.
	Mode CopyMode
	// Force allows overwriting unmanaged (externally placed) skills.
	Force bool
	// DryRun reports what would happen without making any changes.
	DryRun bool
	// Name is the tool name written to .skillsmith.json (installedBy).
	Name string
	// Version is the tool version written to .skillsmith.json.
	Version string
}

// SkillAction describes what happened to a single skill during CopySkills.
type SkillAction struct {
	Dir     string
	Action  string // "installed", "updated", "reinstalled", "skipped", "warned"
	Message string
}

// CopyResult summarizes the outcome of a CopySkills call.
type CopyResult struct {
	Actions []SkillAction
}

// Installed returns the actions whose Action is "installed", "updated", or "reinstalled".
func (r *CopyResult) Installed() []SkillAction {
	return r.filter("installed", "updated", "reinstalled")
}

// Skipped returns the actions whose Action is "skipped".
func (r *CopyResult) Skipped() []SkillAction {
	return r.filter("skipped")
}

// Warned returns the actions whose Action is "warned".
func (r *CopyResult) Warned() []SkillAction {
	return r.filter("warned")
}

func (r *CopyResult) filter(actions ...string) []SkillAction {
	set := make(map[string]bool, len(actions))
	for _, a := range actions {
		set[a] = true
	}
	var out []SkillAction
	for _, a := range r.Actions {
		if set[a.Action] {
			out = append(out, a)
		}
	}
	return out
}

func (r *CopyResult) add(dir, action, msg string) {
	r.Actions = append(r.Actions, SkillAction{Dir: dir, Action: action, Message: msg})
}

// eachError calls fn for each individual error in err. If err wraps multiple
// errors (e.g. from [errors.Join]), fn is called for each element. If err is
// nil, fn is never called.
func eachError(err error, fn func(error)) {
	if err == nil {
		return
	}
	type multiErr interface {
		Unwrap() []error
	}
	if me, ok := err.(multiErr); ok {
		for _, e := range me.Unwrap() {
			fn(e)
		}
		return
	}
	fn(err)
}

// CopySkills copies skills from src (an fs.FS whose top-level directories are
// skill directories) into destDir.
func CopySkills(src fs.FS, destDir string, opts CopyOptions) (*CopyResult, error) {
	skills, discoverErr := agentskill.Discover(src)
	result := &CopyResult{}

	var fatalErr error
	eachError(discoverErr, func(e error) {
		var se *agentskill.SkillError
		if errors.As(e, &se) {
			result.add(se.Dir, "warned", se.Err.Error())
			return
		}
		// Treat non-SkillError discovery failures as fatal.
		if fatalErr == nil {
			fatalErr = e
		} else {
			fatalErr = errors.Join(fatalErr, e)
		}
	})
	if fatalErr != nil {
		return result, fmt.Errorf("discovering skills: %w", fatalErr)
	}

	for _, skill := range skills {
		action, msg, err := copySkill(src, destDir, skill, opts)
		if err != nil {
			return result, fmt.Errorf("copying skill %q: %w", skill.Dir, err)
		}
		result.add(skill.Dir, action, msg)
	}

	return result, nil
}

// copySkill handles the copy logic for a single skill directory.
// It returns the action taken ("installed", "updated", "reinstalled", "skipped", "warned").
func copySkill(src fs.FS, destDir string, skill *agentskill.Skill, opts CopyOptions) (action, msg string, err error) {
	dest := filepath.Join(destDir, skill.Dir)
	managed := IsManaged(dest)

	switch opts.Mode {
	case ModeInstall:
		if _, statErr := os.Stat(dest); statErr == nil {
			// Destination exists.
			if !managed {
				if !opts.Force {
					return "warned", fmt.Sprintf("skill %q exists but is not managed by skillsmith; use --force to overwrite", skill.Dir), nil
				}
				// Force overwrite of unmanaged skill.
			} else {
				// Managed and exists — skip on install.
				return "skipped", fmt.Sprintf("skill %q already installed (use 'update' or 'reinstall' to refresh)", skill.Dir), nil
			}
		}

	case ModeUpdate:
		if !managed {
			// Not managed — skip with a user-visible reason.
			return "skipped", fmt.Sprintf("skill %q is not managed by skillsmith", skill.Dir), nil
		}
		meta, readErr := ReadMeta(dest)
		if readErr == nil && meta.Version == opts.Version {
			// Same version — nothing to do.
			return "skipped", fmt.Sprintf("skill %q is already at version %q", skill.Dir, opts.Version), nil
		}

	case ModeReinstall:
		if !managed {
			if !opts.Force {
				return "warned", fmt.Sprintf("skill %q is not managed by skillsmith; use --force to overwrite", skill.Dir), nil
			}
		}
	}

	if opts.DryRun {
		label := "installed"
		if opts.Mode == ModeUpdate {
			label = "updated"
		} else if opts.Mode == ModeReinstall {
			label = "reinstalled"
		}
		return label, fmt.Sprintf("[dry-run] would %s skill %q", label, skill.Dir), nil
	}

	// Perform the actual file copy.
	if err := copyFSDir(src, skill.Dir, dest); err != nil {
		return "", "", err
	}

	// Write metadata.
	meta := &SkillMeta{
		InstalledBy: opts.Name,
		Version:     opts.Version,
		InstalledAt: time.Now().UTC(),
	}
	if err := WriteMeta(dest, meta); err != nil {
		return "", "", fmt.Errorf("writing metadata for %q: %w", skill.Dir, err)
	}

	label := "installed"
	if opts.Mode == ModeUpdate {
		label = "updated"
	} else if opts.Mode == ModeReinstall {
		label = "reinstalled"
	}
	return label, "", nil
}

// copyFSDir copies the directory srcDir from fsys into destDir on disk,
// preserving the directory structure.
func copyFSDir(fsys fs.FS, srcDir, destDir string) error {
	return fs.WalkDir(fsys, srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, relErr := filepath.Rel(srcDir, path)
		if relErr != nil {
			return relErr
		}
		dest := filepath.Join(destDir, rel)

		if d.IsDir() {
			return os.MkdirAll(dest, 0o755)
		}

		return copyFile(fsys, path, dest)
	})
}

// copyFile copies a single file from fsys to destPath.
func copyFile(fsys fs.FS, srcPath, destPath string) error {
	src, err := fsys.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close() //nolint:errcheck

	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return err
	}

	dst, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer dst.Close() //nolint:errcheck

	_, err = io.Copy(dst, src)
	return err
}
