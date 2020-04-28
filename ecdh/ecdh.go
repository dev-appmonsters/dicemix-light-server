package ecdh

import (
	"crypto"
)

// ECDH - The main interface ECDH.
type ECDH interface {
	ValidateKeypair(privateKey, publicKey []byte) bool
	Unmarshal(publicKey []byte) (crypto.PublicKey, bool)
	UnmarshalSK(privateKey []byte) (crypto.PrivateKey, bool)
	GenerateSharedSecret(crypto.PrivateKey, crypto.PublicKey) ([]byte, error)
}
