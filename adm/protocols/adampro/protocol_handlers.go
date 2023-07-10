package adampro

import "fmt"

func newBlockHandler(admHandler AdamniteHandlerInterface, msg Decoder, peer *Peer) error {
	newBlock := new(NewBlockPacket)
	if err := msg.Decode(newBlock); err != nil {
		return fmt.Errorf("invalid message: %v: %v", msg, err)
	}

	newBlock.Block.ReceivedAt = msg.Time()
	newBlock.Block.ReceivedFrom = peer
	return admHandler.Handle(peer, newBlock)
}
