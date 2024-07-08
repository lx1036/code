package source

// Source describes the source of a definition
type Source string

const (
	// Unspec is used when the source is unspecified
	Unspec Source = "unspec"

	// KubeAPIServer is the source used for state which represents the
	// kube-apiserver, such as the IPs associated with it. This is not to be
	// confused with the Kubernetes source.
	// KubeAPIServer state has the strongest ownership and can only be
	// overwritten by itself.
	KubeAPIServer Source = "kube-apiserver"

	// Local is the source used for state derived from local agent state.
	// Local state has the second strongest ownership, behind KubeAPIServer.
	Local Source = "local"

	// KVStore is the source used for state derived from a key value store.
	// State in the key value stored takes precedence over orchestration
	// system state such as Kubernetes.
	KVStore Source = "kvstore"

	// Kubernetes is the source used for state derived from Kubernetes
	Kubernetes Source = "k8s"

	// CustomResource is the source used for state derived from Kubernetes
	// custom resources
	CustomResource Source = "custom-resource"

	// Generated is the source used for generated state which can be
	// overwritten by all other sources, except for restored (and unspec).
	Generated Source = "generated"

	// Restored is the source used for restored state from data left behind
	// by the previous agent instance. Can be overwritten by all other
	// sources (except for unspec).
	Restored Source = "restored"
)
