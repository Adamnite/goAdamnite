package rpc

type Reply struct {
	//all data is returned in MSGPack formatting of the data within the Reply type.
	Data []byte
}
