package filesystem

import (
	"os"
)

// IsDir checks if the given file/directory is a directory
func IsDir(path string) (bool, error) {
	file, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return file.Mode().IsDir(), nil
}

// GetFileInfo opens the given file/directory for reading and retrieves the fileinfo
func GetFileInfo(path string) (os.FileInfo, error) {
	file, err := os.Open(path)
	if err != nil { // Does not exist / unable to open
		return nil, err
	}

	fileInfo, statErr := file.Stat()

	if err := file.Close(); err != nil {
		return nil, err
	}
	return fileInfo, statErr
}
