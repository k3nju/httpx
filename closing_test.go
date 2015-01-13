package httpx

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func testExpectEOB(readRets ...interface{}) bool {
	bb := readRets[0].([]byte)
	if bb != nil {
		return false
	}
	err := readRets[1].(error)
	if err != EOB {
		return false
	}

	return true
}

func TestClosingBasicUsage(t *testing.T) {
	// basic usage test

	// size of contents in "/etc/resolv.conf" may be
	// bigger than 0 and smaller than DefaultBodyBlockSize
	f, _ := os.Open("/etc/resolv.conf")
	defer f.Close()

	r := NewClosingReader(f)
	// "first read call" will read all of data in f.
	// but first Read() must return "len(bb)" > 0 && err == nil"
	// next Read() will return "bb == nil && err == EOB"
	bb, err := r.Read()
	if err != nil {
		t.Fatal("expected err == nil, but err ==", err)
	}
	if len(bb) > 0 {
		t.Log(string(bb))
	}

	bb, err = r.Read()
	if bb != nil {
		t.Fatal("expected bb == nil, but bb != nil")
	}
	if err != EOB {
		t.Fatal("expected err == EOB, but err == nil")
	}
}

func TestClosingNoRead(t *testing.T) {
	// test for reading no value

	f, _ := os.Open("/dev/null")
	defer f.Close()

	r := NewClosingReader(f)
	if !testExpectEOB(r.Read()) {
		t.Fatal("expected EOB, but not.")
	}
}

func TestClosingTwiceRead(t *testing.T) {
	// test for DefaultBodyBlockSize boundary test

	src0 := strings.Repeat("A", DefaultBodyBlockSize)
	buf := bytes.NewBufferString(src0)

	//
	// first
	//
	r := NewClosingReader(buf)
	bb, err := r.Read()
	if bb == nil || string(bb) != src0 {
		t.Fatal("unexpected result: bb == nil || string(bb.Data) != src0")
	}
	if err != nil {
		t.Fatal("expected err == nil, but err ==", err)
	}

	if !testExpectEOB(r.Read()) {
		t.Fatal("expected EOB, but not.")
	}

	//
	// second
	//
	src1 := src0 + "B"
	buf = bytes.NewBufferString(src1)
	r = NewClosingReader(buf)
	bb, err = r.Read()
	if bb == nil || string(bb) != src0 {
		t.Fatal("unexpected result: bb == nil || string(bb) != src0")
	}
	if err != nil {
		t.Fatal("expected err == nil, but err ==", err)
	}
	bb, err = r.Read()
	if bb == nil || string(bb) != "B" {
		t.Fatal("bb == nil || string(bb) != \"B\"")
	}
	if err != nil {
		t.Fatal("expected err == nil, but err ==", err)
	}
	if !testExpectEOB(r.Read()) {
		t.Fatal("expected EOB, but not.")
	}
}
