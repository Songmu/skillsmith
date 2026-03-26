package skillsmith

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/Songmu/skillsmith/agentskills"
	"golang.org/x/mod/semver"
)

// Smith is the main entry point for the skillsmith skills subcommand.
// Create one with New and call Run to dispatch to the appropriate subcommand.
type Smith struct {
	// FlagSet is the top-level flag set, created by New.
	FlagSet *flag.FlagSet
	// OutWriter is the writer for normal output (defaults to os.Stdout).
	OutWriter io.Writer
	// ErrWriter is the writer for error / diagnostic output (defaults to os.Stderr).
	ErrWriter io.Writer

	name    string // name of the hosting CLI tool
	version string // version stored as provided by the caller
	fs      fs.FS  // auto-detected skills FS
}

// New creates a Smith with the given name, version and skill filesystem.
//
// Version validation: if version does not start with "v", one is prepended
// for validation only. The version is stored as provided.
//
// FS auto-detection: if the root of skillFS contains exactly one directory
// named "skills" (files at root are ignored), that directory is used as the
// skill root via fs.Sub. Otherwise skillFS is used as-is.
func New(name, version string, skillFS fs.FS) (*Smith, error) {
	if skillFS == nil {
		return nil, errors.New("skill filesystem cannot be nil")
	}

	// Validate version using semver (requires "v" prefix).
	vv := version
	if !strings.HasPrefix(vv, "v") {
		vv = "v" + vv
	}
	if !semver.IsValid(vv) {
		return nil, fmt.Errorf("invalid version %q", version)
	}

	// FS auto-detection: strip "skills/" prefix when it is the only directory.
	detectedFS := skillFS
	entries, err := fs.ReadDir(skillFS, ".")
	if err == nil {
		var dirs []string
		for _, e := range entries {
			if e.IsDir() {
				dirs = append(dirs, e.Name())
			}
			// Files at root are intentionally ignored.
		}
		if len(dirs) == 1 && dirs[0] == "skills" {
			if sub, subErr := fs.Sub(skillFS, "skills"); subErr == nil {
				detectedFS = sub
			}
		}
	}

	s := &Smith{
		name:    name,
		version: version,
		fs:      detectedFS,
		FlagSet: flag.NewFlagSet(name+" skills", flag.ContinueOnError),
	}
	s.FlagSet.Usage = func() {
		errW := s.errWriter()
		fmt.Fprintf(errW, "Usage: %s <command> [options]\n\n", s.FlagSet.Name())
		fmt.Fprintf(errW, "Commands:\n")
		for _, cmd := range subcommands {
			fmt.Fprintf(errW, "  %-12s %s\n", cmd.name, cmd.desc)
		}
		fmt.Fprintf(errW, "\nRun '%s <command> --help' for command-specific options.\n", s.FlagSet.Name())
	}
	return s, nil
}

// Name returns the name of the hosting CLI tool.
func (s *Smith) Name() string { return s.name }

// Version returns the version string as provided to New.
func (s *Smith) Version() string { return s.version }

// subcommands lists the available subcommands and their one-line descriptions.
var subcommands = []struct{ name, desc string }{
	{"list", "list embedded skills"},
	{"install", "install skills (skip existing)"},
	{"update", "update skills when version differs"},
	{"reinstall", "reinstall all managed skills"},
	{"uninstall", "uninstall managed skills"},
	{"status", "show install status and version diff"},
}

// Run parses args and dispatches to the matching subcommand.
func (s *Smith) Run(ctx context.Context, args []string) error {
	out := s.outWriter()
	errW := s.errWriter()

	top := s.FlagSet
	top.SetOutput(errW)

	// Parse only the subcommand name; let subcommand parsers handle the rest.
	if err := top.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	subArgs := top.Args()
	if len(subArgs) == 0 {
		top.Usage()
		return nil
	}

	subcmd := subArgs[0]
	rest := subArgs[1:]

	switch subcmd {
	case "list":
		return s.cmdList(ctx, rest, out, errW)
	case "install":
		return s.cmdInstall(ctx, rest, out, errW)
	case "update":
		return s.cmdUpdate(ctx, rest, out, errW)
	case "reinstall":
		return s.cmdReinstall(ctx, rest, out, errW)
	case "uninstall":
		return s.cmdUninstall(ctx, rest, out, errW)
	case "status":
		return s.cmdStatus(ctx, rest, out, errW)
	default:
		fmt.Fprintf(errW, "unknown subcommand %q\n\n", subcmd)
		top.Usage()
		return fmt.Errorf("unknown subcommand %q", subcmd)
	}
}

// Options holds parameters shared by most subcommands.
// CLI frameworks can populate this struct directly and pass it to
// the exported operation methods (Install, Update, etc.) without
// going through Run / flag parsing.
type Options struct {
	DryRun bool
	Prefix string
	Scope  string
	Force  bool
}

// addCommonFlags registers the common flags onto fs.
func addCommonFlags(f *flag.FlagSet, opts *Options) {
	f.BoolVar(&opts.DryRun, "dry-run", false, "print what would happen without making changes")
	f.StringVar(&opts.Prefix, "prefix", "", "skill installation directory (overrides --scope)")
	f.StringVar(&opts.Scope, "scope", "", "target scope: user (~/.agents/skills, default) or repo (<repo-root>/.agents/skills)")
	f.BoolVar(&opts.Force, "force", false, "overwrite unmanaged skills")
}

// installDir returns the effective installation directory for the given options.
// Prefix takes precedence; otherwise the directory is derived from Scope.
func (s *Smith) installDir(opts Options) (string, error) {
	if opts.Prefix != "" {
		return opts.Prefix, nil
	}
	return InstallDirForScope(opts.Scope)
}

// discoverSkills discovers skills from the embedded filesystem, treating
// per-skill errors as non-fatal (they are silently skipped) and returning
// only fatal discovery errors.
func (s *Smith) discoverSkills() ([]*agentskills.Skill, error) {
	skills, discoverErr := agentskills.Discover(s.fs)
	var fatalErr error
	eachError(discoverErr, func(e error) {
		var se *agentskills.SkillError
		if errors.As(e, &se) {
			return
		}
		if fatalErr == nil {
			fatalErr = e
		}
	})
	return skills, fatalErr
}

// List returns the discovered skills from the embedded filesystem.
func (s *Smith) List(ctx context.Context) ([]*agentskills.Skill, error) {
	return s.discoverSkills()
}

// Install installs skills that are not yet present.
func (s *Smith) Install(ctx context.Context, opts Options) (*CopyResult, error) {
	dir, err := s.installDir(opts)
	if err != nil {
		return nil, err
	}
	return CopySkills(s.fs, dir, CopyOptions{
		Mode:    ModeInstall,
		Force:   opts.Force,
		DryRun:  opts.DryRun,
		Name:    s.name,
		Version: s.version,
	})
}

// Update updates managed skills whose version has changed.
func (s *Smith) Update(ctx context.Context, opts Options) (*CopyResult, error) {
	dir, err := s.installDir(opts)
	if err != nil {
		return nil, err
	}
	return CopySkills(s.fs, dir, CopyOptions{
		Mode:    ModeUpdate,
		Force:   opts.Force,
		DryRun:  opts.DryRun,
		Name:    s.name,
		Version: s.version,
	})
}

// Reinstall reinstalls all managed skills regardless of version.
func (s *Smith) Reinstall(ctx context.Context, opts Options) (*CopyResult, error) {
	dir, err := s.installDir(opts)
	if err != nil {
		return nil, err
	}
	return CopySkills(s.fs, dir, CopyOptions{
		Mode:    ModeReinstall,
		Force:   opts.Force,
		DryRun:  opts.DryRun,
		Name:    s.name,
		Version: s.version,
	})
}

// Uninstall removes managed skills.
func (s *Smith) Uninstall(ctx context.Context, opts Options) (*UninstallResult, error) {
	dir, err := s.installDir(opts)
	if err != nil {
		return nil, err
	}

	skills, err := s.discoverSkills()
	if err != nil {
		return nil, err
	}

	result := &UninstallResult{}
	for _, skill := range skills {
		dest := filepath.Join(dir, skill.Dir)
		if !IsManaged(dest) {
			result.Actions = append(result.Actions, UninstallAction{
				Dir:     skill.Dir,
				Action:  "skipped",
				Message: "not managed by skillsmith",
			})
			continue
		}

		if opts.DryRun {
			result.Actions = append(result.Actions, UninstallAction{
				Dir:    skill.Dir,
				Action: "uninstalled",
			})
			continue
		}

		if err := os.RemoveAll(dest); err != nil {
			return result, fmt.Errorf("uninstalling %q: %w", skill.Dir, err)
		}
		result.Actions = append(result.Actions, UninstallAction{
			Dir:    skill.Dir,
			Action: "uninstalled",
		})
	}
	return result, nil
}

// UninstallResult summarizes the outcome of an Uninstall call.
type UninstallResult struct {
	Actions []UninstallAction
}

// UninstallAction describes what happened to a single skill during Uninstall.
type UninstallAction struct {
	Dir     string
	Action  string // "uninstalled", "skipped"
	Message string
}

// StatusResult summarizes the outcome of a Status call.
type StatusResult struct {
	Skills []SkillStatus
}

// SkillStatus describes the installation status of a single skill.
type SkillStatus struct {
	Dir              string
	Installed        bool
	InstalledVersion string
	AvailableVersion string
	UpToDate         bool
	MetadataError    error
}

// Status checks the installation status of skills.
func (s *Smith) Status(ctx context.Context, opts Options) (*StatusResult, error) {
	dir, err := s.installDir(opts)
	if err != nil {
		return nil, err
	}

	skills, err := s.discoverSkills()
	if err != nil {
		return nil, err
	}

	result := &StatusResult{}
	for _, skill := range skills {
		dest := filepath.Join(dir, skill.Dir)
		ss := SkillStatus{
			Dir:              skill.Dir,
			AvailableVersion: s.version,
		}

		if !IsManaged(dest) {
			result.Skills = append(result.Skills, ss)
			continue
		}

		ss.Installed = true
		meta, readErr := ReadMeta(dest)
		if readErr != nil {
			ss.MetadataError = readErr
			result.Skills = append(result.Skills, ss)
			continue
		}

		ss.InstalledVersion = meta.Version
		cmp, ok := compareVersionsSafe(meta.Version, s.version)
		ss.UpToDate = (ok && cmp >= 0) || (!ok && meta.Version == s.version)
		result.Skills = append(result.Skills, ss)
	}
	return result, nil
}

func (s *Smith) outWriter() io.Writer {
	if s.OutWriter != nil {
		return s.OutWriter
	}
	return os.Stdout
}

func (s *Smith) errWriter() io.Writer {
	if s.ErrWriter != nil {
		return s.ErrWriter
	}
	return os.Stderr
}
