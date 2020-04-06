package main

import (
	log "github.com/sirupsen/logrus"
)

type Hub struct {
	// map of all registered clients
	clients map[*Client]bool

	// inbound messages from registered clients
	inbound chan []byte

	// register a new client with this hub
	register chan *Client

	// deregister an existing client with this hub
	deregister chan *Client

	// game identifier for this hub
	identifier string
}

func (hub *Hub) run() {
	for {
		select {
		case client := <-hub.register: // register new client
			log.Infof("registering new client for game %s", hub.identifier)
			hub.clients[client] = true
		case client := <-hub.deregister: // deregister client
			log.Infof("deregistering client for game %s", hub.identifier)
			if _, ok := hub.clients[client]; ok {
				delete(hub.clients, client)
				close(client.outbound)
			}

			// no more clients, cleanup
			if len(hub.clients) == 0 {
				cleanupHub(hub.identifier)
			}
		case message := <-hub.inbound: // send messages to all other clients
			for client := range hub.clients {
				client.outbound <- message
			}
		}
	}
}
