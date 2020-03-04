package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/google/subcommands"

	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclwrite"

	"github.com/tie/internal/renameio"

	"github.com/tie/modpacker/models/curse"
)

type BootstrapCommand struct {
	CursePath  string
	OutputPath string
}

func (*BootstrapCommand) Name() string     { return "bootstrap" }
func (*BootstrapCommand) Synopsis() string { return "migrate existing modpack" }
func (*BootstrapCommand) Usage() string {
	return `Usage: modpacker bootstrap [-o base.pack] [-curse manifest.json]

	Bootstraps a new project using existing CurseForge modpack.

Flags:
`
}

func (cmd *BootstrapCommand) SetFlags(fs *flag.FlagSet) {
	fs.StringVar(&cmd.CursePath, "curse", "manifest.json", "curse manifest path")
	fs.StringVar(&cmd.OutputPath, "o", "base.pack", "output manifest path")
}

func (cmd *BootstrapCommand) Execute(ctx context.Context, fs *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	fpath := cmd.CursePath
	f, err := os.Open(fpath)
	if err != nil {
		log.Printf("open %q: %+v", fpath, err)
		return subcommands.ExitFailure
	}

	var cm curse.Manifest
	if err := json.NewDecoder(f).Decode(&cm); err != nil {
		log.Printf("decode %q: %+v", fpath, err)
		return subcommands.ExitFailure
	}

	var m Manifest
	m.Mods = make([]Mod, 0, len(cm.Files))
	for _, cf := range cm.Files {
		path := fmt.Sprintf("mods/%d-%d.jar", cf.ProjectID, cf.FileID)
		mod := Mod{
			Path:      path,
			Method:    "curse",
			ProjectID: cf.ProjectID,
			FileID:    cf.FileID,
		}
		m.Mods = append(m.Mods, mod)
	}

	opath := filepath.FromSlash(cm.Overrides)
	err = filepath.Walk(cm.Overrides, func(fpath string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.IsDir() {
			return nil
		}
		rpath, err := filepath.Rel(opath, fpath)
		if err != nil {
			return err
		}
		mod := Mod{
			Path: rpath,
			File: fpath,
		}
		m.Mods = append(m.Mods, mod)
		return nil
	})
	if err != nil {
		log.Printf("walk %q: %+v", cm.Overrides, err)
		return subcommands.ExitFailure
	}

	conf := hclwrite.NewEmptyFile()
	body := conf.Body()
	gohcl.EncodeIntoBody(&m, body)
	data := conf.Bytes()

	err = renameio.WriteFile(cmd.OutputPath, data, 0644)
	if err != nil {
		log.Printf("write %q: %+v", cmd.OutputPath, err)
		return subcommands.ExitFailure
	}
	return subcommands.ExitSuccess
}
