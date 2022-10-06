package vm

import (
	"bytes"
	"fmt"
	"io"
)

// See https://www.w3.org/TR/2019/REC-wasm-core-1-20191205/
type Module struct {
	typeSection []*FunctionType
	// importSection []*Import
	functionSection []Index
	tableSection []*Table
	memorySection *Memory
	globalSection []*Global
	// exportSection []*Export
	startSection *Index
	// elementSection []*ElementSegment
	codeSection []*Code
	dataSection []*DataSegment
	dataCountSection *uint32
	ID ModuleID
}


func decode(wasmBytes []byte) *Module {

	if wasmBytes == nil {
		panic("Can't decode nil bytes")
	}

	if len(wasmBytes) < 4 || !bytes.Equal(wasmBytes[0:4], Magic) {
		panic("invalid binary")
	}

	r := bytes.NewReader(wasmBytes)

	// Magic
	buf := make([]byte, 4)
	if _, err := io.ReadFull(r, buf); err != nil || !bytes.Equal(buf, Magic) {
		panic("Invalid bytes")
	}

	// Version.
	if _, err := io.ReadFull(r, buf); err != nil || !bytes.Equal(buf, version) {
		panic("Invalid Version")
	}
	m := &Module{}

	for {
		sectionID, err := r.ReadByte()
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Printf("read section id: %w", err)
		}

		sectionSize, _, err := DecodeUint32(r)
		if err != nil {
			fmt.Printf("get size of section %s: %v", sectionIDName(sectionID), err)
		}
		sectionContentStart := r.Len()

		switch sectionID {
			case sectionIDType:
				vs, _, err := DecodeUint32(r)
				if err != nil {
					fmt.Printf("get size of vector: %w", err)
					panic("Get size of vector")
				}
			
				result := make([]*FunctionType, vs)
				for i := uint32(0); i < vs; i++ {
					if result[i], err = decodeFunctionType(r); err != nil {
						fmt.Printf("read %d-th type: %v", i, err)
					}
				}

				m.typeSection = result

			case sectionIDImport:
				// Not implemented
			
			case sectionIDFunction:
				vs, _, err := DecodeUint32(r)
				if err != nil {
					fmt.Printf("get size of vector: %w", err)
					panic("Get size of vector")
				}
			
				result := make([]uint32, vs)
				for i := uint32(0); i < vs; i++ {
					if result[i], _, err = DecodeUint32(r); err != nil {
						fmt.Printf("get type index: %w", err)
						panic("Get type index")
					}
				}
				m.functionSection = result
			
			case sectionIDTable:
				vs, _, err := DecodeUint32(r)
				if err != nil {
					fmt.Printf("error reading size")
					panic("Error reading size")
				}

				ret := make([]*Table, vs)
				for i := range ret {
					table, err := decodeTable(r)
					if err != nil {
						print(err)
					}
					ret[i] = table
				}
				m.tableSection = ret

			case sectionIDMemory:
				vs, _, err := DecodeUint32(r)
				if err != nil {
					fmt.Printf("error reading size")
				}
				if vs > 1 {
					fmt.Printf("at most one memory allowed in module, but read %d", vs)
					panic("at most one memory allowed in module")
				}
			
				min, maxP, err := decodeLimitsType(r)
				if err != nil {
					panic(err)
				}
		
				mem := &Memory{min: min, cap: min, max: defaultPageSize, isMaxEncoded: maxP != nil}
			
				m.memorySection = mem

			case sectionIDGlobal:
				vs, _, err := DecodeUint32(r)
				if err != nil {
					fmt.Printf("get size of vector: %w", err)
					panic(err)
				}
			
				result := make([]*Global, vs)
				for i := uint32(0); i < vs; i++ {
					if result[i], err = decodeGlobal(r); err != nil {
						fmt.Printf("global[%d]: %w", i, err)
						panic(err)
					}
				}
				m.globalSection = result 
			case sectionIDExport:
				// Not implemented
			
			case sectionIDStart:
				if m.startSection != nil {
					panic("multiple start sections are invalid")
				}
				vs, _, err := DecodeUint32(r)
				if err != nil {
					fmt.Errorf("get function index: %w", err)
					panic(err)
				}
				m.startSection = &vs
			
			case sectionIDElement:
				// m.elementSection, err = decodeElementSection(r)
				// Not implemented
			case sectionIDCode:
				vs, _, err := DecodeUint32(r)
				if err != nil {
					fmt.Errorf("get size of vector: %w", err)
					panic(err)
				}
			
				result := make([]*Code, vs)
				for i := uint32(0); i < vs; i++ {
					if result[i], err = decodeCode(r); err != nil {
						fmt.Errorf("read %d-th code segment: %v", i, err)
						panic(err)
					}
				}
				m.codeSection = result
			case sectionIDData:
				vs, _, err := DecodeUint32(r)
				if err != nil {
					fmt.Errorf("get size of vector: %w", err)
					panic(err)
				}
			
				result := make([]*DataSegment, vs)
				for i := uint32(0); i < vs; i++ {
					if result[i], err = decodeDataSegment(r); err != nil {
						fmt.Errorf("read data segment: %w", err)
						panic(err)
					}
				}

				m.dataSection = result
			case sectionIDDataCount:
				v, _, err := DecodeUint32(r)
				if err != nil && err != io.EOF {
					panic(err)
				}

				m.dataCountSection = &v

			readBytes := sectionContentStart - r.Len()

			if err == nil && int(sectionSize) != readBytes {
				err = fmt.Errorf("invalid section length: expected to be %d but got %d", sectionSize, readBytes)
			}
	
			if err != nil {
				fmt.Printf("section %s: %v", sectionIDName(sectionID), err)
			}
		}
	}

	functionCount, codeCount := m.sectionElementCount(sectionIDFunction), m.sectionElementCount(sectionIDCode)
	if functionCount != codeCount {
		fmt.Printf("function and code section have inconsistent lengths: %d != %d", functionCount, codeCount)
	}

	return m
}

func (m *Module) sectionElementCount(sectionID SectionID) uint32 { // element as in vector elements!
	switch sectionID {
		// case sectionIDCustom:
		// 	if m.nameSection != nil {
		// 		return 1
		// 	}
		// 	return 0
		case sectionIDType:
			return uint32(len(m.typeSection))
		// case sectionIDImport:
		// 	return uint32(len(m.importSection))
		case sectionIDFunction:
			return uint32(len(m.functionSection))
		case sectionIDTable:
			return uint32(len(m.tableSection))
		case sectionIDMemory:
			if m.memorySection != nil {
				return 1
			}
			return 0
		case sectionIDGlobal:
			return uint32(len(m.globalSection))
		// case sectionIDExport:
		// 	return uint32(len(m.exportSection))
		case sectionIDStart:
			if m.startSection != nil {
				return 1
			}
			return 0
		// case sectionIDElement:
		// 	return uint32(len(m.elementSection))
		case sectionIDCode:
			return uint32(len(m.codeSection))
		case sectionIDData:
			return uint32(len(m.dataSection))
		default:
			panic(fmt.Errorf("BUG: unknown section: %d", sectionID))
	}
}