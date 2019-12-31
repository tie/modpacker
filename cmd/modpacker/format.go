package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/pkg/diff"

	"github.com/google/subcommands"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclwrite"

	"github.com/tie/internal/renameio"
	"github.com/tie/internal/robustio"
)

type FormatCommand struct {
	DisableCheck bool
	Overwrite    bool
	ContextSize  int
}

func (*FormatCommand) Name() string     { return "fmt" }
func (*FormatCommand) Synopsis() string { return "format manifests" }
func (*FormatCommand) Usage() string {
	return `Usage: modpacker fmt [-c int] [-w] [-nocheck] [manifest paths]

	Formats manifests using standard syntax. It can either write files
	in-places or generate unified diff with specified context size.

Flags:
`
}

func (cmd *FormatCommand) SetFlags(fs *flag.FlagSet) {
	fs.BoolVar(&cmd.DisableCheck, "nocheck", false, "disable diagnostics")
	fs.BoolVar(&cmd.Overwrite, "w", false, "write result to (source) file instead of stdout")
	fs.IntVar(&cmd.ContextSize, "c", 3, "output n lines of diff context")
}

func (cmd *FormatCommand) Execute(ctx context.Context, fs *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	var color bool
	var parser *hclparse.Parser
	var diagWr hcl.DiagnosticWriter
	if !cmd.DisableCheck {
		parser = hclparse.NewParser()
		diagWr, color = newDiagWr(parser)
	}

	paths := fs.Args()
	if len(paths) <= 0 {
		paths = []string{defaultManifest}
	} else {
		sort.Strings(paths)
	}

	seen := make(map[string]bool, len(paths))
	for _, fpath := range paths {
		if seen[fpath] {
			continue
		}
		seen[fpath] = true
		src, err := robustio.ReadFile(fpath)
		if err != nil {
			log.Printf("read manifest %q: %+v", fpath, err)
			return subcommands.ExitFailure
		}

		if !cmd.DisableCheck {
			file, diags := parser.ParseHCL(src, fpath)
			if diags.HasErrors() {
				err := diagWr.WriteDiagnostics(diags)
				if err != nil {
					log.Printf("write diags: %+v", err)
				}
				return subcommands.ExitFailure
			}
			decodeDiags := gohcl.DecodeBody(file.Body, nil, &Manifest{})
			diags = append(diags, decodeDiags...)
			err := diagWr.WriteDiagnostics(diags)
			if err != nil {
				log.Printf("write diags: %+v", err)
				return subcommands.ExitFailure
			}
			if diags.HasErrors() {
				return subcommands.ExitFailure
			}
		}

		outSrc := hclwrite.Format(src)
		if bytes.Equal(src, outSrc) {
			continue
		}
		if !cmd.Overwrite {
			fpath := filepath.ToSlash(fpath)
			aname := fmt.Sprintf("a/%s", fpath)
			bname := fmt.Sprintf("b/%s", fpath)
			names := diff.Names(aname, bname)
			opts := []diff.WriteOpt{names}
			if color {
				c := diff.TerminalColor()
				opts = append(opts, c)
			}
			a, b := splitLines(src), splitLines(outSrc)
			pair := diff.Bytes(a, b)
			edit := diff.Myers(ctx, pair)
			if cmd.ContextSize >= 0 {
				edit = edit.WithContextSize(cmd.ContextSize)
			}
			_, err := edit.WriteUnified(os.Stdout, pair, opts...)
			if err != nil {
				log.Printf("write diff: %+v", err)
				return subcommands.ExitFailure
			}
			continue
		}
		if err := renameio.WriteFile(fpath, outSrc, 0644); err != nil {
			log.Printf("write file %q: %+v", fpath, err)
			return subcommands.ExitFailure
		}
	}

	return subcommands.ExitSuccess
}

func splitLines(b []byte) [][]byte {
	return bytes.Split(b, []byte("\n"))
}
