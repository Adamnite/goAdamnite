package dial

import "errors"

var (
	errDialSelfNode     = errors.New("cannot dial to self node")
	errDialNoPort       = errors.New("no port")
	errDialAlreadyExsit = errors.New("dial already exist")
	errAlreadyConnected = errors.New("already connected")
	errBlackListNode    = errors.New("blacklist node")
	errNotWhiteListNode = errors.New("not whitelist node")
	errDialedRecently   = errors.New("dialed recently")
)
