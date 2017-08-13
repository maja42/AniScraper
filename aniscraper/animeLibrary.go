package aniscraper

// import (
// 	"fmt"
// 	"io/ioutil"
// 	"os"
// 	"path/filepath"
// 	"sync"

// 	"github.com/maja42/AniScraper/utils"
// )

// // AnimeCollection is a collection of multiple, unique anime folders
// type AnimeCollection interface {
// 	// AddCollection adds all subfolders of 'path' to the collection. The number of added folders is returned.
// 	AddCollection(path string) (int, error)
// 	// AddFolder adds a single anime folder to the collection. The anime folder ID is returned.
// 	AddFolder(path string, name string) (AnimeFolderID, error)
// 	// Contains checks if the given path / anime folder is already part of the collection
// 	Contains(folder os.FileInfo) bool
// 	// Folders returns all existing anime folders by their anime folder ID
// 	//Folders() map[AnimeFolderID]*AnimeFolder
// 	// AfIds() returns a list with allexisting anime folder ids stored in this collection
// 	//AfIds() []AnimeFolderID
// 	// Paths returns a list with absolute paths of all anime folders
// 	//Paths() []string

// 	// Iterate calls the given function for every animeFolder, until false is returned (do not continue) or there are no more folders
// 	Iterate(callback func(folder *AnimeFolder) bool)
// }

// type animeCollection struct {
// 	mutex sync.RWMutex

// 	afidSequence utils.Sequence                 // Used for generating unique anime folder ids
// 	folders      map[AnimeFolderID]*AnimeFolder // afid -> animeFolder

// 	bindingContext ServerBindingContext
// }

// // NewAnimeCollection initialises and returns a new and empty anime collection
// func NewAnimeCollection(bindingContext ServerBindingContext) AnimeCollection {
// 	return &animeCollection{
// 		afidSequence: utils.NewSequenceGenerator(0),
// 		folders:      make(map[AnimeFolderID]*AnimeFolder, 0),

// 		bindingContext: bindingContext,
// 	}
// }

// func (collection *animeCollection) AddCollection(path string) (int, error) {
// 	log.Debugf("Adding anime folders within directory %q...", path)

// 	addCount := 0
// 	files, err := ioutil.ReadDir(path)
// 	if err != nil {
// 		return 0, err
// 	}

// 	for _, file := range files {
// 		fullPath := filepath.Join(path, file.Name())
// 		if !file.IsDir() {
// 			log.Debugf("Ignoring file %q", fullPath)
// 			continue
// 		}
// 		log.Debugf("Found directory %q", fullPath)

// 		_, err := collection.addFolder(path, file.Name())
// 		if err != nil {
// 			log.Warningf("Failed to add anime folder %q: %s", fullPath, err)
// 		} else {
// 			addCount++
// 		}
// 	}
// 	log.Debugf("%d anime folders added", addCount)
// 	return addCount, nil
// }

// func (collection *animeCollection) AddFolder(path string, folderName string) (AnimeFolderID, error) {
// 	fullPath := filepath.Join(path, folderName)

// 	log.Debug("Checking folder...")
// 	exists, isDir := CheckFolder(fullPath)
// 	if !exists {
// 		return -1, fmt.Errorf("The path %q does not exist", fullPath)
// 	}
// 	if !isDir {
// 		return -1, fmt.Errorf("Not a directory: %q", fullPath)
// 	}
// 	return collection.addFolder(path, folderName)
// }

// // addFolder adds a new anime folder (that is ensured to exist)
// func (collection *animeCollection) addFolder(path string, folderName string) (AnimeFolderID, error) {
// 	var err error
// 	path, err = filepath.EvalSymlinks(path)
// 	if err != nil {
// 		return -1, err
// 	}
// 	path, err = filepath.Abs(path) // Nice looking absolute path (propably not canonical!)
// 	if err != nil {
// 		return -1, err
// 	}

// 	afid := AnimeFolderID(collection.afidSequence.Next()) // the sequence maintains its own lock
// 	animeFolder := &AnimeFolder{
// 		ID:   afid,
// 		Path: path,
// 		Name: folderName,
// 	}

// 	// The following uniqueness-checks have been disabled due to performance reasons
// 	//
// 	// log.Debug("Retrieving file info...")
// 	// fileInfo, err := GetFileInfo(fullPath)
// 	// if err != nil {
// 	// 	return -1, err
// 	// }
// 	// log.Debug("Checking existence...")
// 	// if collection.Contains(fileInfo) {
// 	// 	return -1, fmt.Errorf("The anime collection already contains the directory %q", fullPath)
// 	// }

// 	if collection.containsPath(animeFolder.FullPath()) {
// 		return -1, fmt.Errorf("The anime collection already contains a directory with the path %q", animeFolder.FullPath())
// 	}

// 	collection.mutex.Lock()
// 	defer collection.mutex.Unlock()

// 	log.Debug("Appending anime folder...")
// 	collection.folders[afid] = animeFolder

// 	collection.bindingContext.NewAnimeFolder(animeFolder)
// 	return afid, nil
// }

// func (collection *animeCollection) Contains(folder os.FileInfo) bool {
// 	collection.mutex.RLock()
// 	defer collection.mutex.RUnlock()
// 	return collection.contains(folder)
// }

// func (collection *animeCollection) contains(folder os.FileInfo) bool {
// 	for _, animeFolder := range collection.folders {
// 		animeFolderFileInfo, err := GetFileInfo(animeFolder.FullPath())
// 		if err != nil {
// 			log.Errorf("Failed to query existing anime folder %q: %v", animeFolder.FullPath(), err)
// 			continue
// 		}

// 		if os.SameFile(folder, animeFolderFileInfo) {
// 			return true
// 		}
// 	}
// 	return false
// }

// // containsPath checks if the collection already contains a folder with the same path;
// // If false is returned, it does not neccessarily mean that the given folder path does not exist
// func (collection *animeCollection) containsPath(folderPath string) bool {
// 	for _, animeFolder := range collection.folders {
// 		if folderPath == animeFolder.FullPath() {
// 			return true
// 		}
// 	}
// 	return false
// }

// // func (collection *animeCollection) Paths() []string {
// // 	collection.mutex.RLock()
// // 	defer collection.mutex.RUnlock()

// // 	paths := make([]string, 0, len(collection.folders))

// // 	for _, animeFolder := range collection.folders {
// // 		paths = append(paths, animeFolder.FullPath())
// // 	}
// // 	return paths
// // }

// // func (collection *animeCollection) AfIds() []int {
// // 	collection.mutex.RLock()
// // 	defer collection.mutex.RUnlock()

// // 	uids := make([]int, 0, len(collection.folders))

// // 	for uid := range collection.folders {
// // 		uids = append(uids, uid)
// // 	}
// // 	return uids
// // }

// func (collection *animeCollection) Iterate(callback func(folder *AnimeFolder) bool) {
// 	collection.mutex.RLock()
// 	defer collection.mutex.RUnlock()

// 	for _, animeFolder := range collection.folders {
// 		if !callback(animeFolder) {
// 			return
// 		}
// 	}
// }
