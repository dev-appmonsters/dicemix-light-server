package ecdsa

// ECDSA - The main interface P256 curve.
type ECDSA interface {
	Verify([]byte, []byte, []byte) bool
}
