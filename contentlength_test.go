package httpx

import (
	_ "io"
	"os"
	"testing"
)

func TestContentLengthBasicUsage(t *testing.T) {
	f, _ := os.Open("/etc/passwd")
	r := NewContentLengthReader(f, 4)

	// first Read() call will return available data and err == nil
	bb, err := r.Read()
	if !(bb != nil && len(bb) == 4 && err == nil) {
		t.Fatal("unexpected result: !(bb != nil && len(bb) == 4 && err == nil)")
	}

	// Content-Length(4) have been already read,
	// second Read() call will return bb == nil and err == EOB
	if !testExpectEOB(r.Read()) {
		t.Fatal("expected EOB, but not.")
	}
}
