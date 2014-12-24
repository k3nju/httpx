package main

import (
	"fmt"
	. "httpx"
	"io"
	. "log"
	"net"
	"net/url"
	"strings"
)

func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		Fatalln(err)
	}

	i := 0
	for {
		c, err := ln.Accept()
		if err != nil {
			Println(err)
			continue
		}
		go proxy(c, i)
		i++
	}
}

func proxy(cconn net.Conn, no int) {
	var sconn net.Conn
	var s ReadWriter
	defer func() {
		if cconn != nil {
			cconn.Close()
		}
		if sconn != nil {
			sconn.Close()
		}
		Println(no, "returning proxy()")
	}()

	var prevRT string

	c := NewBufConn(cconn)

	for {
		//
		//read request from c(client side connection)
		//
		req, err := ReadRequest(c)
		if err != nil {
			Println(no, "ReadRequest() failed:", err)
			return
		}

		Println(no, req.RequestTarget, req.HTTPVersion.String())

		//
		// connect to a destination server specified by RequestTarget if one of followings is true.
		// * s(server side connection) == nil
		// * or prevRT != req.RequestTarget
		//
		if sconn == nil || prevRT != req.RequestTarget {
			// if sconn != nil && prevRT != req.RequestTarget, which means
			//  client is going to connect a server which is not same previous request.
			if sconn != nil && prevRT != req.RequestTarget {
				sconn.Close()
				sconn = nil
			}

			u, err := url.ParseRequestURI(req.RequestTarget)
			if err != nil {
				Println(no, "url.ParseRequestURI() failed:", err)
				return
			}

			sconn, err = net.Dial("tcp", u.Host+":80")
			if err != nil {
				Println(no, "net.Dial() failed:", err)
				return
			}
			s = NewBufConn(sconn)
			prevRT = req.RequestTarget
			req.RequestTarget = u.RequestURI() // convert request target from absolute-form to origin-form
		}

		//
		// write request to a server
		//
		if err := WriteRequest(s, req); err != nil {
			Println(no, "WriteRequest() failed:", err)
			return
		}

		//
		// read response from a server
		//
		res, err := ReadResponse(s, req)
		if err != nil {
			Println(no, "ReadResponse() failed:", err)
			return
		}

		// write response to a client
		if err := WriteResponse(c, res); err != nil {
			Println(no, "WriteResponse() failed:", err)
			return
		}

		// here, one request-response intermediary finished.
		if shouldClose(res.Headers.Get("Connection"), res.HTTPVersion) {
			Println(no, "closing server side connection")
			sconn.Close()
			sconn = nil
		}
		if shouldClose(req.Headers.Get("Connectioni"), req.HTTPVersion) {
			// TODO: to be gracefull
			Println(no, "closing client side connection")
			cconn.Close()
			cconn = nil
			break
		}
	}
}

func WriteRequest(w io.Writer, req *Request) error {
	// TODO: consider more efficient method
	_, err := fmt.Fprintf(w, "%s %s %s\r\n", req.Method, req.RequestTarget, req.HTTPVersion)
	if err != nil {
		return err
	}

	crlf := []byte("\r\n")
	for _, line := range req.Headers.List() {
		_, err := writeAll(w, line, crlf)
		if err != nil {
			return err
		}
	}

	_, err = w.Write(crlf)
	if err != nil {
		return err
	}

	if req.Body == nil {
		return nil
	}

	for {
		bb, err := req.Body.Read()
		if bb != nil {
			_, err := w.Write(bb.Data)
			if err != nil {
				return err
			}
		}
		if err != nil {
			if err != io.EOF && err != EOB {
				return err
			}
		}
	}

	return nil
}

func WriteResponse(w io.Writer, res *Response) error {
	_, err := fmt.Fprintf(w, "%s %d %s\r\n", res.HTTPVersion, res.StatusCode, res.ReasonPhrase)
	if err != nil {
		Println("StatusLine")
		return err
	}

	crlf := []byte("\r\n")
	for _, line := range res.Headers.List() {
		_, err := writeAll(w, line, crlf)
		if err != nil {
			Println(err)
			return err
		}
	}

	if _, err := w.Write(crlf); err != nil {
		Println("last of headers")
		return err
	}

	for {
		bb, err := res.Body.Read()
		if bb != nil {
			n, err := w.Write(bb.Data)
			if n != len(bb.Data) {
				panic("n != len(bb.Data)")
			}
			if err != nil {
				Println("write body")
				return err
			}
		}
		if err != nil {
			if err != io.EOF && err != EOB {
				Println("ReadBody")
				return err
			}
			break
		}
	}

	if cbb, ok := res.Body.(*ChunkedBodyReader); ok {
		for _, line := range cbb.Trailers.List() {
			if _, err := writeAll(w, line, crlf); err != nil {
				return err
			}
		}

		if _, err := w.Write(crlf); err != nil {
			return err
		}
	}

	return nil
}

func writeAll(w io.Writer, bs ...[]byte) (int64, error) {
	var t int64

	for _, b := range bs {
		for len(b) > 0 {
			n, err := w.Write(b)
			if n > 0 {
				b = b[n:]
				t += int64(n)
			}
			if err != nil {
				return t, err
			}
		}
	}

	return t, nil
}

func shouldClose(conn [][]byte, hv *HTTPVersion) bool {
	if len(conn) > 0 {
		for _, v := range conn {
			if strings.ToLower(string(v)) == "close" {
				return true
			}
		}
	}
	if hv.Major == 1 && hv.Minor == 1 {
		return false
	}

	// client uses HTTP/1.0
	return true
}
