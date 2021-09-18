package mvcc

import (
	"go.etcd.io/etcd/pkg/v3/adt"
)

type watcherSet map[*watcher]struct{}

type watcherSetByKey map[string]watcherSet

// watcherGroup is a collection of watchers organized by their ranges
type watcherGroup struct {
	// keyWatchers has the watchers that watch on a single key
	keyWatchers watcherSetByKey

	// INFO: 红黑树 ranges has the watchers that watch a range; it is sorted by interval
	ranges adt.IntervalTree

	// watchers is the set of all watchers
	watchers watcherSet
}

// size gives the number of unique watchers in the group.
func (wg *watcherGroup) size() int {
	return len(wg.watchers)
}
