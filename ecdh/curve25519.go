package ecdh

import (
	"bytes"
	"crypto"
	"sync"

	ecdh "github.com/wsddn/go-ecdh"
	"golang.org/x/crypto/curve25519"
)

type curve25519ECDH struct {
	ECDH
	sync.Mutex
}

// NewCurve25519ECDH creates a new ECDH instance that uses djb's curve25519
// elliptical curve.
func NewCurve25519ECDH() ECDH {
	return &curve25519ECDH{}
}

// Generate PublicKey from privateKey
func (e *curve25519ECDH) ValidateKeypair(privateKey, publicKey []byte) bool {
	if len(privateKey) != 32 {
		return false
	}

	var priv, pub [32]byte
	copy(priv[:], privateKey)
	curve25519.ScalarBaseMult(&pub, &priv)

	return bytes.Equal(pub[:], publicKey)
}

// Unmarshal converts byte[] to crypto.PublicKey
func (e *curve25519ECDH) Unmarshal(publicKey []byte) (crypto.PublicKey, bool) {
	var ecdhCurve = ecdh.NewCurve25519ECDH()
	return ecdhCurve.Unmarshal(publicKey)
}

// Unmarshal converts byte[] to crypto.PrivateKey
func (e *curve25519ECDH) UnmarshalSK(privateKey []byte) (crypto.PrivateKey, bool) {
	var pri [32]byte
	if len(privateKey) != 32 {
		return nil, false
	}
	copy(pri[:], privateKey)
	return &pri, true
}

// GenerateSharedSecret creates shared key using our private key and others public key
func (e *curve25519ECDH) GenerateSharedSecret(privKey crypto.PrivateKey, pubKey crypto.PublicKey) ([]byte, error) {
	var ecdhCurve = ecdh.NewCurve25519ECDH()
	return ecdhCurve.GenerateSharedSecret(privKey, pubKey)
}
