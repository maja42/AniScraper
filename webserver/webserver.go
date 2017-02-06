package webserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/maja42/AniScraper/utils"
)

type WebServer interface {
	Start(ctx context.Context)
	Exchange() MessageExchange

	Send(cid int, topic string, message interface{}) error
	Broadcast(topic string, message interface{}) error
}

// ClientConnectedCallback is a callback function that is executed after a new client connected
type ClientConnectedCallback func(cid int)

// ClientDisconnectedCallback is a callback function that is executed after an existing client disconnected
type ClientDisconnectedCallback func(cid int)

type webServer struct {
	mutex   sync.RWMutex
	started bool
	ctx     context.Context

	server           http.Server
	clientIdSequence utils.Sequence
	clients          map[int]*Client // clientId => client

	exchange     MessageExchange
	connected    ClientConnectedCallback
	disconnected ClientDisconnectedCallback
}

// Client represents a single connected client that communicates via a websocket
type Client struct {
	socket *websocket.Conn
}

// rawMessage defines the JSON format for sending client messages
type rawMessage struct {
	Type    string      `json:"messageType"`
	Message interface{} `json:"message"`

	ResponseFor int `json:"responseFor"` // If >0, this is a response to the message with 'AnswerAt' set to the given correlation ID
	AnswerAt    int `json:"answerAt"`    // If >0, this message expects a response with the 'ResponseFor' field set to the given correlation ID
}

func NewWebServer(config *WebServerConfig, connected ClientConnectedCallback, disconnected ClientDisconnectedCallback) WebServer {
	webserver := &webServer{
		started:          false,
		clientIdSequence: utils.NewSequenceGenerator(0),
		clients:          make(map[int]*Client),
		exchange:         NewMessageExchange(),
		connected:        connected,
		disconnected:     disconnected,
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

		if server.disconnected != nil {
			server.disconnected(cid)
		}
	}()

	log.Debug("Sending echo message to new client")
	go server.Send(cid, "echo", "Hello there!")

	if server.connected != nil {
		server.connected(cid)
	}
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

		var message rawMessage
		if err := json.Unmarshal(msg, &message); err != nil {
			log.Errorf("Failed to unmarshal incoming websocket message. Disconnecting. \nRaw content: %q (type %d)", string(msg), msgType)
			return fmt.Errorf("Disconnected due to protocol error")
		}

		var respondFunc MessageRespondFunc

		if message.AnswerAt > 0 {
			log.Debugf("Received message (type %q, with response): %v", message.Type, message.Message)

			respondFunc = func(topic string, content interface{}) error {
				return server.respond(cid, message.AnswerAt, topic, content)
			}
		} else {
			log.Debugf("Received message (type %q): %v", message.Type, message.Message)

			respondFunc = func(topic string, content interface{}) error {
				return fmt.Errorf("Unable to send response message: The message (type=%q) didn't expect a response.", message.Type)
			}
		}

		recipients := server.exchange.publish(message.Type, message.Message, cid, respondFunc)
		if recipients == 0 {
			log.Warnf("There are no recipients for messages of type %q", message.Type)
		}
	}
}

func (server *webServer) Exchange() MessageExchange {
	return server.exchange
}

func (server *webServer) Send(cid int, topic string, message interface{}) error {
	textMessage, err := server.marshal(topic, message, 0, 0)
	if err != nil {
		return err
	}
	return server.send(cid, textMessage)
}

func (server *webServer) respond(cid int, correlationId int, topic string, message interface{}) error {
	if correlationId <= 0 {
		return fmt.Errorf("Unable to respond to message: There is no correlation id.")
	}
	textMessage, err := server.marshal(topic, message, correlationId, 0)
	if err != nil {
		return err
	}
	return server.send(cid, textMessage)
}

func (server *webServer) send(cid int, textMessage []byte) error {
	server.mutex.RLock()
	defer server.mutex.RUnlock()

	client, ok := server.clients[cid]
	if !ok {
		return fmt.Errorf("Failed to send message: Destination cid=%d is unknown", cid)
	}

	return client.socket.WriteMessage(websocket.TextMessage, textMessage)
}

// Broadcast sends the given message to all connected clients
func (server *webServer) Broadcast(topic string, message interface{}) error {
	textMessage, err := server.marshal(topic, message, 0, 0)
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

func (server *webServer) marshal(topic string, message interface{}, responseFor int, answerAt int) ([]byte, error) {
	if answerAt > 0 {
		return []byte{}, fmt.Errorf("Unable to marshal message: Messages-with-response are not supported for the server-side yet (not implemented). Currently only the client can request a response.")
	}
	messageObject := &rawMessage{
		Type:        topic,
		Message:     message,
		ResponseFor: responseFor,
		AnswerAt:    answerAt,
	}
	return json.Marshal(messageObject)
}
