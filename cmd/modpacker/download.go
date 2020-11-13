package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"path/filepath"

	"github.com/google/subcommands"

	"github.com/akrylysov/pogreb"
	"github.com/akrylysov/pogreb/fs"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/osfs"

	"github.com/tie/modpacker/fetcher"
	"github.com/tie/modpacker/pack"
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

func (cmd *DownloadCommand) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&cmd.DisableCache, "nocache", false, "disable filesystem cache")
}

func (cmd *DownloadCommand) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	paths := f.Args()

	ms, ok := parseManifests(paths)
	if !ok {
		return subcommands.ExitFailure
	}

	var cachePath string
	if !cmd.DisableCache {
		var err error
		cachePath, err = makeCache(programName)
		if err != nil {
			log.Printf("make cache: %+v", err)
			return subcommands.ExitFailure
		}
	}

	var cachefs billy.Filesystem
	if !cmd.DisableCache {
		cachefs = osfs.New(cachePath)
	} else {
		cachefs = memfs.New()
	}

	var db *pogreb.DB
	if !cmd.DisableCache {
		var err error
		db, err = pogreb.Open(filepath.Join(cachePath, "db"), nil)
		if err != nil {
			log.Printf("open pogreb: %+v", err)
			return subcommands.ExitFailure
		}
	} else {
		// BUG pogreb.Open always calls os.MkdirAll
		var err error
		db, err = pogreb.Open(".", &pogreb.Options{
			FileSystem: fs.Mem,
		})
		if err != nil {
			panic(err)
		}
	}

	c := http.Client{}

	fetcher := fetcher.Fetcher{
		Database: db,
		Files:    cachefs,
		Client:   &c,
	}

	for _, mod := range pack.ModList(ms) {
		err := fetcher.Cache(mod)
		if err != nil {
			log.Printf("download %q mod: %+v", mod.Method, err)
			return subcommands.ExitFailure
		}
	}

	return subcommands.ExitSuccess
}
