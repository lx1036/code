package bpf

import (
	"fmt"
	"sync"

	"k8s.io/klog/v2"
)

var (
	mutex       sync.RWMutex
	mapRegister = map[string]*Map{}
)

func registerMap(path string, m *Map) {
	mutex.Lock()
	mapRegister[path] = m
	mutex.Unlock()

	klog.Infof(fmt.Sprintf("Registered BPF map name %s path %s", m.name, path))
}

func unregisterMap(path string, m *Map) {
	mutex.Lock()
	delete(mapRegister, path)
	mutex.Unlock()

	klog.Infof(fmt.Sprintf("Unregistered BPF map name %s path %s", m.name, path))
}
