package container

// SyncAction indicates different kind of actions in SyncPod() and KillPod(). Now there are only actions
// about start/kill container and setup/teardown network.
type SyncAction string

const (
	// StartContainer action
	StartContainer SyncAction = "StartContainer"
	// KillContainer action
	KillContainer SyncAction = "KillContainer"
	// SetupNetwork action
	SetupNetwork SyncAction = "SetupNetwork"
	// TeardownNetwork action
	TeardownNetwork SyncAction = "TeardownNetwork"
	// InitContainer action
	InitContainer SyncAction = "InitContainer"
	// CreatePodSandbox action
	CreatePodSandbox SyncAction = "CreatePodSandbox"
	// ConfigPodSandbox action
	ConfigPodSandbox SyncAction = "ConfigPodSandbox"
	// KillPodSandbox action
	KillPodSandbox SyncAction = "KillPodSandbox"
)

// SyncResult is the result of sync action.
type SyncResult struct {
	// The associated action of the result
	Action SyncAction
	// The target of the action, now the target can only be:
	//  * Container: Target should be container name
	//  * Network: Target is useless now, we just set it as pod full name now
	Target interface{}
	// Brief error reason
	Error error
	// Human readable error reason
	Message string
}

// PodSyncResult is the summary result of SyncPod() and KillPod()
type PodSyncResult struct {
	// Result of different sync actions
	SyncResults []*SyncResult
	// Error encountered in SyncPod() and KillPod() that is not already included in SyncResults
	SyncError error
}
