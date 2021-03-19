package backend

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type S3Backend struct {
	*s3.S3
	bucket    string
	awsConfig *aws.Config
	sess      *session.Session

	cap Capabilities
}
