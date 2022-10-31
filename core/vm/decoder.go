package vm

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"math"
	"unicode/utf8"
)

type DataSegment struct {
	offsetExpression *ConstantExpression
	init             []byte
}

type ExternType = byte

type Import struct {
	Type ExternType
	Module string
	Name string
	DescFunc Index
	DescTable *Table
	DescMem *Memory
	DescGlobal *GlobalType
}

type Export struct {
	Type ExternType
	// Name is what the host refers to this definition as.
	name string
	// Index is the index of the definition to export, the index namespace is by Type
	// Ex. If ExternTypeFunc, this is a position in the function index namespace.
	index Index
}
type ElementSegment struct {

}

type FunctionType struct {
	// Params are the possibly empty sequence of value types accepted by a function with this signature.
	params []ValueType

	// Results are the possibly empty sequence of value types returned by a function with this signature.
	results []ValueType

	// string is cached as it is used both for String and key
	string string

	// ParamNumInUint64 is the number of uint64 values requires to represent the Wasm param type.
	paramNumInUint64 int

	// ResultsNumInUint64 is the number of uint64 values requires to represent the Wasm result type.
	resultNumInUint64 int
}

type ModuleID = [sha256.Size]byte
type SectionID = byte
type ValueType = byte
type Index = uint32

type Table struct {
	min  uint32
	max  *uint32
	refType RefType
}

type RefType = byte


type Memory struct {
	min, cap, max uint32
	// IsMaxEncoded true if the Max is encoded in the original source (binary or text).
	isMaxEncoded bool
}

type GlobalType struct {
	valType ValueType
	mutable bool
}

type Global struct {
	Type *GlobalType
	init *ConstantExpression
}

type ConstantExpression struct {
	opcode Opcode
	data   []byte
}

type Code struct {
	body []byte
	localTypes []ValueType
}


func sectionIDName(sectionID SectionID) string {
	switch sectionID {
	case sectionIDCustom:
		return "custom"
	case sectionIDType:
		return "type"
	case sectionIDImport:
		return "import"
	case sectionIDFunction:
		return "function"
	case sectionIDTable:
		return "table"
	case sectionIDMemory:
		return "memory"
	case sectionIDGlobal:
		return "global"
	case sectionIDExport:
		return "export"
	case sectionIDStart:
		return "start"
	case sectionIDElement:
		return "element"
	case sectionIDCode:
		return "code"
	case sectionIDData:
		return "data"
	case sectionIDDataCount:
		return "data_count"
	}
	return "unknown"
}

const (
	sectionIDCustom SectionID = iota
	sectionIDType
	sectionIDImport
	sectionIDFunction
	sectionIDTable
	sectionIDMemory
	sectionIDGlobal
	sectionIDExport
	sectionIDStart
	sectionIDElement
	sectionIDCode
	sectionIDData
	sectionIDDataCount
)
const (
	maxVarintLen32 = 5
	maxVarintLen64 = 10
)



var Magic = []byte{0x00, 0x61, 0x73, 0x6D}
var version = []byte{0x01, 0x00, 0x00, 0x00}

func decodeValueTypes(r *bytes.Reader, num uint32) ([]byte, error) {
	if num == 0 {
		return nil, nil
	}
	ret := make([]byte, num)
	buf := make([]byte, num)
	_, err := io.ReadFull(r, buf)
	if err != nil {
		return nil, err
	}

	for i, v := range buf {
		switch v {
		case Op_i32, Op_f32, Op_i64, Op_f64:
			ret[i] = v
		default:
			return nil, fmt.Errorf("invalid value type: %d", v)
		}
	}
	return ret, nil
}

func decodeFunctionType(r *bytes.Reader) (*FunctionType, error) {
	b, err := r.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("read leading byte: %w", err)
	}

	if b != 0x60 {
		return nil, fmt.Errorf("Invalid byte %#x != 0x60", b)
	}

	paramCount, _, err := DecodeUint32(r)
	if err != nil {
		return nil, fmt.Errorf("could not read parameter count: %w", err)
	}

	paramTypes, err := decodeValueTypes(r, paramCount)
	if err != nil {
		return nil, fmt.Errorf("could not read parameter types: %w", err)
	}

	resultCount, _, err := DecodeUint32(r)
	if err != nil {
		return nil, fmt.Errorf("could not read result count: %w", err)
	}

	resultTypes, err := decodeValueTypes(r, resultCount)
	if err != nil {
		return nil, fmt.Errorf("could not read result types: %w", err)
	}

	ret := &FunctionType{
		params:  paramTypes,
		results: resultTypes,
	}
	return ret, nil
}


func decodeLimitsType(r *bytes.Reader) (min uint32, max *uint32, err error) {
	var flag byte
	if flag, err = r.ReadByte(); err != nil {
		err = fmt.Errorf("read leading byte: %v", err)
		return
	}

	switch flag {
	case 0x00:
		min, _, err = DecodeUint32(r)
		if err != nil {
			err = fmt.Errorf("read min of limit: %v", err)
		}
	case 0x01:
		min, _, err = DecodeUint32(r)
		if err != nil {
			err = fmt.Errorf("read min of limit: %v", err)
			return
		}
		var m uint32
		if m, _, err = DecodeUint32(r); err != nil {
			err = fmt.Errorf("read max of limit: %v", err)
		} else {
			max = &m
		}
	default:
		err = fmt.Errorf("%v for limits: %#x != 0x00 or 0x01", errors.New("invalid byte"), flag)
	}
	return
}

func decodeTable(r *bytes.Reader) (*Table, error) {
	_ , err := r.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("read leading byte: %v", err)
	}

	min, max, err := decodeLimitsType(r)
	if err != nil {
		return nil, fmt.Errorf("read limits: %v", err)
	}

	// We do not validate the values of min and max for now
	return &Table{min: min, max: max}, nil
}

func decodeGlobal(r *bytes.Reader) (*Global, error) {
	vt, err := decodeValueTypes(r, 1)
	if err != nil {
		return nil, fmt.Errorf("read value type: %w", err)
	}

	ret := &GlobalType{
		valType: vt[0],
	}

	b, err := r.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("read mutablity: %w", err)
	}

	switch mut := b; mut {
	case 0x00: // not mutable
	case 0x01: // mutable
		ret.mutable = true
	default:
		return nil, fmt.Errorf("%w for mutability: %#x != 0x00 or 0x01", errors.New("Invalid byte"), mut)
	}

	init, err := decodeConstantExpression(r)
	if err != nil {
		return nil, err
	}

	return &Global{Type: ret, init: init}, nil
}

// dataSegmentPrefix represents three types of data segments.
//
// https://www.w3.org/TR/2022/WD-wasm-core-2-20220419/binary/modules.html#data-section
type dataSegmentPrefix = uint32

const (
	// dataSegmentPrefixActive is the prefix for the version 1.0 compatible data segment, which is classified as "active" in 2.0.
	dataSegmentPrefixActive dataSegmentPrefix = 0x0
	// dataSegmentPrefixPassive prefixes the "passive" data segment as in version 2.0 specification.
	dataSegmentPrefixPassive dataSegmentPrefix = 0x1
	// dataSegmentPrefixActiveWithMemoryIndex is the active prefix with memory index encoded which is defined for future use as of 2.0.
	dataSegmentPrefixActiveWithMemoryIndex dataSegmentPrefix = 0x2
)

func decodeDataSegment(r *bytes.Reader) (*DataSegment, error) {
	dataSegmentPrefx, _, err := DecodeUint32(r)
	if err != nil {
		return nil, fmt.Errorf("read data segment prefix: %w", err)
	}

	if dataSegmentPrefx != dataSegmentPrefixActive {
		panic("non-zero prefix for data segment is invalid as")

	}

	var expr *ConstantExpression
	switch dataSegmentPrefx {
	case dataSegmentPrefixActive,
		dataSegmentPrefixActiveWithMemoryIndex:
		// Active data segment as in
		// https://www.w3.org/TR/2022/WD-wasm-core-2-20220419/binary/modules.html#data-section
		if dataSegmentPrefx == 0x2 {
			d, _, err := DecodeUint32(r)
			if err != nil {
				return nil, fmt.Errorf("read memory index: %v", err)
			} else if d != 0 {
				return nil, fmt.Errorf("memory index must be zero but was %d", d)
			}
		}

		expr, err = decodeConstantExpression(r)
		if err != nil {
			return nil, fmt.Errorf("read offset expression: %v", err)
		}
	case dataSegmentPrefixPassive:
		// Passive data segment doesn't need const expr nor memory index encoded.
		// https://www.w3.org/TR/2022/WD-wasm-core-2-20220419/binary/modules.html#data-section
	default:
		return nil, fmt.Errorf("invalid data segment prefix: 0x%x", dataSegmentPrefx)
	}

	vs, _, err := DecodeUint32(r)
	if err != nil {
		return nil, fmt.Errorf("get the size of vector: %v", err)
	}

	b := make([]byte, vs)
	if _, err := io.ReadFull(r, b); err != nil {
		return nil, fmt.Errorf("read bytes for init: %v", err)
	}

	return &DataSegment{
		offsetExpression: expr,
		init:             b,
	}, nil
}

func decodeCode(r *bytes.Reader) (*Code, error) {
	ss, _, err := DecodeUint32(r)
	if err != nil {
		return nil, fmt.Errorf("get the size of code: %w", err)
	}
	remaining := int64(ss)

	// parse locals
	ls, bytesRead, err := DecodeUint32(r)
	remaining -= int64(bytesRead)
	if err != nil {
		return nil, fmt.Errorf("get the size locals: %v", err)
	} else if remaining < 0 {
		return nil, io.EOF
	}

	var nums []uint64
	var types []ValueType
	var sum uint64
	var n uint32
	for i := uint32(0); i < ls; i++ {
		n, bytesRead, err = DecodeUint32(r)
		remaining -= int64(bytesRead) + 1 // +1 for the subsequent ReadByte
		if err != nil {
			return nil, fmt.Errorf("read n of locals: %v", err)
		} else if remaining < 0 {
			return nil, io.EOF
		}

		sum += uint64(n)
		nums = append(nums, uint64(n))

		b, err := r.ReadByte()
		if err != nil {
			return nil, fmt.Errorf("read type of local: %v", err)
		}
		switch vt := b; vt {
		case Op_i32, Op_f32, Op_i64, Op_f64,
			Op_func, Op_anyfunc:
			types = append(types, vt)
		default:
			return nil, fmt.Errorf("invalid local type: 0x%x", vt)
		}
	}

	if sum > math.MaxUint32 {
		return nil, fmt.Errorf("too many locals: %d", sum)
	}

	var localTypes []ValueType
	for i, num := range nums {
		t := types[i]
		for j := uint64(0); j < num; j++ {
			localTypes = append(localTypes, t)
		}
	}

	body := make([]byte, remaining)
	if _, err = io.ReadFull(r, body); err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if endIndex := len(body) - 1; endIndex < 0 || body[endIndex] != Op_end {
		return nil, fmt.Errorf("expr not end with OpcodeEnd")
	}

	return &Code{body: body, localTypes: localTypes}, nil
}

func decodeConstantExpression(r *bytes.Reader) (*ConstantExpression, error) {
	b, err := r.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("read opcode: %v", err)
	}

	remainingBeforeData := int64(r.Len())
	offsetAtData := r.Size() - remainingBeforeData

	opcode := b
	switch opcode {
	case Op_i32_const:
		// Treat constants as signed as their interpretation is not yet known per /RATIONALE.md
		_, _, err = DecodeInt32(r)
	case Op_i64_const:
		// Treat constants as signed as their interpretation is not yet known per /RATIONALE.md
		_, _, err = DecodeInt64(r)
	case Op_f32_const:
		_, err = DecodeFloat32(r)
	case  Op_f64_const:
		_, err = DecodeFloat64(r)
	case Op_get_global:
		_, _, err = DecodeUint32(r)
	case Op_anyfunc:
		_, _, err = DecodeUint32(r)
	default:
		return nil, fmt.Errorf("%v for const expression opt code: %#x", errors.New("Something went wrong"), b)
	}

	if err != nil {
		return nil, fmt.Errorf("read value: %v", err)
	}

	if b, err = r.ReadByte(); err != nil {
		return nil, fmt.Errorf("look for end opcode: %v", err)
	}

	if b != Op_end {
		return nil, fmt.Errorf("constant expression has been not terminated")
	}

	data := make([]byte, remainingBeforeData-int64(r.Len())-1)
	if _, err := r.ReadAt(data, offsetAtData); err != nil {
		return nil, fmt.Errorf("error re-buffering ConstantExpression.Data")
	}

	return &ConstantExpression{opcode: Opcode(opcode), data: data}, nil
}

func ExternTypeName(et ExternType) string {
	switch et {
	case 0x00:
		return "func"
	case 0x01:
		return "table"
	case 0x02:
		return "memory"
	case 0x03:
		return "global"
	}
	return fmt.Sprintf("%#x", et)
}


func decodeUTF8(r *bytes.Reader, contextFormat string, contextArgs ...interface{}) (string, uint32, error) {
	size, sizeOfSize, err := DecodeUint32(r)
	if err != nil {
		return "", 0, fmt.Errorf("failed to read %s size: %w", fmt.Sprintf(contextFormat, contextArgs...), err)
	}

	buf := make([]byte, size)
	if _, err = io.ReadFull(r, buf); err != nil {
		return "", 0, fmt.Errorf("failed to read %s: %w", fmt.Sprintf(contextFormat, contextArgs...), err)
	}

	if !utf8.Valid(buf) {
		return "", 0, fmt.Errorf("%s is not valid UTF-8", fmt.Sprintf(contextFormat, contextArgs...))
	}

	return string(buf), size + uint32(sizeOfSize), nil
}


func decodeImport(r *bytes.Reader, idx uint32) (i *Import, err error) {
	i = &Import{}
	if i.Module, _, err = decodeUTF8(r, "import module"); err != nil {
		return nil, fmt.Errorf("import[%d] error decoding module: %w", idx, err)
	}

	if i.Name, _, err = decodeUTF8(r, "import name"); err != nil {
		return nil, fmt.Errorf("import[%d] error decoding name: %w", idx, err)
	}

	b, err := r.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("import[%d] error decoding type: %w", idx, err)
	}
	i.Type = b
	switch i.Type {
	case 0x00:
		i.DescFunc, _, err = DecodeUint32(r)
	case 0x01:
		i.DescTable, err = decodeTable(r)
	case 0x02:
		i.DescMem, err = decodeMemory(r)
	case 0x03:
		i.DescGlobal, err = decodeGlobalType(r)
	default:
		err = fmt.Errorf("%w: invalid byte for importdesc", b)
	}
	if err != nil {
		return nil, fmt.Errorf("import[%d] %s[%s.%s]: %w", idx, ExternTypeName(i.Type), i.Module, i.Name, err)
	}
	return
}

func decodeMemory(r *bytes.Reader) (*Memory, error) {
	min, maxP, err := decodeLimitsType(r)
	if err != nil {
		return nil, err
	}

	min, capacity, max := min, min, defaultPageSize
	mem := &Memory{min: min, cap: capacity, max: uint32(max), isMaxEncoded: maxP != nil}

	return mem, nil
}

func decodeGlobalType(r *bytes.Reader) (*GlobalType, error) {
	vt, err := decodeValueTypes(r, 1)
	if err != nil {
		return nil, fmt.Errorf("read value type: %w", err)
	}

	ret := &GlobalType{
		valType: vt[0],
	}

	b, err := r.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("read mutablity: %w", err)
	}

	switch mut := b; mut {
	case 0x00: // not mutable
	case 0x01: // mutable
		ret.mutable = true
	default:
		return nil, fmt.Errorf("%w for mutability: %#x != 0x00 or 0x01", errors.New("Invalid bytes"), mut)
	}
	return ret, nil
}

func decodeExport(r *bytes.Reader) (i *Export, err error) {
	i = &Export{}

	if i.name, _, err = decodeUTF8(r, "export name"); err != nil {
		return nil, err
	}

	b, err := r.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("error decoding export kind: %w", err)
	}

	i.Type = b
	switch i.Type {
	case 0x00, 0x01, 0x02, 0x03:
		if i.index, _, err = DecodeUint32(r); err != nil {
			return nil, fmt.Errorf("error decoding export index: %w", err)
		}
	default:
		return nil, fmt.Errorf("%w: invalid byte for exportdesc", b)
	}
	return
}
