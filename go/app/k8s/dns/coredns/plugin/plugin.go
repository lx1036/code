package plugin

import (
	"context"
	"github.com/caddyserver/caddy"
	"github.com/miekg/dns"
)

type Plugin func(Handler) Handler
type Handler interface {
	ServeDNS(ctx context.Context, writer dns.ResponseWriter, msg *dns.Msg) (int, error)
	Name() string
}

func Register(name string, action caddy.SetupFunc) {
	caddy.RegisterPlugin(name, caddy.Plugin{
		ServerType: "dns",
		Action:     action,
	})
}
