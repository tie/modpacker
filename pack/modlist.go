package pack

import (
	"github.com/tie/modpacker/modpacker"
	"github.com/tie/modpacker/pack/hclspec"
)

type modID struct {
	Method    string
	File      string
	ProjectID int
	FileID    int
}

func ModList(ms []hclspec.Manifest) []modpacker.Mod {
	n := 0
	for _, m := range ms {
		n += len(m.Mods)
	}

	if len(ms) <= 0 || n <= 0 {
		return nil
	}

	mods := make([]modpacker.Mod, n)
	refs := make(map[modID]*modpacker.Mod, n)

	// Merge mods and create reference for mod ID.
	for _, m := range ms {
		for i, mod := range m.Mods {
			id := modID{
				Method:    mod.Method,
				File:      mod.File,
				ProjectID: mod.ProjectID,
				FileID:    mod.FileID,
			}
			mods[i] = modpacker.Mod{
				Path:      mod.Path,
				Method:    mod.Method,
				Action:    mod.Action,
				File:      mod.File,
				ProjectID: mod.ProjectID,
				FileID:    mod.FileID,
			}
			refs[id] = &mods[i]
		}
	}

	// Merge check sums into corresponding mods.
	for _, m := range ms {
		for _, check := range m.Checks {
			id := modID{
				Method:    check.Method,
				File:      check.File,
				ProjectID: check.ProjectID,
				FileID:    check.FileID,
			}
			mm, ok := refs[id]
			if !ok {
				continue
			}
			mm.Sums = append(mm.Sums, check.Sums...)
		}
	}

	return mods
}
