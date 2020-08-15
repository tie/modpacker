package archive

import (
	"archive/zip"
	"io"
	"log"
	"path"

	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/util"

	"github.com/tie/modpacker/builder"
	"github.com/tie/modpacker/fetcher"
	"github.com/tie/modpacker/modpacker"
)

var _ builder.Builder = (*ArchiveBuilder)(nil)

type ArchiveBuilder struct {
	Downloader *fetcher.Fetcher
	Archive    *zip.Writer
}

func NewArchiveBuilder(dl *fetcher.Fetcher, w *zip.Writer) *ArchiveBuilder {
	return &ArchiveBuilder{dl, w}
}

func (b *ArchiveBuilder) Add(m modpacker.Mod) error {
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
	case modpacker.ActionNone:
		return b.AddReader(src, m.Path)
	case modpacker.ActionUnzip:
		return b.AddUnzip(src, m.Path)
	}
	return builder.ErrUnknownModAction
}

func (b *ArchiveBuilder) AddUnzip(f billy.File, dir string) error {
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
		// FIXME don’t ignore empty directories.
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
		err := b.AddZipFile(f, name)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *ArchiveBuilder) AddZipFile(f *zip.File, name string) error {
	r, err := f.Open()
	if err != nil {
		return err
	}
	return b.AddReader(r, name)
}

func (b *ArchiveBuilder) AddReader(r io.Reader, name string) error {
	w, err := b.Archive.Create(name)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, r)
	return err
}

func (b *ArchiveBuilder) Close() error {
	return nil
}
