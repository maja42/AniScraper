package filesystem

import (
	"fmt"
	"os"
)

// MediaFolder is a single directory, containing exacly one series
type MediaFolder struct {
	Uid  int
	Path string
	Name string
}

func (folder *MediaFolder) FullPath() string {
	return fmt.Sprintf("%s%c%s", folder.Path, os.PathSeparator, folder.Name)
}
