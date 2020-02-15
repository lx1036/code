package basicauth

import (
	caddy "k8s-lx1036/k8s/dns/caddyserver"
	"k8s-lx1036/k8s/dns/caddyserver/caddyhttp/httpserver"
)

func init() {
	caddy.RegisterPlugin("basicauth", caddy.Plugin{
		ServerType: "http",
		Action:     setup,
	})
}

// setup configures a new BasicAuth middleware instance.
func setup(c *caddy.Controller) error {
	cfg := httpserver.GetConfig(c)
	root := cfg.Root

	rules, err := basicAuthParse(c)
	if err != nil {
		return err
	}

	basic := BasicAuth{Rules: rules}

	cfg.AddMiddleware(func(next httpserver.Handler) httpserver.Handler {
		basic.Next = next
		basic.SiteRoot = root
		return basic
	})

	return nil
}
