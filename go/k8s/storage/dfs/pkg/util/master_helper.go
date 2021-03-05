package util

// MasterHelper defines the helper struct to manage the master.
type MasterHelper interface {
	AddNode(address string)
	Nodes() []string
	Leader() string
	Request(method, path string, param map[string]string, body []byte) (data []byte, err error)
}
