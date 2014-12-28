package main

import (
	"bytes"
	"fmt"
	"httpx"
	"io"
	"log"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
)

var printStdout = false
var crlf = []byte("\r\n")

func dprint(v ...interface{}) {
	fmt.Println(v...)
}

func main() {
	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalln(err)
	}
	dprint("listening...")

	for {
		c, err := l.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		dprint("accepted", c.RemoteAddr())
		cc := &clientConn{httpx.NewBufConn(c)}
		go handle(cc)
	}
}

type clientConn struct {
	*httpx.BufConn
}

func (cc *clientConn) Close() {
	cc.C.Close()
}

type serverConn struct {
	*httpx.BufConn
	daddr string
}

func (sc *serverConn) Close() {
	sc.C.Close()
}

func handle(cc *clientConn) {
	var sc *serverConn

	defer func() {
		dprint("exiting handle")
		cc.Close()
		if sc != nil {
			sc.Close()
		}
	}()

	for {
		dprint("reading request")
		req, err := httpx.ReadRequest(cc.BufConn)
		if err != nil {
			if err != io.EOF {
				log.Println(err)
			}
			break
		}

		dprint("raw request target:", req.RequestTarget)
		daddr, rt, err := parseRequestTarget(req.RequestTarget)
		if err != nil {
			log.Println(err)
			break
		}
		req.RequestTarget = rt
		dprint("proxy request target:", daddr, req.RequestTarget)

		if sc == nil || sc.daddr != daddr {
			if sc != nil && sc.daddr != daddr {
				dprint("closing from", sc.daddr, "connecting to", daddr)
				sc.Close()
				sc = nil
			}

			sc, err = connect(daddr)
			if err != nil {
				log.Println(err)
				break
			}
			dprint("connected to", sc.daddr)
		}

		dprint("writing request")
		if err := writeRequest(sc, req); err != nil {
			log.Println(err)
			break
		}

		dprint("reading response")
		res, err := httpx.ReadResponse(sc.BufConn, req)
		if err != nil {
			log.Println(err)
			break
		}

		dprint("writing resposne")
		if err := writeResponse(cc, res); err != nil {
			log.Println(err)
			break
		}

		if !isPersist(res.HTTPVersion, res.Headers) {
			dprint("server side is not persistent connection. closing")
			sc.Close()
			sc = nil
		}
		if !isPersist(req.HTTPVersion, req.Headers) {
			dprint("client side is not persistent. closing")
			break
		}
	}
}

func writeResponse(w io.Writer, res *httpx.Response) error {
	if printStdout {
		w = io.MultiWriter(os.Stdout, w)
	}

	tmp := &bytes.Buffer{}
	fmt.Fprintf(tmp, "%s %d %s\r\n", res.HTTPVersion, res.StatusCode, res.ReasonPhrase)
	if err := writeHeaders(tmp, res.Headers); err != nil {
		return err
	}
	if err := write(w, tmp.Bytes()); err != nil {
		return err
	}
	if res.Body == nil {
		return nil
	}

	return writeBody(w, res.Body)
}

func writeRequest(w io.Writer, req *httpx.Request) error {
	if printStdout {
		w = io.MultiWriter(os.Stdout, w)
	}

	tmp := &bytes.Buffer{}
	fmt.Fprintf(tmp, "%s %s %s\r\n", req.Method, req.RequestTarget, req.HTTPVersion)
	if err := writeHeaders(tmp, req.Headers); err != nil {
		return err
	}
	if err := write(w, tmp.Bytes()); err != nil {
		return err
	}
	if req.Body == nil {
		return nil
	}

	return writeBody(w, req.Body)
}

func connect(daddr string) (*serverConn, error) {
	s, err := net.Dial("tcp", daddr)
	if err != nil {
		return nil, err
	}

	return &serverConn{
		BufConn: httpx.NewBufConn(s),
		daddr:   daddr,
	}, nil
}

func parseRequestTarget(rt string) (string, string, error) {
	u, err := url.ParseRequestURI(rt)
	if err != nil {
		return "", "", err
	}

	host := u.Host
	if !hasPort(host) {
		host = host + ":80"
	}

	uri := u.RequestURI()

	return host, uri, nil
}

func hasPort(host string) bool {
	i := strings.LastIndex(host, ":")
	if i < 0 {
		return false
	}
	if _, err := strconv.ParseUint(host[i+1:], 0, 16); err != nil {
		return false
	}
	return true
}

func write(w io.Writer, buf ...[]byte) error {
	for _, b := range buf {
		for len(b) > 0 {
			n, err := w.Write(b)
			if n > 0 {
				b = b[n:]
			}
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func writeHeaders(tmp *bytes.Buffer, headers *httpx.Headers) error {
	h := bytes.Join(headers.List(), []byte("\r\n"))
	if err := write(tmp, h, crlf, crlf); err != nil {
		return err
	}

	return nil
}

func writeBody(w io.Writer, br httpx.BodyReader) error {
	for {
		bb, err := br.Read()
		if bb != nil {
			if err := write(w, bb.Data); err != nil {
				return err
			}
		}
		if err != nil {
			if err != io.EOF && err != httpx.EOB {
				return err
			}
			break
		}
	}

	cb, ok := br.(*httpx.ChunkedBodyReader)
	if !ok || cb.Trailers == nil {
		return nil
	}

	tmp := &bytes.Buffer{}
	if err := writeHeaders(tmp, cb.Trailers); err != nil {
		return err
	}
	if err := write(w, tmp.Bytes()); err != nil {
		return err
	}

	return nil
}

func isPersist(v *httpx.HTTPVersion, headers *httpx.Headers) bool {
	for _, v := range headers.Get("connection") {
		if strings.ToLower(string(v)) == "close" {
			return false
		}
	}

	if v.Major >= 1 && v.Minor >= 1 {
		return false
	}

	// work around
	return false
}
