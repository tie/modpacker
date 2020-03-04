package main

import (
	"context"
	"flag"
	"log"
	"net/http"

	"github.com/google/subcommands"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"

	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/memfs"
	"gopkg.in/src-d/go-billy.v4/osfs"

	"github.com/tie/internal/renameio"

	"github.com/tie/modpacker/fetcher"
	"github.com/tie/modpacker/models"
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

func (cmd *SumsCommand) SetFlags(fs *flag.FlagSet) {
	fs.BoolVar(&cmd.DisableCache, "nocache", false, "disable filesystem cache")
	fs.StringVar(&cmd.OutputPath, "o", "sums.hcl", "manifest output path")
}

func (cmd *SumsCommand) Execute(ctx context.Context, fs *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
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
	c := http.Client{}
	fetcher := fetcher.Fetcher{
		Files:  cacheDir,
		Client: &c,
	}

	f := hclwrite.NewEmptyFile()
	body := f.Body()
	sb := SumsBuilder{
		Body: body,
	}

	for _, mod := range m.ModList() {
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
	outSrc := f.Bytes()
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

func (b *SumsBuilder) Add(m models.Mod, sums []string) {
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
