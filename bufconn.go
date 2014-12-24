package httpx

import (
	"bufio"
	"errors"
	"net"
)

var (
	ErrLineTooLong = errors.New("line too long")
)

type BufConn struct {
	C  net.Conn
	br *bufio.Reader
}

func NewBufConn(c net.Conn) *BufConn {
	return &BufConn{
		C:  c,
		br: bufio.NewReader(c),
	}
}

func (bc *BufConn) Read(p []byte) (int, error) {
	return bc.br.Read(p)
}

func (bc *BufConn) ReadLine() ([]byte, error) {
	tmp, isPrefix, err := bc.br.ReadLine()
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
		tmp, isPrefix, err := bc.br.ReadLine()
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
