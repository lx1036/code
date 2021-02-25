package util

// These constants are PVC condition types related to resize operation.
const (
	VolumeResizing           = "Resizing"
	VolumeResizeFailed       = "VolumeResizeFailed"
	VolumeResizeSuccess      = "VolumeResizeSuccessful"
	FileSystemResizeRequired = "FileSystemResizeRequired"
)

const (
	// If CSI migration is enabled, the value will be CSI driver name
	// Otherwise, it will be in-tree storage plugin name
	VolumeResizerKey = "volume.kubernetes.io/storage-resizer"
)
