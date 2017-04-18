package main

import (
	"flag"
	"github.com/ginuerzh/pht"
	"io"
	"log"
	"os"
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
	client := pht.NewClient(host, key)
	conn, err := client.Dial()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		defer conn.Close()
		for {
			if _, err := conn.Write([]byte("Hello world! ")); err != nil {
				break
			}
			time.Sleep(1000 * time.Millisecond)
		}
	}()
	io.Copy(os.Stderr, conn)
}
