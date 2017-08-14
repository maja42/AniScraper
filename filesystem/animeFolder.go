package filesystem

import (
	"fmt"
	"path/filepath"

	"github.com/google/uuid"
)

// AnimeFolder is a single directory, containing exacly one anime
type AnimeFolder struct {
	ID         uuid.UUID
	Path       string // path without the folder name and without a trailing path separator
	FolderName string
}

func NewAnimeFolder(path string, folderName string) *AnimeFolder {
	return &AnimeFolder{
		ID:         uuid.New(),
		Path:       path,
		FolderName: folderName,
	}
}

// FullPath Returns the full path of the anime folder, without a trailing path separator
func (f *AnimeFolder) FullPath() string {
	return filepath.Join(f.Path, f.FolderName)
}

func (f *AnimeFolder) String() string {
	return fmt.Sprintf("ID: %v, FolderName: %-20s, Path: %s", f.ID, f.FolderName, f.Path)
}
