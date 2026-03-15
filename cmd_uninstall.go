package skillsmith

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Songmu/skillsmith/agentskills"
)

func (s *Smith) cmdUninstall(_ context.Context, args []string, out, errW io.Writer) error {
	f := flag.NewFlagSet("uninstall", flag.ContinueOnError)
	f.SetOutput(errW)
	var cf commonFlags
	addCommonFlags(f, &cf)
	if err := f.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	dir, err := s.installDir(cf)
	if err != nil {
		return err
	}

	skills, discoverErr := agentskills.Discover(s.fs)
	if discoverErr != nil {
		var skillErr *agentskills.SkillError
		if !errors.As(discoverErr, &skillErr) {
			return discoverErr
		}
		eachError(discoverErr, func(e error) {
			fmt.Fprintf(errW, "warning: %v\n", e)
		})
	}

	for _, skill := range skills {
		dest := filepath.Join(dir, skill.Dir)
		if !IsManaged(dest) {
			fmt.Fprintf(out, "skipped:     %s — not managed by skillsmith\n", skill.Dir)
			continue
		}

		if cf.dryRun {
			fmt.Fprintf(out, "uninstalled (dry-run): %s\n", skill.Dir)
			continue
		}

		if err := os.RemoveAll(dest); err != nil {
			return fmt.Errorf("uninstalling %q: %w", skill.Dir, err)
		}
		fmt.Fprintf(out, "uninstalled: %s\n", skill.Dir)
	}

	if cf.dryRun {
		fmt.Fprintln(out, "[dry-run] no changes were made")
	}
	return nil
}
