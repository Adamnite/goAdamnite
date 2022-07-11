package admnode

import "errors"

var (
	ErrInvalidSig = errors.New("invalid signature on node information")
)
