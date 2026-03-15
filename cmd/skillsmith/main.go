package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/Songmu/skillsmith"
)

func main() {
	log.SetFlags(0)

	args := os.Args[1:]
	if len(args) > 0 && args[0] == "skills" {
		s, err := skillsmith.New("skillsmith", skillsmith.Version(), skillsmith.DemoFS())
		if err != nil {
			log.Fatal(err)
		}
		err = s.Run(context.Background(), args[1:])
		if err != nil && err != flag.ErrHelp {
			log.Println(err)
			os.Exit(1)
		}
		return
	}

	err := skillsmith.Run(context.Background(), args, os.Stdout, os.Stderr)
	if err != nil && err != flag.ErrHelp {
		log.Println(err)
		exitCode := 1
		if ecoder, ok := err.(interface{ ExitCode() int }); ok {
			exitCode = ecoder.ExitCode()
		}
		os.Exit(exitCode)
	}
}
