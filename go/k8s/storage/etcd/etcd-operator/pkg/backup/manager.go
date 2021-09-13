package backup

import (
	"context"
	"fmt"
	"strings"
	"time"

	"k8s-lx1036/k8s/storage/etcd/etcd-operator/pkg/controller/backup/writer"

	clientv3 "go.etcd.io/etcd/client/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	DefaultDialTimeout = 5 * time.Second
)

type BackupManager struct {
	endpoints []string

	writer writer.Writer
}

// SaveSnap uses backup writer to save etcd snapshot to a specified S3 path
// and returns backup etcd server's kv store revision and its version.
func (backupManager *BackupManager) SaveSnapshot(ctx context.Context, s3Path string, isPeriodic bool) (int64, string, *metav1.Time, error) {
	now := time.Now()
	etcdClient, maxRevision, err := backupManager.getEtcdClientWithMaxRevision(ctx)
	if err != nil {

	}
	defer etcdClient.Close()

	statusResponse, err := etcdClient.Status(ctx, etcdClient.Endpoints()[0])
	if err != nil {

	}

	snapshotReadCloser, err := etcdClient.Snapshot(ctx)
	if err != nil {

	}
	defer snapshotReadCloser.Close()

	if isPeriodic {
		s3Path = fmt.Sprintf("%s_v%d_%s", s3Path, maxRevision, now.Format("2006_01_02_15_04_05"))
	}
	_, err = backupManager.writer.Write(ctx, s3Path, snapshotReadCloser)
	if err != nil {

	}

	return maxRevision, statusResponse.Version, &metav1.Time{Time: now}, nil
}

// INFO: 这里获取 max revision 的那个 etcd client，然后让那个 etcd 去做备份!!!
// getEtcdClientWithMaxRevision gets the etcd endpoint with the maximum kv store revision
// and returns the etcd client of that member.
func (backupManager *BackupManager) getEtcdClientWithMaxRevision(ctx context.Context) (*clientv3.Client, int64, error) {
	var errors []string
	maxRevision := int64(0)
	var maxRevisionClient *clientv3.Client
	var etcdClients []*clientv3.Client
	for _, endpoint := range backupManager.endpoints {
		config := clientv3.Config{
			Endpoints:   []string{endpoint},
			DialTimeout: DefaultDialTimeout,
			TLS:         nil,
		}

		etcdClient, err := clientv3.New(config)
		if err != nil {
			errors = append(errors, err.Error())
			continue
		}

		etcdClients = append(etcdClients, etcdClient)

		response, err := etcdClient.Get(ctx, "/", clientv3.WithSerializable())
		if err != nil {
			errors = append(errors, err.Error())
			continue
		}

		if response.Header.Revision > maxRevision {
			maxRevision = response.Header.Revision
			maxRevisionClient = etcdClient
		}
	}

	// INFO: 需要关闭已经开启连接的 etcd client
	for _, etcdClient := range etcdClients {
		if etcdClient == maxRevisionClient {
			continue
		}

		etcdClient.Close()
	}

	return maxRevisionClient, maxRevision, fmt.Errorf(fmt.Sprintf("[etcdClientWithMaxRevision]err: %s", strings.Join(errors, " ")))
}
