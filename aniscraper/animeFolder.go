package aniscraper

import (
	"fmt"
	"os"
)

// AnimeFolder is a single directory, containing exacly one anime
type AnimeFolder struct {
	Afid int    // anime folder ID
	Path string // path without the folder name and without a traling path separator
	Name string
}

func (folder *AnimeFolder) FullPath() string {
	return fmt.Sprintf("%s%c%s", folder.Path, os.PathSeparator, folder.Name)
}
