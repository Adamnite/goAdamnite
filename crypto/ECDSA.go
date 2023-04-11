package crypto

import(
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"encoding/pem"
)


func Create_Key()(*ecdsa.PrivateKey, error){
	curve := Secp256k1()
    privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)

}

func To_ECDSA(d []byte)(*ecdsa.PrivateKey, error) {
	curve := Secp256k1()
    privateKey := &ecdsa.PrivateKey{
        D: d,
        PublicKey: ecdsa.PublicKey{
            Curve: curve,
            X:     curve.Params().Gx,
            Y:     curve.Params().Gy,
        },
	}
}



func Save_ECDSA(name string, key *ecdsa.PrivateKey) error {
	privateKeyBytes := make([]byte, 32)
    copy(privateKeyBytes[32-len(key.D.Bytes()):], privateKey.D.Bytes())
    privateKeyPem := &pem.Block{
        Type:  "EC PRIVATE KEY",
        Bytes: privateKeyBytes,
    }

    // Save the private key to a file
    file, err := os.Create(name+".pem")
    if err != nil {
        panic(err)
    }
    defer file.Close()
}

func d_fromECDSA(key *ecdsa.PrivateKey){
	if key == nil {
		return nil
	}
	return math.PaddedBigBytes(priv.D, priv.Params().BitSize/8)
}

func HexToECDSA(privateKeyHex string) (*ecdsa.PrivateKey, error) {
    // Decode the private key from hex format
    privateKeyBytes, err := hex.DecodeString(privateKeyHex)
    if err != nil {
        return nil, err
    }
	return To_ECDSA(privateKeyBytes)
}


func LoadECDSA(filename string) (*ecdsa.PrivateKey, error) {
	file, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    // Read the private key from the file
    privateKeyBytes, err := hex.DecodeString(file.Name())
    if err != nil {
        return nil, err
    }
	return HexToECDSA(privateKeyBytes)
}
