package httpx

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
)

var (
	ErrMalformedHTTPVersion = errors.New("malformed HTTP version")
)

type HTTPVersion struct {
	Major uint
	Minor uint
}

func (v *HTTPVersion) String() string {
	return fmt.Sprintf("HTTP/%d.%d", v.Major, v.Minor)
}

// v = []byte("HTTP/1.X")
func ParseHTTPVersion(v []byte) (*HTTPVersion, error) {
	s1 := bytes.Index(v, []byte("/"))
	if s1 < 0 {
		return nil, ErrMalformedHTTPVersion
	}
	s2 := bytes.Index(v[s1+1:], []byte("."))
	if s2 < 0 {
		return nil, ErrMalformedHTTPVersion
	}
	s2 += s1 + 1

	major, err := strconv.ParseUint(string(v[s1+1:s2]), 0, 8)
	if err != nil {
		return nil, err
	}

	minor, err := strconv.ParseUint(string(v[s2+1:]), 0, 8)
	if err != nil {
		return nil, err
	}

	return &HTTPVersion{
		Major: uint(major),
		Minor: uint(minor),
	}, nil
}
