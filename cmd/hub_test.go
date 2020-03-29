package cmd

import (
	"bytes"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestHub(t *testing.T) {

	hub := newHub()
	go hub.run()

	pseudoServer := &Client{hub: hub,
		ID:    "*",
		send:  make(chan message, 256),
		topic: "abcd",
	}

	hub.register <- pseudoServer

	pseudoClient1 := &Client{hub: hub,
		ID:    "client1",
		send:  make(chan message, 256),
		topic: "abcd",
	}

	hub.register <- pseudoClient1

	pseudoClient2 := &Client{hub: hub,
		ID:    "client2",
		send:  make(chan message, 256),
		topic: "abcd",
	}

	hub.register <- pseudoClient2

	m1 := []byte("message1")
	m2 := []byte("message2")
	m3 := []byte("message3")

	go expectOneMessage(pseudoClient1.send, m1, t)
	go expectOneMessage(pseudoClient2.send, m2, t)
	go expectOneMessage(pseudoServer.send, m3, t)

	hub.broadcast <- message{sender: *pseudoServer, data: m1, mt: websocket.BinaryMessage, ID: "client1"}
	hub.broadcast <- message{sender: *pseudoServer, data: m2, mt: websocket.BinaryMessage, ID: "client2"}
	hub.broadcast <- message{sender: *pseudoClient1, data: m3, mt: websocket.BinaryMessage, ID: "*"}

}

func expectOneMessage(channel chan message, expected []byte, t *testing.T) {

	// want one message
	select {
	case <-time.After(10 * time.Millisecond):
		t.Errorf("Timeout awaiting %s", expected)
	case msg, ok := <-channel:
		if ok {
			if bytes.Compare(msg.data, expected) != 0 {
				t.Errorf("Wrong message. Want: %s\nGot : %s\n", expected, msg.data)
			}
		} else {
			t.Errorf("Channel closed awaiting %s", expected)
		}
	}
	// expect a time out waiting for a second message
	select {
	case <-time.After(10 * time.Millisecond):
		// expect to get here, all good
	case msg, ok := <-channel:
		if ok {
			t.Errorf("Got extra message of %s", msg.data)
		}
	}
}
