package server

import (
	"sync"

	"github.com/dev-appmonsters/dicemix-light-server/messages"
	"github.com/dev-appmonsters/dicemix-light-server/utils"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

// Client is a middleman between the websocket connection and the hub.
type client struct {
	hub *hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte
}

// to isolate clients from parallel dicemix executations
type run struct {
	sessionID uint64
	run       int
	peers     []*messages.PeersInfo
	nextState int
	messages  [][]byte
	sync.Mutex
}

type waitingClient struct {
	id        int32
	publicKey []byte
}

// hub maintains the set of active clients and broadcasts messages to the
// clients.
type hub struct {
	clients      map[*client]int32
	runs         map[uint64]*run
	waitingQueue []*waitingClient
	request      chan []byte
	register     chan *client
	unregister   chan *client
	sync.Mutex
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func newHub() *hub {
	return &hub{
		clients:      make(map[*client]int32),
		runs:         make(map[uint64]*run),
		waitingQueue: make([]*waitingClient, 0),
		request:      make(chan []byte),
		register:     make(chan *client),
		unregister:   make(chan *client),
	}
}

func newRun() *run {
	return &run{
		sessionID: 0,
		run:       -1,
		peers:     make([]*messages.PeersInfo, utils.MinPeers),
		nextState: 0,
		messages:  make([][]byte, 0),
	}
}

// starts a run
// registers a peer when he want to participate in TX
// unregisters a peer
// listens for requests from peers and calls its corresponding handler
func (h *hub) listener() {
	for {
		select {
		case client := <-h.register:
			if h.registration(client) {
				log.Info("INCOMING C_JOIN_REQUEST - SUCCESSFUL")
			} else {
				log.Info("INCOMING C_JOIN_REQUEST - FAILED")
			}

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				log.Info("INCOMING - USER UN-REGISTRATION - ", h.clients[client])
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.request:
			handleRequest(message, h)
		}
	}
}

// adds a peer in h.peers if |h.peers| < MaxPeers
// send a failure message response to other peers
func (h *hub) registration(client *client) bool {
	h.Lock()
	defer h.Unlock()

	// generates a random user id for new client
	userID := utils.RandInt31()

	// send registration response to client
	header := responseHeader(messages.S_JOIN_RESPONSE, 0, "Welcome to CoinShuffle++. Waiting for other peers to join ...", "")
	registration, err := proto.Marshal(&messages.RegisterResponse{
		Header: header,
		Id:     userID,
	})

	if checkError(err) {
		return false
	}

	client.send <- registration

	// map client with its userID
	h.clients[client] = userID

	// store client in waiting queue
	h.waitingQueue = append(h.waitingQueue, &waitingClient{id: userID})
	return true
}

// initiates DiceMix-Light protocol
// send all peers ID's
func (h *hub) startDicemix() {
	// generate session id for clients involved in current dicemix execution
	sessionID := utils.RandUint64()

	// create new run
	run := newRun()
	run.peers = make([]*messages.PeersInfo, utils.MinPeers)
	run.sessionID = sessionID
	run.run = 0

	// maintains list of clients which have registered
	// but have not sent their long term public key yet
	waitingClients := make([]*waitingClient, 0)
	i := 0

	// copy peersInfo from waiting queue to start dicemix run
	for _, waitingClient := range h.waitingQueue {
		if len(waitingClient.publicKey) == 0 {
			// if client has not sent long term public key yet
			// add him to waitingClients
			waitingClients = append(waitingClients, waitingClient)
			continue
		}

		// clients those have sent their long term public key
		// add them to newly created dicemix run
		run.peers[i] = &messages.PeersInfo{Id: waitingClient.id}
		run.peers[i].LTPublicKey = waitingClient.publicKey
		run.peers[i].MessageReceived = true
		i++
	}

	// creates an association between sessionID and run
	h.runs[sessionID] = run

	// replace waitingQueue with waitingClients
	// i.e. store only those clients in waitingQueue which
	// have not sent their long term public key yet
	h.waitingQueue = waitingClients

	// broadcasts - initiates DiceMix-Light protocol
	go broadcastDiceMixResponse(h, sessionID, messages.S_START_DICEMIX, "Initiate DiceMix Protocol", "")
}
