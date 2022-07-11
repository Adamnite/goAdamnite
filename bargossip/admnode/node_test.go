package admnode

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewWithParams(t *testing.T) {
	node := NewWithParams(pubKey, net.IPv4(192, 168, 109, 100), 90, 90)

	require.Equal(t, node.Pubkey(), pubKey)
	require.Equal(t, node.TCP(), uint16(90))
	require.Equal(t, node.UDP(), uint16(90))
	require.Equal(t, node.info.GetType(), TypeCompatURLV1)
}
