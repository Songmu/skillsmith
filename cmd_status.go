package skillsmith

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
)

func (s *Smith) cmdStatus(ctx context.Context, args []string, out, errW io.Writer) error {
	f := flag.NewFlagSet("status", flag.ContinueOnError)
	f.SetOutput(errW)
	var opts Options
	addCommonFlags(f, &opts)
	if err := f.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	result, err := s.Status(ctx, opts)
	if err != nil {
		return err
	}

	for _, ss := range result.Skills {
		switch {
		case !ss.Installed:
			fmt.Fprintf(out, "%-30s not installed\n", ss.Dir)
		case ss.MetadataError != nil:
			fmt.Fprintf(out, "%-30s installed (metadata unreadable: %v)\n", ss.Dir, ss.MetadataError)
		case ss.UpToDate:
			fmt.Fprintf(out, "%-30s installed %s (up to date)\n", ss.Dir, ss.InstalledVersion)
		default:
			fmt.Fprintf(out, "%-30s installed %s → available %s\n", ss.Dir, ss.InstalledVersion, ss.AvailableVersion)
		}
	}
	return nil
}
