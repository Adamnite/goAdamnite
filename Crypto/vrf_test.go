package crypto

import (
	"bytes"
	"testing"
)

func TestFull(t *testing.T) {
	sk, err := GenerateVRFKey(nil)
	if err != nil {
		t.Fatal(err)
	}

	pk, _ := sk.Public()
	msg := []byte("adamnite vrf message")
	msgVRF := sk.Compute(msg)

	vrf, proof := sk.Prove(msg)

	if !pk.Verify(msg, vrf, proof) {
		t.Error("Verify failed")
	}

	if !bytes.Equal(vrf, msgVRF) {
		t.Error("Compute != Prove")
	}
}
