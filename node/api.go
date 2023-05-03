package node

import "github.com/adamnite/go-adamnite/rpc"

func (n *Node) apis() []rpc.API {
	return []rpc.API{
		{
			Namespace: "admin",
			Version:   "1.0",
			Service:   &privateAdminAPI{n},
		},
	}
}

type privateAdminAPI struct {
	node *Node
}

func (api *privateAdminAPI) AddPeer(url string) (bool, error) {
	return true, nil
}
