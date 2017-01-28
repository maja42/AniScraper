package filesystem

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/maja42/AniScraper/utils"
)

// MediaCollection is a collection of multiple, unique mediafolders
type MediaCollection interface {
	// AddCollection returns all subfolders of path to the collection. The number of added folders is returned.
	AddCollection(path string) (int, error)
	// AddFolder adds a single media folder to the collection.
	AddFolder(path string, name string) error
	// Contains checks if the given path / media folder is already part of the collection
	Contains(folder os.FileInfo) bool
	Paths() []string
}

type mediaCollection struct {
	folders     map[int]*MediaFolder // uid -> mediaFolder
	mutex       sync.RWMutex
	uidSequence utils.Sequence // Used for generating unique ids
}

// NewMediaCollection initialises and returns a new and empty media collection
func NewMediaCollection() MediaCollection {
	return &mediaCollection{
		folders:     make(map[int]*MediaFolder, 0),
		uidSequence: utils.NewSequenceGenerator(0),
	}
}

func (collection *mediaCollection) AddCollection(path string) (int, error) {
	log.Debugf("Adding media folders within directory %q...", path)

	addCount := 0
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return 0, err
	}

	for _, file := range files {
		fullPath := FullPath(path, file.Name())
		if !file.IsDir() {
			log.Debugf("Ignoring file %q", fullPath)
			continue
		}
		log.Debugf("Found directory %q", fullPath)

		err := collection.AddFolder(path, file.Name())
		if err != nil {
			log.Warningf("Failed to add media folder %q: %v", fullPath, err)
		} else {
			addCount++
		}
	}
	return addCount, nil
}

func (collection *mediaCollection) AddFolder(path string, folderName string) error {
	var err error
	fullPath := FullPath(path, folderName)

	log.Debug("Checking folder...")
	exists, isDir := CheckFolder(fullPath)
	if !exists {
		return fmt.Errorf("The path %q does not exist", fullPath)
	}

	if !isDir {
		return fmt.Errorf("Not a directory: %q", fullPath)
	}

	if path, err = filepath.Abs(path); err != nil { // convert to absolute path
		return err
	}
	fullPath = FullPath(path, folderName)

	// The following uniqueness-checks have been disabled due to performance reasons

	// log.Debug("Retrieving file info...")
	// fileInfo, err := GetFileInfo(fullPath)
	// if err != nil {
	// 	return err
	// }
	// log.Debug("Checking existence...")
	// if collection.Contains(fileInfo) {
	// 	return fmt.Errorf("The media collection already contains the directory %q", fullPath)
	// }

	log.Debug("Appending media folder...")
	uid := collection.uidSequence.Next() // the sequence maintains its own lock
	media := &MediaFolder{
		Uid:  uid,
		Path: path,
		Name: folderName,
	}

	collection.mutex.Lock()
	defer collection.mutex.Unlock()

	collection.folders[uid] = media
	return nil
}

func (collection *mediaCollection) Contains(folder os.FileInfo) bool {
	collection.mutex.Lock()
	defer collection.mutex.Unlock()
	return collection.contains(folder)
}

func (collection *mediaCollection) contains(folder os.FileInfo) bool {
	for _, mediaFolder := range collection.folders {

		mediaFolderFileInfo, err := GetFileInfo(mediaFolder.FullPath())
		if err != nil {
			log.Errorf("Failed to query existing media folder %q: %v", mediaFolder.FullPath(), err)
			continue
		}

		if os.SameFile(folder, mediaFolderFileInfo) {
			return true
		}

	}
	return false
}

func (collection *mediaCollection) Paths() []string {
	collection.mutex.Lock()
	defer collection.mutex.Unlock()

	paths := make([]string, len(collection.folders))

	for _, mediaFolder := range collection.folders {
		paths = append(paths, mediaFolder.FullPath())
	}

	return paths
}
