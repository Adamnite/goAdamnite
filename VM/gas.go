package VM

// This file will be used for OPCODES for which gas computation depends on
// different parameters

// Common Gas costs
const (
	GasQuickStep   uint64 = 1
	GasFastestStep uint64 = 3
	GasFastStep    uint64 = 5
	GasMidStep     uint64 = 20
	GasSlowStep    uint64 = 50
	GasExtStep     uint64 = 100
)

func gasStorageStore() uint64 {
	return 1000
}
