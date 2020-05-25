package whoami

import (
	"github.com/caddyserver/caddy"
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
