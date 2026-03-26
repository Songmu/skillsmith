package skillsmith

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
)

func (s *Smith) cmdList(ctx context.Context, args []string, out, errW io.Writer) error {
	f := flag.NewFlagSet("list", flag.ContinueOnError)
	f.SetOutput(errW)
	if err := f.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	skills, err := s.List(ctx)
	if err != nil {
		return err
	}

	if len(skills) == 0 {
		fmt.Fprintln(out, "no skills found")
		return nil
	}

	for _, sk := range skills {
		if sk.Description != "" {
			fmt.Fprintf(out, "%-30s %s\n", sk.Dir, sk.Description)
		} else {
			fmt.Fprintln(out, sk.Dir)
		}
	}
	return nil
}
