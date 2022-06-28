package bpf

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"k8s.io/klog/v2"
	"sync"
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

	log.WithField("path", path).Debug("Registered BPF map")
}
