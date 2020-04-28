package nike

import (
	"sync"

	"github.com/dev-appmonsters/dicemix-light-server/ecdh"
	"github.com/dev-appmonsters/dicemix-light-server/rng"

	log "github.com/sirupsen/logrus"
)

type nike struct {
	NIKE
	sync.Mutex
}

// NewNike creates a new NIKE instance
func NewNike() NIKE {
	return &nike{}
}

// DeriveSharedKeys - derives shared keys from My Private Key and Peers Public Key
// generates RNG based on shared key using ChaCha20
func (n *nike) DeriveSharedKeys(priv []byte, pub []byte) ([]byte, rng.DiceMixRng) {
	ecdh := ecdh.NewCurve25519ECDH()
	privateKey, _ := ecdh.UnmarshalSK(priv)
	publicKey, res := ecdh.Unmarshal(pub)

	if !res {
		log.Fatalf("Error: generating NIKE Shared Keys %v", res)
	}

	sharedKey, err := ecdh.GenerateSharedSecret(privateKey, publicKey)

	if err != nil {
		log.Fatalf("Error: generating NIKE Shared Keys %v", err)
	}

	dicemix := rng.NewRng(sharedKey)
	return sharedKey, dicemix
}
