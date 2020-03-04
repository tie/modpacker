package main

import (
	"github.com/tie/modpacker/models"
)

type Manifest struct {
	Mods   []Mod   `hcl:"mod,block"`
	Checks []Check `hcl:"check,block"`
}

type Mod struct {
	Path      string `hcl:"path,label"`
	Action    string `hcl:"action,optional"`
	Method    string `hcl:"method,optional"`
	File      string `hcl:"file,optional"`
	ProjectID int    `hcl:"projectID,optional"`
	FileID    int    `hcl:"fileID,optional"`
}

type Check struct {
	Method    string   `hcl:"method,attr"`
	File      string   `hcl:"file,optional"`
	ProjectID int      `hcl:"projectID,optional"`
	FileID    int      `hcl:"fileID,optional"`
	Sums      []string `hcl:"sums,attr"`
}

type ModID struct {
	Method    string
	File      string
	ProjectID int
	FileID    int
}

func (m *Mod) ID() ModID {
	return ModID{
		Method:    m.Method,
		File:      m.File,
		ProjectID: m.ProjectID,
		FileID:    m.FileID,
	}
}

func (c *Check) ID() ModID {
	return ModID{
		Method:    c.Method,
		File:      c.File,
		ProjectID: c.ProjectID,
		FileID:    c.FileID,
	}
}

func (m *Manifest) ModList() []models.Mod {
	// Merge check sums into corresponding mods.
	n := len(m.Mods)
	mods := make([]models.Mod, n)
	refs := make(map[ModID]*models.Mod, n)

	for i, mod := range m.Mods {
		mm := models.Mod{
			Path:      mod.Path,
			Method:    mod.Method,
			Action:    mod.Action,
			File:      mod.File,
			ProjectID: mod.ProjectID,
			FileID:    mod.FileID,
		}
		mods[i] = mm

		id := mod.ID()
		refs[id] = &mm
	}

	for _, check := range m.Checks {
		id := check.ID()
		mm, ok := refs[id]
		if !ok {
			continue
		}
		mm.Sums = append(mm.Sums, check.Sums...)
	}

	return mods
}
