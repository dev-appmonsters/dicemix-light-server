package dc

import (
	"github.com/dev-appmonsters/dicemix-light-server/field"
	"github.com/dev-appmonsters/dicemix-light-server/messages"
	"github.com/dev-appmonsters/dicemix-light-server/solver"
	"github.com/dev-appmonsters/dicemix-light-server/utils"
)

type dcNet struct {
	DC
}

// NewDCNetwork creates a new DC instance
func NewDCNetwork() DC {
	return &dcNet{}
}

// obtains all peers DC-EXP vectors
// combines them and generates DC-COMBINED vector
// solves DC-COMBINED vector and obtain's its roots using Flint
func (d *dcNet) SolveDCExponential(peers []*messages.PeersInfo) []uint64 {
	var i, totalMsgsCount uint32
	var dcCombined = make([]uint64, len(peers[0].DCVector))
	copy(dcCombined, peers[0].DCVector)

	// obtain Total Messages Count
	for _, peer := range peers {
		totalMsgsCount += peer.NumMsgs
	}

	// generates DC-COMBINED vector
	for j := 1; j < len(peers); j++ {
		for i = 0; i < totalMsgsCount; i++ {
			dcCombined[i] = field.NewField(dcCombined[i]).Add(field.NewField(peers[j].DCVector[i])).Value()
		}
	}

	// NOTE: totalMsgsCount should be less than 1000 or else FLINT would fail to obtain roots
	// and [0,0,......] will be considered as roots
	// Basic sanity check to avoid weird inputs
	// check - solver/solver_flint.cpp (46)
	return solver.Solve(dcCombined, int(totalMsgsCount))
}

// Resolve the DC-net
func (d *dcNet) ResolveDCNet(peers []*messages.PeersInfo, totalMsgsCount int) [][]byte {
	var allMessages = make([][]byte, len(peers[0].DCSimpleVector))

	// copies DCSimpleVector
	for i, vector := range peers[0].DCSimpleVector {
		allMessages[i] = make([]byte, len(vector))
		copy(allMessages[i], vector)
	}

	// decode messages
	for i := 1; i < len(peers); i++ {
		for j := 0; j < totalMsgsCount; j++ {
			// decodes messages from slots by cancelling out randomness introduced in DC-Simple
			// xor operation - all_messages[j] = dc_simple_vector[j] + <randomness for chacha20>
			utils.XorBytes(allMessages[j], allMessages[j], peers[i].DCSimpleVector[j])

		}
	}
	return allMessages
}
