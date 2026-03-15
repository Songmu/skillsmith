package skillsmith

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
)

// Smith is the main entry point for the skillsmith skills subcommand.
// Embed an fs.FS containing skill directories and call Run to dispatch
// to the appropriate subcommand.
type Smith struct {
	// FS holds the embedded or in-memory skill files.
	FS fs.FS
	// Version is the version string of the hosting CLI tool.
	Version string
	// Name is the name of the hosting CLI tool (written to .skillsmith.json).
	Name string
	// OutWriter is the writer for normal output (defaults to os.Stdout).
	OutWriter io.Writer
	// ErrWriter is the writer for error / diagnostic output (defaults to os.Stderr).
	ErrWriter io.Writer
}

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

	top := flag.NewFlagSet("skills", flag.ContinueOnError)
	top.SetOutput(errW)
	top.Usage = func() {
		fmt.Fprintf(errW, "Usage: skills <command> [options]\n\n")
		fmt.Fprintf(errW, "Commands:\n")
		for _, cmd := range subcommands {
			fmt.Fprintf(errW, "  %-12s %s\n", cmd.name, cmd.desc)
		}
		fmt.Fprintf(errW, "\nRun 'skills <command> --help' for command-specific options.\n")
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
	f.StringVar(&cf.scope, "scope", "", "target scope: user (~/.agents/skills, default) or repo (<repo-root>/.agents/skills)")
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
