package dnsserver

import (
	"context"
	"github.com/miekg/dns"
	"k8s-lx1036/k8s/dns/coredns/plugin"
	"testing"
)

func mockConfig(transport string, handler plugin.Handler) *Config {
	config := &Config{
		Plugin:      nil,
		ListenHosts: []string{"127.0.0.1"},
		Port:        "53",
		Transport:   "",
		registry:    map[string]plugin.Handler{},
	}
	config.AddPlugin(func(next plugin.Handler) plugin.Handler {
		return handler
	})

	return config
}

type MockPlugin struct{}

func (plugin MockPlugin) Name() string {
	return "mockPlugin"
}
func (plugin MockPlugin) ServeDNS(ctx context.Context, writer dns.ResponseWriter, msg *dns.Msg) (int, error) {
	return 0, nil
}

func TestHandler(test *testing.T) {
	mockPlugin := MockPlugin{}
	config := mockConfig("dns", mockPlugin)
	if _, err := NewServer("127.0.0.1:53", []*Config{config}); err != nil {
		test.Errorf("Expected no error, got %s", err.Error())
	}
	if handler := config.Handler("mockPlugin"); handler != mockPlugin {
		test.Errorf("Expected mockPlugin from Handler, got %T", handler)
	}
	if handler := config.Handler("noPlugin"); handler != nil {
		test.Errorf("Expected mockPlugin from Handler, got %T", handler)
	}
}

func TestHandlers(test *testing.T) {
	mockPlugin := MockPlugin{}
	config := mockConfig("dns", mockPlugin)
	if _, err := NewServer("127.0.0.1:53", []*Config{config}); err != nil {
		test.Errorf("Expected no error, got %s", err.Error())
	}
	handlers := config.Handlers()
	if len(handlers) != 1 || handlers[0] != mockPlugin {
		test.Errorf("Expected mockPlugin, got %v", handlers)
	}
}
