package server

import (
	"github.com/dev-appmonsters/dicemix-light-server/ecdsa"
	"github.com/dev-appmonsters/dicemix-light-server/messages"
	"github.com/dev-appmonsters/dicemix-light-server/utils"

	"github.com/jinzhu/copier"
	log "github.com/sirupsen/logrus"
)

// handles non responsive peers
// after responseWait seconds if all peers have not submitted their response
// then remove them and consider those peers as offline
// and broadcast message to active peers
func registerDelayHandler(h *hub, sessionID uint64, state int, run int) {
	h.Lock()
	defer h.Unlock()

	// if session exists
	if _, ok := h.runs[sessionID]; !ok {
		return
	}

	// if round has been completed successfully
	if h.runs[sessionID].nextState != state || h.runs[sessionID].run != run {
		return
	}

	log.Info("Round has not done ", state, ", SessionId - ", sessionID)

	switch state {
	case messages.C_KEY_EXCHANGE:
		// if some peers have not submitted their PublicKey
		broadcastDiceMixResponse(h, sessionID, messages.S_KEY_EXCHANGE, "Key Exchange Response", "")
	case messages.C_EXP_DC_VECTOR:
		// if some peers have not submitted their DC-EXP vector
		broadcastDCExponentialResponse(h, sessionID, messages.S_EXP_DC_VECTOR, "Solved DC Exponential Roots", "")
	case messages.C_SIMPLE_DC_VECTOR:
		// if some peers have not submitted their DC-SIMPLE vector
		broadcastDCSimpleResponse(h, sessionID, messages.S_SIMPLE_DC_VECTOR, "DC Simple Response", "")
	case messages.C_TX_CONFIRMATION:
		// if some peers have not submitted their CONFIRMATION
		checkConfirmations(h, sessionID)
	case messages.C_KESK_RESPONSE:
		// if some peers have not submitted their KESK
		// initiate blame
		startBlame(h, sessionID)
	}
}

// removes offline peers from h.runs[sessionID].peers
// returns true if removed any offline peer
func filterPeers(h *hub, sessionID uint64) bool {
	var allPeers []*messages.PeersInfo
	copier.Copy(&allPeers, &h.runs[sessionID].peers)
	h.runs[sessionID].peers = make([]*messages.PeersInfo, 0)

	for _, peer := range allPeers {
		// check if client is active and has submitted response
		if peer.MessageReceived {
			peer.MessageReceived = false
			h.runs[sessionID].peers = append(h.runs[sessionID].peers, peer)
			continue
		}

		// if client is offline and not submitted response
		removePeer(h, peer.Id)
	}
	// removed any offline peer?
	return len(allPeers) != len(h.runs[sessionID].peers)
}

// checks if all peers have submitted a valid confirmation for msgs
// if yes then DiceMix protocol is considered as successful
// else moves to BLAME stage
func checkConfirmations(h *hub, sessionID uint64) {
	// removes offline peers
	// returns true if removed any offline peers
	if res := filterPeers(h, sessionID); res {
		// if any P_Excluded trace back to KE Stage
		h.runs[sessionID].run++
		broadcastKEResponse(h, sessionID)
		return
	}

	// TODO: [change] as we consider that if peer doesnt want to send his confirmation and want to go to blame stage
	// sends |confirmation| = 0, which would be changed in future

	// check if any of peers does'nt agree to continue
	for _, peer := range h.runs[sessionID].peers {
		if !peer.Confirmation {
			// Blame stage - INIT KESK
			log.Info("BLAME - Peer ", peer.Id, " does'nt provide correct corfirmation")
			h.runs[sessionID].run++
			broadcastKESKRequest(h, sessionID)
			return
		}

		log.Info("CONFIRMATION - Peer ", peer.Id, " sent correct confirmation")

	}

	// DiceMix is successful
	broadcastTXDone(h, sessionID)
}

// predicts next expected RequestCodes from client againts current ResponseCode
func nextState(responseCode int) int {
	switch responseCode {
	case messages.S_START_DICEMIX:
		return messages.C_KEY_EXCHANGE
	case messages.S_KEY_EXCHANGE:
		return messages.C_EXP_DC_VECTOR
	case messages.S_EXP_DC_VECTOR:
		return messages.C_SIMPLE_DC_VECTOR
	case messages.S_SIMPLE_DC_VECTOR:
		return messages.C_TX_CONFIRMATION
	case messages.S_KESK_REQUEST:
		return messages.C_KESK_RESPONSE
	}

	return 0
}

// checks if peer incorrectly signed message or not
// if incorrectly signed discard the message.
func validateMessage(message *messages.SignedRequest, h *hub, id int32, sessionID uint64) bool {
	// id session id is valid
	if _, ok := h.runs[sessionID]; !ok {
		return false
	}

	// get long term public key of peer to verify signed message
	// if publickey found verify message
	if publicKey, found := publicKey(h.runs[sessionID].peers, id); found {
		ecdsa := ecdsa.NewCurveECDSA()
		return ecdsa.Verify(publicKey, message.RequestData, message.Signature)
	}
	return false
}

// return long term public key of peer with specified id
// if found -> returns publickey, true
// else -> returns nil, false
func publicKey(peers []*messages.PeersInfo, id int32) ([]byte, bool) {
	for _, peer := range peers {
		if peer.Id == id && len(peer.LTPublicKey) > 0 {
			return peer.LTPublicKey, true
		}
	}
	return nil, false
}

// generates ResponseHeader required in any response message
// broadcasted by server to all active peers
func responseHeader(code uint32, sessionID uint64, message, err string) *messages.ResponseHeader {
	return &messages.ResponseHeader{
		Code:      code,
		SessionId: sessionID,
		Timestamp: utils.Timestamp(),
		Message:   message,
		Err:       err,
	}
}

// to keep track of number of clients which have already
// submitted the request for corresponding RequestCode (for current run)
func counter(peers []*messages.PeersInfo) (counter int) {
	for _, peer := range peers {
		if peer.MessageReceived {
			counter++
		}
	}
	return
}

// returns total number of messages
func totalMessageCount(peers []*messages.PeersInfo) uint32 {
	var count uint32
	for _, peer := range peers {
		count += peer.NumMsgs
	}
	return count
}

// returns client connection object from client id
func getClient(m map[*client]int32, data int32) (*client, bool) {
	for key, value := range m {
		if data == value {
			return key, true
		}
	}
	return nil, false
}
