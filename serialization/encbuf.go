package serialization

import (
	"io"
	"reflect"
)

type encbuf struct {
	str      []byte        // string data, contains everything except list headers
	lheads   []listhead    // all list headers
	lhsize   int           // sum of sizes of all encoded list headers
	sizebuf  [9]byte       // auxiliary buffer for uint encoding
	bufvalue reflect.Value // used in writeByteArrayCopy
}

func (w *encbuf) reset() {
	w.lhsize = 0
	w.str = w.str[:0]
	w.lheads = w.lheads[:0]
}

// encbuf implements io.Writer so it can be passed it into Encodeserialization.
func (w *encbuf) Write(b []byte) (int, error) {
	w.str = append(w.str, b...)
	return len(b), nil
}

func (w *encbuf) encode(val interface{}) error {
	rval := reflect.ValueOf(val)
	writer, err := cachedWriter(rval.Type())
	if err != nil {
		return err
	}
	return writer(rval, w)
}

func (w *encbuf) encodeStringHeader(size int) {
	if size < 56 {
		w.str = append(w.str, 0x80+byte(size))
	} else {
		sizesize := putint(w.sizebuf[1:], uint64(size))
		w.sizebuf[0] = 0xB7 + byte(sizesize)
		w.str = append(w.str, w.sizebuf[:sizesize+1]...)
	}
}

func (w *encbuf) encodeString(b []byte) {
	if len(b) == 1 && b[0] <= 0x7F {
		// fits single byte, no string header
		w.str = append(w.str, b[0])
	} else {
		w.encodeStringHeader(len(b))
		w.str = append(w.str, b...)
	}
}

func (w *encbuf) encodeUint(i uint64) {
	if i == 0 {
		w.str = append(w.str, 0x80)
	} else if i < 128 {
		// fits single byte
		w.str = append(w.str, byte(i))
	} else {
		s := putint(w.sizebuf[1:], i)
		w.sizebuf[0] = 0x80 + byte(s)
		w.str = append(w.str, w.sizebuf[:s+1]...)
	}
}

// list adds a new list header to the header stack. It returns the index
// of the header. The caller must call listEnd with this index after encoding
// the content of the list.
func (w *encbuf) list() int {
	w.lheads = append(w.lheads, listhead{offset: len(w.str), size: w.lhsize})
	return len(w.lheads) - 1
}

func (w *encbuf) listEnd(index int) {
	lh := &w.lheads[index]
	lh.size = w.size() - lh.offset - lh.size
	if lh.size < 56 {
		w.lhsize++ // length encoded into kind tag
	} else {
		w.lhsize += 1 + intsize(uint64(lh.size))
	}
}

func (w *encbuf) size() int {
	return len(w.str) + w.lhsize
}

func (w *encbuf) toBytes() []byte {
	out := make([]byte, w.size())
	strpos := 0
	pos := 0
	for _, head := range w.lheads {
		// write string data before header
		n := copy(out[pos:], w.str[strpos:head.offset])
		pos += n
		strpos += n
		// write the header
		enc := head.encode(out[pos:])
		pos += len(enc)
	}
	// copy string data after the last list header
	copy(out[pos:], w.str[strpos:])
	return out
}

func (w *encbuf) toWriter(out io.Writer) (err error) {
	strpos := 0
	for _, head := range w.lheads {
		// write string data before header
		if head.offset-strpos > 0 {
			n, err := out.Write(w.str[strpos:head.offset])
			strpos += n
			if err != nil {
				return err
			}
		}
		// write the header
		enc := head.encode(w.sizebuf[:])
		if _, err = out.Write(enc); err != nil {
			return err
		}
	}
	if strpos < len(w.str) {
		// write string data after the last list header
		_, err = out.Write(w.str[strpos:])
	}
	return err
}
