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
	pongWait = 45 * time.Second

	// send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
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
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
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
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
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
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

