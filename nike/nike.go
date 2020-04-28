package nike

import "github.com/dev-appmonsters/dicemix-light-server/rng"

// NIKE - The main interface for Non-interactive Key Exchange (NIKE).
type NIKE interface {
	DeriveSharedKeys([]byte, []byte) ([]byte, rng.DiceMixRng)
}
