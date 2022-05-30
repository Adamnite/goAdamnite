package rpc

type API struct {
	Namespace string
	Public    bool
	Version   string
	Service   interface{}
}
