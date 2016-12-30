package awslambdaproxy

import (
	"io"
	"sync"
	"log"
	"io/ioutil"
	"bytes"
	"net/http"

	"github.com/pkg/errors"
)

const (
	getIPUrl = "http://checkip.amazonaws.com/"
)

func bidirectionalCopy(dst io.ReadWriteCloser, src io.ReadWriteCloser) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		io.Copy(dst, src)
		dst.Close()
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		io.Copy(src, dst)
		src.Close()
		wg.Done()
	}()
}

func getPublicIp() (string, error) {
	log.Println("Getting public IP address..")
	resp, err := http.Get(getIPUrl)
	if err != nil {
		return "", errors.Wrap(err, "Failed to get IP address from " + getIPUrl)
	}
	defer resp.Body.Close()
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "Failed to read IP address from " + getIPUrl)
	}
	return string(bytes.TrimSpace(buf)), nil
}