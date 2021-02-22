package cache

type PluginHandler interface {
	// Validate returns an error if the information provided by
	// the potential plugin is erroneous (unsupported version, ...)
	ValidatePlugin(pluginName string, endpoint string, versions []string) error
	// RegisterPlugin is called so that the plugin can be register by any
	// plugin consumer
	// Error encountered here can still be Notified to the plugin.
	RegisterPlugin(pluginName, endpoint string, versions []string) error
	// DeRegister is called once the pluginwatcher observes that the socket has
	// been deleted.
	DeRegisterPlugin(pluginName string)
}
