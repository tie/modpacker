package builder

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"path"

	"github.com/tie/modpacker/fetcher"
	"github.com/tie/modpacker/models"
	"github.com/tie/modpacker/models/curse"
)

type curseBuilder struct {
	archiveBuilder

	CurseFiles []curse.File
}

func NewCurseBuilder(dl *fetcher.Fetcher, w *zip.Writer) Builder {
	b := archiveBuilder{dl, w}
	return &curseBuilder{b, nil}
}

func (b *curseBuilder) Add(m models.Mod) error {
	if m.Method != models.MethodCurse || m.Action != models.ActionNone {
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
