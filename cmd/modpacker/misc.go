package main

import (
	"log"
	"os"
	"path/filepath"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"

	"github.com/tie/internal/robustio"
)

func cacheDir(p string) (string, error) {
	c, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(c, p), nil
}

func makeCache(p string) (string, error) {
	c, err := cacheDir(p)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(c, 0700); err != nil {
		return "", err
	}
	return c, nil
}

func newDiagWr(p *hclparse.Parser) (diagWr hcl.DiagnosticWriter, color bool) {
	files := p.Files()
	stderr := os.Stderr
	fd := int(stderr.Fd())
	istty, color := fdinfo(fd)
	if !istty {
		diagWr := hcl.NewDiagnosticTextWriter(stderr, files, 80, color)
		return diagWr, color
	}
	var width uint
	if w, _, err := terminal.GetSize(fd); err != nil {
		log.Printf("get term size: %+v", err)
	} else if w >= 0 {
		width = uint(w)
	} else {
		width = 80
	}
	return hcl.NewDiagnosticTextWriter(stderr, files, width, color), color
}

func fdinfo(fd int) (istty, color bool) {
	istty = terminal.IsTerminal(fd)
	if istty {
		color = true
	}
	// See https://no-color.org
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		color = false
	}
	return
}

func mergeManifests(paths []string) (Manifest, bool) {
	var m Manifest
	var files []*hcl.File
	var diags hcl.Diagnostics
	parser := hclparse.NewParser()
	diagWr, _ := newDiagWr(parser)
	for _, fpath := range paths {
		src, err := robustio.ReadFile(fpath)
		if err != nil {
			log.Printf("read %q: %+v", fpath, err)
			return m, false
		}
		file, parseDiags := parser.ParseHCL(src, fpath)
		diags = append(diags, parseDiags...)
		if parseDiags.HasErrors() {
			err := diagWr.WriteDiagnostics(diags)
			if err != nil {
				log.Printf("write diags: %+v", err)
			}
			return m, false
		}
		files = append(files, file)
	}
	if diags.HasErrors() {
		err := diagWr.WriteDiagnostics(diags)
		if err != nil {
			log.Printf("write diags: %+v", err)
		}
		return m, false
	}
	body := hcl.MergeFiles(files)
	decodeDiags := gohcl.DecodeBody(body, nil, &m)
	diags = append(diags, decodeDiags...)
	err := diagWr.WriteDiagnostics(diags)
	if err != nil {
		log.Printf("write diags: %+v", err)
		return m, false
	}
	if diags.HasErrors() {
		return m, false
	}
	return m, true
}
