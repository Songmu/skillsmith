package skillsmith

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"

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
	version string // version without "v" prefix
	fs      fs.FS  // auto-detected skills FS
}

// New creates a Smith with the given name, version and skill filesystem.
//
// Version validation: if version does not start with "v", one is prepended
// for validation only. The version is stored without the "v" prefix.
//
// FS auto-detection: if the root of skillFS contains exactly one directory
// named "skills" (files at root are ignored), that directory is used as the
// skill root via fs.Sub. Otherwise skillFS is used as-is.
func New(name, version string, skillFS fs.FS) (*Smith, error) {
	// Validate version using semver (requires "v" prefix).
	vv := version
	if !strings.HasPrefix(vv, "v") {
		vv = "v" + vv
	}
	if !semver.IsValid(vv) {
		return nil, fmt.Errorf("invalid version %q", version)
	}
	// Store without "v" prefix.
	stored := strings.TrimPrefix(vv, "v")

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
		version: stored,
		fs:      detectedFS,
		FlagSet: flag.NewFlagSet(name+" skills", flag.ContinueOnError),
	}
	return s, nil
}

// Name returns the name of the hosting CLI tool.
func (s *Smith) Name() string { return s.name }

// Version returns the version string (without the "v" prefix).
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
	top.Usage = func() {
		fmt.Fprintf(errW, "Usage: %s <command> [options]\n\n", top.Name())
		fmt.Fprintf(errW, "Commands:\n")
		for _, cmd := range subcommands {
			fmt.Fprintf(errW, "  %-12s %s\n", cmd.name, cmd.desc)
		}
		fmt.Fprintf(errW, "\nRun '%s <command> --help' for command-specific options.\n", top.Name())
	}

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

// commonFlags holds flags shared by most subcommands.
type commonFlags struct {
	dryRun bool
	prefix string
	scope  string
	force  bool
}

// addCommonFlags registers the common flags onto fs.
func addCommonFlags(f *flag.FlagSet, cf *commonFlags) {
	f.BoolVar(&cf.dryRun, "dry-run", false, "print what would happen without making changes")
	f.StringVar(&cf.prefix, "prefix", "", "skill installation directory (overrides --scope)")
	f.StringVar(&cf.scope, "scope", "", "target scope: user (~/.agents/skills, default) or repo (.agents/skills)")
	f.BoolVar(&cf.force, "force", false, "overwrite unmanaged skills")
}

// installDir returns the effective installation directory for the given flags.
// --prefix takes precedence; otherwise the directory is derived from --scope.
func (s *Smith) installDir(cf commonFlags) (string, error) {
	if cf.prefix != "" {
		return cf.prefix, nil
	}
	return InstallDirForScope(cf.scope)
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
