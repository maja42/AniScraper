package aniscraper

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

// // GetFileInfo opens the given file/directory for reading, retrieves the fileinfo and closes the file/directory
// func GetFileInfo(path string) (os.FileInfo, error) {
// 	file, err := os.Open(path)
// 	if err != nil { // Does not exist / unable to open
// 		return nil, err
// 	}

// 	fileInfo, statErr := file.Stat()

// 	if err := file.Close(); err != nil {
// 		return nil, err
// 	}
// 	return fileInfo, statErr
// }

// // CheckFolder checks if the given path exists and can be opened (yes: true) and if it is a directory (true) or a file (false)
// func CheckFolder(path string) (bool, bool) {
// 	file, err := os.Open(path)
// 	if err != nil { // Does not exist / unable to open
// 		return false, false
// 	}

// 	defer func() {
// 		if err := file.Close(); err != nil {
// 			log.Errorf("Failed to close file/folder %q: %s", path, err)
// 		}
// 	}()

// 	stats, err := file.Stat()
// 	if err != nil {
// 		log.Errorf("Failed to get file info of opened file/folder %q: %s", path, err)
// 		return false, false
// 	}

// 	return true, stats.IsDir()
// }
