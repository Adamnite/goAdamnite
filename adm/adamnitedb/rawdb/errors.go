package rawdb

import (
	"errors"
)
var (
	ErrNoData  		 	= errors.New("db errors! no data!")
	ErrUnknown 			= errors.New("db unknown errors!")
	ErrMsgPackDecode 	= errors.New("msgpack decode errors!")
)