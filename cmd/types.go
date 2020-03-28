package cmd

import (
	"sync"
	"time"

	"github.com/eclesh/welford"
	"github.com/gorilla/websocket"
)

// Client is a middleperson between the websocket connection and the hub.
type Client struct {
	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan message

	// string representing the path the client connected to
	topic string

	// is client allowed to provide input to the crossbar?
	broadcaster bool

	stats *Stats

	name string

	userAgent string

	remoteAddr string
}

type RxTx struct {
	Tx ReportStats `json:"tx"`
	Rx ReportStats `json:"rx"`
}

type ReportStats struct {
	Last string `json:"last"` //how many seconds ago...

	Size float64 `json:"size"`

	Fps float64 `json:"fps"`
}

// Client report omits non-serialisable internal references
type ClientReport struct {
	Topic string `json:"topic"`

	Broadcaster bool `json:"broadcaster"`

	Connected string `json:"connected"`

	RemoteAddr string `json:"remoteAddr"`

	UserAgent string `json:"userAgent"`

	Stats RxTx `json:"stats"`
}

type Stats struct {
	connectedAt time.Time

	rx *Frames

	tx *Frames
}

type Frames struct {
	last time.Time

	size *welford.Stats

	ns *welford.Stats
}

// messages will be wrapped in this struct for muxing
type message struct {
	sender Client
	mt     int
	data   []byte //text data are converted to/from bytes as needed
}

// TODO - remove unused types below this line (some still in use)

type clientDetails struct {
	name         string
	topic        string
	messagesChan chan message
}

// requests to add or delete subscribers are represented by this struct
type clientAction struct {
	action clientActionType
	client clientDetails
}

// userActionType represents the type of of action requested
type clientActionType int

// clientActionType constants
const (
	clientAdd clientActionType = iota
	clientDelete
)

type topicDirectory struct {
	sync.Mutex
	directory map[string][]clientDetails
}

// gobwas/ws
type readClientDataReturns struct {
	msg []byte
	mt  int
	err error
}

type summaryStats struct {
	topic map[string]topicStats
}

type topicStats struct {
	audience *welford.Stats
	size     *welford.Stats
	rx       map[string]int
}

type messageStats struct {
	topic string
	rx    []string
	size  int
}
