package main

import (
	"time"

	"github.com/adamnite/go-adamnite/rpc"
)

//this is entirely setup to run a adamnite RPC server on local host. Do not use for any further purpose.

func main() {
	as := rpc.NewAdamniteServer(nil, nil)
	foo := "[127.0.0.1]:12345"
	as.Launch(&foo)
	time.Sleep(time.Minute * 5) //change this time for how long you want the server to stay up
	// for {}//uncomment to have run forever.
}
