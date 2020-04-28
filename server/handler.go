package server

import (
	"github.com/dev-appmonsters/dicemix-light-server/messages"
	"github.com/dev-appmonsters/dicemix-light-server/utils"

	"github.com/golang/protobuf/proto"
	log "github.com/sirupsen/logrus"
)

// handles any request message from peers
func handleRequest(message []byte, h *hub) {
	h.Lock()
	defer h.Unlock()

	// decode protobuf sent by peer via network
	signedRequest := &messages.SignedRequest{}
	if err := proto.Unmarshal(message, signedRequest); checkError(err) {
		return
	}

	// used to obtain info about peerId, code and sessionID from signedRequest
	r := &messages.GenericRequest{}
	if err := proto.Unmarshal(signedRequest.RequestData, r); checkError(err) {
		return
	}

	// if client has sent his long term public key in message
	if r.Header.Code == messages.C_LTPK_REQUEST {
		handleLTSKRequest(signedRequest.RequestData, h)
		return
	}

	// checks if peer incorrectly signed message or not
	// if incorrectly signed discard the message.
	if !validateMessage(signedRequest, h, r.Header.Id, r.Header.SessionId) {
		log.Info("Recv: Wrong Signature Code - ", r.Header.Code, ", PeerId - ", r.Header.Id)
		return
	}

	runInfo := h.runs[r.Header.SessionId]

	// check if request from client was one of
	// the expected Requests or not
	if runInfo.nextState != int(r.Header.Code) {
		return
	}

	// to keep track of number of clients which have already
	// submitted this request (for current run)
	var counter = counter(runInfo.peers)

	if counter >= len(runInfo.peers) {
		return
	}

	switch r.Header.Code {
	case messages.C_KEY_EXCHANGE:
		request := &messages.KeyExchangeRequest{}
		if err := proto.Unmarshal(signedRequest.RequestData, request); checkError(err) {
			return
		}
		handleKeyExchangeRequest(request, h, counter)

	case messages.C_EXP_DC_VECTOR:
		request := &messages.DCExpRequest{}
		if err := proto.Unmarshal(signedRequest.RequestData, request); checkError(err) {
			return
		}
		handleDCExponentialRequest(request, h, counter)

	case messages.C_SIMPLE_DC_VECTOR:
		request := &messages.DCSimpleRequest{}
		if err := proto.Unmarshal(signedRequest.RequestData, request); checkError(err) {
			return
		}
		handleDCSimpleRequest(request, h, counter)

	case messages.C_TX_CONFIRMATION:
		request := &messages.ConfirmationRequest{}
		if err := proto.Unmarshal(signedRequest.RequestData, request); checkError(err) {
			return
		}
		handleConfirmationRequest(request, h, counter)

	case messages.C_KESK_RESPONSE:
		request := &messages.InitiaiteKESKResponse{}
		if err := proto.Unmarshal(signedRequest.RequestData, request); checkError(err) {
			return
		}
		handleInitiateKESKResponse(request, h, counter)

	}
}

// obtains PublicKeys and NumberOfMsgs sent by peers
func handleLTSKRequest(message []byte, h *hub) {
	request := &messages.LtpkExchangeRequest{}
	if err := proto.Unmarshal(message, request); checkError(err) {
		return
	}

	// TODO: check if public key is valid or not
	counter := 0
	for i := 0; i < len(h.waitingQueue); i++ {
		if len(h.waitingQueue[i].publicKey) > 0 {
			counter++
		} else if h.waitingQueue[i].id == request.Header.Id && len(request.PublicKey) > 0 {
			log.Info("Recv: handleLTSKRequest PeerId - ", request.Header.Id)
			h.waitingQueue[i].publicKey = request.PublicKey
			counter++
		}
	}

	// if MinPeers have registered and sent their long term public key
	// create a new dicemix run
	if counter == utils.MinPeers {
		h.startDicemix()
	}
}

// obtains PublicKeys and NumberOfMsgs sent by peers
func handleKeyExchangeRequest(request *messages.KeyExchangeRequest, h *hub, counter int) {
	sessionID := request.Header.SessionId
	for i := 0; i < len(h.runs[sessionID].peers); i++ {
		if h.runs[sessionID].peers[i].Id == request.Header.Id {
			// TODO: check if public key is valid or not
			if request.NumMsgs < 1 {
				return
			}

			h.runs[sessionID].peers[i].PublicKey = request.PublicKey
			h.runs[sessionID].peers[i].NumMsgs = request.NumMsgs
			h.runs[sessionID].peers[i].MessageReceived = true

			log.Info("Recv: handleKeyExchangeRequest PeerId - ", request.Header.Id)
			counter++
			break
		}
	}

	// if all active peers have submitted their response
	if counter == len(h.runs[sessionID].peers) {
		broadcastDiceMixResponse(h, sessionID, messages.S_KEY_EXCHANGE, "Key Exchange Response", "")
	}
}

// obtains DC-EXP vector sent by peers
func handleDCExponentialRequest(request *messages.DCExpRequest, h *hub, counter int) {
	sessionID := request.Header.SessionId
	msgCount := int(totalMessageCount(h.runs[sessionID].peers))
	for i := 0; i < len(h.runs[sessionID].peers); i++ {
		if h.runs[sessionID].peers[i].Id == request.Header.Id {
			if len(request.DCExpVector) != msgCount {
				return
			}

			h.runs[sessionID].peers[i].DCVector = request.DCExpVector
			h.runs[sessionID].peers[i].MessageReceived = true

			log.Info("Recv: handleDCExponentialRequest PeerId - ", request.Header.Id)
			counter++
			break
		}
	}

	// if all active peers have submitted their response
	if counter == len(h.runs[sessionID].peers) {
		broadcastDCExponentialResponse(h, sessionID, messages.S_EXP_DC_VECTOR, "Solved DC Exponential Roots", "")
	}
}

// obtains DC-SIMPLE vector sent by peers
func handleDCSimpleRequest(request *messages.DCSimpleRequest, h *hub, counter int) {
	sessionID := request.Header.SessionId
	msgCount := int(totalMessageCount(h.runs[sessionID].peers))
	for i := 0; i < len(h.runs[sessionID].peers); i++ {
		if h.runs[sessionID].peers[i].Id == request.Header.Id {
			if len(request.DCSimpleVector) != msgCount {
				return
			}

			h.runs[sessionID].peers[i].DCSimpleVector = request.DCSimpleVector
			h.runs[sessionID].peers[i].OK = request.MyOk
			h.runs[sessionID].peers[i].MessageReceived = true
			h.runs[sessionID].peers[i].NextPublicKey = request.NextPublicKey

			log.Info("Recv: handleDCSimpleRequest PeerId - ", request.Header.Id)
			counter++
			break
		}
	}

	// if all active peers have submitted their response
	if counter == len(h.runs[sessionID].peers) {
		broadcastDCSimpleResponse(h, sessionID, messages.S_SIMPLE_DC_VECTOR, "DC Simple Response", "")
	}
}

// obtains confirmations from peers
// if all peers provided valid confirmations then Dicemix is successful
// else moved to BLAME stage
func handleConfirmationRequest(request *messages.ConfirmationRequest, h *hub, counter int) {
	sessionID := request.Header.SessionId
	for i := 0; i < len(h.runs[sessionID].peers); i++ {
		if h.runs[sessionID].peers[i].Id == request.Header.Id {
			h.runs[sessionID].peers[i].Confirmation = request.Confirmation
			h.runs[sessionID].peers[i].MessageReceived = true

			log.Info("Recv: Confirmation Request PeerId - ", request.Header.Id)
			counter++
			break
		}
	}

	// if all active peers have submitted their response
	if counter == len(h.runs[sessionID].peers) {
		checkConfirmations(h, sessionID)
	}
}

// obtains KESK of peers
// used in BLAME stage to identify malicious peer
func handleInitiateKESKResponse(request *messages.InitiaiteKESKResponse, h *hub, counter int) {
	sessionID := request.Header.SessionId
	for i := 0; i < len(h.runs[sessionID].peers); i++ {
		if h.runs[sessionID].peers[i].Id == request.Header.Id {
			h.runs[sessionID].peers[i].PrivateKey = request.PrivateKey
			h.runs[sessionID].peers[i].MessageReceived = true

			log.Info("Recv: handleInitiateKESKResponse PeerId - ", request.Header.Id)
			counter++
			break
		}
	}

	// if all active peers have submitted their kesk
	if counter == len(h.runs[sessionID].peers) {
		// initiate blame
		startBlame(h, sessionID)
	}
}
