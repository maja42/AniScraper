package aniscraper

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/maja42/AniScraper/utils"
)

// AnimeCollection is a collection of multiple, unique anime folders
type AnimeCollection interface {
	// AddCollection returns all subfolders of 'path' to the collection. The number of added folders is returned.
	AddCollection(path string) (int, error)
	// AddFolder adds a single anime folder to the collection. The afid (anime folder ID) is returned.
	AddFolder(path string, name string) (int, error)
	// Contains checks if the given path / anime folder is already part of the collection
	Contains(folder os.FileInfo) bool
	// Folders returns all existing anime folders by their afid
	//Folders() map[int]*AnimeFolder
	// AfIds() returns a list with all existing afids stored in this collection
	//AfIds() []int
	// Paths returns a list with absolute paths of all anime folders
	//Paths() []string

	// Iterate calles the given function for every animeFolder, until false is returned (do not continue) or there are no more folders
	Iterate(callback func(afid int, folder *AnimeFolder) bool)
}

type animeCollection struct {
	mutex sync.RWMutex

	afidSequence utils.Sequence       // Used for generating unique anime folder ids
	folders      map[int]*AnimeFolder // afid -> animeFolder

	bindingContext ServerBindingContext
}

// NewAnimeCollection initialises and returns a new and empty anime collection
func NewAnimeCollection(bindingContext ServerBindingContext) AnimeCollection {
	return &animeCollection{
		afidSequence: utils.NewSequenceGenerator(0),
		folders:      make(map[int]*AnimeFolder, 0),

		bindingContext: bindingContext,
	}
}

func (collection *animeCollection) AddCollection(path string) (int, error) {
	log.Debugf("Adding anime folders within directory %q...", path)

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

		_, err := collection.addFolder(path, file.Name())
		if err != nil {
			log.Warningf("Failed to add anime folder %q: %v", fullPath, err)
		} else {
			addCount++
		}
	}
	log.Debugf("%d anime folders added", addCount)
	return addCount, nil
}

func (collection *animeCollection) AddFolder(path string, folderName string) (int, error) {
	fullPath := FullPath(path, folderName)

	log.Debug("Checking folder...")
	exists, isDir := CheckFolder(fullPath)
	if !exists {
		return -1, fmt.Errorf("The path %q does not exist", fullPath)
	}
	if !isDir {
		return -1, fmt.Errorf("Not a directory: %q", fullPath)
	}
	return collection.addFolder(path, folderName)
}

// addFolder adds a new anime folder (that is ensured to exist)
func (collection *animeCollection) addFolder(path string, folderName string) (int, error) {
	var err error
	path, err = filepath.EvalSymlinks(path)
	if err != nil {
		return -1, err
	}
	path, err = filepath.Abs(path) // Nice looking absolute path (propably not canonical!)
	if err != nil {
		return -1, err
	}

	// The following uniqueness-checks have been disabled due to performance reasons

	// log.Debug("Retrieving file info...")
	// fileInfo, err := GetFileInfo(fullPath)
	// if err != nil {
	// 	return err
	// }
	// log.Debug("Checking existence...")
	// if collection.Contains(fileInfo) {
	// 	return fmt.Errorf("The anime collection already contains the directory %q", fullPath)
	// }

	log.Debug("Appending anime folder...")
	afid := collection.afidSequence.Next() // the sequence maintains its own lock
	animeFolder := &AnimeFolder{
		Afid: afid,
		Path: path,
		Name: folderName,
	}

	collection.mutex.Lock()
	defer collection.mutex.Unlock()

	collection.folders[afid] = animeFolder

	collection.bindingContext.NewAnimeFolder(afid, animeFolder)
	return afid, nil
}

func (collection *animeCollection) Contains(folder os.FileInfo) bool {
	collection.mutex.RLock()
	defer collection.mutex.RUnlock()
	return collection.contains(folder)
}

func (collection *animeCollection) contains(folder os.FileInfo) bool {
	for _, animeFolder := range collection.folders {

		animeFolderFileInfo, err := GetFileInfo(animeFolder.FullPath())
		if err != nil {
			log.Errorf("Failed to query existing anime folder %q: %v", animeFolder.FullPath(), err)
			continue
		}

		if os.SameFile(folder, animeFolderFileInfo) {
			return true
		}

	}
	return false
}

// func (collection *animeCollection) Paths() []string {
// 	collection.mutex.RLock()
// 	defer collection.mutex.RUnlock()

// 	paths := make([]string, 0, len(collection.folders))

// 	for _, animeFolder := range collection.folders {
// 		paths = append(paths, animeFolder.FullPath())
// 	}
// 	return paths
// }

// func (collection *animeCollection) AfIds() []int {
// 	collection.mutex.RLock()
// 	defer collection.mutex.RUnlock()

// 	uids := make([]int, 0, len(collection.folders))

// 	for uid := range collection.folders {
// 		uids = append(uids, uid)
// 	}
// 	return uids
// }

func (collection *animeCollection) Iterate(callback func(afid int, folder *AnimeFolder) bool) {
	collection.mutex.RLock()
	defer collection.mutex.RUnlock()

	for afid, animeFolder := range collection.folders {
		if !callback(afid, animeFolder) {
			return
		}
	}
}
