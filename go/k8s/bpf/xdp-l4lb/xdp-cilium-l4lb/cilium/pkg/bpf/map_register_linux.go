package bpf

import (
	"github.com/cilium/cilium/pkg/lock"
	log "github.com/sirupsen/logrus"
)

var (
	mutex lock.RWMutex

	// 全局变量存储所有打开的 maps
	mapRegister = map[string]*Map{}
)

func registerMap(path string, m *Map) {
	mutex.Lock()
	mapRegister[path] = m
	mutex.Unlock()

	log.WithField("path", path).Debug("Registered BPF map")
}

func unregisterMap(path string, m *Map) {
	mutex.Lock()
	delete(mapRegister, path)
	mutex.Unlock()

	log.WithField("path", path).Debug("Unregistered BPF map")
}
