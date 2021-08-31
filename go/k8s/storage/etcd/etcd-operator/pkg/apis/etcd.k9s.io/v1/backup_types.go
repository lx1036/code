package v1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// EtcdBackupList is a list of EtcdBackup.
type EtcdBackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []EtcdBackup `json:"items"`
}


// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,singular=etcdbackup
// +kubebuilder:printcolumn:name="StorageType",type="string",JSONPath=".spec.storageType"

// EtcdBackup represents a Kubernetes EtcdBackup Custom Resource.
type EtcdBackup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              BackupSpec   `json:"spec"`
	Status            BackupStatus `json:"status,omitempty"`
}



type BackupStorageType string
const (
	// AWS S3 related consts
	BackupStorageTypeS3          BackupStorageType = "S3"
)

// BackupSpec contains a backup specification for an etcd cluster.
type BackupSpec struct {
	// EtcdEndpoints specifies the endpoints of an etcd cluster.
	// When multiple endpoints are given, the backup operator retrieves
	// the backup from the endpoint that has the most up-to-date state.
	// The given endpoints must belong to the same etcd cluster.
	EtcdEndpoints []string `json:"etcdEndpoints,omitempty"`
	// StorageType is the etcd backup storage type.
	// We need this field because CRD doesn't support validation against invalid fields
	// and we cannot verify invalid backup storage source.
	StorageType BackupStorageType `json:"storageType"`
	// BackupPolicy configures the backup process.
	BackupPolicy *BackupPolicy `json:"backupPolicy,omitempty"`
	// BackupSource is the backup storage source.
	BackupSource `json:",inline"`
	// ClientTLSSecret is the secret containing the etcd TLS client certs and
	// must contain the following data items:
	// data:
	//    "etcd-client.crt": <pem-encoded-cert>
	//    "etcd-client.key": <pem-encoded-key>
	//    "etcd-client-ca.crt": <pem-encoded-ca-cert>
	ClientTLSSecret string `json:"clientTLSSecret,omitempty"`
}

// BackupPolicy defines backup policy.
type BackupPolicy struct {
	// TimeoutInSecond is the maximal allowed time in second of the entire backup process.
	TimeoutInSecond int64 `json:"timeoutInSecond,omitempty"`
	// BackupIntervalInSecond is to specify how often operator take snapshot
	// 0 is magic number to indicate one-shot backup
	BackupIntervalInSecond int64 `json:"backupIntervalInSecond,omitempty"`
	// MaxBackups is to specify how many backups we want to keep
	// 0 is magic number to indicate un-limited backups
	MaxBackups int `json:"maxBackups,omitempty"`
}

// BackupSource contains the supported backup sources.
type BackupSource struct {
	// S3 defines the S3 backup source spec.
	S3 *S3BackupSource `json:"s3,omitempty"`
}

// S3BackupSource provides the spec how to store backups on S3.
type S3BackupSource struct {
	// Path is the full s3 path where the backup is saved.
	// The format of the path must be: "<s3-bucket-name>/<path-to-backup-file>"
	// e.g: "mybucket/etcd.backup"
	Path string `json:"path"`
	
	// The name of the secret object that stores the AWS credential and config files.
	// The file name of the credential MUST be 'credentials'.
	// The file name of the config MUST be 'config'.
	// The profile to use in both files will be 'default'.
	//
	// AWSSecret overwrites the default etcd operator wide AWS credential and config.
	AWSSecret string `json:"awsSecret"`
	
	// Endpoint if blank points to aws. If specified, can point to s3 compatible object
	// stores.
	Endpoint string `json:"endpoint,omitempty"`
	
	// ForcePathStyle forces to use path style over the default subdomain style.
	// This is useful when you have an s3 compatible endpoint that doesn't support
	// subdomain buckets.
	ForcePathStyle bool `json:"forcePathStyle"`
}


// BackupStatus represents the status of the EtcdBackup Custom Resource.
type BackupStatus struct {
	// Succeeded indicates if the backup has Succeeded.
	Succeeded bool `json:"succeeded"`
	// Reason indicates the reason for any backup related failures.
	Reason string `json:"Reason,omitempty"`
	// EtcdVersion is the version of the backup etcd server.
	EtcdVersion string `json:"etcdVersion,omitempty"`
	// EtcdRevision is the revision of etcd's KV store where the backup is performed on.
	EtcdRevision int64 `json:"etcdRevision,omitempty"`
	// LastSuccessDate indicate the time to get snapshot last time
	LastSuccessDate metav1.Time `json:"lastSuccessDate,omitempty"`
}

