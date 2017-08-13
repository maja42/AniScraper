package aniscraper

// import (
// 	"context"
// 	"fmt"
// 	"sync"

// 	"github.com/maja42/AniScraper/webserver"
// )

// // ServerBindingContext contains all methods that are callable by the server-side (golang)
// type ServerBindingContext interface {
// 	NewAnimeFolder(folder *AnimeFolder)
// }

// // ClientBindingContext contains all methods that are callable by the client side (incoming webserver messages)
// type ClientBindingContext interface {
// 	NewClient(cid int)
// }

// // BindingContext contains special methods for dealing with server-client-bindings
// type BindingContext interface {
// 	Initialize(webserver webserver.WebServer, animeCollection AnimeCollection) error

// 	ServerBindingContext() ServerBindingContext
// 	ClientBindingContext() ClientBindingContext
// }

// type bindingContext struct {
// 	mutex       sync.RWMutex
// 	initialized bool
// 	ctx         context.Context

// 	webserver       webserver.WebServer
// 	wsExchange      webserver.MessageExchange
// 	animeCollection AnimeCollection
// }

// func NewBindingContext(ctx context.Context) BindingContext {
// 	bindingContext := &bindingContext{
// 		ctx: ctx,
// 	}
// 	return bindingContext
// }

// func (binding *bindingContext) Initialize(webserver webserver.WebServer, animeCollection AnimeCollection) error {
// 	binding.mutex.Lock()
// 	defer binding.mutex.Unlock()
// 	if binding.initialized {
// 		return fmt.Errorf("The binding context has already been initialized")
// 	}

// 	binding.webserver = webserver
// 	binding.wsExchange = webserver.Exchange()
// 	binding.animeCollection = animeCollection

// 	go binding.echoHandling()

// 	binding.initialized = true
// 	return nil
// }

// func (binding *bindingContext) echoHandling() {
// 	echoChannel := binding.wsExchange.Subscribe([]string{"echo"})
// 	echoReplyChannel := binding.wsExchange.Subscribe([]string{"echo-reply"})

// 	for {
// 		select {
// 		case <-binding.ctx.Done():
// 			return
// 		case message, ok := <-echoChannel:
// 			if !ok {
// 				return
// 			}
// 			log.Infof("Responding to echo message: %v", message.Content)
// 			if err := message.Respond("echo-reply", message.Content); err != nil {
// 				log.Error(err)
// 			}
// 		case message, ok := <-echoReplyChannel:
// 			if !ok {
// 				return
// 			}
// 			log.Infof("Received echo reply message: %v", message.Content)
// 		}
// 	}
// }

// func (binding *bindingContext) ServerBindingContext() ServerBindingContext {
// 	return binding
// }

// func (binding *bindingContext) ClientBindingContext() ClientBindingContext {
// 	return binding
// }

// func (srvBinding *bindingContext) NewAnimeFolder(folder *AnimeFolder) {
// 	if err := srvBinding.synchronizeAnimeFolder(folder, -1); err != nil {
// 		log.Errorf("Failed to broadcast new anime folder %s", err)
// 	}
// }

// func (binding *bindingContext) synchronizeAnimeFolder(folder *AnimeFolder, cid int) error {
// 	messageType := "newAnimeFolder"
// 	message := struct {
// 		AfId       AnimeFolderID
// 		FolderName string
// 	}{
// 		folder.ID,
// 		folder.Name,
// 	}

// 	return binding.webserver.Transmit(cid, messageType, message)
// }

// func (cliBinding *bindingContext) NewClient(cid int) {
// 	// Synchronize AnimeCollection
// 	cliBinding.webserver.Send(cid, "clearAnimeCollection", nil)

// 	cliBinding.animeCollection.Iterate(func(folder *AnimeFolder) bool {
// 		if err := cliBinding.synchronizeAnimeFolder(folder, cid); err != nil {
// 			log.Errorf("Failed to synchronize anime collection %s", err)
// 			return false
// 		}
// 		return true // continue iterating
// 	})
// }
