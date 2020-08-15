package curse

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"path"

	"github.com/tie/modpacker/builder"
	"github.com/tie/modpacker/builder/archive"
	"github.com/tie/modpacker/builder/curse/jsonspec"
	"github.com/tie/modpacker/fetcher"
	"github.com/tie/modpacker/modpacker"
)

var _ builder.Builder = (*CurseBuilder)(nil)

type CurseBuilder struct {
	archive.ArchiveBuilder

	CurseFiles []jsonspec.File
}

func NewCurseBuilder(dl *fetcher.Fetcher, w *zip.Writer) *CurseBuilder {
	b := archive.ArchiveBuilder{dl, w}
	return &CurseBuilder{b, nil}
}

func (b *CurseBuilder) Add(m modpacker.Mod) error {
	if m.Method != modpacker.MethodCurse || m.Action != modpacker.ActionNone {
		m.Path = path.Join("overrides", m.Path)
		return b.ArchiveBuilder.Add(m)
	}

	// FIXME How does Curse handle manifests with other modpack projects files?
	// I doubt it supports inheritance.

	b.CurseFiles = append(b.CurseFiles, jsonspec.File{
		ProjectID: m.ProjectID,
		FileID:    m.FileID,
		Required:  true,
	})
	return nil
}

func (b *CurseBuilder) Close() error {
	// TODO so many things to do.
	m := jsonspec.Manifest{
		Minecraft: jsonspec.MinecraftInstance{
			// TODO
			Version: "",
			ModLoaders: []jsonspec.ModLoader{
				// TODO
				jsonspec.ModLoader{},
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
	return b.AddReader(&buf, "manifest.json")
}
