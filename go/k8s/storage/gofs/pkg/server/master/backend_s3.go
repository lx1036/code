package master

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"k8s.io/klog/v2"
)

//DeleteBucketInfo defines bucket info
type DeleteBucketInfo struct {
	ID         uint64
	AccessKey  string
	SecretKey  string
	Endpoint   string
	Region     string
	BucketName string
}

// NewDeleteBucketInfo creates a new bucket info for deleting
func NewDeleteBucketInfo(id uint64, accessKey, secretKey, endpoint, region, bucketName string) *DeleteBucketInfo {
	return &DeleteBucketInfo{
		ID:         id,
		AccessKey:  accessKey,
		SecretKey:  secretKey,
		Endpoint:   endpoint,
		Region:     region,
		BucketName: bucketName,
	}
}

// CreateBucket creates a new bucket in s3
func (cluster *Cluster) CreateBucket(accessKey, secretKey, endpoint, region, bucketName string) (err error) {
	credential := credentials.NewStaticCredentials(accessKey, secretKey, "")
	config := &aws.Config{
		Region:           aws.String(region),
		Endpoint:         aws.String(endpoint),
		S3ForcePathStyle: aws.Bool(true),
		Credentials:      credential,
	}
	sess := session.Must(session.NewSession(config))
	s3Client := s3.New(sess)

	_, err = s3Client.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		klog.Errorf("Failed to create bucket[%v], error: %v", bucketName, err)
		return err
	}
	err = s3Client.WaitUntilBucketExists(&s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		klog.Errorf("Failed to wait for bucket to exist %s, %v\n", bucketName, err)
		return err
	}

	return nil
}
