package vm

// This should return both the typeinfo about the function and the body/code
// The mapping inside the storage should be something like [funcHash] => [ [typeinfo], [code]]

func getCode(hash []byte) (FunctionType, []OperationCommon, []ControlBlock) {
	// TODO: get the code from the API/local cache, and parse then return it.

	funcType := FunctionType{}
	retrievedCode := []byte{}
	code, ctrlStack := parseBytes(retrievedCode)

	return funcType, code, ctrlStack
}
