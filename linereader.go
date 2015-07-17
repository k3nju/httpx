package httpx

import (
	"bufio"
	"errors"
	"io"
	"strings"
)

var (
	ErrLineTooLong = errors.New("line too long")
)

type LineReader interface {
	ReadLine() ([]byte, error)
}

func ReadLine(br *bufio.Reader) ([]byte, error) {
	tmp, isPrefix, err := br.ReadLine()
	if err != nil {
		return nil, err
	}

	// NOTE: tmp references to inner buffer in bufio.Reader.
	//       must copy from tmp byte slice to own buffer
	line := make([]byte, len(tmp))
	copy(line, tmp)

	if !isPrefix {
		return line, nil
	}

	// read continued lines
	for i := 0; i < 10; i++ {
		tmp, isPrefix, err := br.ReadLine()
		if err != nil {
			return nil, err
		}
		line = append(line, tmp...)
		if !isPrefix {
			return line, nil
		}
	}

	return line, ErrLineTooLong
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
