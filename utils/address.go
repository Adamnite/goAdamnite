import (
	"crypto/sha512"
	"fmt"
	"github.com/vmihailenco/msgpack/v5"
	"golang.org/x/crypto/ripemd160"
)

const (
	AddressLength  = 28
	HashLength     = 20
	ChecksumLength = 4
	HexPrefix     = "0x"
)

type Address [AddressLength]byte

type Hash [HashLength]byte

func (a Address) MarshalMsgpack() ([]byte, error) {
	return a[:], nil
}

func (a *Address) UnmarshalMsgpack(data []byte) error {
	if len(data) != AddressLength {
		return fmt.Errorf("invalid address length")
	}
	copy(a[:], data)
	return nil
}

func (a Address) Bytes() []byte {
	return a[:]
}

func (a Address) Hash() []byte {
	hash := sha512.Sum512(a.Bytes())
	ripemd := ripemd160.New()
	ripemd.Write(hash[:])
	return ripemd.Sum(nil)
}

func (a Address) Hex() string {
	return HexPrefix + hex.EncodeToString(a.Bytes())
}

func HashToAddress(hash []byte) (Address, error) {
	if len(hash) != HashLength {
		return Address{}, fmt.Errorf("invalid hash length")
	}
	var addr Address
	copy(addr[:], hash)
	return addr, nil
}

func HexToAddress(hexStr string) (Address, error) {
	hexStr = trimHexPrefix(hexStr)
	decoded, err := hex.DecodeString(hexStr)
	if err != nil {
		return Address{}, err
	}
	if len(decoded) != AddressLength {
		return Address{}, fmt.Errorf("invalid address length")
	}
	var addr Address
	copy(addr[:], decoded)
	return addr, nil
}

func AddressToString(addr Address) string {
	return addr.Hex()
}

func StringToAddress(addrStr string) (Address, error) {
	return HexToAddress(addrStr)
}

func hexToBytes(hexStr string) ([]byte, error) {
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode hexadecimal string: %w", err)
	}
	return bytes, nil
}