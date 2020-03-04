package builder

import (
	"archive/zip"
	"io"
	"log"
	"path"

	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/util"

	"github.com/tie/modpacker/fetcher"
	"github.com/tie/modpacker/models"
)

type archiveBuilder struct {
	Downloader *fetcher.Fetcher
	Pack       *zip.Writer
}

func NewArchiveBuilder(dl *fetcher.Fetcher, w *zip.Writer) Builder {
	return &archiveBuilder{dl, w}
}

func (b *archiveBuilder) Add(m models.Mod) error {
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
	case models.ActionNone:
		return b.addReader(src, m.Path)
	case models.ActionUnzip:
		return b.addUnzip(src, m.Path)
	}
	return models.ErrUnknownModAction
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
