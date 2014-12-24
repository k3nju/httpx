package httpx

import (
	"io"
)

type Reader interface {
	LineReader
	io.Reader
}

type ReadWriter interface {
	Reader
	io.Writer
}
