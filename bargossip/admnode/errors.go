package admnode

import "errors"

var (
	ErrInvalidSig 	= errors.New("invalid signature on node information")
	ErrDBIdMismatch = errors.New("P2P NodeDB ID mismatch")
)
