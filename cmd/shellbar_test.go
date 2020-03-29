package cmd

import (
	"bytes"
	"context"
	"encoding/gob"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/phayes/freeport"
	log "github.com/sirupsen/logrus"
	"github.com/timdrysdale/reconws"
	"github.com/timdrysdale/srgob"
)

func TestShellbar(t *testing.T) {
	suppressLog()
	defer displayLog()
	// setup shellbar on local (free) port

	closed := make(chan struct{})
	var wg sync.WaitGroup

	port, err := freeport.GetFreePort()
	if err != nil {
		log.Fatal(err)
	}

	addr := ":" + strconv.Itoa(port)

	wg.Add(1)
	go shellbar(addr, closed, &wg)

	time.Sleep(10 * time.Millisecond)

	// set up test server and two clients

	ctx, cancel := context.WithCancel(context.Background())

	uc := "ws://127.0.0.1" + addr + "/some/location"
	us := "ws://127.0.0.1" + addr + "/serve/some/location"

	c0 := reconws.New()
	c1 := reconws.New()
	s := reconws.New()

	go c0.Reconnect(ctx, uc)
	go c1.Reconnect(ctx, uc)
	go s.Reconnect(ctx, us)

	timeout := 50 * time.Millisecond

	time.Sleep(timeout)

	payload0 := []byte("Hello from client0")
	payload1 := []byte("Hello from client1")

	mtype := websocket.TextMessage

	c0.Out <- reconws.WsMessage{Data: payload0, Type: mtype}

	srxc0 := expectOneGob(s.In, payload0, timeout, t)

	c1.Out <- reconws.WsMessage{Data: payload1, Type: mtype}

	srxc1 := expectOneGob(s.In, payload1, timeout, t)

	expectNoMsg(s.In, timeout, t)

	// reply from the server
	var buf0 bytes.Buffer
	var buf1 bytes.Buffer

	reply0 := []byte("Greetings to client0")
	reply1 := []byte("Greetings to client1")

	sr0 := srgob.Message{
		ID:   srxc0.ID, //ID is randomly assigned so get from gob received by server
		Data: reply0,
	}

	sr1 := srgob.Message{
		ID:   srxc1.ID, //ID is randomly assigned so get from gob received by server
		Data: reply1,
	}

	encoder0 := gob.NewEncoder(&buf0)
	err = encoder0.Encode(sr0)

	if err != nil {
		t.Errorf("Encoding gob for server reply")
	} else {

		s.Out <- reconws.WsMessage{Data: buf0.Bytes(), Type: websocket.BinaryMessage}
	}

	encoder1 := gob.NewEncoder(&buf1)
	err = encoder1.Encode(sr1)

	if err != nil {
		t.Errorf("Encoding gob for server reply")
	} else {

		s.Out <- reconws.WsMessage{Data: buf1.Bytes(), Type: websocket.BinaryMessage}
	}

	_ = expectOneSlice(c0.In, reply0, timeout, t)

	expectNoMsg(c0.In, timeout, t)

	_ = expectOneSlice(c1.In, reply1, timeout, t)

	expectNoMsg(c1.In, timeout, t)

	time.Sleep(timeout)

	cancel()

	time.Sleep(timeout)

	close(closed)

	wg.Wait()

}

func expectOneGob(channel chan reconws.WsMessage, expected []byte, timeout time.Duration, t *testing.T) srgob.Message {

	var receivedGob srgob.Message

	select {
	case <-time.After(timeout):
		t.Errorf("timeout receiving message (expected %s)", expected)
	case msg, ok := <-channel:
		if ok {
			bufReader := bytes.NewReader(msg.Data)
			decoder := gob.NewDecoder(bufReader)
			var sr srgob.Message
			err := decoder.Decode(&sr)
			if err != nil {
				t.Errorf("Decoding gob from relay %v", err)
			} else {
				receivedGob = sr
				if bytes.Compare(sr.Data, expected) != 0 {
					t.Errorf("Messages don't match: Want: %s\nGot : %s\n", expected, sr.Data)
				}
			}
		} else {
			t.Error("Channel problem")
		}
	}
	return receivedGob
}

func expectNoMsg(channel chan reconws.WsMessage, timeout time.Duration, t *testing.T) {

	select {
	case <-time.After(timeout):
		return //we are expecting to timeout, this is good
	case msg, ok := <-channel:
		if ok {
			t.Errorf("Receieved unexpected message %v", msg)
		} else {
			//just a channel problem, not an unexpected message
		}
	}
}

func expectOneSlice(channel chan reconws.WsMessage, expected []byte, timeout time.Duration, t *testing.T) []byte {

	var receivedSlice []byte

	select {
	case <-time.After(timeout):
		t.Errorf("timeout receiving message (expected %s)", expected)
	case msg, ok := <-channel:
		if ok {
			receivedSlice = msg.Data
			if bytes.Compare(receivedSlice, expected) != 0 {
				t.Errorf("Messages don't match: Want: %s\nGot : %s\n", expected, receivedSlice)
			}
		} else {
			t.Error("Channel problem")
		}
	}
	return receivedSlice
}
