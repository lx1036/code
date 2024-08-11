package api

import "sync"

var IgnoredDevicesList = ignoredDevicesList{}

type ignoredDevicesList struct {
	sync.RWMutex
	ignoredDevices []string
}

func (l *ignoredDevicesList) Set(deviceLists ...[]string) {
	l.Lock()
	defer l.Unlock()
	l.ignoredDevices = l.ignoredDevices[:0]
	for _, devices := range deviceLists {
		l.ignoredDevices = append(l.ignoredDevices, devices...)
	}
}

func (l *ignoredDevicesList) Range(f func(i int, device string) bool) {
	l.RLock()
	defer l.RUnlock()
	for i, device := range l.ignoredDevices {
		if !f(i, device) {
			break
		}
	}
}
