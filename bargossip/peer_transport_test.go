package bargossip

import (
	"testing"
	"crypto/rand"
	"github.com/adamnite/go-adamnite/crypto"
	"github.com/adamnite/go-adamnite/crypto/ecies"
	"github.com/adamnite/go-adamnite/bargossip/utils"
	"github.com/adamnite/go-adamnite/common/hexutil"
	"github.com/vmihailenco/msgpack/v5"
)

func TestConst(t *testing.T) {
	maxUint24 := int(^uint32(0) >> 8)
	t.Log(maxUint24)
}

func TestHandshake(t *testing.T) {
	prvKey, _ := crypto.GenerateKey()
	remotePrivKey, _ := crypto.GenerateKey()
	remotePeerPubKey := &remotePrivKey.PublicKey

	encKeys := &handshakeEncKeys{remotePubKey: ecies.ImportECDSAPublic(remotePeerPubKey)}
	encKeys.initNonce = make([]byte, 32)
	_, err := rand.Read(encKeys.initNonce)
	if err != nil {
		t.Error(err)
	}

	encKeys.oneTimePrivKey, err = ecies.GenerateKey(rand.Reader, crypto.S256(), nil)
	if err != nil {
		t.Error(err)
	}

	sk, err := ecies.ImportECDSA(prvKey).GenerateShared(ecies.ImportECDSAPublic(remotePeerPubKey), 16, 16) // generate secret key to use in both side to encrypt
	if err != nil {
		t.Error(err)
	}

	signed := utils.XOR(sk, encKeys.initNonce)
	signature, err := crypto.Sign(signed, encKeys.oneTimePrivKey.ExportECDSA())
	if err != nil {
		t.Error(err)
	}

	hMsg := new(handshakeMsg)
	copy(hMsg.Nonce[:], encKeys.initNonce)
	copy(hMsg.SenderPubKey[:], crypto.FromECDSAPub(&prvKey.PublicKey)[1:])
	copy(hMsg.Signature[:], signature)
	hMsg.Version = AdamniteTCPHandshakeVersion

	// 2. Create a packet from handshake message
	packedMsg, err := msgpack.Marshal(hMsg)
	if err != nil {
		t.Error(err)
	}

	enc, err := ecies.Encrypt(rand.Reader, ecies.ImportECDSAPublic(remotePeerPubKey), packedMsg, nil, nil)
	if err != nil {
		t.Error(err)
	}

	t.Log(hexutil.Encode(enc[:]))
	t.Log(crypto.PubkeyToAddress(*remotePeerPubKey))
	t.Log("enc", enc)

	ehMsg := new(handshakeMsg)

	key := ecies.ImportECDSA(remotePrivKey)
	if dec, err := key.Decrypt(enc, nil, nil); err == nil {
		if compareBytes(dec, packedMsg) {
			ehMsg.Decode(dec)
			t.Log("dec", dec)
		}
	} else {
		t.Error(err)
	}

	encKeys = new(handshakeEncKeys)
	if err := encKeys.handleHandshakeMsg(hMsg, prvKey); err != nil {
		t.Error(err)
	}

	respHMsg, err := encKeys.makeRespHandshakeMsg()
	if err != nil {
		t.Error(err)
	}

	// 2. Create packet
	respBytes, err := msgpack.Marshal(respHMsg)
	if err != nil {
		t.Error(err)
	}

	enc, err = ecies.Encrypt(rand.Reader, encKeys.remotePubKey, respBytes, nil, nil)
	if err != nil {
		t.Error(err)
	}
}

func compareBytes (a []byte, b []byte) bool {
	for i := 0; i < len(a); i++ {
		if (a[i] != b[i]) {
			return false
		}
	}
	return true
}

func TestMsgPackSerialize(t *testing.T) {
	var uInt uint64 
	uInt = 5

	byUint, err := msgpack.Marshal(uInt)
	if err != nil {
		t.Error(err)
	}

	t.Log(byUint)
	uInt = 0x88888888888888;
	byUint1, err := msgpack.Marshal(uInt)
	if err != nil {
		t.Error(err)
	}
	t.Log(byUint1)
}