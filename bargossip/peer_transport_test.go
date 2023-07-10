package bargossip

import "testing"

func TestConst(t *testing.T) {
	maxUint24 := int(^uint32(0) >> 8)
	t.Log(maxUint24)
}
