package ws

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message.
	writeTimeout = 10 * time.Second

	// Time allowed to receive a ping response.
	pongTimeout = 60 * time.Second

	// Frequency of outgoing pings. This must be less then pongTimeout.
	pingInterval = (pongTimeout * 9) / 10

	// Message size limit. We aren't trying to receive messages, send only.
	incomingMsgLimit = 1
)

var (
	lf    = []byte{'\n'}
	space = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client manages a single ws connection.
type Client struct {
	// Reference to the Hub.
	hub *Hub

	// The tcp conn.
	conn *websocket.Conn

	// Buffered channel of outgoing messages.
	send chan []byte
}

func (c *Client) close() {
	c.hub.unregister <- c
	if err := c.conn.Close(); err != nil {
		log.Println(
			fmt.Sprintf("Error closing ws connection: %v", err),
		)
	}
}

func (c *Client) handleIncoming() {
	defer c.close()

	c.conn.SetReadLimit(incomingMsgLimit)
	c.conn.SetReadDeadline(time.Now().Add(pongTimeout))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongTimeout))
		return nil
	})

	// We don't want to read any messages. In theory, gorilla/ws should just
	// discard a message when a new one is received, so resources won't congest.
	// We still need a blocking loop to keep our defer func from running.
	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			log.Println(fmt.Sprintf("Error receiving message (possible close): %v", err))
			return
		}
	}
}

func (c *Client) handleOutgoing() {
	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		c.close()
	}()

	for {
		select {
		// watch ok for a channel close
		case m, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			if !ok {
				// closed channel, so send a close message
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			wc, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Println(fmt.Sprintf("Error acquiring connection writer: %v", err))
			}

			_, err = wc.Write(m)
			if err != nil {
				log.Println(fmt.Sprintf("Error writing to connection: %v", err))
			}

			// Dequeue send messages buffered on the channel
			for n := len(c.send); n > 0; n-- {
				wc.Write(lf)
				wc.Write(<-c.send)
			}

			if err := wc.Close(); err != nil {
				// trigger the close logic
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ServeWS upgrades a connection to ws and handles messaging with the hub.
// If the connection cannot be upgraded, a non-nil error is returned.
func ServeWS(h *Hub, w http.ResponseWriter, r *http.Request) error {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}

	client := &Client{
		hub:  h,
		conn: conn,
		send: make(chan []byte, 1<<4),
	}

	h.register <- client

	go client.handleIncoming()
	go client.handleOutgoing()

	return nil
}
