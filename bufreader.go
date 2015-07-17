package httpx

import (
	"bufio"
	"io"
)

type BufferedReader struct {
	*bufio.Reader
}

func NewBufferedReader(r io.Reader) *BufferedReader {
	br, ok := r.(*bufio.Reader)
	if !ok {
		br = bufio.NewReader(r)
	}

	return &BufferedReader{Reader: br}
}

func (r *BufferedReader) ReadLine() ([]byte, error) {
	return ReadLine(r.Reader)
}
