package crypto


import (
	"bytes"
	"testing"
	"time"
)

func TestVRF(t *testing.T) {
	// Test key generation
	pk, sk, err := GenerateKey()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Test Compute
	message := []byte("hello")
	vrf := Compute(message, sk)

	// Test Prove and Verify
	vrfBytes, proof := Prove(message, sk)
	if !Verify(pk, message, vrfBytes, proof) {
		t.Errorf("Failed to verify VRF proof")
	}

	// Test bitshift forgery
	forgedProof := make([]byte, len(proof))
	copy(forgedProof, proof)
	forgedProof[0] ^= 1

	if Verify(pk, message, vrfBytes, forgedProof) {
		t.Errorf("Bitshift forgery not detected")
	}

	// Test vector testing with modified first byte
	vectors := []struct {
		message []byte
		vrf     []byte
	}{
		{[]byte{0x01, 0x65, 0x73, 0x6b, 0x65, 0x74, 0x63, 0x68}, []byte{0x16, 0x4a, 0xc1, 0x33, 0x2a, 0x84, 0x27, 0x92, 0xc4, 0x54, 0x42, 0x89, 0xe9, 0x0b, 0x5d, 0x21, 0x3b, 0x57, 0x88, 0x7a, 0x2a, 0x07, 0x6b, 0x62, 0x9e, 0x97, 0xda, 0x19, 0x50, 0xd0, 0x21, 0x8b}},
		{[]byte{0x02, 0x76, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x32}, []byte{0x1f, 0x06, 0x11, 0x3b, 0x5e, 0x70, 0x80, 0x0f, 0x62, 0x23, 0xad, 0x66, 0x5d, 0x6b, 0x1d, 0x24, 0xf9, 0x57, 0xf5, 0xd9, 0xa0, 0x0e, 0x33, 0x42, 0x71, 0x8c, 0x44, 0xd0, 0x10, 0x60, 0x2d, 0xc3}},
	}

	for _, vector := range vectors {
		vrfBytes = Compute(vector.message, sk)
		if !bytes.Equal(vrfBytes, vector.vrf) {
			t.Errorf("Computed VRF does not match expected value")
		}
	}
}

func BenchmarkCompute(b *testing.B) {
	message := []byte("benchmark")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
			Compute(message, sk)
	}
}

func BenchmarkProve(b *testing.B) {
	message := []byte("benchmark")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
			Prove(message, sk)
	}
}

func BenchmarkVerify(b *testing.B) {
	message := []byte("benchmark")
	vrfBytes, proof := Prove(message, sk)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
			Verify(pk, message, vrfBytes, proof)
	}
}
