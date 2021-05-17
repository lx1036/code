package store

// Store provides the interface for storing keyed data.
// Store must be thread-safe
type Store interface {
	// key must contain one or more characters in [A-Za-z0-9]
	// Write writes data with key.
	Write(key string, data []byte) error
	// Read retrieves data with key
	// Read must return ErrKeyNotFound if key is not found.
	Read(key string) ([]byte, error)
	// Delete deletes data by key
	// Delete must not return error if key does not exist
	Delete(key string) error
	// List lists all existing keys.
	List() ([]string, error)
}
