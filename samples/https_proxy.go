package main

import (
	. "httpx"
	. "log"
	"net"
	"sync"
)

const (
	CONNECTED = "HTTP/1.1 200 Connection established\r\n\r\n"
)

func main() {
	ln, err := net.Listen("tcp", ":8081")
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
		go proxy(c.(*net.TCPConn), i)
		i++
	}
}

func proxy(cc *net.TCPConn, no int) {
	cbconn := NewBufConn(cc)
	var s *net.TCPConn
	defer func() {
		cc.Close()

		if s != nil {
			s.Close()
		}

		Println(no, "returning proxy()")
	}()

	//
	//read request from c(client side connection)
	//
	req, err := ReadRequest(cbconn)
	if err != nil {
		Println(no, "ReadRequest() failed:", err)
		return
	}

	Println(no, req.RequestTarget, req.HTTPVersion.String())

	//
	// connect to a destination server
	//
	if sc, err := net.Dial("tcp", req.RequestTarget); err != nil {
		Println(no, "net.Dial() failed:", err)
		return
	} else {
		s = sc.(*net.TCPConn)
	}
	sbconn := NewBufConn(s)

	//
	// return 200 to a client
	//

	if n, err := cbconn.Write([]byte(CONNECTED)); true {
		if err != nil {
			Println(no, "Write() failed:", err)
			return
		}
		if n < len(CONNECTED) {
			Println(no, "Write() failed:(short write")
			return
		}
	}

	//
	// here, become a tunneling mode
	//

	var wg sync.WaitGroup
	wg.Add(2)
	go transport(no, "client to server", cbconn, sbconn, &wg)
	go transport(no, "server to client", sbconn, cbconn, &wg)

	wg.Wait()
}

func transport(no int, dir string, src, dst *BufConn, wg *sync.WaitGroup) {
	buf := make([]byte, 8192)
	for {
		rn, err := src.Read(buf)
		if err != nil {
			Println(no, dir, "Read() failed:", err)
			break
		}

		sn, err := dst.Write(buf[:rn])
		if err != nil {
			Println(no, dir, "Write() failed:", err)
			break
		}
		if sn < rn {
			Println(no, dir, "Write() failed:(short write)")
			break
		}
	}
	c := dst.C.(*net.TCPConn)
	c.CloseWrite()
	wg.Done()
}
