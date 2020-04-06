package main

import (
	"time"

	"github.com/gorilla/websocket"

	log "github.com/sirupsen/logrus"
)

const (
	// time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second
)


type Client struct {
	hub *Hub

	// socket connection
	conn *websocket.Conn

	// buffered channel for outbound messages
	outbound chan []byte
}

// readMessage reads an inbound message from the websocket
func (c *Client) readMessage() {
	defer func() {
		c.hub.deregister <- c
		c.conn.Close()
	}()
	for {
		_, message, err := c.conn.ReadMessage()
		c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
		if err != nil { // websocket unexpectedly closed
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("unexpected socket close: %v", err)
			}
			break
		}
		c.hub.inbound <- message
	}
}

// writeMessage writes an outbound message to the websocket
func (c *Client) writeMessage() {
	defer func() {
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.outbound:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok { // the hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			writer, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Errorf("unable to get connection writer: %v", message, err)
				return
			}

			// write outbound message to socket
			writer.Write(message)
			if err := writer.Close(); err != nil {
				log.Errorf("unable to close connection writer: %v", message, err)
			}
		}
	}
}

