package main

import (
	"context"
	"flag"
	"log"
	"net/http"

	"github.com/google/subcommands"

	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/memfs"
	"gopkg.in/src-d/go-billy.v4/osfs"

	"github.com/tie/modpacker/fetcher"
)

type DownloadCommand struct {
	DisableCache bool
}

func (*DownloadCommand) Name() string     { return "download" }
func (*DownloadCommand) Synopsis() string { return "download mods to local cache" }
func (*DownloadCommand) Usage() string {
	return `Usage: modpacker download [-nocache] [manifest paths]

	Downloads mods from manifest to local cache.
	Useful for pre-filling local cache and checking download availability.

Flags:
`
}

func (cmd *DownloadCommand) SetFlags(fs *flag.FlagSet) {
	fs.BoolVar(&cmd.DisableCache, "nocache", false, "disable filesystem cache")
}

func (cmd *DownloadCommand) Execute(ctx context.Context, fs *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	paths := fs.Args()
	m, ok := mergeManifests(paths)
	if !ok {
		return subcommands.ExitFailure
	}

	var cacheDir billy.Filesystem
	if !cmd.DisableCache {
		cache, err := makeCache(programName)
		if err != nil {
			log.Printf("make cache: %+v", err)
			return subcommands.ExitFailure
		}
		cacheDir = osfs.New(cache)
	} else {
		cacheDir = memfs.New()
	}
	c := http.Client{}
	fetcher := fetcher.Fetcher{
		Files:  cacheDir,
		Client: &c,
	}

	for _, mod := range m.ModList() {
		err := fetcher.Cache(mod)
		if err != nil {
			log.Printf("download %q mod: %+v", mod.Method, err)
			return subcommands.ExitFailure
		}
	}

	return subcommands.ExitSuccess
}
