package webserver

import (
	"context"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/maja42/AniScraper/utils"
)

type WebServer interface {
	Start(ctx context.Context)
}

type webServer struct {
	mutex   sync.RWMutex
	started bool
	ctx     context.Context

	server           http.Server
	clientIdSequence utils.Sequence
	clients          map[int]*client // clientId => client
}

func NewWebServer(config *WebServerConfig) WebServer {
	webserver := &webServer{
		started:          false,
		clientIdSequence: utils.NewSequenceGenerator(0),
		clients:          make(map[int]*client),
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
}

func (server *webServer) websocketHandler(res http.ResponseWriter, req *http.Request) {
	conn, err := (&websocket.Upgrader{CheckOrigin: server.checkOrigin}).Upgrade(res, req, nil)
	if err != nil {
		http.NotFound(res, req)
		log.Warnf("Failed to accept client: %v", err)
		return
	}
	log.Info("Client connected")

	client := &client{
		socket: conn,
	}

	cid := server.clientIdSequence.Next()

	server.mutex.Lock()
	defer server.mutex.Unlock()

	server.clients[cid] = client

	log.Debugf("Total clients connected: %d", len(server.clients))
	go server.handleClient(cid, client)
}

func (server *webServer) checkOrigin(r *http.Request) bool {
	return true
}

func (server *webServer) handleClient(cid int, client *client) {
	clientCtx, cancel := context.WithCancel(server.ctx)

	go func() {
		err := client.Handle(clientCtx)
		log.Infof("Client disconnected: %v", err)

		cancel()

		server.mutex.Lock()
		defer server.mutex.Unlock()
		delete(server.clients, cid)

		log.Debugf("Total clients connected: %d", len(server.clients))
	}()
}
