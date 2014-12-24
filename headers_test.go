package httpx

import (
	"strings"
	"testing"
)

var src = strings.Replace(`Accept-Ranges: bytes
Cache-Control: max-age=604800
Content-Type: text/html
Date: Tue, 23 Dec 2014 21:26:34 GMT
Etag: "359670651"
Expires: Tue, 30 Dec 2014 21:26:34 GMT
Last-Modified: Fri, 09 Aug 2013 23:54:35 GMT
Server: ECS (rhv/818F)
X-Cache: HIT
x-ec-custom-error: 1
Content-Length: 1270

`, "\n", "\r\n", -1)

func TestUsage(t *testing.T) {
	s := newStringLineReader(src)

	// create headers
	h, err := ReadHeaders(s)
	if err != nil {
		t.Fatal(err)
	}
	_ = h

	// List() lists all fields
	a := strings.Split(src, "\r\n")
	for i, v := range h.List() {
		if a[i] != string(v) {
			t.Fatalf("expected %s, got %s", a[i], string(v))
		}
	}

	// Get() gets field by name
	vs := h.Get("content-length")
	if len(vs) != 1 {
		t.Fatalf("len(vs) != 1")
	}
	if string(vs[0]) != "1270" {
		t.Fatal("expected 1270", "got", string(vs[0]))
	}

	// Del() deletes field by name
	h.Del("X-Cache")
	if h.Get("x-CACHE") != nil {
		t.Fatal("expected X-Cache key deleted, but exists")
	}

	// Set() sets new field
	h.Set("X-TestName", []byte("hoge"))
	if string(h.Get("x-testname")[0]) != "hoge" {
		t.Fatal("expected X-TestName key been set, but not exists")
	}

	for _, l := range h.List() {
		t.Log(string(l))
	}
}
