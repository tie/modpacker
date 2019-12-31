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
	fs := flag.NewFlagSet(programName, flag.ContinueOnError)
	fs.Bool("h", false, "alias for help")
	fs.Bool("help", false, "print usage")

	cdr := subcommands.NewCommander(fs, programName)
	cdr.Register(&BootstrapCommand{}, "")
	cdr.Register(&CleanCommand{}, "")
	cdr.Register(&CompileCommand{}, "")
	cdr.Register(&DownloadCommand{}, "")
	cdr.Register(&FormatCommand{}, "")
	cdr.Register(&SumsCommand{}, "")
	cdr.Register(cdr.HelpCommand(), "help")
	cdr.Register(cdr.FlagsCommand(), "help")
	cdr.Register(cdr.CommandsCommand(), "help")

	if err := fs.Parse(os.Args[1:]); err != nil {
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
