package webserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/maja42/AniScraper/utils"
	"github.com/maja42/AniScraper/webserver/exchange"
)

type WebServer interface {
	Start(ctx context.Context)
	Exchange() exchange.MessageExchange
	Send(destination int, topic string, message interface{}) error
	Broadcast(topic string, message interface{}) error
}

type webServer struct {
	mutex   sync.RWMutex
	started bool
	ctx     context.Context

	server           http.Server
	clientIdSequence utils.Sequence
	clients          map[int]*Client // clientId => client

	exchange exchange.MessageExchange
}

// Client represents a single connected client that communicates via a websocket
type Client struct {
	socket *websocket.Conn
}

// Message defines the JSON format for client communication
type Message struct {
	Type    string      `json:"messageType"`
	Message interface{} `json:"message"`
}

func NewWebServer(config *WebServerConfig) WebServer {
	webserver := &webServer{
		started:          false,
		clientIdSequence: utils.NewSequenceGenerator(0),
		clients:          make(map[int]*Client),
		exchange:         exchange.NewMessageExchange(),
	}

	var handler = http.NewServeMux()

	// static content
	handler.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("webapp"))))

	// websocket communication
	handler.HandleFunc("/websocket", webserver.websocketHandler)

	webserver.server = http.Server{
		Addr:         config.AddressBinding,
		Handler:      handler,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	}

	return webserver
}

func (server *webServer) Start(ctx context.Context) {
	server.mutex.Lock()
	defer server.mutex.Unlock()

	if server.started {
		panic("The webserver already started")
	}
	server.started = true
	server.ctx = ctx

	// Start webserver
	go func() {
		log.Info("Starting webserver")
		err := server.server.ListenAndServe()
		if err != nil {
			panic(err)
		}
	}()

	go func() {
		<-server.ctx.Done()
		log.Info("Shutting down")
		server.exchange.Shutdown()
	}()
}

// websocketHandler receives incomming websocket connections and starts client handlers
func (server *webServer) websocketHandler(res http.ResponseWriter, req *http.Request) {
	conn, err := (&websocket.Upgrader{CheckOrigin: server.checkOrigin}).Upgrade(res, req, nil)
	if err != nil {
		http.NotFound(res, req)
		log.Warnf("Failed to accept client: %v", err)
		return
	}
	log.Info("Client connected")

	client := &Client{
		socket: conn,
	}

	cid := server.clientIdSequence.Next()

	server.mutex.Lock()
	defer server.mutex.Unlock()

	server.clients[cid] = client

	log.Debugf("Total clients connected: %d", len(server.clients))

	// Handle client:
	go func() {
		err := server.handleIncomingClientMessages(cid, client)
		log.Infof("Client disconnected: %v", err)

		server.mutex.Lock()
		defer server.mutex.Unlock()
		delete(server.clients, cid)

		log.Debugf("Total clients connected: %d", len(server.clients))
	}()

	log.Debug("Sending echo message to new client")
	go server.Send(cid, "echo", "Hello there!")
}

// checkOrigin decides if the given request is allowed to be processed
func (server *webServer) checkOrigin(r *http.Request) bool {
	return true
}

// handleIncomingClientMessages receives all incomming websocket messages, unmarshals them and broadcasts them on the message exchange.
func (server *webServer) handleIncomingClientMessages(cid int, client *Client) error {
	defer client.socket.Close()
	for {
		msgType, msg, err := client.socket.ReadMessage()
		if err != nil {
			return err
		}
		log.Debugf("Received message. Raw content: %q (type %d)", string(msg), msgType)

		var message Message
		if err := json.Unmarshal(msg, &message); err != nil {
			log.Errorf("Failed to unmarshal incoming websocket message. Disconnecting. \nMessage: %q", string(msg))
			return fmt.Errorf("Disconnected due to protocol error")
		}
		log.Infof("Received message (type %q): %v", message.Type, message.Message)

		recipients := server.exchange.Publish(message.Type, message.Message, cid)
		if recipients == 0 {
			log.Warnf("There are no recipients for messages of type %q", message.Type)
		}
	}
}

func (server *webServer) Exchange() exchange.MessageExchange {
	return server.exchange
}

func (server *webServer) Send(destination int, topic string, message interface{}) error {
	textMessage, err := server.marshal(topic, message)
	if err != nil {
		return err
	}

	server.mutex.RLock()
	defer server.mutex.RUnlock()

	client, ok := server.clients[destination]
	if !ok {
		return fmt.Errorf("Failed to send message: Destination %d is unknown", destination)
	}

	return client.socket.WriteMessage(websocket.TextMessage, textMessage)
}

// Broadcast sends the given message to all connected clients
func (server *webServer) Broadcast(topic string, message interface{}) error {
	textMessage, err := server.marshal(topic, message)
	if err != nil {
		return err
	}

	server.mutex.RLock()
	defer server.mutex.RUnlock()

	for _, client := range server.clients {
		if err := client.socket.WriteMessage(websocket.TextMessage, textMessage); err != nil {
			return err
		}
	}
	return nil
}

func (server *webServer) marshal(topic string, message interface{}) ([]byte, error) {
	messageObject := &Message{
		Type:    topic,
		Message: message,
	}
	return json.Marshal(messageObject)
}
