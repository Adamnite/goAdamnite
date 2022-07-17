package dial

import "errors"

var (
	errDialSelfNode = errors.New("cannot dial to self node")
)
