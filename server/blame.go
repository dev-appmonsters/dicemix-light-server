package server

import (
	"github.com/dev-appmonsters/dicemix-light-server/ecdh"
	"github.com/dev-appmonsters/dicemix-light-server/field"
	"github.com/dev-appmonsters/dicemix-light-server/nike"
	"github.com/dev-appmonsters/dicemix-light-server/rng"
	"github.com/dev-appmonsters/dicemix-light-server/utils"

	op "github.com/adam-hanna/arrayOperations"
	log "github.com/sirupsen/logrus"
)

// Peer - contains information of peers of a Participant
type peerInfo struct {
	ID        int32
	SharedKey []byte
	Dicemix   rng.DiceMixRng
}

// Participant - contains details of participant in Blame stage
type participant struct {
	ID           int32
	Peers        []*peerInfo
	MessagesHash []uint64
}

func startBlame(h *hub, sessionID uint64) {
	var participants = make([]*participant, 0)
	var roots = iDcNet.SolveDCExponential(h.runs[sessionID].peers)

	// identifies honest peers (who have expected protocol messages)
	participants = initBlame(h, sessionID, participants, roots)

	// identify and exclude peers involved in slot collision
	if collisions, found := slotCollision(h, sessionID, participants); found {
		eliminatePeers(collisions, h, sessionID)
	}

	// removes malicious and offline peers
	// i.e. those peers who have sent unexpected protocol messages
	filterPeers(h, sessionID)

	rotateKeys(h, sessionID)
	broadcastKEResponse(h, sessionID)
}

// identifies honest peers (who have expected protocol messages)
// Exclude peers in next run who have sent unexpected protocol messages
func initBlame(h *hub, sessionID uint64, participants []*participant, roots []uint64) []*participant {
	nike := nike.NewNike()

	for i := 0; i < len(h.runs[sessionID].peers); i++ {
		peer := h.runs[sessionID].peers[i]

		// do not perform following actions for
		// those peers whcih have not sent their KESK
		if !peer.MessageReceived {
			continue
		}

		// check if peer has sent correct private key
		// validate(privateKey, publicKey) key pairs
		// if sent wrong keys exclude clients
		ecdh := ecdh.NewCurve25519ECDH()
		if ok := ecdh.ValidateKeypair(peer.PrivateKey, peer.PublicKey); !ok {
			h.runs[sessionID].peers[i].MessageReceived = false
			continue
		}

		// create participant object to store info of a valid
		// participant of blame (i.e. whcih has sent his KESK)
		var participant = &participant{}
		participant.ID = peer.Id
		participant.Peers = make([]*peerInfo, 0)

		privateKey := peer.PrivateKey
		totalMsgsCount := int(peer.NumMsgs)

		// for every peer active till confirmation
		// irrespective of he has sent his kesk or not
		for _, otherPeer := range h.runs[sessionID].peers {
			if peer.Id == otherPeer.Id {
				continue
			}

			// derive sharedSecret with otherPeers
			var peer = &peerInfo{}
			peer.ID = otherPeer.Id
			peer.SharedKey, peer.Dicemix = nike.DeriveSharedKeys(privateKey, otherPeer.PublicKey)
			totalMsgsCount += int(otherPeer.NumMsgs)

			// append peer to peers of participant
			participant.Peers = append(participant.Peers, peer)
		}

		// recover messages - obtains messages of participant from his DC-Simple broadcast
		messages := recoverMessages(participant.Peers, peer.DCSimpleVector)

		// number of msg sent by client and number of msgs he promised to send are not equal
		// then remove client
		if uint32(len(messages)) != peer.NumMsgs {
			h.runs[sessionID].peers[i].MessageReceived = false
			continue
		}

		// recover message hashes - obtains hashes sent by participant from his DC-Exp broadcast
		// verify messages checks if peer has sent
		// unexpected message and corresponding hash
		hashes, ok := verifyMessageHashes(participant.ID, participant.Peers, messages, totalMsgsCount, peer.DCVector)
		participant.MessagesHash = hashes

		// check if user sent confirmation = false but his msg was in generated Dc-Simple vector
		// if so then remove malicious peer
		allMessages := h.runs[sessionID].messages

		// if message and message hashes do not correspond
		// check validity of ok sent by client in DC-SIMPLE round
		// case: if user has sent actual dc-simple-vector and ok=false
		// then remove client
		if !ok || (!peer.OK && utils.IsSubset(participant.MessagesHash, roots)) || (ok && utils.ContainBytes(messages, allMessages) && !peer.Confirmation) {
			// set peer.MessageReceived to false
			// so it would be removed by filterPeers()
			h.runs[sessionID].peers[i].MessageReceived = false
			continue
		}

		// if participant is valid add to valid participants
		participants = append(participants, participant)
	}

	return participants
}

// to identify peers who are involved in slot collision
// Exclude peers who are involved in a slot collision,
// i.e., a message hash collision
func slotCollision(h *hub, sessionID uint64, participants []*participant) ([]int32, bool) {
	// store id's of peers involved in slot collision
	var collisions = make([]int32, 0)

	// for all (p1, p2) in P^2
	for i := 0; i < len(participants); i++ {
		p1 := participants[i]
		for j := i + 1; j < len(participants); j++ {
			p2 := participants[j]

			intersection, ok := op.Intersect(p1.MessagesHash, p2.MessagesHash)
			slice, ok := intersection.Interface().([]uint64)

			if !ok {
				log.Warn("Error: Cannot convert reflect.Value to []uint64")
				return nil, false
			}

			// if |intersection| == 0 {no collision occured between peer1 and peer2}
			if len(slice) == 0 {
				continue
			}

			// collision occured between peer1 and peer2
			// add them to collisions slice i.e.
			// P_exclude := P_exclude U {p1, p2}
			collisions = append(collisions, p1.ID)
			collisions = append(collisions, p2.ID)
		}
	}

	return collisions, len(collisions) != 0
}

// removes peers involed in slot collision
func eliminatePeers(collisions []int32, h *hub, sessionID uint64) {
	// remove every peer involved in slot collision
	for i := 0; i < len(collisions); i++ {
		for j := 0; j < len(h.runs[sessionID].peers); j++ {
			// if peer is not involved in collision
			if collisions[i] != h.runs[sessionID].peers[j].Id {
				continue
			}

			// set peer.MessageReceived to false
			// so it would be removed by filterPeers()
			h.runs[sessionID].peers[j].MessageReceived = false
		}
	}
}

// recovers honest peers messages from his DC-SIMPLE vector
// by cancelling out randomness
func recoverMessages(peers []*peerInfo, messages [][]byte) [][]byte {
	messages = decodeMessages(peers, messages)
	messages = utils.RemoveEmptyBytes(messages)
	return messages
}

// decodes messages from slots
func decodeMessages(peers []*peerInfo, messages [][]byte) [][]byte {
	for i := 0; i < len(peers); i++ {
		for j := 0; j < len(messages); j++ {
			// decodes messages
			// xor operation - messages[j] = dc_simple_vector[j] + <randomness for chacha20>
			utils.XorBytes(messages[j], messages[j], peers[i].Dicemix.GetBytes(20))
		}
	}
	return messages
}

// checks if message sent by peer in DC-Simple
// and Hash sent by him in DC-EXP are related or not
// todo so first generates dc-exp vector from mesages
// returns hashes of messages from roots (if valid)
// bool - roots contains valid hashes of messages or not
func verifyMessageHashes(myID int32, peers []*peerInfo, messages [][]byte, totalMsgsCount int, peerDC []uint64) ([]uint64, bool) {
	messageHashes := make([]uint64, len(messages))
	dc := make([]uint64, totalMsgsCount)
	peersCount := len(peers)

	// generates power sums of message_hashes
	// my_dc[i] := my_dc[i] (+) (my_msg_hashes[j] ** (i + 1))
	for j := 0; j < len(messages); j++ {
		// generates 64 bit hash of my_message[j]
		messageHashes[j] = utils.ShortHash(utils.BytesToBase58String(messages[j]))
		var pow uint64 = 1
		for i := 0; i < totalMsgsCount; i++ {
			pow = utils.Power(messageHashes[j], pow)
			dc[i] = field.NewField(dc[i]).Add(field.NewField(pow)).Value()
		}
	}

	// encode power sums
	// my_dc[i] := my_dc[i] (+) (sgn(my_id - p.id) (*) p.dicemix.get_field_element())
	for j := 0; j < peersCount; j++ {
		for i := 0; i < totalMsgsCount; i++ {
			var op2 = field.NewField(peers[j].Dicemix.GetFieldElement())
			if myID < peers[j].ID {
				op2 = op2.Neg()
			}
			dc[i] = field.NewField(dc[i]).Add(op2).Value()
		}
	}

	// check if peer sent correct dc-exp vector
	// by checking generated DC vector and received DC vector for equality
	if utils.CheckEqualUint64(dc, peerDC) {
		return messageHashes, true
	}
	return nil, false
}

// rotate keys to be used in next run
// (kepk) := (my_next_kepk)
// (my_next_kepk) := (undef)
func rotateKeys(h *hub, sessionID uint64) {
	for i := 0; i < len(h.runs[sessionID].peers); i++ {
		h.runs[sessionID].peers[i].PublicKey = h.runs[sessionID].peers[i].NextPublicKey
		h.runs[sessionID].peers[i].NextPublicKey = nil
	}
}
