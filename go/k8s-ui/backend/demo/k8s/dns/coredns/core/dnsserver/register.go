package dnsserver

import (
	"fmt"
	"github.com/caddyserver/caddy"
	"github.com/caddyserver/caddy/caddyfile"
	"k8s-lx1036/k8s-ui/backend/demo/k8s/dns/coredns/plugin"
	"net"
)

type Config struct {
	Plugin      []plugin.Plugin
	ListenHosts []string
	Port        string
	Transport   string
	registry    map[string]plugin.Handler
}

func (config *Config) Handler(name string) plugin.Handler {
	if config.registry == nil {
		return nil
	}
	if handler, ok := config.registry[name]; ok {
		return handler
	}

	return nil
}

func (config *Config) Handlers() []plugin.Handler {
	var handlers []plugin.Handler
	for _, handler := range config.registry {
		handlers = append(handlers, handler)
	}

	return handlers
}

func (config *Config) RegisterHandler(handler plugin.Handler) {
	if config.registry == nil {
		config.registry = make(map[string]plugin.Handler)
	}

	config.registry[handler.Name()] = handler
}

func GetConfig(controller *caddy.Controller) *Config {
	context := controller.Context().(*dnsContext)
	key := fmt.Sprintf("%d:%d", controller.ServerBlockIndex, controller.ServerBlockKeyIndex)

	if config, ok := context.keysToConfigs[key]; ok {
		return config
	}

	context.saveConfig(key, &Config{ListenHosts: []string{""}})

	return GetConfig(controller)
}

func (config *Config) AddPlugin(plugin plugin.Plugin) {
	config.Plugin = append(config.Plugin, plugin)
}

type dnsContext struct {
	keysToConfigs map[string]*Config
	configs       []*Config
}

func (context *dnsContext) saveConfig(key string, config *Config) {
	context.configs = append(context.configs, config)
	context.keysToConfigs[key] = config
}

func (context *dnsContext) InspectServerBlocks(sourceFile string, serverBlocks []caddyfile.ServerBlock) ([]caddyfile.ServerBlock, error) {
	return nil, nil
}

func (context *dnsContext) MakeServers() ([]caddy.Server, error) {
	/*groups, err := groupConfigsByListenAddr(context.configs)
	if err != nil {
		return nil, err
	}

	var servers []caddy.Server
	for addr, group := range groups {
		switch tr, _ := parse.Transport(addr); tr {
		case parse.DNS:
			server, err := NewServer(addr, group)
			servers = append(servers, server)

		}
	}*/

	return nil, nil
}

func groupConfigsByListenAddr(configs []*Config) (map[string][]*Config, error) {
	groups := make(map[string][]*Config)
	for _, config := range configs {
		for _, host := range config.ListenHosts {
			addr, err := net.ResolveTCPAddr("tcp", net.JoinHostPort(host, config.Port))
			if err != nil {
				return nil, err
			}
			addrStr := config.Transport + "://" + addr.String()
			groups[addrStr] = append(groups[addrStr], config)
		}
	}

	return groups, nil
}
