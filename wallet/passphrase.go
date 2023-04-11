package wallet


import (
	"bytes"
	"crypto/rand"
	"crypto/sha512"
	"encoding/binary"
	"fmt"
	"math/big"
)

//Converts a 32 byte Adamnite private-key to a 24 word mnemonic
func KeytoMnemonic(key []byte) (string, error) {
	if len(key) != 32 {
		return "", fmt.Errorf("private key must be 32 bytes long")
	}
	h := sha512.Sum256(key)
	hBits := new(big.Int).SetBytes(h[:])
	numWords := len(wordList)
	wordIndices := make([]int, 24)

	for i := 0; i < 24; i++ {
		wordIndex := int(new(big.Int).Mod(hBits, big.NewInt(int64(numWords))).Int64())
		wordIndices[i] = wordIndex
		hBits.Div(hBits, big.NewInt(int64(numWords)))
	}

	var b bytes.Buffer
	for i := 0; i < 23; i++ {
		b.WriteString(wordList[wordIndices[i]])
		b.WriteString(" ")
	}
	b.WriteString(wordList[wordIndices[23]])

	return b.String(), nil
}


//Converts the 24 word mnemonic to a 32 byte private key
func MnemonictoKey(mnemonic string) ([]byte, error) {
	words := bytes.Split([]byte(mnemonic), []byte(" "))
	if len(words) != 24 {
		return nil, fmt.Errorf("mnemonic must contain 24 words")
	}

	numWords := len(wordList)
	hBits := big.NewInt(0)

	for i, word := range words {
		wordIndex := -1
		for j, candidate := range wordList {
			if bytes.Equal([]byte(candidate), word) {
				wordIndex = j
				break
			}
		}

		if wordIndex == -1 {
			return nil, fmt.Errorf("invalid mnemonic word: %s", string(word))
		}

		hBits.Mul(hBits, big.NewInt(int64(numWords)))
		hBits.Add(hBits, big.NewInt(int64(wordIndex)))
	}

	h := make([]byte, 32)
	binary.BigEndian.PutUint64(h, hBits.Uint64())
	binary.BigEndian.PutUint64(h[8:], hBits.Uint64())
	binary.BigEndian.PutUint64(h[16:], hBits.Uint64())
	binary.BigEndian.PutUint64(h[24:], hBits.Uint64())

	return h, nil
}