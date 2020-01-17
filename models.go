package modpacker

type Mod struct {
	// Path is the file name in modpack archive.
	Path string

	// Method is the method used for downloading the mod.
	// Possible values: "curse", "optifine".
	Method string

	// Action is the additional action to perform
	// on the downloaded file (e.g. "unzip" world save).
	Action string

	// File specifies the OptiFine file name.
	File string

	// ProjectID specifies the project ID on CurseForge.
	ProjectID int
	// FileID specifies the file ID of the CurseForge project.
	FileID int

	// Sums is a list of expected file checksums.
	Sums []string
}
