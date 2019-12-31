package modpacker

import (
	"archive/zip"
	"io"
	"log"
	"path"

	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/util"
)

type Builder struct {
	Downloader *Downloader
	Pack       *zip.Writer
}

func (b *Builder) Add(m Mod) error {
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
	case "":
		return b.addReader(src, m.Path)
	case "unzip":
		return b.addUnzip(src, m.Path)
	}
	return ErrUnknownModAction
}

func (b *Builder) addUnzip(f billy.File, dir string) error {
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
		name := path.Join(dir, f.Name)
		err := b.addZipFile(f, name)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *Builder) addZipFile(f *zip.File, name string) error {
	r, err := f.Open()
	if err != nil {
		return err
	}
	return b.addReader(r, name)
}

func (b *Builder) addReader(r io.Reader, name string) error {
	w, err := b.Pack.Create(name)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, r)
	return err
}
