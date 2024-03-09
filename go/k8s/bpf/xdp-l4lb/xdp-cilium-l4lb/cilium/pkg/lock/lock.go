package lock

// RWMutex is equivalent to sync.RWMutex but applies deadlock detection if the
// built tag "lockdebug" is set
type RWMutex struct {
	internalRWMutex
}
