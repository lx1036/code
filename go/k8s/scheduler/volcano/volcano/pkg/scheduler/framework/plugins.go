package framework

import "sync"

var pluginMutex sync.RWMutex

var actionMap = map[string]Action{}

func GetAction(name string) (Action, bool) {
	pluginMutex.RLock()
	defer pluginMutex.RUnlock()

	act, found := actionMap[name]
	return act, found
}
