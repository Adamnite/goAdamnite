package adamclient

import (
	"errors"
	"math/big"

	"github.com/adamnite/go-adamnite/utils"
	"github.com/adamnite/go-adamnite/core/types"
)

// senderFromServer is a types.Signer that remembers the sender address returned by the RPC
// server. It is stored in the transaction's sender address cache to avoid an additional
// request in TransactionSender.
type senderFromServer struct {
	addr      utils.Address
	blockhash utils.Hash
}

var errNotCached = errors.New("sender not cached")

func setSenderFromServer(tx *types.Transaction, addr utils.Address, block utils.Hash) {
	// Use types.Sender for side-effect to store our signer into the cache.
	types.Sender(&senderFromServer{addr, block}, tx)
}

func (s *senderFromServer) Equal(other types.Signer) bool {
	os, ok := other.(*senderFromServer)
	return ok && os.blockhash == s.blockhash
}

func (s *senderFromServer) Sender(tx *types.Transaction) (utils.Address, error) {
	if s.blockhash == (utils.Hash{}) {
		return utils.Address{}, errNotCached
	}
	return s.addr, nil
}

func (s *senderFromServer) Hash(tx *types.Transaction) utils.Hash {
	panic("can't sign with senderFromServer")
}
func (s *senderFromServer) SignatureValues(tx *types.Transaction, sig []byte) (R, S, V *big.Int, err error) {
	panic("can't sign with senderFromServer")
}

func (s senderFromServer) ChainType() *big.Int {
	return nil
}
