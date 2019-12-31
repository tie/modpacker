package main

import (
	"archive/zip"
	"bufio"
	"context"
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/google/subcommands"

	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/memfs"
	"gopkg.in/src-d/go-billy.v4/osfs"

	"github.com/tie/modpacker"
)

type CompileCommand struct {
	OutputPath   string
	DisableCache bool
}

func (*CompileCommand) Name() string     { return "compile" }
func (*CompileCommand) Synopsis() string { return "compile the modpack" }
func (*CompileCommand) Usage() string {
	return `Usage: modpacker compile [-o modpack.zip] [-nocache] [manifest paths]

	Compiles the modpack from manifests. The output is a zip archive
	containing files specified by "mod" blocks. For each corresponding
	"check" block the integrity of the mods is verified. Use "sums"
	subcommand to generate sums manifest for an existing set of files.

Flags:
`
}

func (cmd *CompileCommand) SetFlags(fs *flag.FlagSet) {
	fs.BoolVar(&cmd.DisableCache, "nocache", false, "disable filesystem cache")
	fs.StringVar(&cmd.OutputPath, "o", "modpack.zip", "modpack output path")
}

func (cmd *CompileCommand) Execute(ctx context.Context, fs *flag.FlagSet, args ...interface{}) (rc subcommands.ExitStatus) {
	paths := fs.Args()
	if len(paths) <= 0 {
		paths = []string{defaultManifest}
	}

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

	fpath := cmd.OutputPath
	f, err := os.Create(fpath)
	if err != nil {
		log.Printf("create %q: %+v", fpath, err)
		return subcommands.ExitFailure
	}
	defer func() {
		err := f.Close()
		if err != nil {
			log.Printf("close %q: %+v", fpath, err)
			rc = subcommands.ExitFailure
		}
	}()

	w := bufio.NewWriter(f)
	z := zip.NewWriter(w)
	defer func() {
		err := z.Close()
		if err != nil {
			log.Printf("close archive: %+v", err)
			rc = subcommands.ExitFailure
		}
	}()

	c := &http.Client{}
	dl := &modpacker.Downloader{
		Files:  cacheDir,
		Client: c,
	}
	b := &modpacker.Builder{
		Downloader: dl,
		Pack:       z,
	}

	for _, mod := range m.ModList() {
		err := b.Add(mod)
		if err != nil {
			log.Printf("add %q mod: %+v", mod.Method, err)
			return subcommands.ExitFailure
		}
	}

	return subcommands.ExitSuccess
}
