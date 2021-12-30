package client

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"syscall"
	"time"

	"k8s-lx1036/k8s/storage/fuse"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"k8s.io/klog/v2"
)

const (
	DefaultMinPartSize = 5 << 20
	DefaultMaxParts    = 60
)

var defaultHTTPTransport = http.Transport{
	Proxy: http.ProxyFromEnvironment,
	DialContext: (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext,
	MaxIdleConns:          1000,
	MaxIdleConnsPerHost:   1000,
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 10 * time.Second,
}

type S3Config struct {
	// Common Backend Config
	Region              string
	Endpoint            string
	AccessKey           string
	SecretKey           string
	Version             string
	DisableSSL          bool
	S3ForcePathStyle    bool
	MergeIOVector       bool
	HTTPTimeout         time.Duration
	NoParallelMultipart bool
}

type MultipartCommitInput struct {
	sync.Mutex
	wg sync.WaitGroup

	Key        *string
	Metadata   map[string]*string
	UploadId   *string
	Parts      []*string
	NumParts   int
	NextOffset int64
	Offset     int64
}

type S3Client struct {
	*s3.S3

	bucket string
	agent  string

	awsConfig *aws.Config
	session   *session.Session

	maxParts    int
	minPartSize int

	uploads map[string]*MultipartCommitInput

	mergeIoVector bool

	cap Capabilities
}

func (s3Client *S3Client) Write(file string, offset int64, data []byte) (wsize int, err error) {
	panic("implement me")
}

func (s3Client *S3Client) WriteStream(file string, offset int64, length int64, reader io.ReadSeeker) (wsize int, err error) {
	panic("implement me")
}

func (s3Client *S3Client) ReadStream(file string, offset int64, length int64, writer io.Writer) (rsize int, err error) {
	panic("implement me")
}

func (s3Client *S3Client) WriteStreamWithCallBack(file string, offset int64, length int64, reader io.ReadSeeker, cb IOCallback) {
	panic("implement me")
}

func (s3Client *S3Client) ReadStreamWithCallBack(file string, offset int64, length int64, writer io.Writer, cb IOCallback) {
	panic("implement me")
}

func (s3Client *S3Client) WriteV(file string, vec *IOVector, reader io.ReadSeeker) (wsize int, err error) {
	panic("implement me")
}

func (s3Client *S3Client) ReadV(file string, vec *IOVector, writer io.Writer) (rsize int, err error) {
	panic("implement me")
}

func (s3Client *S3Client) Truncate(file string, offset int64) (err error) {
	panic("implement me")
}

func (s3Client *S3Client) Fallocate(file string, op int, off int64, len int64) (err error) {
	panic("implement me")
}

func (s3Client *S3Client) Flush(file string) (err error) {
	panic("implement me")
}

func (s3Client *S3Client) Rename(src string, dst string, dir bool) (err error) {
	panic("implement me")
}

func (s3Client *S3Client) Delete(file string) (err error) {
	panic("implement me")
}

func (s3Client *S3Client) Deletes(files []string) (err error) {
	panic("implement me")
}

func (s3Client *S3Client) SupportCallBack() bool {
	panic("implement me")
}

// INFO: 从 file 文件读取数据写到 data
//  aws s3api get-object --key=2 --bucket pvc-f73f7c99-0b5c-40ee-b57c-acdebcebed34 --endpoint-url ${endpoint-url} --range bytes=1-100 2.txt
//  => "asdfadfasdfasdfasdf"
func (s3Client *S3Client) Read(file string, offset int64, data []byte) (int, error) {
	rNeed := len(data)
	end := offset + int64(rNeed) - 1
	bytes := fmt.Sprintf("bytes=%v-%v", offset, end)
	input := &s3.GetObjectInput{
		Bucket: aws.String(s3Client.bucket),
		Key:    aws.String(file),
	}
	input.Range = &bytes
	klog.Infof(fmt.Sprintf("[Read]send input %s to s3", input.String()))
	reader, err := s3Client.getObject(input)
	if err != nil {
		return 0, err
	}
	defer reader.Close()

	n, err := io.ReadFull(reader, data)
	if err != nil {
		if err != io.EOF {
			return 0, err
		}
	}

	return n, nil
}

func (s3Client *S3Client) getObject(input *s3.GetObjectInput) (io.ReadCloser, error) {
	req, resp := s3Client.GetObjectRequest(input)
	req.HTTPRequest.Header.Add("User-Agent", s3Client.agent)
	err := req.Send()
	if err != nil {
		return nil, mapAwsError(err)
	}

	return resp.Body, nil
}

func (s3Client *S3Client) AuthBucket() error {
	input := &s3.HeadBucketInput{
		Bucket: aws.String(s3Client.bucket),
	}

	bucketOutput, err := s3Client.HeadBucket(input)
	if err != nil {
		return err
	}

	klog.Infof(fmt.Sprintf("[AuthBucket]bucketOutput %s", bucketOutput.String()))

	return nil
}

func NewS3Backend(bucket string, cfg *S3Config) (*S3Client, error) {
	credential := credentials.NewStaticCredentials(cfg.AccessKey, cfg.SecretKey, "")
	awsConfig := (&aws.Config{
		Region:           &cfg.Region,
		Endpoint:         &cfg.Endpoint,
		DisableSSL:       &cfg.DisableSSL,
		S3ForcePathStyle: &cfg.S3ForcePathStyle,
		Credentials:      credential,
	}).WithHTTPClient(&http.Client{
		Transport: &defaultHTTPTransport,
		Timeout:   cfg.HTTPTimeout,
	})
	s3Backend := &S3Client{
		bucket:      bucket,
		awsConfig:   awsConfig,
		minPartSize: DefaultMinPartSize,
		maxParts:    DefaultMaxParts,
		uploads:     make(map[string]*MultipartCommitInput),
	}
	s3Backend.mergeIoVector = cfg.MergeIOVector
	s3Backend.agent = fmt.Sprintf("fuseFS/%v", cfg.Version)

	// create new session
	var err error
	s3Backend.session, err = session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		return nil, err
	}
	s3Backend.S3 = s3.New(s3Backend.session, s3Backend.awsConfig)
	err = s3Backend.AuthBucket()
	if err != nil {
		klog.Infof(fmt.Sprintf("[NewS3Backend]auth bucket %s failed with err %v", s3Backend.bucket, err))
		return nil, err
	}

	//s3Backend.replicators = util.Ticket{Total: 128}.Init()

	return s3Backend, nil
}

func mapAwsError(err error) error {
	if err == nil {
		return nil
	}
	if awsErr, ok := err.(awserr.Error); ok {
		switch awsErr.Code() {
		case "BucketRegionError":
			return err
		case "NoSuchBucket":
			return syscall.ENXIO
		case "BucketAlreadyOwnedByYou":
			return fuse.EEXIST
		case "NoSuchKey":
			return fuse.ENOENT
		}
		if reqErr, ok := err.(awserr.RequestFailure); ok {
			err = mapHttpError(reqErr.StatusCode())
			if err != nil {
				return err
			} else {
				klog.Errorf("http=%v %v s3=%v request=%v\n", reqErr.StatusCode(), reqErr.Message(), awsErr.Code(), reqErr.RequestID())
				return reqErr
			}
		} else {
			klog.Errorf("code=%v msg=%v, err=%v\n", awsErr.Code(), awsErr.Message(), awsErr.OrigErr())
			return awsErr
		}
	} else {
		return err
	}
}

func mapHttpError(status int) error {
	switch status {
	case 400:
		return fuse.EINVAL
	case 401:
		return syscall.EACCES
	case 403:
		return syscall.EACCES
	case 404:
		return fuse.ENOENT
	case 405:
		return syscall.ENOTSUP
	case 429:
		return syscall.EAGAIN
	case 500:
		return syscall.EAGAIN
	default:
		return nil
	}
}
