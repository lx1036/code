package controller

// Controller interface
type Controller interface {
	// Run method
	Run(workers int, stopCh chan struct{})
}
