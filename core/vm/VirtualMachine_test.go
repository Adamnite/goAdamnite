package vm

import (
	"testing"
)

func Test_newVirtualMachine(t *testing.T) {
	var codes []string
	vm := newVirtualMachine(codes, Storage{})
	println(vm)
	vm.step()
	println(len(vm.vmMemory) == 0)
	vm.run()
	println(len(vm.vmMemory) == 0)
	// if we make it this far, ill count it as a success
}
