package server

import (
	"io"
	"sync"

	"github.com/aws/aws-sdk-go/aws/endpoints"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

func bidirectionalCopy(src io.ReadWriteCloser, dst io.ReadWriteCloser) {
	defer dst.Close()
	defer src.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		io.Copy(dst, src)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		io.Copy(src, dst)
		wg.Done()
	}()
	wg.Wait()
}

func GetSessionAWS() (*session.Session, error) {
	sess, err := session.NewSession(aws.NewConfig())
	if err != nil {
		return nil, err
	}
	if _, err = sess.Config.Credentials.Get(); err != nil {
		return nil, err
	}
	return sess, nil
}

func GetValidLambdaRegions() []string {
	resolver := endpoints.DefaultResolver()
	partitions := resolver.(endpoints.EnumPartitions).Partitions()
	var validLambdaRegions []string
	for _, p := range partitions {
		if p.ID() == "aws" {
			for k := range p.Regions() {
				validLambdaRegions = append(validLambdaRegions, k)
			}
		}
	}
	return validLambdaRegions
}
