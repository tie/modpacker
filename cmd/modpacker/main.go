package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/google/subcommands"
)

const (
	programName     = "modpacker"
	defaultManifest = "base.pack"
)

func init() {
	log.SetFlags(0)
}

func main() {
	f := flag.NewFlagSet(programName, flag.ContinueOnError)
	f.Bool("h", false, "alias for help")
	f.Bool("help", false, "print usage")

	cdr := subcommands.NewCommander(f, programName)
	cdr.Register(&BootstrapCommand{}, "")
	cdr.Register(&CleanCommand{}, "")
	cdr.Register(&CompileCommand{}, "")
	cdr.Register(&DownloadCommand{}, "")
	cdr.Register(&FormatCommand{}, "")
	cdr.Register(&ModlistCommand{}, "")
	cdr.Register(&SumsCommand{}, "")
	cdr.Register(cdr.HelpCommand(), "help")
	cdr.Register(cdr.FlagsCommand(), "help")
	cdr.Register(cdr.CommandsCommand(), "help")

	if err := f.Parse(os.Args[1:]); err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	switch cdr.Execute(ctx) {
	case subcommands.ExitFailure:
		os.Exit(1)
	case subcommands.ExitUsageError:
		os.Exit(2)
	}
}
