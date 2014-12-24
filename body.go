package httpx

import (
	"errors"
	"strconv"
)

var (
	EOB           = errors.New("end of body")
	IncompleteEOB = errors.New("incomplete body read")
)

const (
	DefaultBodyBlockSize = 8192
)

type BodyBlock struct {
	Data []byte
}

type BodyReader interface {
	Read() (*BodyBlock, error)
}

func SetRequestBodyReader(req *Request, r Reader) error {
	if vs := req.Headers.Get("transfer-encoding"); vs != nil {
		// TODO: consider handing "identity" encoding for backword compatibility
		// NOTE: identity encoding has been removed in RFC7230
		if !isChunked(vs) {
			return errors.New("encoding is not chunked")
		}

		req.Body = NewChunkedBodyReader(r)
		return nil
	}

	if vs := req.Headers.Get("content-length"); vs != nil {
		cl, err := parseContentLength(vs)
		if err != nil {
			return err
		}

		req.Body = NewContentLengthReader(r, cl)
		return nil
	}

	return nil
}

func SetResponseBodyReader(res *Response, r Reader, req *Request) error {
	if req.Method == "HEAD" {
		return nil
	}
	if (100 <= res.StatusCode && res.StatusCode <= 100) ||
		res.StatusCode == 204 ||
		res.StatusCode == 304 {
		return nil
	}

	if req.Method == "CONNECT" && res.StatusCode == 200 {
		res.Body = NewClosingReader(r)
		return nil
	}

	if vs := res.Headers.Get("transfer-encoding"); vs != nil {
		if isChunked(vs) {
			res.Body = NewChunkedBodyReader(r)
		} else {
			res.Body = NewClosingReader(r)
		}
		return nil
	}

	if vs := res.Headers.Get("content-length"); vs != nil {
		cl, err := parseContentLength(vs)
		if err != nil {
			return err
		}

		res.Body = NewContentLengthReader(r, cl)
		return nil
	}

	res.Body = NewClosingReader(r)
	return nil
}

func isChunked(values [][]byte) bool {
	var i int
	if i = len(values); i == 0 {
		return false
	}

	if string(values[i-1]) == "chunked" {
		return true
	}

	return false
}

func parseContentLength(values [][]byte) (uint64, error) {
	if len(values) != 1 {
		// multiple value
		return 0, errors.New("multiple Content-Length value found")
	}

	cl, err := strconv.ParseUint(string(values[0]), 0, 64)
	if err != nil {
		return 0, err
	}

	return cl, nil
}
