package filesystem

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/maja42/AniScraper/utils"
	"github.com/stretchr/testify/assert"
)

func TestNewAnimeLibrary(t *testing.T) {
	lib, err := NewAnimeLibrary(0, utils.NewNoLogger())
	animeLib := lib.(*animeLibrary)
	assert.NotNil(t, animeLib.logger)
	assert.False(t, animeLib.isWatchingFilesystem)
	assert.NoError(t, err)
}

func TestAddCollections(t *testing.T) {
	fileStructure := makeDefaultFileStructure()
	tmpDir := setupTempFilesystem(t, fileStructure)
	defer teardownTempFileSystem(t, tmpDir)

	lib, err := NewAnimeLibrary(0, utils.NewNoLogger())
	assert.NoError(t, err)
	addCollectionsToAnimeLibrary(t, lib, tmpDir, fileStructure)

	assert.Equal(t, len(fileStructure), lib.CollectionCount())
	assert.Equal(t, 0, lib.AnimeFolderCount())
}

func TestLoadFromFileSystem(t *testing.T) {
	fileStructure := makeDefaultFileStructure()
	tmpDir := setupTempFilesystem(t, fileStructure)
	defer teardownTempFileSystem(t, tmpDir)

	lib, err := NewAnimeLibrary(0, utils.NewNoLogger())
	assert.NoError(t, err)
	addCollectionsToAnimeLibrary(t, lib, tmpDir, fileStructure)

	err = lib.LoadFromFilesystem()
	assert.NoError(t, err)
	assert.False(t, lib.IsWatching())

	assert.Equal(t, len(fileStructure), lib.CollectionCount())
	assert.Equal(t, 6, lib.AnimeFolderCount())

	iteratedFolderPaths := make([]string, 0)
	completed := lib.IterateAnimeFolders(func(folder *AnimeFolder) bool {
		iteratedFolderPaths = append(iteratedFolderPaths, folder.FullPath())
		return true
	})
	assert.True(t, completed)

	for _, c := range fileStructure {
		for _, s := range c.Subfolders {
			assert.Contains(t, iteratedFolderPaths, filepath.Join(tmpDir, c.FolderName, s))
		}
	}
}

func TestLoadFromFileSystemViaWatches(t *testing.T) {
	fileStructure := makeDefaultFileStructure()
	tmpDir := setupTempFilesystem(t, fileStructure)
	defer teardownTempFileSystem(t, tmpDir)

	lib, err := NewAnimeLibrary(0, utils.NewNoLogger())
	assert.NoError(t, err)
	addCollectionsToAnimeLibrary(t, lib, tmpDir, fileStructure)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	errors, err := lib.WatchFilesystem(ctx, true)

	var wg sync.WaitGroup
	errorCount := 0
	wg.Add(1)
	go func() {
		defer wg.Done()
		for range errors {
			errorCount++
		}
	}()

	assert.NoError(t, err)
	assert.True(t, lib.IsWatching())

	assert.Equal(t, len(fileStructure), lib.CollectionCount())
	assert.Equal(t, 6, lib.AnimeFolderCount())

	cancel()
	lib.Clear()
	lib.Wait()
	assert.False(t, lib.IsWatching())

	assert.Equal(t, 0, lib.CollectionCount())
	assert.Equal(t, 0, lib.AnimeFolderCount())

	wg.Wait()
	assert.Equal(t, 0, errorCount)
}

func TestLoadFromFileSystem_Async(t *testing.T) {
	fileStructure := makeDefaultFileStructure()

	lib, err := NewAnimeLibrary(0, utils.NewNoLogger())
	assert.NoError(t, err)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	errors, err := lib.WatchFilesystem(ctx, true)
	var wg sync.WaitGroup
	errorCount := 0
	wg.Add(1)
	go func() {
		defer wg.Done()
		for range errors {
			errorCount++
		}
	}()
	assert.NoError(t, err)
	assert.True(t, lib.IsWatching())

	assert.Equal(t, 0, lib.CollectionCount())
	assert.Equal(t, 0, lib.AnimeFolderCount())

	tmpDir := setupTempFilesystem(t, fileStructure)
	defer teardownTempFileSystem(t, tmpDir)
	addCollectionsToAnimeLibrary(t, lib, tmpDir, fileStructure)

	<-time.After(500 * time.Millisecond)
	assert.Equal(t, len(fileStructure), lib.CollectionCount())
	assert.Equal(t, 6, lib.AnimeFolderCount())

	cancel()
	lib.Clear()
	lib.Wait()
	assert.False(t, lib.IsWatching())

	assert.Equal(t, 0, lib.CollectionCount())
	assert.Equal(t, 0, lib.AnimeFolderCount())

	wg.Wait()
	assert.Equal(t, 0, errorCount)
}

type CollectionFileStructure struct {
	FolderName string
	Subfolders []string
}

func makeDefaultFileStructure() []CollectionFileStructure {
	return []CollectionFileStructure{
		{"Collection1", []string{"Col1 AF1"}},
		{"Collection2", []string{"Col2 AF1", "Col2 AF2"}},
		{"Collection3", []string{"Col3 AF1", "Col3 AF2", "Col3 AF3"}},
	}
}

func setupTempFilesystem(t *testing.T, filestructure []CollectionFileStructure) string {
	tmpDir, err := ioutil.TempDir("", "AniScraperLib")
	assert.NoError(t, err)

	for _, c := range filestructure {
		err := os.Mkdir(filepath.Join(tmpDir, c.FolderName), os.ModePerm)
		assert.NoError(t, err)

		for _, s := range c.Subfolders {
			err := os.Mkdir(filepath.Join(tmpDir, c.FolderName, s), os.ModePerm)
			assert.NoError(t, err)
		}
	}
	return tmpDir
}

func teardownTempFileSystem(t *testing.T, tmpDir string) {
	err := os.RemoveAll(tmpDir)
	assert.NoError(t, err)
}

func addCollectionsToAnimeLibrary(t *testing.T, lib AnimeLibrary, tmpDir string, filestructure []CollectionFileStructure) {
	for key, c := range filestructure {
		_, err := lib.AddCollection("name "+strconv.Itoa(key), filepath.Join(tmpDir, c.FolderName))
		assert.NoError(t, err)
	}
}
