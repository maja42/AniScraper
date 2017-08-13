package aniscraper

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/google/uuid"

	"github.com/maja42/AniScraper/utils"
)

// AnimeLibrary contains a set of different collections, each collection representing a folder with multiple animes
type AnimeLibrary interface {
	// AddCollection adds all subfolders of 'path' to the collection. The number of added folders is returned.
	AddCollection(name string, path string) (uuid.UUID, error)
	// Contains checks if the given path / anime folder is already part of the collection
	Contains(folder os.FileInfo) AnimeCollection
	// RemoveCollection removes a single collection from the library
	RemoveCollection(collection uuid.UUID) error
	// Clear removes all anime collections
	Clear() error

	// LoadFromFilesystem (re-)initializes all collections with the data from the filesystem; fails if the filesystem is currently watched
	LoadFromFilesystem() error
	// WatchFileSystem (re-)initializes  all collections, starts to monitor the underlying filesystem folder and updates the collections automatically; fails if the filesystem is already watched
	WatchFilesystem(ctx context.Context, alsoWatchFolders bool) (<-chan error, error)

	// // IterateCollections calls the given function for every animeCollection, until false is returned (do not continue) or there are no more collections; Returns false if the iteration was aborted
	// IterateCollections(callback func(AnimeCollection) bool) bool
	// // IterateAnimeFolders calls the given function for every animeFolder in every collection, until false is returned (do not continue) or there are no more folders; Returns false if the iteration was aborted
	// IterateAnimeFolders(callback func(AnimeCollection, *AnimeFolder) bool) bool
	// IterateAnimeFolders calls the given function for every animeFolder in every collection, until false is returned (do not continue) or there are no more folders; Returns false if the iteration was aborted
	IterateAnimeFolders(callback func(*AnimeFolder) bool) bool
}

type animeLibrary struct {
	libraryCollections map[uuid.UUID]libraryCollection

	isWatchingFilesystem bool

	mutex  sync.RWMutex
	logger utils.Logger
}

type libraryCollection struct {
	collection   AnimeCollection
	watchContext context.Context
}

// NewAnimeLibrary returns a new and empty anime library
func NewAnimeLibrary(logger utils.Logger) (AnimeLibrary, error) {
	return &animeLibrary{
		libraryCollections: make(map[uuid.UUID]libraryCollection, 0),
		logger:             logger.New("AnimeLibrary"),
	}, nil
}

func (l *animeLibrary) AddCollection(name string, path string) (uuid.UUID, error) {
	collection, err := NewAnimeCollection(name, path, l.logger)
	if err != nil {
		return uuid.Nil, err
	}

	l.mutex.Lock()
	defer l.mutex.Unlock()

	fileInfo, err := GetFileInfo(collection.Path())
	if err != nil {
		return uuid.Nil, fmt.Errorf("Cannot add collection to library: %s", err)
	}
	if existing := l.contains(fileInfo); existing != nil {
		return existing.ID(), nil
	}

	l.logger.Debugf("Adding collection %q", collection.Name())
	l.libraryCollections[collection.ID()] = libraryCollection{
		collection: collection,
	}
	return collection.ID(), nil
}

func (l *animeLibrary) RemoveCollection(collectionID uuid.UUID) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	return l.removeCollection(collectionID)
}

// removeFolder removes an existing anime collection from the library; the caller needs to lock the mutex beforehand
func (l *animeLibrary) removeCollection(collectionID uuid.UUID) error {
	libraryCollection, ok := l.libraryCollections[collectionID]
	if !ok {
		return fmt.Errorf("Unable to remove collection. ID %v not found", collectionID)
	}
	l.logger.Debugf("Removing collection %q from the library", libraryCollection.collection.Name())
	delete(l.libraryCollections, collectionID)
	return nil
}

func (l *animeLibrary) Clear() error {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.logger.Infof("Removing all collections from the anime library")
	for id := range l.libraryCollections {
		if err := l.removeCollection(id); err != nil {
			return err
		}
	}
	if len(l.libraryCollections) != 0 {
		l.logger.Panicf("Failed to clear anime library. Remaining data: %v", l.libraryCollections)
	}
	return nil
}

func (l *animeLibrary) LoadFromFilesystem() error {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	if l.isWatchingFilesystem {
		return fmt.Errorf("The filesystem is currently watched")
	}

	for _, libraryCollection := range l.libraryCollections {
		if err := libraryCollection.collection.LoadFromFilesystem(); err != nil {
			return err
		}
	}
	return nil
}

func (l *animeLibrary) WatchFilesystem(ctx context.Context, alsoWatchFolders bool) (<-chan error, error) {
	return nil, fmt.Errorf("Not implemented yet")
}

// func (l *animeLibrary) IterateCollections(callback func(AnimeCollection) bool) bool {
// 	l.mutex.RLock()
// 	defer l.mutex.RUnlock()

// 	for _, collection := range l.collections {
// 		if !callback(collection) {
// 			return false
// 		}
// 	}
// 	return true
// }

// func (l *animeLibrary) IterateAnimeFolders(callback func(AnimeCollection, *AnimeFolder) bool) bool {
// 	l.mutex.RLock()
// 	defer l.mutex.RUnlock()

// 	for _, collection := range l.collections {
// 		cont := collection.Iterate(func(folder *AnimeFolder) bool {
// 			return callback(collection, folder)
// 		})
// 		if !cont {
// 			return false
// 		}
// 	}
// 	return true
// }

func (l *animeLibrary) IterateAnimeFolders(callback func(*AnimeFolder) bool) bool {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	for _, libraryCollection := range l.libraryCollections {
		if !libraryCollection.collection.Iterate(callback) {
			return false
		}
	}
	return true
}

func (l *animeLibrary) Contains(folder os.FileInfo) AnimeCollection {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	return l.contains(folder)
}

func (l *animeLibrary) contains(folder os.FileInfo) AnimeCollection {
	for _, libraryCollection := range l.libraryCollections {
		collection := libraryCollection.collection
		collectionFileInfo, err := GetFileInfo(collection.Path())
		if err != nil {
			log.Errorf("Failed to query file info of collection %q: %v", collection.Name(), err)
			continue
		}

		if os.SameFile(folder, collectionFileInfo) {
			return collection
		}
	}
	return nil
}
