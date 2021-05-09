package value

// Transformer allows a value to be transformed before being read from or written to the underlying store. The methods
// must be able to undo the transformation caused by the other.
type Transformer interface {
	// TransformFromStorage may transform the provided data from its underlying storage representation or return an error.
	// Stale is true if the object on disk is stale and a write to etcd should be issued, even if the contents of the object
	// have not changed.
	TransformFromStorage(data []byte, context Context) (out []byte, stale bool, err error)
	// TransformToStorage may transform the provided data into the appropriate form in storage or return an error.
	TransformToStorage(data []byte, context Context) (out []byte, err error)
}

// Context is additional information that a storage transformation may need to verify the data at rest.
type Context interface {
	// AuthenticatedData should return an array of bytes that describes the current value. If the value changes,
	// the transformer may report the value as unreadable or tampered. This may be nil if no such description exists
	// or is needed. For additional verification, set this to data that strongly identifies the value, such as
	// the key and creation version of the stored data.
	AuthenticatedData() []byte
}
