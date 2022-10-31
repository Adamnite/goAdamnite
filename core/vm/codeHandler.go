package vm

func getCode(hash []byte) []OperationCommon {
	// TODO: get the code from the API/local cache, and parse then return it.
	if len(hash) != 64 {
		return nil
	}

	return []OperationCommon{}
}
