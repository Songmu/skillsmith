package skillsmith

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/Songmu/skillsmith/agentskills"
)

func (s *Smith) cmdList(_ context.Context, args []string, out, errW io.Writer) error {
	f := flag.NewFlagSet("list", flag.ContinueOnError)
	f.SetOutput(errW)
	if err := f.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	skills, discoverErr := agentskills.Discover(s.fs)
	var fatalErr error
	eachError(discoverErr, func(e error) {
		var se *agentskills.SkillError
		if errors.As(e, &se) {
			fmt.Fprintf(errW, "warning: %v\n", e)
			return
		}
		// Treat non-*SkillError errors as fatal.
		if fatalErr == nil {
			fatalErr = e
		}
	})
	if fatalErr != nil {
		return fatalErr
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
