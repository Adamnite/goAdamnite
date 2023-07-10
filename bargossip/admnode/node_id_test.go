package admnode

import "testing"

func Test_Hex_ID(t *testing.T) {
	orgID := NodeID{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 221, 222, 223, 224, 225, 226, 227, 228, 229, 230, 231}

	hexID := StringToHexID("0x000102030405060708090a0b0c0d0e0f1011121314dddedfe0e1e2e3e4e5e6e7")
	hexIDWithoutPrefix := StringToHexID("000102030405060708090a0b0c0d0e0f1011121314dddedfe0e1e2e3e4e5e6e7")

	if orgID != hexID {
		t.Errorf("wrong hexID, wants %x", orgID)
	}

	if orgID != hexIDWithoutPrefix {
		t.Errorf("wrong hexIDWithoutPrefix, wants %x", orgID)
	}
}
