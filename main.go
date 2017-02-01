package main

import (
	"context"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/maja42/AniScraper/filesystem"
	"github.com/maja42/AniScraper/webserver"
)

func main() {
	SetupLogger(logrus.DebugLevel)
	log.Info("AniScraper started")

	config := DefaultConfig()

	ctx := context.Background()

	server := webserver.NewWebServer(&config.WebServerConfig)
	exchange := server.Exchange()

	echoChannel := exchange.Subscribe([]string{"echo"})
	echoReplyChannel := exchange.Subscribe([]string{"echo-reply"})
	go func() {
		for {
			select {
			case message, ok := <-echoChannel:
				if !ok {
					return
				}
				log.Infof("Responding to echo message: %v", message.Content)
				server.Send(message.Sender, "echo-reply", message.Content)
			case message, ok := <-echoReplyChannel:
				if !ok {
					return
				}
				log.Infof("Received echo reply message: %v", message.Content)
			}
		}
	}()

	server.Start(ctx)

	var mediaCollection = filesystem.NewMediaCollection()

	newFolderChannel := mediaCollection.Exchange().Subscribe([]string{"newMediaFolder"})
	go func() {
		for {
			message, ok := <-newFolderChannel
			if !ok {
				return
			}

			server.Send(message.Sender, "newMediaFolder", message.Content)

		}
	}()

	time.Sleep(3 * time.Second)

	folder := "M:"

	count, err := mediaCollection.AddCollection(folder)
	if err != nil {
		log.Fatalf("Failed to add folder %q: %v", folder, err)
	}
	log.Infof("Added %d folders", count)

	<-ctx.Done()
}
