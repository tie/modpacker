package main

import (
	"archive/zip"
	"bufio"
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/subcommands"

	"github.com/akrylysov/pogreb"
	"github.com/akrylysov/pogreb/fs"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/osfs"

	"github.com/tie/modpacker/builder"
	"github.com/tie/modpacker/builder/archive"
	"github.com/tie/modpacker/builder/curse"
	"github.com/tie/modpacker/fetcher"
	"github.com/tie/modpacker/pack"
)

const (
	OutputModeStandalone = "standalone"
	OutputModeCurse      = "curse"
)

type CompileCommand struct {
	OutputMode   string
	OutputPath   string
	DisableCache bool
}

func (*CompileCommand) Name() string     { return "compile" }
func (*CompileCommand) Synopsis() string { return "compile the modpack" }
func (*CompileCommand) Usage() string {
	return `Usage: modpacker compile [-o modpack.zip] [-mode standalone] [-nocache] [manifest paths]

	Compiles the modpack from manifests. The output is a zip archive
	containing files specified by "mod" blocks. For each corresponding
	"check" block the integrity of the mods is verified. Use "sums"
	subcommand to generate sums manifest for an existing set of files.

        The layout of the files in output archive is specified by -mode
        option. The supported modes are:

            standalone
                Standalone archive containing all specified files.
                This is the default mode.
            curse
                Archive compatible with CurseForge/Twitch launcher.
                In this mode "mod" blocks with "curse" method will be added
                to manifest.json file instead of being downloaded. The sums
                for those blocks are therefore ignored.

Flags:
`
}

func (cmd *CompileCommand) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&cmd.DisableCache, "nocache", false, "disable filesystem cache")
	f.StringVar(&cmd.OutputPath, "o", "modpack.zip", "modpack output path")
	f.StringVar(&cmd.OutputMode, "mode", OutputModeStandalone, "modpack output mode")
}

func (cmd *CompileCommand) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) (rc subcommands.ExitStatus) {
	paths := f.Args()
	if len(paths) <= 0 {
		paths = []string{defaultManifest}
	}

	switch cmd.OutputMode {
	case OutputModeStandalone:
	case OutputModeCurse:
	default:
		log.Printf("unknown output mode: %q", cmd.OutputMode)
		return subcommands.ExitFailure
	}

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

	fpath := cmd.OutputPath
	file, err := os.Create(fpath)
	if err != nil {
		log.Printf("create %q: %+v", fpath, err)
		return subcommands.ExitFailure
	}
	defer func() {
		err := file.Close()
		if err != nil {
			log.Printf("close %q: %+v", fpath, err)
			rc = subcommands.ExitFailure
		}
	}()

	w := bufio.NewWriter(file)
	z := zip.NewWriter(w)
	defer func() {
		err := z.Close()
		if err != nil {
			log.Printf("close archive: %+v", err)
			rc = subcommands.ExitFailure
		}
	}()

	var b builder.Builder
	switch cmd.OutputMode {
	case OutputModeStandalone:
		b = archive.NewArchiveBuilder(&fetcher, z)
	case OutputModeCurse:
		b = curse.NewCurseBuilder(&fetcher, z)
	}

	for _, mod := range pack.ModList(ms) {
		err := b.Add(mod)
		if err != nil {
			log.Printf("add %q mod: %+v", mod.Method, err)
			return subcommands.ExitFailure
		}
	}

	return subcommands.ExitSuccess
}
