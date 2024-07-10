package framework

type Action interface {
	Name() string

	// Initialize initializes the allocator plugins.
	Initialize()

	// Execute allocates the cluster's resources into each queue.
	Execute(ssn *Session)

	// UnInitialize the allocator plugins.
	UnInitialize()
}
