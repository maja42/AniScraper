package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/maja42/AniScraper/filesystem"
)

func main() {
	SetupLogger(logrus.DebugLevel)
	log.Info("AniScraper started")

	var mediaCollection = filesystem.NewMediaCollection()

	folder := "M:"

	count, err := mediaCollection.AddCollection(folder)
	if err != nil {
		log.Fatalf("Failed to add folder %q: %v", folder, err)
	}
	log.Infof("Added %d folders", count)
}
