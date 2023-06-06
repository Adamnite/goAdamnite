package fdutils

import "fmt"

const processHandleLimit = 16384 // 2 ^ 14

func GetCurrentHandles() (int, error) {
	return processHandleLimit, nil
}

func GetMaxHandles() (int, error) {
	return GetCurrentHandles()
}

func GetRaise(max uint64) (uint64, error) {
	if max > processHandleLimit {
		return processHandleLimit, fmt.Errorf("file descriptor limit (%d) reached", processHandleLimit)
	}
	return max, nil
}
