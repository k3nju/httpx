package httpx

import (
	"io"
)

type ContentLengthReader struct {
	r      io.Reader
	remain uint64
	err    error
}

func NewContentLengthReader(r io.Reader, length uint64) *ContentLengthReader {
	return &ContentLengthReader{
		r:      r,
		remain: length,
	}
}

func (r *ContentLengthReader) Read() (*BodyBlock, error) {
	if r.err != nil {
		return nil, r.err
	}

	var buf []byte
	if r.remain > DefaultBodyBlockSize {
		buf = make([]byte, DefaultBodyBlockSize)
	} else {
		buf = make([]byte, r.remain)
	}

	var n int
	n, r.err = r.r.Read(buf)
	if n > 0 {
		if r.remain -= uint64(n); r.remain == 0 {
			r.err = EOB // for next call
		}
		return &BodyBlock{
			Data: buf[:n],
		}, nil
	}

	if n <= 0 && r.err == nil {
		panic("unexpected condition: n <= 0 && r.err == nil")
	}

	return nil, r.err
}
