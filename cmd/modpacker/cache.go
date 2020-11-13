package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/google/subcommands"
)

type CleanCommand struct {
}

func (*CleanCommand) Name() string     { return "clean" }
func (*CleanCommand) Synopsis() string { return "remove cached files" }
func (*CleanCommand) Usage() string {
	return `Usage: modpacker clean

	Cleans cached files.
`
}

func (cmd *CleanCommand) SetFlags(f *flag.FlagSet) {
}

func (cmd *CleanCommand) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	path, err := cacheDir(programName)
	if err != nil {
		log.Printf("cache path: %+v", err)
		return subcommands.ExitFailure
	}
	if err := os.RemoveAll(path); err != nil {
		log.Printf("clean %q: %+v", path, err)
		return subcommands.ExitFailure
	}
	return subcommands.ExitSuccess
}
