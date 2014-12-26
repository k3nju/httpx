package httpx

import (
	"errors"
	"fmt"
	"io"
	"strconv"
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

func ReadResponse(r Reader, req *Request) (*Response, error) {
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
