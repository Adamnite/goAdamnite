package crypto

import (
	"crypto/sha512"
	"golang.org/x/crypto/ripemd160"
	"hash"
	"errors"
	"fmt"
	"crypto/sha256"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/common/math"
)


var incorrect_hash_type = errors.New("Invalid Hash Type")

const (
	Sha512_truncated
	Sha512
	Ripemd160
	Sumhash
)

cconst (
	Sha512_truncatedSize = sha512.Size256
	Sha512Size 			 = sha512.Size
	SumhashDigestSize    = sumhash.Sumhash512DigestSize
	Ripemd160Size		 = ripemd160.Size

)

type Hash_State struct {
	_struct struct{} 'codec:", omityempty, omitemptyarray'
}

type HashKind uint16
func (h HashKind) String() string {
	switch h {
	case Sha512_truncated:
		return "sha512_truncated"
	case Sha512:
		return "sha512"
	case Sumhash:
		return "sumhash"
	case Ripemd160:
		return "ripmed160"
	default:
		return ""
	}
}

func (z HashState) NewHash() hash.Hash {
	switch z.HashKindW {

	case Sha512_truncated:
		return sha512.New512_256()
	case Sha512:
		return sha512.New()
	case Sumhash:
		return sumhash.New512(nil)
	case Ripemd160:
		return Ripemd160.New()
	default:
		return incorrect_hash_type{}
	}
}



