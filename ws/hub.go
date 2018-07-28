package ws

// Hub manages client registration and plumbing messages to/from clients.
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// An unbuffered channel of requests to register.
	// No buffer ensures clients are registered before message handling starts.
	register chan *Client

	// Requests to unregister
	unregister chan *Client

	done chan struct{}
}

// NewHub returns a Hub.
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		done:       make(chan struct{}),
	}
}

// RunLoop registers/unregisters clients.
func (h *Hub) RunLoop() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case <-h.done:
			return
		}
	}
}

// Broadcast emits the message on all registered clients.
func (h *Hub) Broadcast(b []byte) {
	for client := range h.clients {
		client.send <- b
	}
}

// Close unregisters all connected clients.
func (h *Hub) Close() error {
	for client := range h.clients {
		h.unregister <- client
	}

	close(h.done)

	return nil
}
