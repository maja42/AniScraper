package main

import (
	"context"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/maja42/AniScraper/aniscraper"
	"github.com/maja42/AniScraper/webserver"
)

func main() {
	SetupLogger(logrus.DebugLevel)
	log.Info("AniScraper started")

	config := DefaultConfig()
	ctx := context.Background()

	bindingContext := aniscraper.NewBindingContext(ctx)

	server := webserver.NewWebServer(&config.WebServerConfig,
		bindingContext.ClientBindingContext().NewClient, // Client connected callback
		nil) // Client disconnected callback

	animeCollection := aniscraper.NewAnimeCollection(bindingContext.ServerBindingContext())

	bindingContext.Initialize(server, animeCollection)

	// exchange := server.Exchange()

	// echoChannel := exchange.Subscribe([]string{"echo"})
	// echoReplyChannel := exchange.Subscribe([]string{"echo-reply"})
	// go func() {
	// 	for {
	// 		select {
	// 		case message, ok := <-echoChannel:
	// 			if !ok {
	// 				return
	// 			}
	// 			log.Infof("Responding to echo message: %v", message.Content)
	// 			server.Send(message.Sender, "echo-reply", message.Content)
	// 		case message, ok := <-echoReplyChannel:
	// 			if !ok {
	// 				return
	// 			}
	// 			log.Infof("Received echo reply message: %v", message.Content)
	// 		}
	// 	}
	// }()

	server.Start(ctx)

	// newFolderChannel := animeCollection.Exchange().Subscribe([]string{"newAnimeFolder"})
	// go func() {
	// 	for {
	// 		message, ok := <-newFolderChannel
	// 		if !ok {
	// 			return
	// 		}

	// 		server.Send(message.Sender, "newAnimeFolder", message.Content)

	// 	}
	// }()

	time.Sleep(3 * time.Second)

	folder := "M:"

	count, err := animeCollection.AddCollection(folder)
	if err != nil {
		log.Fatalf("Failed to add folder %q: %v", folder, err)
	}
	log.Infof("Added %d folders", count)

	<-ctx.Done()
}
