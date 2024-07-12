package framework

import "sync"

// Plugin is the interface of scheduler plugin
type Plugin interface {
	Name() string

	OnSessionOpen(ssn *Session)
	OnSessionClose(ssn *Session)
}

var pluginMutex sync.RWMutex

type Arguments map[string]interface{}
type PluginBuilder = func(Arguments) Plugin

var pluginBuilders = map[string]PluginBuilder{}

func RegisterPluginBuilder(name string, pc PluginBuilder) {
	pluginMutex.Lock()
	defer pluginMutex.Unlock()

	pluginBuilders[name] = pc
}

func GetPluginBuilder(name string) (PluginBuilder, bool) {
	pluginMutex.RLock()
	defer pluginMutex.RUnlock()

	pb, found := pluginBuilders[name]
	return pb, found
}

var actionMap = map[string]Action{}

func RegisterAction(act Action) {
	pluginMutex.Lock()
	defer pluginMutex.Unlock()

	actionMap[act.Name()] = act
}

func GetAction(name string) (Action, bool) {
	pluginMutex.RLock()
	defer pluginMutex.RUnlock()

	act, found := actionMap[name]
	return act, found
}
