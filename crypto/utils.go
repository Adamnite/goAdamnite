package crypto

import (
    "bufio"
    "bytes"
    "io"
    "unicode"
	"errors"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"encoding/hex"
	"math/big"
	"os"
	"strings"
	"crypto/rand"
	"crypto/ecdsa"
	"crypto/elliptic"
)


func readASCII(reader io.Reader, bufferSize int) ([]byte, error) {
    // Create a new scanner for the reader
    scanner := bufio.NewScanner(reader)

    // Set the scanner to split tokens by bytes
    scanner.Split(bufio.ScanBytes)

    // Create a buffer to store the read data
    buffer := bytes.Buffer{}

    // Read each byte of the input and append it to the buffer
    for scanner.Scan() {
        b := scanner.Bytes()[0]
        if !unicode.IsPrint(rune(b)) {
            break
        }
        buffer.WriteByte(b)
        if buffer.Len() == bufferSize {
            break
        }
    }

    // Return the buffer contents as a byte slice
    return buffer.Bytes(), scanner.Err()
}

const BASE58_TABLE = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

func Base58Encode(input []byte) string {
    // Convert input bytes to a big integer
    x := big.Int{}
    x.SetBytes(input)

    // Create a base58 encoder with the given BASE58_TABLE
    base := len(BASE58_TABLE)
    encoded := make([]byte, 0, len(input)*138/100) // maximum possible size of the output
    for x.Sign() > 0 {
        // Divide x by the base, and get the remainder
        mod := new(big.Int)
        x.DivMod(&x, big.NewInt(int64(base)), mod)

        // Append the corresponding character from the BASE58_TABLE
        encoded = append(encoded, BASE58_TABLE[mod.Int64()])
    }

    // Reverse the encoded bytes
    for i, j := 0, len(encoded)-1; i < j; i, j = i+1, j-1 {
        encoded[i], encoded[j] = encoded[j], encoded[i]
    }

    return string(encoded)
}
func Base58Decode(input string) ([]byte, error) {
    // Convert input string to a big integer
    x := big.Int{}
    for _, r := range input {
        // Find the index of the character in the BASE58_TABLE
        i := strings.IndexRune(BASE58_TABLE, r)
        if i < 0 {
            return nil, fmt.Errorf("invalid character '%c' in input string", r)
        }

        // Multiply x by the base, and add the character value
        x.Mul(&x, big.NewInt(int64(len(BASE58_TABLE))))
        x.Add(&x, big.NewInt(int64(i)))
    }

    // Convert the big integer to a byte slice
    decoded := x.Bytes()

    // Trim leading zeros from the byte slice
    i := 0
    for i < len(decoded) && decoded[i] == 0 {
        i++
    }

    // Append leading zeros to the output slice
    output := make([]byte, 0, i+len(decoded))
    output = append(output, make([]byte, i)...)
    output = append(output, decoded...)

    return output, nil
}

func Clear(b []byte) {
    // Use the bytes package to fill the byte slice with zeros
    bytes.Fill(b, 0)
}

func ToPublic(b []byte) (*ecdsa.PublicKey, error) {
    // Parse the byte slice as an ECDSA private key
    priv := new(ecdsa.PrivateKey)
    priv.Curve = elliptic.P256k1
    priv.D = new(big.Int).SetBytes(b)

    // Generate the corresponding public key
    pub := new(ecdsa.PublicKey)
    pub.Curve = priv.Curve
    pub.X, pub.Y = priv.Curve.ScalarBaseMult(priv.D.Bytes())

    return pub, nil
}

func Validate(v, r, s []byte, pubKey *ecdsa.PublicKey) bool {
    // Parse the signature values as big.Int
    V := new(big.Int).SetBytes(v)
    R := new(big.Int).SetBytes(r)
    S := new(big.Int).SetBytes(s)

    // Validate the signature
    if V.Cmp(big.NewInt(27)) != 0 && V.Cmp(big.NewInt(28)) != 0 {
        return false
    }
    if R.Sign() <= 0 || S.Sign() <= 0 {
        return false
    }
    if R.Cmp(pubKey.Params().N) >= 0 || S.Cmp(pubKey.Params().N) >= 0 {
        return false
    }

    // Compute the hash of the message
    // and verify the signature
    hash := []byte("my message")
    e := new(big.Int).SetBytes(hash)
    w := new(big.Int).ModInverse(S, pubKey.Params().N)
    u1 := new(big.Int).Mul(e, w)
    u2 := new(big.Int).Mul(R, w)
    u2.Mod(u2, pubKey.Params().N)
    x, y := pubKey.Curve.ScalarBaseMult(u1.Bytes())
    x1, y1 := pubKey.Curve.ScalarMult(pubKey.X, pubKey.Y, u2.Bytes())
    x, y = pubKey.Curve.Add(x, y, x1, y1)

    return R.Cmp(x) == 0
}