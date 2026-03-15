package skillsmith

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
)

func (s *Smith) cmdUpdate(_ context.Context, args []string, out, errW io.Writer) error {
	f := flag.NewFlagSet("update", flag.ContinueOnError)
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

	result, err := CopySkills(s.FS, dir, CopyOptions{
		Mode:    ModeUpdate,
		Force:   cf.force,
		DryRun:  cf.dryRun,
		Name:    s.Name,
		Version: s.Version,
	})
	if err != nil {
		return err
	}

	for _, a := range result.Actions {
		switch a.Action {
		case "updated":
			if cf.dryRun {
				fmt.Fprintf(out, "updated (dry-run): %s\n", a.Dir)
			} else {
				fmt.Fprintf(out, "updated:   %s\n", a.Dir)
			}
		case "skipped":
			fmt.Fprintf(out, "skipped:   %s — %s\n", a.Dir, a.Message)
		case "warned":
			fmt.Fprintf(errW, "warning:   %s — %s\n", a.Dir, a.Message)
		}
	}

	if cf.dryRun {
		fmt.Fprintln(out, "[dry-run] no changes were made")
	}
	return nil
}
