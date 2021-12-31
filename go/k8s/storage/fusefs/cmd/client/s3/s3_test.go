package s3

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

// INFO: aws s3api get-object --key=2 --bucket pvc-f73f7c99-0b5c-40ee-b57c-acdebcebed34 --endpoint-url ${endpoint-url} --range bytes=1-100 2.txt
//  => "asdfadfasdfasdfasdf"
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

	data := make([]byte, 100)
	size, err := s3Backend.Read(*file, 1, data)
	if err != nil {
		klog.Fatal(err)
	}
	klog.Infof(fmt.Sprintf("data %s, size %d", data, size))
}
