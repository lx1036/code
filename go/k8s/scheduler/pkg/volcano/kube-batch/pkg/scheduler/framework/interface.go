package framework

// Action is the interface of scheduler action.
type Action interface {
	// The unique name of Action.
	Name() string

	// Initialize initializes the allocator plugins.
	Initialize()

	// Execute allocates the cluster's resources into each queue.
	Execute(ssn *Session)

	// UnIntialize un-initializes the allocator plugins.
	UnInitialize()
}
