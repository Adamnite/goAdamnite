package vm

import (
	"bytes"
	"fmt"
	"io"
)

// See https://www.w3.org/TR/2019/REC-wasm-core-1-20191205/
type Module struct {
	typeSection     []*FunctionType
	importSection   []*Import
	functionSection []Index

	// Tables contents references to functions. This can be used to achieve dynamic function calling.
	// It will be used by the `call_inderect` opcode
	// https://developer.mozilla.org/en-US/docs/WebAssembly/Understanding_the_text_format#webassembly_tables
	tableSection     []*Table
	memorySection    *Memory
	globalSection    []*Global
	exportSection    []*Export
	startSection     *Index
	elementSection   []*ElementSegment
	codeSection      []*Code
	dataSection      []*DataSegment
	dataCountSection *uint32
	ID               ModuleID
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
			panic(fmt.Errorf("read section id: %w", err))
		}

		sectionSize, _, err := DecodeUint32(r)
		if err != nil {
			msg := fmt.Errorf("get size of section %s: %v", sectionIDName(sectionID), err)
			panic(msg)
		}
		sectionContentStart := r.Len()

		switch sectionID {
		case sectionIDType:
			vs, _, err := DecodeUint32(r)
			if err != nil {
				msg := fmt.Errorf("get size of vector: %w", err)
				panic(msg)
			}

			result := make([]*FunctionType, vs)
			for i := uint32(0); i < vs; i++ {
				if result[i], err = decodeFunctionType(r); err != nil {
					panic(fmt.Errorf("read %d-th type: %v", i, err))
				}
			}

			m.typeSection = result

		case sectionIDImport:
			vs, _, err := DecodeUint32(r)
			if err != nil {
				panic(fmt.Errorf("get size of vector: %w", err))
			}

			result := make([]*Import, vs)
			for i := uint32(0); i < vs; i++ {
				if result[i], err = decodeImport(r, i); err != nil {
					panic(err)
				}
			}
			m.importSection = result

		case sectionIDFunction:
			vs, _, err := DecodeUint32(r)
			if err != nil {
				panic(fmt.Errorf("get size of vector: %w", err))
			}

			result := make([]uint32, vs)
			for i := uint32(0); i < vs; i++ {
				if result[i], _, err = DecodeUint32(r); err != nil {
					panic(fmt.Errorf("get type index: %w", err))
				}
			}
			m.functionSection = result

		case sectionIDTable:
			vs, _, err := DecodeUint32(r)
			if err != nil {
				panic(fmt.Errorf("error reading size %w", err))
			}

			ret := make([]*Table, vs)
			exitFlag := false
			for i := 0; i < int(vs) && !exitFlag; i++ {
				table, err := decodeTable(r)
				if err != nil {
					fmt.Println(err)
					exitFlag = err.Error() == "read leading byte: EOF"
				}
				ret[i] = table

			}
			m.tableSection = ret

		case sectionIDMemory:
			vs, _, err := DecodeUint32(r)
			if err != nil {
				panic("error reading size")
			}
			if vs > 1 {
				panic(fmt.Sprintf("at most one memory allowed in module, but read %d", vs))
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
				panic(fmt.Errorf("get size of vector: %w", err))
			}

			result := make([]*Global, vs)
			for i := uint32(0); i < vs; i++ {
				if result[i], err = decodeGlobal(r); err != nil {
					panic(fmt.Errorf("global[%d]: %w", i, err))
				}
			}
			m.globalSection = result

		case sectionIDExport:
			vs, _, sizeErr := DecodeUint32(r)
			if sizeErr != nil {
				panic(fmt.Errorf("get size of vector: %v", sizeErr))
			}

			usedName := make(map[string]struct{}, vs)
			exportSection := make([]*Export, 0, vs)
			for i := Index(0); i < vs; i++ {
				export, err := decodeExport(r)
				if err != nil {
					panic(fmt.Errorf("read export: %s", err))
				}

				if _, ok := usedName[export.name]; ok {
					panic(fmt.Errorf("export[%d] duplicates name %q", i, export.name))
				} else {
					usedName[export.name] = struct{}{}
				}
				exportSection = append(exportSection, export)
			}
			m.exportSection = exportSection

		case sectionIDStart:
			if m.startSection != nil {
				panic("multiple start sections are invalid")
			}
			vs, _, err := DecodeUint32(r)
			if err != nil {
				panic(fmt.Errorf("get function index: %w", err))
			}
			m.startSection = &vs

		case sectionIDElement:
			// m.elementSection, err = decodeElementSection(r)
			// Not implemented
			panic("SECTION ID ELEMENT FOUND")
		case sectionIDCode:
			vs, _, err := DecodeUint32(r)
			if err != nil {
				panic(fmt.Errorf("get size of vector: %w", err))
			}

			result := make([]*Code, vs)
			for i := uint32(0); i < vs; i++ {
				if result[i], err = decodeCode(r); err != nil {
					panic(fmt.Errorf("read %d-th code segment: %v", i, err))
				}
			}
			m.codeSection = result
		case sectionIDData:
			vs, _, err := DecodeUint32(r)
			if err != nil {
				panic(fmt.Errorf("get size of vector: %s", err))
			}

			result := make([]*DataSegment, vs)
			for i := uint32(0); i < vs; i++ {
				if result[i], err = decodeDataSegment(r); err != nil {
					panic(fmt.Sprintf("read data segment: %s", err))
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
				panic(err)
			}

			if err != nil {
				panic(fmt.Errorf("section %s: %v", sectionIDName(sectionID), err))
			}
		}
	}

	functionCount, codeCount := m.sectionElementCount(sectionIDFunction), m.sectionElementCount(sectionIDCode)
	if functionCount != codeCount {
		panic(fmt.Sprintf("function and code section have inconsistent lengths: %d != %d", functionCount, codeCount))
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
	case sectionIDImport:
		return uint32(len(m.importSection))
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
	case sectionIDExport:
		return uint32(len(m.exportSection))
	case sectionIDStart:
		if m.startSection != nil {
			return 1
		}
		return 0
	case sectionIDElement:
		return uint32(len(m.elementSection))
	case sectionIDCode:
		return uint32(len(m.codeSection))
	case sectionIDData:
		return uint32(len(m.dataSection))
	default:
		panic(fmt.Errorf("BUG: unknown section: %d", sectionID))
	}

}
