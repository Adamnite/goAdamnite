package adm

import "github.com/adamnite/go-adamnite/dpos"

type AdamniteAPI interface {
	DposEngine() dpos.DPOS
}
