package utils

import (
	"time"
)

const (
	// JoinTXRound - rounds
	JoinTXRound = 1

	// WriteWait - Time allowed to write a message to the peer.
	WriteWait = 10 * time.Second

	// PongWait - Time allowed to read the next pong message from the peer.
	PongWait = 60 * time.Second

	// PingPeriod - Send pings to peer with this period. Must be less than pongWait.
	PingPeriod = (PongWait * 9) / 10

	// MinPeers - number of peers required to start DiceMix protocol
	MinPeers = 3

	// ResponseWait - Time to wait for response from peers.
	ResponseWait = 5
)

var (
	// Newline - represents new line char
	Newline = []byte{'\n'}

	// Space - represents space char
	Space = []byte{' '}
)
