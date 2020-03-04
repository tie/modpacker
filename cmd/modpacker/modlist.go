package main

import (
	"bytes"
	"context"
	"flag"
	"html/template"
	"log"

	"github.com/google/subcommands"

	"github.com/tie/internal/renameio"
)

const modlistTemplate = `<!doctype html>
<ul>
{{range .}}
<li>TODO {{.}}</li>
{{end}}
</ul>
`

type ModlistCommand struct {
	OutputPath   string
	DisableCache bool
}

func (*ModlistCommand) Name() string     { return "modlist" }
func (*ModlistCommand) Synopsis() string { return "generate modlist page" }
func (*ModlistCommand) Usage() string {
	return `Usage: modpacker sums [-o sums.pack] [-nocache] [manifest paths]

	Generates modlist page for all CurseForge mods.

Flags:
`
}

func (cmd *ModlistCommand) SetFlags(fs *flag.FlagSet) {
	fs.BoolVar(&cmd.DisableCache, "nocache", false, "disable filesystem cache")
	fs.StringVar(&cmd.OutputPath, "o", "modlist.html", "modlist page output path")
}

func (cmd *ModlistCommand) Execute(ctx context.Context, fs *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	tpl, err := template.New("modlist").Parse(modlistTemplate)
	if err != nil {
		log.Printf("parse modlist template: %+v", err)
		return subcommands.ExitFailure
	}

	paths := fs.Args()
	if len(paths) <= 0 {
		paths = []string{defaultManifest}
	}

	m, ok := mergeManifests(paths)
	if !ok {
		return subcommands.ExitFailure
	}

	modlist := m.ModList()

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, modlist); err != nil {
		log.Printf("execute template: %+v", err)
		return subcommands.ExitFailure
	}

	fpath := cmd.OutputPath
	outSrc := buf.Bytes()
	if err := renameio.WriteFile(fpath, outSrc, 0644); err != nil {
		log.Printf("write file %q: %+v", fpath, err)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}
