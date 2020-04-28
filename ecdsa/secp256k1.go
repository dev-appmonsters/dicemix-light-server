package ecdsa

import (
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

type curveP256 struct {
	ECDSA
}

// NewCurveECDSA creates a new Elliptic Curve Digital Signature Algorithm  instance
func NewCurveECDSA() ECDSA {
	return &curveP256{}
}

// Verify reports whether sig is a valid signature of message by publicKey.
func (e *curveP256) Verify(publicKeyBytes, message, signatureBytes []byte) bool {
	publicKey, err := btcec.ParsePubKey(publicKeyBytes, btcec.S256())
	signature, err := btcec.ParseDERSignature(signatureBytes, btcec.S256())
	messageHash := chainhash.DoubleHashB(message)

	if err != nil {
		return false
	}

	// Verify the signature for the message using the public key.
	return signature.Verify(messageHash, publicKey)
}
