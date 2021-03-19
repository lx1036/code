package heap

import "k8s.io/kubernetes/pkg/scheduler/metrics"

// KeyFunc is a function type to get the key from an object.
type KeyFunc func(obj interface{}) (string, error)

type heapItem struct {
	obj   interface{} // The object which is stored in the heap.
	index int         // The index of the object's key in the Heap.queue.
}

type itemKeyValue struct {
	key string
	obj interface{}
}

// data is an internal struct that implements the standard heap interface
// and keeps the data stored in the heap.
type data struct {
	// items is a map from key of the objects to the objects and their index.
	// We depend on the property that items in the map are in the queue and vice versa.
	items map[string]*heapItem
	// queue implements a heap data structure and keeps the order of elements
	// according to the heap invariant. The queue keeps the keys of objects stored
	// in "items".
	queue []string
	
	// keyFunc is used to make the key used for queued item insertion and retrieval, and
	// should be deterministic.
	keyFunc KeyFunc
	// lessFunc is used to compare two objects in the heap.
	lessFunc lessFunc
}

// Heap is a producer/consumer queue that implements a heap data structure.
// It can be used to implement priority queues and similar data structures.
type Heap struct {
	// data stores objects and has a queue that keeps their ordering according
	// to the heap invariant.
	data *data
	// metricRecorder updates the counter when elements of a heap get added or
	// removed, and it does nothing if it's nil
	metricRecorder metrics.MetricRecorder
}

