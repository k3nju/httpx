package httpx

import (
	"bufio"
	"net"
)

type BufConn struct {
	*bufio.Reader
	C net.Conn
}

func NewBufConn(c net.Conn) *BufConn {
	return &BufConn{
		Reader: bufio.NewReader(c),
		C:      c,
	}
}

/*
func (bc *BufConn) Read(p []byte) (int, error) {
	return bc.br.Read(p)
}
*/

func (bc *BufConn) ReadLine() ([]byte, error) {
	return ReadLine(bc.Reader)
}

func (bc *BufConn) Write(p []byte) (int, error) {
	t := len(p)
	for len(p) > 0 {
		n, err := bc.C.Write(p)
		if n > 0 {
			p = p[n:]
		}
		if err != nil {
			return t - len(p), err
		}
	}

	return t - len(p), nil
}
