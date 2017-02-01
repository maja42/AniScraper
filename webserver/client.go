package webserver

import (
	"context"
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
)

// Client represents a single connected client that communicates over a websocket
type Client interface {
	Handle(ctx context.Context) error
}

type client struct {
	mutex   sync.RWMutex
	handled bool

	ctx    context.Context
	socket *websocket.Conn
}

func (client *client) Handle(ctx context.Context) error {
	client.mutex.Lock()

	if client.handled {
		client.mutex.Unlock()
		return fmt.Errorf("The client is already handled")
	}
	client.handled = true
	client.ctx = ctx
	client.mutex.Unlock()

	return client.handle()
}

func (client *client) handle() error {
	for {
		msgType, msg, err := client.socket.ReadMessage()
		if err != nil {
			return err
		}

		log.Infof("Received message: %s (type=%v)", string(msg), msgType)
		err = client.socket.WriteMessage(websocket.TextMessage, []byte("got it"))
		if err != nil {
			return err
		}
	}
}
