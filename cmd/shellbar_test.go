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
	payload1 := []byte("Hello from client0")

	mtype := websocket.TextMessage

	c0.Out <- reconws.WsMessage{Data: payload0, Type: mtype}

	_ = expectOneGob(s.In, payload0, timeout, t)

	c1.Out <- reconws.WsMessage{Data: payload1, Type: mtype}

	_ = expectOneGob(s.In, payload1, timeout, t)

	time.Sleep(timeout)

	cancel()

	time.Sleep(timeout)

	close(closed)

	wg.Wait()

}

func expectOneGob(channel chan reconws.WsMessage, expected []byte, timeout time.Duration, t *testing.T) (data []byte) {

	var receivedData []byte

	select {
	case <-time.After(timeout):
		t.Errorf("timeout receiving client0 message at server")
	case msg, ok := <-channel:
		if ok {
			bufReader := bytes.NewReader(msg.Data)
			decoder := gob.NewDecoder(bufReader)
			var sr srgob.Message
			err := decoder.Decode(&sr)
			if err != nil {
				t.Errorf("Decoding gob from relay %v", err) //(log.WithField("Error", err.Error())
			} else {
				receivedData = sr.Data
				if bytes.Compare(sr.Data, expected) != 0 {
					t.Errorf("Messages don't match: Want: %s\nGot : %s\n", expected, sr.Data)
				}
			}
		} else {
			t.Error("Channel problem")
		}
	}
	return receivedData
}
