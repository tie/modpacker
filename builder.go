package modpacker

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"log"
	"path"

	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/util"

	"github.com/tie/modpacker/internal/curse"
)

const (
	actionNone  = ""
	actionUnzip = "unzip"
)

type Builder interface {
	Add(m Mod) error
	Close() error
}

type archiveBuilder struct {
	Downloader *Downloader
	Pack       *zip.Writer
}

func NewArchiveBuilder(dl *Downloader, w *zip.Writer) Builder {
	return &archiveBuilder{dl, w}
}

func (b *archiveBuilder) Add(m Mod) error {
	src, err := b.Downloader.Open(m)
	if err != nil {
		return err
	}
	defer func() {
		err := src.Close()
		if err != nil {
			log.Printf("close: %+v", err)
		}
	}()
	switch m.Action {
	case actionNone:
		return b.addReader(src, m.Path)
	case actionUnzip:
		return b.addUnzip(src, m.Path)
	}
	return ErrUnknownModAction
}

func (b *archiveBuilder) addUnzip(f billy.File, dir string) error {
	fi, err := util.Stat(f)
	if err != nil {
		return err
	}
	size := fi.Size()
	z, err := zip.NewReader(f, size)
	if err != nil {
		return err
	}
	for _, f := range z.File {
		// If last char in file name is slash,
		// then the file is empty and represents
		// a directory. We skip those for brevity.
		l := len(f.Name)
		if l > 0 {
			if f.Name[l-1] == '/' {
				continue
			}
		}
		// TODO should we sanitize name?
		name := path.Join(dir, f.Name)
		err := b.addZipFile(f, name)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *archiveBuilder) addZipFile(f *zip.File, name string) error {
	r, err := f.Open()
	if err != nil {
		return err
	}
	return b.addReader(r, name)
}

func (b *archiveBuilder) addReader(r io.Reader, name string) error {
	w, err := b.Pack.Create(name)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, r)
	return err
}

func (b *archiveBuilder) Close() error {
	return nil
}

type curseBuilder struct {
	archiveBuilder

	CurseFiles []curse.File
}

func NewCurseBuilder(dl *Downloader, w *zip.Writer) Builder {
	b := archiveBuilder{dl, w}
	return &curseBuilder{b, nil}
}

func (b *curseBuilder) Add(m Mod) error {
	if m.Method != methodCurse || m.Action != actionNone {
		m.Path = path.Join("overrides", m.Path)
		return b.archiveBuilder.Add(m)
	}

	// FIXME How does Curse handle manifests with other modpack projects files?
	// I doubt it supports inheritance.

	b.CurseFiles = append(b.CurseFiles, curse.File{
		ProjectID: m.ProjectID,
		FileID:    m.FileID,
		Required:  true,
	})
	return nil
}

func (b *curseBuilder) Close() error {
	// TODO so many things to do.
	m := curse.Manifest{
		Minecraft: curse.MinecraftInstance{
			Version: "", // TODO
			ModLoaders: []curse.ModLoader{
				curse.ModLoader{},
				// TODO
			},
		},
		ManifestType:    "minecraftModpack",
		ManifestVersion: 1,
		Name:            "", // TODO
		Version:         "", // TODO
		Author:          "", // TODO
		Desc:            "", // TODO
		ProjectID:       0,  // TODO
		Files:           b.CurseFiles,
	}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(&m); err != nil {
		return err
	}
	return b.addReader(&buf, "manifest.json")
}
