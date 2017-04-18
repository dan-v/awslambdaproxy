package main

import (
	"flag"
	"github.com/ginuerzh/pht"
	"log"
	"net"
	"time"
)

var (
	host string
	key  string
)

func init() {
	flag.StringVar(&host, "h", ":8080", "host address")
	flag.StringVar(&key, "key", "", "key")
	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	server := pht.Server{
		Addr:    host,
		Key:     key,
		Handler: handleConnection,
	}
	log.Fatal(server.ListenAndServe())
}

func handleConnection(conn net.Conn) {
	b := make([]byte, 1*1024)
	for {
		conn.SetDeadline(time.Now().Add(time.Second * 60))
		n, err := conn.Read(b)
		if err != nil {
			log.Println(err)
			return
		}
		_, err = conn.Write(b[:n])
		if err != nil {
			log.Println(err)
			return
		}
	}
}
