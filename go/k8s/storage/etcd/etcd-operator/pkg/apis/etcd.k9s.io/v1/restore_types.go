package v1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// EtcdRestoreList is a list of EtcdRestore.
type EtcdRestoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []EtcdRestore `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,singular=etcdrestore
// +kubebuilder:printcolumn:name="BackupStorageType",type="string",JSONPath=".spec.backupStorageType"
// +kubebuilder:printcolumn:name="EtcdClusterRef",type="string",JSONPath=".spec.etcdCluster.name"

// EtcdRestore represents a Kubernetes EtcdRestore Custom Resource.
// The EtcdRestore CR name will be used as the name of the new restored cluster.
type EtcdRestore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              RestoreSpec   `json:"spec"`
	Status            RestoreStatus `json:"status,omitempty"`
}

// RestoreSpec defines how to restore an etcd cluster from existing backup.
type RestoreSpec struct {
	// BackupStorageType is the type of the backup storage which is used as RestoreSource.
	BackupStorageType BackupStorageType `json:"backupStorageType"`

	// RestoreSource tells the where to get the backup and restore from.
	RestoreSource `json:",inline"`

	// EtcdCluster references an EtcdCluster resource whose metadata and spec
	// will be used to create the new restored EtcdCluster CR.
	// This reference EtcdCluster CR and all its resources will be deleted before the
	// restored EtcdCluster CR is created.
	EtcdCluster EtcdClusterRef `json:"etcdCluster"`
}

// EtcdClusterRef references an EtcdCluster resource whose metadata and spec
// will be used to create the new restored EtcdCluster CR.
// This reference EtcdCluster CR and all its resources will be deleted before the
// restored EtcdCluster CR is created.
type EtcdClusterRef struct {
	// Name is the EtcdCluster resource name.
	// This reference EtcdCluster must be present in the same namespace as the restore-operator
	Name string `json:"name"`
}

type RestoreSource struct {
	// S3 tells where on S3 the backup is saved and how to fetch the backup.
	S3 *S3RestoreSource `json:"s3,omitempty"`
}

type S3RestoreSource struct {
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
	Endpoint string `json:"endpoint"`

	// ForcePathStyle forces to use path style over the default subdomain style.
	// This is useful when you have an s3 compatible endpoint that doesn't support
	// subdomain buckets.
	ForcePathStyle bool `json:"forcePathStyle"`
}

// RestoreStatus reports the status of this restore operation.
type RestoreStatus struct {
	// Succeeded indicates if the backup has Succeeded.
	Succeeded bool `json:"succeeded"`
	// Reason indicates the reason for any backup related failures.
	Reason string `json:"reason,omitempty"`
}
