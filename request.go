package httpx

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
)

var (
	ErrMalformedRequestLine = errors.New("malformed request line")
)

type Request struct {
	Method        string
	RequestTarget string
	HTTPVersion   *HTTPVersion

	Headers *Headers

	Body BodyReader
}

func (req *Request) HeaderBytes() []byte {
	rl := strings.Join([]string{req.Method, req.RequestTarget, req.HTTPVersion.String()}, " ")
	return bytes.Join(
		[][]byte{
			[]byte(rl),          // request line
			req.Headers.Bytes(), // headers
			[]byte("\r\n"),      // last line
		},
		[]byte("\r\n"))
}

func (req *Request) BodyReader() BodyReader {
	return req.Body
}

func parseRequestLine(line []byte) (string, string, *HTTPVersion, error) {
	m, rt, v, ok := parseStartLine(line)
	if !ok {
		return "", "", nil, ErrMalformedRequestLine
	}

	hv, err := ParseHTTPVersion(v)
	if err != nil {
		return "", "", nil, err
	}

	return string(m), string(rt), hv, nil
}

func ReadRequest(r Reader) (*Request, error) {
	line, err := r.ReadLine()
	// LineReader.ReadLine returns
	// * valid line data and nil error
	// OR
	// * invalid line data and non-nil error
	// It never returns
	// * valid line data and non-nil error
	if err != nil {
		return nil, err
	}

	req := &Request{}
	req.Method, req.RequestTarget, req.HTTPVersion, err = parseRequestLine(line)
	if err != nil {
		return nil, err
	}

	req.Headers, err = ReadHeaders(r)
	if err != nil {
		return nil, NewErrorFrom("ReadHeaders() failed", err)
	}

	if err := SetRequestBodyReader(req, r); err != nil {
		return nil, NewErrorFrom("SetRequestBodyReader() failed", err)
	}

	return req, nil
}

func DumpRequest(w io.Writer, req *Request) {
	fmt.Fprintf(w, "%s %s %s\r\n", req.Method, req.RequestTarget, req.HTTPVersion)
	for _, line := range req.Headers.List() {
		fmt.Fprintf(w, "%s\r\n", line)
	}
}
