package aniscraper

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sync"

	"github.com/maja42/AniScraper/utils"

	"golang.org/x/net/context"

	"github.com/fsnotify/fsnotify"
	"github.com/google/uuid"
)

// AnimeCollection represents a single directory that can contain multiple individual anime folders
type AnimeCollection interface {
	// AnimeFolder returns the anime folder with the given folder name (not the path); no trailing slash
	AnimeFolder(folderName string) *AnimeFolder
	// LoadFromFilesystem (re-)initializes the anime-folder with the data from the filesystem; fails if the filesystem is currently watched
	LoadFromFilesystem() error
	// WatchFileSystem (re-)initializes the anime-folder, starts to monitor the underlying filesystem folder and updates the collection automatically; fails if the filesystem is already watched
	WatchFilesystem(ctx context.Context, alsoWatchFolders bool) (<-chan error, error)
	// Clear removes all anime folders; fails if the filesystem is currently watched
	Clear() error

	// Wait blocks until all go routines (if any) have finished
	Wait()

	// Iterate calls the given function for every animeFolder, until false is returned (do not continue) or there are no more folders
	Iterate(callback func(folder *AnimeFolder) bool)
}

// animeCollection is a single directory that can contain multiple individual anime folders
type animeCollection struct {
	id   uuid.UUID
	name string // user defined name
	path string

	animeFolders           map[uuid.UUID]*AnimeFolder
	isWatchingCollection   bool
	isWatchingAnimeFolders bool
	animeFolderWatcher     *fsnotify.Watcher // also exists if anime folders are not watched

	mutex  sync.RWMutex
	wg     sync.WaitGroup
	logger utils.Logger
}

// NewAnimeCollection initialises and returns a new and empty anime collection
func NewAnimeCollection(name string, path string, logger utils.Logger) (AnimeCollection, error) {
	var err error
	path, err = filepath.EvalSymlinks(path)
	if err != nil {
		return nil, err
	}
	path, err = filepath.Abs(path) // Nice looking absolute path (propably not canonical!)
	if err != nil {
		return nil, err
	}

	return &animeCollection{
		id:   uuid.New(),
		name: name,
		path: path,

		animeFolders: make(map[uuid.UUID]*AnimeFolder),
		logger:       logger.New("AnimeCollection[" + name + "]"),
	}, nil
}

func (c *animeCollection) WatchFilesystem(ctx context.Context, alsoWatchFolders bool) (<-chan error, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if err := c.clear(); err != nil {
		return nil, err
	}

	collectionWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	c.animeFolderWatcher, err = fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)
	errOut := make(chan error)
	c.wg.Add(1)
	go func() {
		defer func() {
			c.mutex.Lock()
			c.isWatchingCollection = false
			c.isWatchingAnimeFolders = false
			if err := c.animeFolderWatcher.Close(); err != nil {
				errOut <- fmt.Errorf("Failed to close anime folder watcher: %s", err)
			}
			c.animeFolderWatcher = nil
			if err := collectionWatcher.Close(); err != nil {
				errOut <- fmt.Errorf("Failed to close collection watcher: %s", err)
			}
			c.mutex.Unlock()
			close(errOut)
			c.wg.Done()
		}()

		logger := c.logger.New("watcher")
		for {
			select {
			// collection events:
			case event := <-collectionWatcher.Events:
				if err := c.onCollectionWatcherEvent(event, logger); err != nil {
					errOut <- fmt.Errorf("Failed to process event %v: %s", event, err)
				}
			case err := <-collectionWatcher.Errors:
				errOut <- fmt.Errorf("Error while watching anime collection: %s", err)

			// anime folder events:
			case event := <-c.animeFolderWatcher.Events:
				if err := c.onAnimeFolderWatcherEvent(event, logger); err != nil {
					errOut <- fmt.Errorf("Failed to process event %v: %s", event, err)
				}
			case err := <-c.animeFolderWatcher.Errors:
				errOut <- fmt.Errorf("Error while watching anime collection: %s", err)

			// abort:
			case <-ctx.Done():
				logger.Debugf("Stopping anime collection watcher")
				return
			}
		}
	}()
	c.isWatchingCollection = true

	if err := collectionWatcher.Add(c.path); err != nil {
		cancel()
		return nil, fmt.Errorf("Failed to watch directory: %s", err)
	}

	c.isWatchingAnimeFolders = alsoWatchFolders
	if err := c.loadFromFilesystem(); err != nil {
		return nil, err
	}
	return errOut, nil
}

func (c *animeCollection) LoadFromFilesystem() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if err := c.clear(); err != nil {
		return err
	}
	return c.loadFromFilesystem()
}

// loadFromFilesystem adds all folders from the filesystem; the caller might want to clear() beforehand; the caller needs to lock the mutex beforehand
func (c *animeCollection) loadFromFilesystem() error {
	c.logger.Debugf("Initializing anime collection...")
	files, err := ioutil.ReadDir(c.path)
	if err != nil {
		return err
	}

	for _, file := range files {
		fullPath := filepath.Join(c.path, file.Name())
		if !file.IsDir() {
			c.logger.Debugf("Ignoring file %q", fullPath)
			continue
		}
		_, err := c.addFolder(file.Name())
		if err != nil {
			c.logger.Warnf("Failed to add anime folder %q: %s", fullPath, err)
		}
	}
	c.logger.Debugf("Found %d anime folders", len(c.animeFolders))
	return nil
}

func (c *animeCollection) Clear() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.clear()
}

// clear removes all folders from the collection; fails if the directory is watched; the caller needs to lock the mutex beforehand
func (c *animeCollection) clear() error {
	if c.isWatchingCollection {
		return fmt.Errorf("There is an active filesystem watcher")
	}

	c.logger.Debug("Clearing anime collection...")
	for folderID := range c.animeFolders {
		if err := c.removeFolder(folderID); err != nil {
			c.logger.Errorf("Failed to remove anime folder: %s", err)
		}
	}
	if len(c.animeFolders) != 0 {
		c.logger.Panicf("Failed to clear anime collection. Remaining data: %v", c.animeFolders)
	}
	return nil
}

func (c *animeCollection) onCollectionWatcherEvent(event fsnotify.Event, logger utils.Logger) error {
	fileName := filepath.Base(event.Name)

	if event.Op&fsnotify.Remove != 0 || event.Op&fsnotify.Rename != 0 { // After a rename, a create follows automatically
		logger.Debugf("Detected the removal of %q", fileName)
		if err := c.removeFolderByName(fileName); err != nil {
			// If the file was deleted before the addFolder added it, the remove will fail as well --> ignore it
			logger.Debugf("Failed to remove anime folder %q. "+
				"Maybe it was alreaedy removed before it could get added to the collection?", fileName)
		}
	} else if event.Op&fsnotify.Create != 0 {
		fullPath := filepath.Join(c.path, fileName)
		isDir, err := IsDir(fullPath)
		if err != nil {
			// Maybe the file / directory has already been deleted since the event --> ignore it
			logger.Debugf("Failed to determine if the created element (%q) is a directory. "+
				"Maybe it has been deleted in the mean time?", fileName)
			return nil
		}
		if !isDir { // Not interested
			return nil
		}

		logger.Debugf("Detected the creation of a new directory named %q", fileName)
		_, err = c.addFolder(fileName)
		return err
	}
	return nil
}

func (c *animeCollection) onAnimeFolderWatcherEvent(event fsnotify.Event, logger utils.Logger) error {
	fileName := filepath.Base(event.Name)

	if event.Op&(fsnotify.Remove|fsnotify.Rename|fsnotify.Create) == 0 {
		return nil // Not interested
	}

	logger.Debugf("Detected content modification of anime folder %q", fileName)
	return nil
}

// addFolder adds a new anime folder, or returns the id if the folder has already been added; the caller needs to lock the mutex beforehand
func (c *animeCollection) addFolder(folderName string) (uuid.UUID, error) {
	if folder := c.animeFolder(folderName); folder != nil {
		c.logger.Debugf("Ignoring anime folder %q, because it is already part of the collection", folderName)
		return folder.ID, nil
	}

	c.logger.Debugf("Appending anime folder %q", folderName)
	animeFolder := NewAnimeFolder(c.path, folderName)

	c.animeFolders[animeFolder.ID] = animeFolder

	if c.isWatchingAnimeFolders {
		if err := c.animeFolderWatcher.Add(animeFolder.FullPath()); err != nil {
			// Maybe the directory has already been deleted again --> ignore it
			c.logger.Debugf("Failed to create watcher for the anime folder %q. "+
				"Maybe it has been deleted in the mean time?", folderName)
		}
	}
	return animeFolder.ID, nil
}

// removeFolder removes an existing anime folder from the collection
func (c *animeCollection) removeFolderByName(folderName string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	folder := c.animeFolder(folderName)
	if folder == nil {
		return fmt.Errorf("Folder %q not found", folderName)
	}
	return c.removeFolder(folder.ID)
}

// removeFolder removes an existing anime folder from the collection; the caller needs to lock the mutex beforehand
func (c *animeCollection) removeFolder(folderID uuid.UUID) error {
	folder, ok := c.animeFolders[folderID]
	if !ok {
		return fmt.Errorf("Unable to remove folder. ID %v not found", folderID)
	}

	delete(c.animeFolders, folderID)

	if c.isWatchingAnimeFolders {
		if err := c.animeFolderWatcher.Remove(folder.FullPath()); err != nil {
			// Maybe the file / directory has already been deleted--> ignore it
			c.logger.Debugf("Could not remove watcher for anime folder %q; "+
				"Maybe it has been deleted in the mean time? - %s", folder.FolderName, err)
		}
	}
	return nil
}

func (c *animeCollection) AnimeFolder(folderName string) *AnimeFolder {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.animeFolder(folderName)
}

// animeFolder returns an existing anime folder from the collection; the caller needs to lock the mutex beforehand
func (c *animeCollection) animeFolder(folderName string) *AnimeFolder {
	for _, animeFolder := range c.animeFolders {
		if folderName == animeFolder.FolderName {
			return animeFolder
		}
	}
	return nil
}

func (c *animeCollection) Iterate(callback func(folder *AnimeFolder) bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	for _, animeFolder := range c.animeFolders {
		if !callback(animeFolder) {
			return
		}
	}
}

func (c *animeCollection) Wait() {
	c.wg.Wait()
}
