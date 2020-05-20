package whoami

import (
	"github.com/caddyserver/caddy"
	"k8s-lx1036/k8s/dns/coredns/core/dnsserver"
	"k8s-lx1036/k8s/dns/coredns/plugin"
)

func init() {
	plugin.Register("whoami", setup)
}

func setup(controller *caddy.Controller) error {
	controller.Next()

	dnsserver.GetConfig(controller).AddPlugin(func(handler plugin.Handler) plugin.Handler {
		return Whoami{}
	})

	return nil
}
