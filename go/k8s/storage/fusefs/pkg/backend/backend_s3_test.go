package backend

import (
	"flag"
	"fmt"
	"k8s.io/klog/v2"
	"testing"
)

var (
	endpoint  = flag.String("endpoint", "", "")
	accessKey = flag.String("ak", "", "")
	secretKey = flag.String("sk", "", "")
	bucket    = flag.String("bucket", "", "")
	file      = flag.String("file", "", "")
)

func TestReadFile(test *testing.T) {
	flag.Parse()

	if len(*endpoint) == 0 || len(*accessKey) == 0 || len(*secretKey) == 0 || len(*bucket) == 0 {
		return
	}

	s3Config := &S3Config{
		Region:           "us-east-1",
		Endpoint:         *endpoint,
		AccessKey:        *accessKey,
		SecretKey:        *secretKey,
		S3ForcePathStyle: true, // 必须为 true
	}

	s3Backend, err := NewS3Backend(*bucket, s3Config)
	if err != nil {
		klog.Fatal(err)
	}

	var data []byte
	size, err := s3Backend.Read(*file, 0, data)
	if err != nil {
		klog.Fatal(err)
	}
	klog.Infof(fmt.Sprintf("data %s, size %d", data, size))
}
