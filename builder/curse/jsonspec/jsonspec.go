package jsonspec

type Manifest struct {
	ManifestType    string `json:"manifestType"`
	ManifestVersion int    `json:"manifestVersion"`

	Minecraft MinecraftInstance `json:"minecraft"`

	Name      string `json:"name"`
	Version   string `json:"version"`
	Author    string `json:"author"`
	Desc      string `json:"description"`
	ProjectID int    `json:"projectID"`

	Files     []File `json:"files"`
	Overrides string `json:"overrides"`
}

type MinecraftInstance struct {
	Version    string      `json:"version"`
	ModLoaders []ModLoader `json:"modLoaders"`
}

type ModLoader struct {
	ID      string `json:"id"`
	Primary bool   `json:"primary"`
}

type File struct {
	ProjectID int  `json:"projectID"`
	FileID    int  `json:"fileID"`
	Required  bool `json:"required"`
}
