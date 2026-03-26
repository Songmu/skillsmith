package skillsmith

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
)

func (s *Smith) cmdReinstall(ctx context.Context, args []string, out, errW io.Writer) error {
	f := flag.NewFlagSet("reinstall", flag.ContinueOnError)
	f.SetOutput(errW)
	var opts Options
	addCommonFlags(f, &opts)
	if err := f.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	result, err := s.Reinstall(ctx, opts)
	if err != nil {
		return err
	}

	for _, a := range result.Actions {
		switch a.Action {
		case "reinstalled":
			if opts.DryRun {
				fmt.Fprintf(out, "reinstalled (dry-run): %s\n", a.Dir)
			} else {
				fmt.Fprintf(out, "reinstalled: %s\n", a.Dir)
			}
		case "warned":
			fmt.Fprintf(errW, "warning:     %s — %s\n", a.Dir, a.Message)
		}
	}

	if opts.DryRun {
		fmt.Fprintln(out, "[dry-run] no changes were made")
	}
	return nil
}
