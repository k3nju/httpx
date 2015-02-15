package httpx

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

var (
	ErrMalformedResponseLine = errors.New("malformed response line")
)

type Response struct {
	HTTPVersion  *HTTPVersion
	StatusCode   uint
	ReasonPhrase string

	Headers *Headers

	Body BodyReader
}

func (res *Response) HeaderBytes() []byte {
	sl := strings.Join(
		[]string{
			res.HTTPVersion.String(),
			strconv.Itoa(int(res.StatusCode)),
			res.ReasonPhrase},
		" ")
	return bytes.Join(
		[][]byte{
			[]byte(sl),          // status line
			res.Headers.Bytes(), // headers
			[]byte("\r\n"),      // last line
		},
		[]byte("\r\n"))
}

func (res *Response) BodyReader() BodyReader {
	return res.Body
}

func parseStatusLine(line []byte) (*HTTPVersion, uint, string, error) {
	v, sc, rp, ok := parseStartLine(line)
	if !ok {
		return nil, 0, "", ErrMalformedResponseLine
	}

	hv, err := ParseHTTPVersion(v)
	if err != nil {
		return nil, 0, "", err
	}

	t, err := strconv.ParseUint(string(sc), 0, 16)
	if err != nil {
		return nil, 0, "", ErrMalformedResponseLine
	}

	return hv, uint(t), string(rp), nil
}

func ReadResponseHeader(r Reader) (*Response, error) {
	line, err := r.ReadLine()
	if err != nil {
		return nil, err
	}

	res := &Response{}
	res.HTTPVersion, res.StatusCode, res.ReasonPhrase, err = parseStatusLine(line)
	if err != nil {
		return nil, err
	}

	res.Headers, err = ReadHeaders(r)
	if err != nil {
		return nil, NewErrorFrom("ReadHeaders() failed", err)
	}

	return res, nil
}

func ReadResponse(r Reader, req *Request) (*Response, error) {
	res, err := ReadResponseHeader(r)
	if err != nil {
		return nil, err
	}

	if err := SetResponseBodyReader(res, r, req); err != nil {
		return nil, NewErrorFrom("SetResponseBodyReader() failed", err)
	}

	return res, nil
}

func DumpResponse(w io.Writer, res *Response) {
	fmt.Fprintf(w, "%s %d %s", res.HTTPVersion, res.StatusCode, res.ReasonPhrase)
	for _, line := range res.Headers.List() {
		fmt.Fprintf(w, "%s\r\n", line)
	}

}
