package httpx

import (
	"io"
	"strings"
)

type LineReader interface {
	ReadLine() ([]byte, error)
}

//-----------------------------------------------------------------------------------------//
// for test use

type stringLineReader struct {
	i     int
	lines []string
}

func newStringLineReader(src string) *stringLineReader {
	var lines []string
	for _, l := range strings.Split(src, "\n") {
		lines = append(lines, strings.TrimRight(l, "\r"))
	}

	return &stringLineReader{
		lines: lines,
	}
}

func (r *stringLineReader) ReadLine() ([]byte, error) {
	if r.i >= len(r.lines) {
		return nil, io.EOF
	}

	ret := r.lines[r.i]
	r.i++

	return []byte(ret), nil
}
