package main

import (
	"github.com/ginuerzh/gosocks5"
	"io"
	"log"
	"net"
)

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	server := &gosocks5.Server{
		Addr:   ":9999",
		Handle: handle,
	}

	server.ListenAndServe()
}

func handle(conn net.Conn, method uint8) error {
	defer conn.Close()

	req, err := gosocks5.ReadRequest(conn)
	if err != nil {
		log.Println(err)
		return err
	}
	tconn, err := Connect(req.Addr.String())
	if err != nil {
		log.Println(err)
		return err
	}
	defer tconn.Close()

	rep := gosocks5.NewReply(gosocks5.Succeeded, nil)
	if err := rep.Write(conn); err != nil {
		return err
	}

	if err := Transport(conn, tconn); err != nil {
		log.Println(err)
	}

	return nil
}

func Connect(addr string) (net.Conn, error) {
	taddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return net.DialTCP("tcp", nil, taddr)
}

func Copy(dst io.Writer, src io.Reader) (written int64, err error) {
	buf := make([]byte, 32*1024)
	for {
		nr, er := src.Read(buf)
		//log.Println("cp r", nr, er)
		if nr > 0 {
			nw, ew := dst.Write(buf[:nr])
			//log.Println("cp w", nw, ew)
			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				err = ew
				break
			}
			/*
				if nr != nw {
					err = io.ErrShortWrite
					break
				}
			*/
		}
		if er == io.EOF {
			break
		}
		if er != nil {
			err = er
			break
		}
	}
	return
}

func Pipe(src io.Reader, dst io.Writer, c chan<- error) {
	_, err := Copy(dst, src)
	c <- err
}

func Transport(conn, conn2 net.Conn) (err error) {
	rChan := make(chan error, 1)
	wChan := make(chan error, 1)

	go Pipe(conn, conn2, wChan)
	go Pipe(conn2, conn, rChan)

	select {
	case err = <-wChan:
		//log.Println("w exit", err)
	case err = <-rChan:
		//log.Println("r exit", err)
	}

	return
}
