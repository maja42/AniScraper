package main

import (
	"context"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/maja42/AniScraper/aniscraper"
)

func main() {
	logger := SetupLogger(logrus.DebugLevel)
	logger.Info("AniScraper started")
	var wg sync.WaitGroup

	// config := DefaultConfig()
	ctx := context.Background()
	ctx, _ = context.WithTimeout(ctx, 8*time.Second)

	identificationChannel := make(chan *aniscraper.AnimeFolder, 100)

	webClient := aniscraper.NewWebClient()
	identifier := aniscraper.NewAnimeIdentifier(webClient, identificationChannel)

	// bindingContext := aniscraper.NewBindingContext(ctx)

	folder := "D:\\deleteme"

	// server := webserver.NewWebServer(&config.WebServerConfig,
	// 	bindingContext.ClientBindingContext().NewClient, // Client connected callback
	// 	nil) // Client disconnected callback

	animeCollection, err := aniscraper.NewAnimeCollection("Test", folder, logger)
	if err != nil {
		logger.Panicf("Failed to create anime collection %q: %s", folder, err)
	}

	errors, err := animeCollection.WatchFilesystem(ctx, true)
	if err != nil {
		logger.Panicf("Failed to create anime collection %q: %s", folder, err)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			err, ok := <-errors
			if !ok {
				return
			}
			logger.Errorf("AC watcher error: %s", err)
		}
	}()

	// bindingContext.Initialize(server, animeCollection)

	// Starting...
	identifier.Start(ctx, 8) // 8 animes at once
	// server.Start(ctx)

	// count, err := animeCollection.AddCollection(folder)
	// if err != nil {
	// 	logger.Fatalf("Failed to add folder %q: %v", folder, err)
	// }
	// logger.Infof("Added %d folders", count)

	// watcher, err := fsnotify.NewWatcher()
	// if err != nil {
	// 	logger.Fatal(err)
	// }
	// defer watcher.Close()

	// go func() {
	// 	for {
	// 		select {
	// 		case event := <-watcher.Events:
	// 			logger.Println("event:", event)
	// 			if event.Op&fsnotify.Write == fsnotify.Write {
	// 				logger.Println("modified file:", event.Name)
	// 			}
	// 		case err := <-watcher.Errors:
	// 			logger.Println("error:", err)
	// 		}
	// 	}
	// }()

	// err = watcher.Add(folder)
	// if err != nil {
	// 	logger.Fatal(err)
	// }

	<-ctx.Done()

	logger.Infof("Waiting for anime collection to finish")
	animeCollection.Wait()

	wg.Wait()

	animeCollection.Iterate(func(folder *aniscraper.AnimeFolder) bool {
		logger.Warnf("%s", folder.FolderName)
		return true
	})
}
