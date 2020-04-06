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

	"github.com/go-redis/redis/v7"
)

// wrapper struct to hold dictionary words
type Dictionary struct {
	Words []string `json:"words"`
}

// body of any update board request
type UpdateRequest struct {
	Clicked int `json:"clicked"`
}

var (
	// current working directory used for path lookups
	cwd, _ = os.Getwd()

	// serve static files in production environment
	staticAssets = filepath.Join(cwd, "/static/web")

	// dictionary for games
	words = loadWords()

	// socket upgrader
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
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
	w.Write([]byte(state))
}

// updates a given game after a click event
func UpdateGameHandler(w http.ResponseWriter, r *http.Request) {
	identifier := mux.Vars(r)["id"]
	log.Infof("received update request for game %s", identifier)

	// marshal update request into struct
	var updateRequest UpdateRequest
	err := json.NewDecoder(r.Body).Decode(&updateRequest)
	processError("unable to marshal update request body", err)

	// fetch GameBoard state from redis
	var gameBoard codenames.GameBoard
	state, _ := redisConnection.GetKey(identifier)

	// marshal redis value into GameBoard struct
	err = json.Unmarshal([]byte(state), &gameBoard)
	processError("unable to unmarshal gameboard value", err)

	// update board and propagate to redis
	gameBoard.Words[updateRequest.Clicked].Revealed = true
	marshaledGame, _ := json.Marshal(gameBoard)
	redisConnection.SetKey(identifier, marshaledGame)

	// send empty response
	w.Write([]byte{})
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
		hub:      hub,
		conn:     conn,
		outbound: make(chan []byte),
	}

	// register client
	hub.register <- client

	go client.readMessage()
	go client.writeMessage()
}

// checks whether a given identifier exists, surfaces 404 if not
func gameExistenceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identifier := mux.Vars(r)["id"]
		_, err := redisConnection.GetKey(identifier)
		if err == redis.Nil {
			w.WriteHeader(http.StatusNotFound)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

func CreateGameHandler(w http.ResponseWriter, r *http.Request) {
	identifier, game := codenames.CreateGame(words, redisConnection.GetKeys())
	marshaledGame, _ := json.Marshal(game)
	redisConnection.SetKey(identifier, marshaledGame)
	json.NewEncoder(w).Encode(map[string]string{"identifier": identifier})
}

// when all client connects are gone, remove hub and redis subscription
func cleanupHub(identifier string) {
	log.Infof("cleaning up hub for game %s", identifier)
	delete(gameConnections, identifier)
	redisConnection.Unsubscribe(identifier)
}

// handle serving static assets in bundled environment
func SpaHandler(w http.ResponseWriter, r *http.Request) {
	path, _ := filepath.Abs(r.URL.Path)
	filename := filepath.Join(staticAssets, path)

	// check whether a file exists at the given path
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		// file does not exist, serve index.html
		http.ServeFile(w, r, filepath.Join(staticAssets, "index.html"))
		return
	}

	// otherwise, use http.FileServer to serve the static dir
	http.FileServer(http.Dir(staticAssets)).ServeHTTP(w, r)
}

func main() {
	router := mux.NewRouter().StrictSlash(true)

	// routes that require game existence
	subrouter := router.PathPrefix("/api/games/{id}").Subrouter()
	subrouter.Use(gameExistenceMiddleware)
	subrouter.HandleFunc("/", GetGameHandler).Methods("GET")
	subrouter.HandleFunc("/update", UpdateGameHandler).Methods("POST")
	subrouter.HandleFunc("/socket", CreateGameSocketHandler).Methods("GET")

	router.HandleFunc("/api/games", CreateGameHandler).Methods("POST")
	router.HandleFunc("/health", HealthHandler).Methods("GET")

	// static html/js/css assets
	router.PathPrefix("/").HandlerFunc(SpaHandler)

	// setup CORS
	originsOk := handlers.AllowedOrigins([]string{"*"})
	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})

	server := &http.Server{
		Handler:      handlers.CORS(originsOk, headersOk)(router),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		Addr:         fmt.Sprintf(":%s", os.Getenv("PORT")),
	}

	// generate random seed for this server
	rand.Seed(time.Now().UTC().UnixNano())

	// start redis client
	go redisConnection.PropagateUpdate()
	redisConnection.SetKeyspaceEvents()

	// start server
	log.Infof("server started")
	log.Fatal(server.ListenAndServe())
}
