package configs

import "os"

const (
	Wildcard = -1
)

type DeviceType rune

const (
	WildcardDevice DeviceType = 'a'
	BlockDevice    DeviceType = 'b'
	CharDevice     DeviceType = 'c' // or 'u'
	FifoDevice     DeviceType = 'p'
)

// DevicePermissions is a cgroupv1-style string to represent device access. It
// has to be a string for backward compatibility reasons, hence why it has
// methods to do set operations.
type DevicePermissions string

type DeviceRule struct {
	// Type of device ('c' for char, 'b' for block). If set to 'a', this rule
	// acts as a wildcard and all fields other than Allow are ignored.
	Type DeviceType `json:"type"`

	// Major is the device's major number.
	Major int64 `json:"major"`

	// Minor is the device's minor number.
	Minor int64 `json:"minor"`

	// Permissions is the set of permissions that this rule applies to (in the
	// cgroupv1 format -- any combination of "rwm").
	Permissions DevicePermissions `json:"permissions"`

	// Allow specifies whether this rule is allowed.
	Allow bool `json:"allow"`
}

type Device struct {
	DeviceRule

	// Path to the device.
	Path string `json:"path"`

	// FileMode permission bits for the device.
	FileMode os.FileMode `json:"file_mode"`

	// Uid of the device.
	Uid uint32 `json:"uid"`

	// Gid of the device.
	Gid uint32 `json:"gid"`
}
