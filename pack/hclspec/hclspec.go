package hclspec

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
