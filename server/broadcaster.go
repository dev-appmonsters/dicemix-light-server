package server

import (
	"time"

	"github.com/dev-appmonsters/dicemix-light-server/messages"
	"github.com/dev-appmonsters/dicemix-light-server/utils"

	"github.com/golang/protobuf/proto"
	log "github.com/sirupsen/logrus"
)

// Removes offline peers
// Broadcasts message to active peers
// Broadcasts responses for -
// S_START_DICEMIX, S_KEY_EXCHANGE, S_SIMPLE_DC_VECTOR, S_TX_CONFIRMATION
func broadcastDiceMixResponse(h *hub, sessionID uint64, state uint32, message string, errMessage string) {
	// removes offline peers
	// returns true if removed any offline peers
	res := filterPeers(h, sessionID)

	if res {
		// if any P_Excluded go back to KE Stage
		if state == messages.S_SIMPLE_DC_VECTOR {
			broadcastKEResponse(h, sessionID)
			return
		}
	}

	// broadcast response to all active peers
	header := responseHeader(state, sessionID, message, errMessage)
	peers, err := proto.Marshal(&messages.DiceMixResponse{
		Header: header,
		Peers:  h.runs[sessionID].peers,
	})

	broadcast(h, sessionID, peers, err, state)
}

func broadcastDCSimpleResponse(h *hub, sessionID uint64, state uint32, message string, errMessage string) {
	// removes offline peers
	// returns true if removed any offline peers
	res := filterPeers(h, sessionID)

	if res {
		// if any P_Excluded go back to KE Stage
		if state == messages.S_SIMPLE_DC_VECTOR {
			h.runs[sessionID].run++
			broadcastKEResponse(h, sessionID)
			return
		}
	}

	count := int(totalMessageCount(h.runs[sessionID].peers))
	h.runs[sessionID].messages = iDcNet.ResolveDCNet(h.runs[sessionID].peers, count)

	// broadcast response to all active peers
	header := responseHeader(state, sessionID, message, errMessage)
	peers, err := proto.Marshal(&messages.DCSimpleResponse{
		Header:   header,
		Messages: h.runs[sessionID].messages,
		Peers:    h.runs[sessionID].peers,
	})

	broadcast(h, sessionID, peers, err, state)
}

// Removes offline peers and broadcasts message to active peers
// Broadcasts responses for -
// S_EXP_DC_VECTOR
func broadcastDCExponentialResponse(h *hub, sessionID uint64, state uint32, message string, errMessage string) {
	// removes offline peers
	// returns true if removed any offline peers
	res := filterPeers(h, sessionID)

	if res {
		if state == messages.S_EXP_DC_VECTOR {
			h.runs[sessionID].run++
			broadcastKEResponse(h, sessionID)
			return
		}
	}

	// broadcast response to all active peers
	header := responseHeader(state, sessionID, message, errMessage)
	peers, err := proto.Marshal(&messages.DCExpResponse{
		Header: header,
		Roots:  iDcNet.SolveDCExponential(h.runs[sessionID].peers),
	})

	broadcast(h, sessionID, peers, err, state)
}

// creates a new run by broadcast KE Exchange Respose to active peers
// when previous run has been discarded due to some offline peers
func broadcastKEResponse(h *hub, sessionID uint64) {
	// broadcast response to all active peers
	header := responseHeader(messages.S_KEY_EXCHANGE, sessionID, "Key Exchange Response", "")
	peers, err := proto.Marshal(&messages.DiceMixResponse{
		Header: header,
		Peers:  h.runs[sessionID].peers,
	})

	broadcast(h, sessionID, peers, err, messages.S_KEY_EXCHANGE)
}

// sent if all peers agrees to continue
// and have submitted confirmations
func broadcastTXDone(h *hub, sessionID uint64) {
	// broadcast response to all active peers
	header := responseHeader(messages.S_TX_SUCCESSFUL, sessionID, "DiceMix Successful Response", "")
	peers, err := proto.Marshal(&messages.TXDoneResponse{
		Header: header,
	})

	broadcast(h, sessionID, peers, err, messages.S_TX_SUCCESSFUL)
}

// sent if all peers agrees to continue
// and have submitted confirmations
func broadcastKESKRequest(h *hub, sessionID uint64) {
	// broadcast response to all active peers
	header := responseHeader(messages.S_KESK_REQUEST, sessionID, "Blame - send your kesk to identify culprit", "")
	peers, err := proto.Marshal(&messages.InitiaiteKESK{
		Header: header,
	})

	broadcast(h, sessionID, peers, err, messages.S_KESK_REQUEST)
}

// Broadcasts messages to active peers
// sets lastRoundUUID to roundUUID of current Response
// Registers a go routine to handled non responsive peers
func broadcast(h *hub, sessionID uint64, message []byte, err error, statusCode uint32) {
	if checkError(err) {
		return
	}

	// minimum peer check
	if len(h.runs[sessionID].peers) < 2 {
		log.Warn("MinPeers: Less than two peers. SessionId - ", sessionID, ", Peers - ", len(h.runs[sessionID].peers))
		// terminate run
		terminate(h, sessionID)
		return
	}

	// wait for 1 sec before broadcasting
	time.Sleep(time.Second)

	for _, peerInfo := range h.runs[sessionID].peers {
		if client, ok := getClient(h.clients, peerInfo.Id); ok {
			select {
			case client.send <- message:
			default:
				close(client.send)
				delete(h.clients, client)
			}
		}

		log.Info("SENT: SessionId - ", sessionID, ", ResponseCode - ", statusCode, ", PeerId - ", peerInfo.Id)
	}

	if statusCode == messages.S_TX_SUCCESSFUL {
		// run is successful
		// successfull termination
		terminate(h, sessionID)
		log.Info("RUN Successful ", sessionID)
		return
	}

	// predict next expected RequestCode from client againts current ResponseCode
	h.runs[sessionID].nextState = nextState(int(statusCode))

	log.Info("SessionId - ", sessionID, ", Expected Next State - ", h.runs[sessionID].nextState)

	// registers a go-routine to handle offline peers
	go registerWorker(h, sessionID, uint32(h.runs[sessionID].nextState), h.runs[sessionID].run)
}

// registers a go-routine to handle offline peers
func registerWorker(h *hub, sessionID uint64, statusCode uint32, run int) {
	select {
	// wait for responseWait seconds then run registerDelayHandler()
	case <-time.After(utils.ResponseWait * time.Second):
		registerDelayHandler(h, sessionID, int(statusCode), run)
	}
}
