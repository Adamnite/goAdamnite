package admpacket

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"reflect"
	"strings"
	"testing"

	"github.com/adamnite/go-adamnite/bargossip/admnode"
	"github.com/adamnite/go-adamnite/common/hexutil"
	"github.com/adamnite/go-adamnite/crypto"
)

func Test_ECDH(t *testing.T) {
	var (
		staticKey = hexPrivkey("0xfb757dc581730490a1d7a00deea65e9b1936924caaea8f44d476014856b68736")
		publicKey = hexPubkey(crypto.S256(), "0x039961e4c2356d61bedb83052c115d311acb3a96f5777296dcf297351130266231")
		want      = hexutil.MustDecode("0x033b11a2a1f214567e1537ce5e509ffd9b21373247f2a3ff6841f4976f53165e7e")
	)
	result := ecdh(staticKey, publicKey)
	check(t, "shared-secret", result, want)
}

func TestDeriveKeys(t *testing.T) {
	t.Parallel()

	var (
		n1    = admnode.NodeID{1}
		n2    = admnode.NodeID{2}
		cdata = []byte{1, 2, 3, 4}
	)
	sec1 := deriveKeys(sha256.New, testKeyA, &testKeyB.PublicKey, n1, n2, cdata)
	sec2 := deriveKeys(sha256.New, testKeyB, &testKeyA.PublicKey, n1, n2, cdata)
	if sec1 == nil || sec2 == nil {
		t.Fatal("key agreement failed")
	}
	if !reflect.DeepEqual(sec1, sec2) {
		t.Fatalf("keys not equal:\n  %+v\n  %+v", sec1, sec2)
	}

	var ctbuf []byte
	var pt = []byte{'a', 'b', 'c'}
	ctbuf = make([]byte, 100)
	var nonce = []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}

	ct, err := encryptGCM(ctbuf[:0], sec1.writekey, nonce, pt, cdata)
	if err != nil {
		t.Fatal("error")
	}
	rpt, err := decryptGCM(sec1.writekey, nonce, ct, cdata)
	if err != nil {
		t.Fatal("error")
	}

	t.Log(rpt)
}

func check(t *testing.T, what string, x, y []byte) {
	t.Helper()

	if !bytes.Equal(x, y) {
		t.Errorf("wrong %s: %#x != %#x", what, x, y)
	} else {
		t.Logf("%s = %#x", what, x)
	}
}

func hexPrivkey(input string) *ecdsa.PrivateKey {
	key, err := crypto.HexToECDSA(strings.TrimPrefix(input, "0x"))
	if err != nil {
		panic(err)
	}
	return key
}

func hexPubkey(curve elliptic.Curve, input string) *ecdsa.PublicKey {
	key, err := DecodePubkey(curve, hexutil.MustDecode(input))
	if err != nil {
		panic(err)
	}
	return key
}
