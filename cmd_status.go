package skillsmith

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"path/filepath"

	"github.com/Songmu/skillsmith/agentskill"
)

func (s *Smith) cmdStatus(_ context.Context, args []string, out, errW io.Writer) error {
	f := flag.NewFlagSet("status", flag.ContinueOnError)
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

	skills, errs := agentskill.Discover(s.FS)
	for _, e := range errs {
		fmt.Fprintf(errW, "warning: %v\n", e)
	}

	for _, skill := range skills {
		dest := filepath.Join(dir, skill.Dir)
		if !IsManaged(dest) {
			fmt.Fprintf(out, "%-30s not installed\n", skill.Dir)
			continue
		}

		meta, err := ReadMeta(dest)
		if err != nil {
			fmt.Fprintf(out, "%-30s installed (metadata unreadable: %v)\n", skill.Dir, err)
			continue
		}

		if meta.Version == s.Version {
			fmt.Fprintf(out, "%-30s installed %s (up to date)\n", skill.Dir, meta.Version)
		} else {
			fmt.Fprintf(out, "%-30s installed %s → available %s\n", skill.Dir, meta.Version, s.Version)
		}
	}
	return nil
}
