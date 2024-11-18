package admpacket

import "bytes"

func bytesCopy(r *bytes.Buffer) []byte {
	b := make([]byte, r.Len())
	copy(b, r.Bytes())
	return b
}
