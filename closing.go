package httpx

import (
	"io"
)

type ClosingReader struct {
	r   io.Reader
	err error
}

func NewClosingReader(r io.Reader) *ClosingReader {
	return &ClosingReader{
		r: r,
	}
}

func (r *ClosingReader) Read() (*BodyBlock, error) {
	if r.err != nil {
		return nil, r.err
	}

	buf := make([]byte, DefaultBodyBlockSize)
	n, err := r.r.Read(buf)
	if err == io.EOF {
		err = EOB
	}
	r.err = err

	if n > 0 {
		return &BodyBlock{
			Data: buf[:n],
		}, nil
	}

	return nil, r.err
}
