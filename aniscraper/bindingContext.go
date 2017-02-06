package aniscraper

import (
	"context"
	"fmt"
	"sync"

	"github.com/maja42/AniScraper/webserver"
)

// ServerBindingContext contains all methods that are callable by the server-side (golang)
type ServerBindingContext interface {
	NewAnimeFolder(afid int, folder *AnimeFolder)
}

// ClientBindingContext contains all methods that are callable by the client side (incoming webserver messages)
type ClientBindingContext interface {
	NewClient(cid int)
}

// BindingContext contains special methods for dealing with server-client-bindings
type BindingContext interface {
	Initialize(webserver webserver.WebServer, animeCollection AnimeCollection) error

	ServerBindingContext() ServerBindingContext
	ClientBindingContext() ClientBindingContext
}

type bindingContext struct {
	mutex       sync.RWMutex
	initialized bool
	ctx         context.Context

	webserver       webserver.WebServer
	wsExchange      webserver.MessageExchange
	animeCollection AnimeCollection
}

func NewBindingContext(ctx context.Context) BindingContext {
	bindingContext := &bindingContext{
		ctx: ctx,
	}
	return bindingContext
}

func (binding *bindingContext) Initialize(webserver webserver.WebServer, animeCollection AnimeCollection) error {
	binding.mutex.Lock()
	defer binding.mutex.Unlock()
	if binding.initialized {
		return fmt.Errorf("The binding context has already been initialized")
	}

	binding.webserver = webserver
	binding.wsExchange = webserver.Exchange()
	binding.animeCollection = animeCollection

	go binding.echoHandling()

	binding.initialized = true
	return nil
}

func (binding *bindingContext) echoHandling() {
	echoChannel := binding.wsExchange.Subscribe([]string{"echo"})
	echoReplyChannel := binding.wsExchange.Subscribe([]string{"echo-reply"})

	for {
		select {
		case <-binding.ctx.Done():
			return
		case message, ok := <-echoChannel:
			if !ok {
				return
			}
			log.Infof("Responding to echo message: %v", message.Content)
			if err := message.Respond("echo-reply", message.Content); err != nil {
				log.Error(err)
			}
		case message, ok := <-echoReplyChannel:
			if !ok {
				return
			}
			log.Infof("Received echo reply message: %v", message.Content)
		}
	}
}

func (binding *bindingContext) ServerBindingContext() ServerBindingContext {
	return binding
}

func (binding *bindingContext) ClientBindingContext() ClientBindingContext {
	return binding
}

func (srvBinding *bindingContext) NewAnimeFolder(afid int, folder *AnimeFolder) {

}

func (cliBinding *bindingContext) NewClient(cid int) {

}
