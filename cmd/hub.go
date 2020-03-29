package cmd

// Hub maintains the set of active clients and broadcasts messages to the
// clients.

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[string]map[*Client]bool),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			if _, ok := h.clients[client.topic]; !ok {
				h.clients[client.topic] = make(map[*Client]bool)
				//log.WithFields(log.Fields{"ID": client.ID, "topic": client.topic}).Info("Registered") //TODO delete log for performance
			}
			h.clients[client.topic][client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client.topic]; ok {
				delete(h.clients[client.topic], client)
				close(client.send)
			}
		case message := <-h.broadcast:
			topic := message.sender.topic
			for client := range h.clients[topic] {
				if client.ID == message.ID {
					select {
					case client.send <- message:
						//log.WithFields(log.Fields{"ID": client.ID, "topic": client.topic, "payload": message}).Info("Send") //TODO delete log for performance
					default:
						close(client.send)
						delete(h.clients[topic], client)
					}
				}
			}
		}
	}
}
