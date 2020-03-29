package cmd

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
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
	//suppressLog()
	//defer displayLog()
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
	mtype := websocket.TextMessage
	c0.Out <- reconws.WsMessage{Data: payload0, Type: mtype}

	select {
	case <-time.After(timeout):
		t.Errorf("timeout receiving client0 message at server")
	case msg, ok := <-s.In:
		if ok {
			fmt.Printf("msg received %v", msg.Data)

			bufReader := bytes.NewReader(msg.Data)
			decoder := gob.NewDecoder(bufReader)
			var sr srgob.Message
			err = decoder.Decode(&sr)
			if err != nil {
				log.WithField("Error", err.Error()).Error("Decoding gob from relay")
			} else {
				// send only the payload to the non-server client, not a gob
				if bytes.Compare(sr.Data, payload0) != 0 {
					t.Errorf("Messages don't match: Want: %s\nGot : %s\n", payload0, sr.Data)
				}
			}

		} else {
			t.Error("Channel problem")
		}
	}
	/*
		reply := <-r.In

		if bytes.Compare(reply.Data, payload) != 0 {
			t.Errorf("Got unexpected response: %s, wanted %s\n", reply.Data, payload)
		}
	*/

	time.Sleep(1 * time.Second)

	cancel()

	time.Sleep(10 * time.Millisecond)

	close(closed)

	wg.Wait()

}
