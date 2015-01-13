package httpx

import (
	"bytes"
	"errors"
	"strconv"
)

const (
	MaxChunkHeaderSize = 512
)

var (
	ErrTooLargeChunkHeader = errors.New("too large chunk header")
	ErrInsufficientBuffer  = errors.New("insufficient buffer")
)

type cbReadResult struct {
	data []byte
	err  error
}

type ChunkedBodyReader struct {
	r        Reader
	Trailers *Headers
	err      error
	ch       chan *cbReadResult
}

func NewChunkedBodyReader(r Reader) *ChunkedBodyReader {
	cbr := &ChunkedBodyReader{
		r:  r,
		ch: make(chan *cbReadResult, 1),
	}

	go cbr.read()

	return cbr
}

func (r *ChunkedBodyReader) read() {
	defer close(r.ch)
	buf := NewTmpBuf(DefaultBodyBlockSize)

	for {
		// read chunk header
		line, size, err := cbReadChunkHeader(r.r)
		if err != nil {
			r.ch <- &cbReadResult{err: NewErrorFrom("reading chunk header failed", err)}
			return
		}

		// chunk-size == 0 means end of chunks(last-chunk)
		if size == 0 {
			// copy raw chunk header and break to read trailers
			// copying raw chunk header for keeping chunk-ext if it exists.
			tmp := make([]byte, len(line)+2)
			copy(tmp[copy(tmp, line):], []byte("\r\n"))
			r.ch <- &cbReadResult{data: tmp}
			break
		}

		// prepare to read chunk-data
		tmp, detached := buf.Prepare(uint64(len(line) + 2))
		if detached != nil {
			// new memory allocated and remained data is detached in TmpBuf
			// send detached data to receiver.
			r.ch <- &cbReadResult{data: detached}
		}
		if tmp == nil {
			// required size is larger than TmpBuf inner buffer size
			r.ch <- &cbReadResult{err: ErrTooLargeChunkHeader}
			return
		}

		// copy chunk header and notify copied size to TmpBuf
		copy(tmp[copy(tmp, line):], []byte("\r\n"))
		buf.Consume(uint64(len(line) + 2))

		// chunk-data size. "+= 2" means "\r\n" at end of chunk-data
		size += 2

		// read chunk
		for size > 0 {
			// limiting max read size up to DefaultBodyBlockSize
			rsize := size
			if rsize > DefaultBodyBlockSize {
				rsize = DefaultBodyBlockSize
			}

			// ensure to read data up to rsize
			rbuf, detached := buf.Prepare(rsize)
			if detached != nil {
				r.ch <- &cbReadResult{data: detached}
			}
			if rbuf == nil {
				r.ch <- &cbReadResult{err: ErrInsufficientBuffer}
				return
			}

			n, err := r.r.Read(rbuf)
			if err != nil {
				r.ch <- &cbReadResult{err: NewErrorFrom("Read() failed(reading chunk)", err)}
				return
			}
			if n > 0 {
				buf.Consume(uint64(n))
				r.ch <- &cbReadResult{data: buf.Detach()}
				if size -= uint64(n); size == 0 {
					break
				}
			}
		}
	}

	// read trailers
	t, err := ReadHeaders(r.r)
	if err != nil {
		r.ch <- &cbReadResult{
			err: NewErrorFrom("ReadHeaders() failed(reading trailer)", err),
		}
		return
	}
	r.Trailers = t
}

func cbReadChunkHeader(lr LineReader) ([]byte, uint64, error) {
	line, err := lr.ReadLine()
	if err != nil {
		return nil, 0, err
	}
	if len(line)+2 > MaxChunkHeaderSize {
		return nil, 0, ErrTooLargeChunkHeader
	}

	s := line
	if i := bytes.Index(s, []byte(";")); i != -1 {
		s = s[:i]
	}

	size, err := strconv.ParseUint(string(s), 16, 64)
	if err != nil {
		return nil, 0, err
	}

	return line, size, nil
}

func (r *ChunkedBodyReader) Read() ([]byte, error) {
	if r.err != nil {
		return nil, r.err
	}

	res := <-r.ch
	if res == nil {
		r.err = EOB
		return nil, EOB
	}
	if res.err != nil {
		if res.data != nil {
			// when res.err != nil, res.data must be nil
			panic("unexpected read result: res.data != nil && res.err != nil")
		}

		r.err = res.err
		return nil, r.err
	}

	return res.data, nil
}
