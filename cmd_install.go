package skillsmith

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
)

func (s *Smith) cmdInstall(ctx context.Context, args []string, out, errW io.Writer) error {
	f := flag.NewFlagSet("install", flag.ContinueOnError)
	f.SetOutput(errW)
	var opts Options
	addCommonFlags(f, &opts)
	if err := f.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	result, err := s.Install(ctx, opts)
	if err != nil {
		return err
	}

	for _, a := range result.Actions {
		switch a.Action {
		case "installed":
			if opts.DryRun {
				fmt.Fprintf(out, "installed (dry-run): %s\n", a.Dir)
			} else {
				fmt.Fprintf(out, "installed: %s\n", a.Dir)
			}
		case "skipped":
			fmt.Fprintf(out, "skipped:   %s — %s\n", a.Dir, a.Message)
		case "warned":
			fmt.Fprintf(errW, "warning:   %s — %s\n", a.Dir, a.Message)
		}
	}

	if opts.DryRun {
		fmt.Fprintln(out, "[dry-run] no changes were made")
	}
	return nil
}
