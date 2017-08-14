package main

import (
	"context"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/maja42/AniScraper/filesystem"
	"github.com/maja42/AniScraper/taskplanner"
)

func main() {
	logger := SetupLogger(logrus.DebugLevel)
	logger.Info("AniScraper started")
	var wg sync.WaitGroup

	// config := DefaultConfig()
	ctx := context.Background()
	ctx, _ = context.WithTimeout(ctx, 8*time.Second)

	// identificationChannel := make(chan *aniscraper.AnimeFolder, 100)

	// webClient := aniscraper.NewWebClient()
	// identifier := aniscraper.NewAnimeIdentifier(webClient, identificationChannel)

	// bindingContext := aniscraper.NewBindingContext(ctx)

	library, err := filesystem.NewAnimeLibrary(100, logger)
	if err != nil {
		logger.Panicf("Failed to create anime library: %s", err)
	}

	taskPlanner := taskplanner.NewTaskPlanner(library, logger)

	// server := webserver.NewWebServer(&config.WebServerConfig,
	// 	bindingContext.ClientBindingContext().NewClient, // Client connected callback
	// 	nil) // Client disconnected callback

	if _, err = library.AddCollection("Test", "D:\\deleteme1"); err != nil {
		logger.Panicf("Failed to add anime collection: %s", err)
	}

	errors, err := library.WatchFilesystem(ctx, true)
	if err != nil {
		logger.Panicf("Failed to watch filesystem: %s", err)
	}

	if _, err = library.AddCollection("Test", "D:\\deleteme2"); err != nil {
		logger.Panicf("Failed to add anime collection: %s", err)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			err, ok := <-errors
			if !ok {
				return
			}
			logger.Errorf("Filesystem watcher error: %s", err)
		}
	}()

	if err := taskPlanner.Start(ctx); err != nil {
		logger.Panicf("Failed to start task planner: %s", err)
	}
	// bindingContext.Initialize(server, animeCollection)

	// Starting...
	// identifier.Start(ctx, 8) // 8 animes at once
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

	if err := library.Clear(); err != nil {
		logger.Panicf("Failed to clear library: %s", err)
	}
	logger.Infof("Waiting for library to finish")
	library.Wait()

	logger.Infof("Waiting for task planner to finish")
	taskPlanner.Wait()

	wg.Wait()
}
