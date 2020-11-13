package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"path/filepath"

	"github.com/google/subcommands"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"

	"github.com/akrylysov/pogreb"
	"github.com/akrylysov/pogreb/fs"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/osfs"

	"github.com/tie/internal/renameio"

	"github.com/tie/modpacker/fetcher"
	"github.com/tie/modpacker/modpacker"
	"github.com/tie/modpacker/pack"
)

type SumsCommand struct {
	OutputPath   string
	DisableCache bool
}

func (*SumsCommand) Name() string     { return "sums" }
func (*SumsCommand) Synopsis() string { return "generate checksum manifest" }
func (*SumsCommand) Usage() string {
	return `Usage: modpacker sums [-o sums.pack] [-nocache] [manifest paths]

	Generates checksum manifest for all mods. The resulting manifest will contain
	"check" block for each distinct mod from input manifests. That is,
	adding the same mod to different paths won’t produce multiple "check" blocks.

Flags:
`
}

func (cmd *SumsCommand) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&cmd.DisableCache, "nocache", false, "disable filesystem cache")
	f.StringVar(&cmd.OutputPath, "o", "sums.pack", "manifest output path")
}

func (cmd *SumsCommand) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	paths := f.Args()
	if len(paths) <= 0 {
		paths = []string{defaultManifest}
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

	sumsFile := hclwrite.NewEmptyFile()
	body := sumsFile.Body()
	sb := SumsBuilder{
		Body: body,
	}

	for _, mod := range pack.ModList(ms) {
		sums, err := fetcher.Sums(mod)
		if err != nil {
			log.Printf("sum %q mod: %+v", mod.Method, err)
			return subcommands.ExitFailure
		}
		if len(sums) <= 0 {
			continue
		}
		sb.Add(mod, sums)
	}

	fpath := cmd.OutputPath
	outSrc := sumsFile.Bytes()
	if err := renameio.WriteFile(fpath, outSrc, 0644); err != nil {
		log.Printf("write file %q: %+v", fpath, err)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}

type SumsBuilder struct {
	*hclwrite.Body
	Length int
}

func (b *SumsBuilder) Add(m modpacker.Mod, sums []string) {
	if b.Length > 0 {
		b.AppendNewline()
	}
	b.Length++

	block := b.AppendNewBlock("check", nil)
	body := block.Body()

	method := cty.StringVal(m.Method)
	body.SetAttributeValue("method", method)

	if f := m.File; f != "" {
		file := cty.StringVal(f)
		body.SetAttributeValue("file", file)
	}

	if id := int64(m.ProjectID); id > 0 {
		projectID := cty.NumberIntVal(id)
		body.SetAttributeValue("projectID", projectID)
	}

	if id := int64(m.FileID); id > 0 {
		fileID := cty.NumberIntVal(int64(m.FileID))
		body.SetAttributeValue("fileID", fileID)
	}

	vals := make([]cty.Value, len(sums))
	for i, sum := range sums {
		vals[i] = cty.StringVal(sum)
	}
	list := cty.ListVal(vals)
	body.SetAttributeValue("sums", list)
}
