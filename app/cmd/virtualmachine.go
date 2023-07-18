package cmd

import "github.com/adamnite/go-adamnite/VM"

type vmHandler struct {
	ocdbLink    string
	debuggingDB *VM.SpoofedDBCache
	trueDBCache *VM.DBCache
}

// creates a new vm handler
func NewVMHandler(dblink string) *vmHandler {
	vmh := vmHandler{
		ocdbLink: dblink,
		// localDB: ,
	}
	VM.NewDBCache(dblink)
	VM.NewSpoofedDBCache(nil, nil)

	return &vmh
}
