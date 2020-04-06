package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"codenames/codenames"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type Dictionary struct {
	Words []string `json:"words"`
}

const (
	port = 8080
)

var (
	// current working directory used for path lookups
	cwd, _ = os.Getwd()

	// dictionary for games
	words = loadWords()

	// socket upgrader
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	// in memory store of current game connections
	gameConnections = make(map[string]*Hub)

	// initialize redis client
	redisConnection = &RedisConnection{
		gameConnections: &gameConnections,
	}
)

// load words from static dictionary
func loadWords() []string {
	dictionary, err := os.Open(filepath.Join(cwd, "static", "words.json"))
	processError("unable to open dictionary file", err)

	jsonBytes, err := ioutil.ReadAll(dictionary)
	processError("unable to read dictionary file", err)

	var unmarshaled Dictionary
	json.Unmarshal([]byte(jsonBytes), &unmarshaled)

	return unmarshaled.Words
}

// generic function to process errors
func processError(message string, err error) {
	if err != nil {
		log.Fatalf("%s: %v", message, err)
		os.Exit(1)
	}
}

// simple health check
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"message": "healthy"})
}

// returns the current state of the game from redis
func GetGameHandler(w http.ResponseWriter, r *http.Request) {
	state, _ := redisConnection.GetKey(mux.Vars(r)["id"])
	json.NewEncoder(w).Encode(state)
}

// creates a new game socket
func CreateGameSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	processError("unable to upgrade connection to socket", err)

	identifier := mux.Vars(r)["id"]
	hub, ok := gameConnections[identifier]
	if !ok { // create new hub
		log.Infof("creating new hub for identifier %s", identifier)
		hub = &Hub{
			register:   make(chan *Client),
			deregister: make(chan *Client),
			clients:    make(map[*Client]bool),
			inbound:    make(chan []byte),
			identifier: identifier,
		}

		// store hub in gameConnections
		gameConnections[identifier] = hub

		// subscribe to identifier updates
		redisConnection.Subscribe(identifier)

		// start running the hub
		go hub.run()
	}

	client := &Client{
		hub: hub,
		conn: conn,
		outbound: make(chan []byte),
	}

	// register client
	hub.register <- client

	go client.readMessage()
	go client.writeMessage()
}

func CreateGameHandler(w http.ResponseWriter, r *http.Request) {
	identifier, _ := codenames.CreateGame(words, redisConnection.GetKeys())
	json.NewEncoder(w).Encode(map[string]string{"identifier": identifier})
}

// when all client connects are gone, remove hub and redis subscription
func cleanupHub(identifier string) {
	log.Infof("cleaning up hub for game %s", identifier)
	delete(gameConnections, identifier)
	redisConnection.Unsubscribe(identifier)
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/games/{id}", GetGameHandler).Methods("GET")
	router.HandleFunc("/games/{id}/socket", CreateGameSocketHandler).Methods("GET")
	router.HandleFunc("/games", CreateGameHandler).Methods("POST")
	router.HandleFunc("/health", HealthHandler).Methods("GET")

	server := &http.Server{
		Handler: handlers.CORS()(router),
		ReadTimeout: 5 * time.Second,
		WriteTimeout: 10 * time.Second,
		Addr: fmt.Sprintf(":%d", port),
	}

	// generate random seed for this server
	rand.Seed(time.Now().UTC().UnixNano())

	// start redis client
	go redisConnection.PropagateUpdate()

	// start server
	log.Infof("server started")
	log.Fatal(server.ListenAndServe())
}
