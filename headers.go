package httpx

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

const (
	maxLineCount = 200
)

var (
	//ErrMalformedHeader = errors.New("malformed header")
	ErrEOHNotFound   = errors.New("end of header(empty line) not found")
	ErrColonNotFound = errors.New("header field delimiter(':') not found")
)

/* Data structure in Headers struct
* fields                * index
  [0] []byte("A: B") <----key:"A", values:{
  [1] nil               |            *fieldIndex{field:0, value:2, contCount:0}
  [2] []byte("A: C") <--|            *fieldIndex{field:2, value:2, contCount:0} }
  [3] []byte("D: E") <----key:"D", values:{
  [4] []byte(" F")                   *fieldIndex{field:3, value:2, contCount:1} }

  fiels[1] is deleted entry.
*/

type fieldIndex struct {
	field     int // line index to Headers.fields
	value     int // index starting value in Headers.fields[N]
	contCount int // continued line count
}

type Headers struct {
	fields [][]byte
	index  map[string][]*fieldIndex
}

func (h *Headers) Set(name string, value []byte) {
	if h == nil {
		return
	}

	lname := strings.ToLower(name)
	h.Del(lname)

	f, valpos := newHeaderField([]byte(name), value)
	h.fields = append(h.fields, f)
	fidx := &fieldIndex{
		field: len(h.fields) - 1,
		value: valpos,
	}

	h.index[lname] = append(h.index[lname], fidx)
}

func (h *Headers) Get(name string) [][]byte {
	if h == nil {
		return nil
	}

	fidxs, ok := h.index[strings.ToLower(name)]
	if !ok || len(fidxs) == 0 {
		return nil
	}

	var b bytes.Buffer
	for _, fidx := range fidxs {
		// write 1st line
		b.Write(h.fields[fidx.field][fidx.value:])
		// write continued lines
		for i := 0; i < fidx.contCount; i++ {
			b.Write([]byte(" "))
			b.Write(bytes.TrimLeft(h.fields[fidx.field+1+i], " \t"))
		}
	}

	parts := bytes.Split(b.Bytes(), []byte(","))
	for i := 0; i < len(parts); i++ {
		parts[i] = trimAsFieldValue(parts[i])
	}

	return parts
}

func (h *Headers) Del(name string) {
	if h == nil {
		return
	}
	name = strings.ToLower(name)

	fidxs, ok := h.index[name]
	if !ok || len(fidxs) == 0 {
		return
	}

	for _, fidx := range fidxs {
		h.fields[fidx.field] = nil
		for i := 0; i < fidx.contCount; i++ {
			h.fields[fidx.field+1+i] = nil
		}
	}

	delete(h.index, name)
}

func (h *Headers) List() [][]byte {
	if h == nil {
		return nil
	}

	var ret [][]byte
	for _, f := range h.fields {
		if f != nil {
			ret = append(ret, f)
		}
	}

	return ret
}

func newHeaderField(name, value []byte) ([]byte, int) {
	// length = len(name) + ": " + len(value)
	l := make([]byte, len(name)+2+len(value))
	n := copy(l, name)
	n += copy(l[n:], []byte(": "))
	copy(l[n:], value)

	return l, n
}

func ReadHeaders(lr LineReader) (*Headers, error) {
	reachedEOH := false
	fields := make([][]byte, 0, 20)
	index := make(map[string][]*fieldIndex)
	lineIdx := 0
	var prev *fieldIndex

	for i := 0; i < maxLineCount; i++ {
		line, err := lr.ReadLine()
		if err != nil {
			// if err is non-nil, which means we didn't reached to the end of header.
			// so no need to parse uncompleted header, just return.
			return nil, err
		}
		if len(line) == 0 {
			reachedEOH = true
			break
		}

		fields = append(fields, line)

		if isNewLine(line[0]) {
			name, valpos, err := parseField(line)
			if err != nil {
				return nil, NewErrorFrom(
					fmt.Sprintf("parsing header field failed at %d", i),
					err)
			}

			fidx := &fieldIndex{
				field: lineIdx,
				value: valpos,
			}
			index[name] = append(index[name], fidx)
			prev = fidx
			lineIdx++
		} else {
			if prev == nil {
				panic("unexpected condition: prev == nil")
			}

			prev.contCount += 1
		}
	}

	if !reachedEOH {
		return nil, ErrEOHNotFound
	}

	if len(fields) == 0 {
		return nil, nil
	}

	return &Headers{
		fields: fields,
		index:  index,
	}, nil
}

func isNewLine(b byte) bool {
	return b != 0x09 && b != 0x20
}

func parseField(line []byte) (string, int, error) {
	i := bytes.Index(line, []byte(":"))
	if i == -1 {
		return "", 0, ErrColonNotFound
	}
	name := line[:i]

	/* postpone
	name := trimAsToken(name)
	if len(name) == 0 {
		return "", 0, ErrMalformedHeader
	}
	*/

	name = bytes.ToLower(name)

	// "field-name", field-value index in line, no-error
	return string(name), i + 1, nil
}
