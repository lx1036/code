package store

import (
	"fmt"
	"regexp"
)

const (
	keyMaxLength = 250

	keyCharFmt      string = "[A-Za-z0-9]"
	keyExtCharFmt   string = "[-A-Za-z0-9_.]"
	qualifiedKeyFmt string = "(" + keyCharFmt + keyExtCharFmt + "*)?" + keyCharFmt
)

var (
	// Key must consist of alphanumeric characters, '-', '_' or '.', and must start
	// and end with an alphanumeric character.
	keyRegex = regexp.MustCompile("^" + qualifiedKeyFmt + "$")

	// ErrKeyNotFound is the error returned if key is not found in Store.
	ErrKeyNotFound = fmt.Errorf("key is not found")
)

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

// ValidateKey returns an error if the given key does not meet the requirement
// of the key format and length.
func ValidateKey(key string) error {
	if len(key) <= keyMaxLength && keyRegex.MatchString(key) {
		return nil
	}
	return fmt.Errorf("invalid key: %q", key)
}
