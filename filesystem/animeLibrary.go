package filesystem

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/google/uuid"

	"github.com/maja42/AniScraper/utils"
)

// AnimeLibrary contains a set of different collections, each collection representing a folder with multiple animes
// The AnimeLibrary represents all animes the user has. These animes can be scattered accross multiple locations, called "AnimeCollection".
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
	// WatchFileSystem (re-)initializes  all collections, starts to monitor the underlying filesystem folder and updates the collections automatically; fails if the filesystem is already watched; can abort at any time in case of a fatal error
	WatchFilesystem(ctx context.Context, alsoWatchFolders bool) (<-chan error, error)
	// IsWatching returns true if the filesystem is currently watched
	IsWatching() bool

	// Wait blocks until all go routines (if any) have finished
	Wait()

	// Subscribe returns a new event channel which emits any changes to the library from now on; Optionally emits create-events for any existing folders when the call was made
	Subscribe(ctx context.Context, initializeWithExistingData bool) <-chan Event

	// CollectionCount returns the number of anime collections
	CollectionCount() int
	// CollectionCount returns the total number of anime folders
	AnimeFolderCount() int

	// IterateAnimeFolders calls the given function for every animeFolder in every collection, until false is returned (do not continue) or there are no more folders; Returns false if the iteration was aborted
	IterateAnimeFolders(callback func(*AnimeFolder) bool) bool
}

type animeLibrary struct {
	libraryCollections map[uuid.UUID]libraryCollection

	isWatchingFilesystem   bool
	isWatchingAnimeFolders bool               // If individual anime folders are watched as well
	watchContext           context.Context    // The context that is used for watching
	watchCancelFunc        context.CancelFunc // Used to cancel watching
	watchErrors            chan error         // error output channel for watchers
	watchWg                sync.WaitGroup     // all go routines that are running as part of the filesystem watchers

	subMutex               sync.Mutex
	subscribers            []subscriber
	eventChannelBufferSize int

	mutex  sync.RWMutex
	wg     sync.WaitGroup
	logger utils.Logger
}

type libraryCollection struct {
	collection            AnimeCollection
	cancelEventForwarding chan struct{} // Closed as soon as the collection is removed from the library
	cancelWatching        context.CancelFunc
}

type subscriber struct {
	events chan Event
}

// NewAnimeLibrary returns a new and empty anime library
func NewAnimeLibrary(eventChannelBufferSize int, logger utils.Logger) (AnimeLibrary, error) {
	return &animeLibrary{
		libraryCollections:     make(map[uuid.UUID]libraryCollection, 0),
		subscribers:            make([]subscriber, 0),
		eventChannelBufferSize: eventChannelBufferSize,
		logger:                 logger.New("AnimeLibrary"),
	}, nil
}

func (l *animeLibrary) AddCollection(name string, path string) (uuid.UUID, error) {
	collection, err := NewAnimeCollection(name, path, l.eventChannelBufferSize, l.logger)
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

	if err := l.addCollection(collection); err != nil {
		return uuid.Nil, err
	}
	return collection.ID(), nil
}

// removeFolder adds a new anime collection to the library; the caller needs to lock the mutex beforehand
func (l *animeLibrary) addCollection(collection AnimeCollection) error {
	l.logger.Debugf("Adding collection %q", collection.Name())

	libCollection := libraryCollection{
		collection: collection,
	}

	libCollection.cancelEventForwarding = make(chan struct{})
	l.libraryCollections[collection.ID()] = libCollection

	l.wg.Add(1)
	go func() { // forward events
		defer l.wg.Done()
		for {
			select {
			case event := <-libCollection.collection.Events():
				l.informSubscribers(event)
			case _, ok := <-libCollection.cancelEventForwarding:
				if !ok {
					return
				}
			}
		}
	}()

	if l.isWatchingAnimeFolders {
		if err := l.watchCollectionFilesystem(libCollection); err != nil {
			close(libCollection.cancelEventForwarding)
			delete(l.libraryCollections, collection.ID())
			return fmt.Errorf("Could not add collection %q, because watching the filesystem failed: %s", collection.Name(), err)
		}
	}
	return nil
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

	if libraryCollection.cancelWatching != nil {
		libraryCollection.cancelWatching()
	}
	close(libraryCollection.cancelEventForwarding)
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
	l.mutex.Lock()
	defer l.mutex.Unlock()
	if l.isWatchingFilesystem {
		return nil, fmt.Errorf("The filesystem is already watched")
	}

	l.isWatchingFilesystem = true
	l.isWatchingAnimeFolders = alsoWatchFolders
	l.watchContext, l.watchCancelFunc = context.WithCancel(ctx)
	l.watchErrors = make(chan error)

	for _, libraryCollection := range l.libraryCollections {
		if err := l.watchCollectionFilesystem(libraryCollection); err != nil {
			l.watchErrors <- fmt.Errorf("Aborting filesystem watching - could not watch collection %q: %s",
				libraryCollection.collection.Name(), err)
			l.watchCancelFunc() // abort already running watches
			break
		}
	}

	l.wg.Add(1)
	go func() {
		defer l.wg.Done()
		l.watchWg.Wait()
		l.mutex.Lock()
		defer l.mutex.Unlock()
		l.isWatchingFilesystem = false
		close(l.watchErrors)
	}()
	return l.watchErrors, nil

}

func (l *animeLibrary) watchCollectionFilesystem(libCollection libraryCollection) error {
	watchCtx, cancel := context.WithCancel(l.watchContext)

	errors, err := libCollection.collection.WatchFilesystem(watchCtx, l.isWatchingAnimeFolders)
	if err != nil {
		cancel()
		return err
	}
	libCollection.cancelWatching = cancel

	forwardErrors := func(in <-chan error, out chan<- error) {
		for err := range in {
			out <- err
		}
		l.mutex.Lock()
		defer l.mutex.Unlock()
		libCollection.cancelWatching = nil
		l.watchWg.Done()
	}

	l.watchWg.Add(1)
	go forwardErrors(errors, l.watchErrors)
	return nil
}

func (l *animeLibrary) IsWatching() bool {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	return l.isWatchingFilesystem
}

func (l *animeLibrary) Wait() {
	if l.CollectionCount() > 0 {
		// If there are collections, there are event-forwarding routines. So the library needs to be clear() before wait can be called
		l.logger.Panicf("Cannot wait() for AnimeLibrary if it still contains collections")
	}
	l.wg.Wait()
}

func (l *animeLibrary) Subscribe(ctx context.Context, initializeWithExistingData bool) <-chan Event {
	l.subMutex.Lock()
	defer l.subMutex.Unlock()

	newSub := subscriber{
		events: make(chan Event, l.eventChannelBufferSize),
	}

	if initializeWithExistingData {
		l.mutex.Lock() // Do not change the collections / anime folders before we registered the subscriber
		defer l.mutex.Unlock()
		l.iterateAnimeFolders(func(animeFolder *AnimeFolder) bool {
			newSub.events <- Event{
				EventType:   FOLDER_ADDED,
				AnimeFolder: animeFolder,
			}
			return true
		})
	}
	l.subscribers = append(l.subscribers, newSub)

	l.wg.Add(1)
	go func() {
		defer l.wg.Done()
		<-ctx.Done()

		l.subMutex.Lock()
		defer l.subMutex.Unlock()

		found := false
		for i, sub := range l.subscribers {
			if sub == newSub {
				subscriberCount := len(l.subscribers)
				l.subscribers[i] = l.subscribers[subscriberCount-1]
				l.subscribers = l.subscribers[:subscriberCount-1]
				found = true
				break
			}
		}
		if !found {
			l.logger.Panicf("Unable to find subscriber")
		}
		close(newSub.events)
	}()
	return newSub.events
}

func (l *animeLibrary) informSubscribers(event Event) {
	for _, sub := range l.subscribers {
		sub.events <- event
	}
}

func (l *animeLibrary) CollectionCount() int {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	return len(l.libraryCollections)
}

func (l *animeLibrary) AnimeFolderCount() int {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	cnt := 0
	for _, libraryCollection := range l.libraryCollections {
		cnt += libraryCollection.collection.AnimeFolderCount()
	}
	return cnt
}

func (l *animeLibrary) IterateAnimeFolders(callback func(*AnimeFolder) bool) bool {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	return l.iterateAnimeFolders(callback)
}

func (l *animeLibrary) iterateAnimeFolders(callback func(*AnimeFolder) bool) bool {
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
