package backup

import (
	"context"
	v1 "k8s-lx1036/k8s/storage/etcd/etcd-operator/pkg/apis/etcd.k9s.io/v1"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	AccessKey = "accessKey"
	SecretKey = "secretKey"
)

func handleS3Backup(etcdBackup *v1.EtcdBackup, kubeClient *kubernetes.Clientset) {

	s3Client, err := NewS3ClientFromSecret(etcdBackup, kubeClient)
	if err != nil {
		return
	}

}

func NewS3ClientFromSecret(etcdBackup *v1.EtcdBackup, kubeClient *kubernetes.Clientset) (*s3.S3, error) {
	secret, err := kubeClient.CoreV1().Secrets(etcdBackup.Namespace).Get(context.TODO(),
		etcdBackup.Spec.BackupSource.S3.AWSSecret, metav1.GetOptions{})
	if err != nil {
		return
	}

	accessKey, ok := secret.Data[AccessKey]
	if !ok {
		return
	}
	secretKey, ok := secret.Data[SecretKey]
	if !ok {
		return
	}

	credential := credentials.NewStaticCredentials(string(accessKey), string(secretKey), "")
	endpoint := etcdBackup.Spec.S3.Endpoint
	s3ForcePathStyle := etcdBackup.Spec.S3.ForcePathStyle
	options := &session.Options{
		Config: aws.Config{
			Endpoint:         &endpoint,
			S3ForcePathStyle: &s3ForcePathStyle,
			Credentials:      credential,
		},
		SharedConfigState: session.SharedConfigEnable,
	}
	sess, err := session.NewSessionWithOptions(*options)
	if err != nil {
		return nil, err
	}
	S3 := s3.New(sess)

	return S3, nil
}
