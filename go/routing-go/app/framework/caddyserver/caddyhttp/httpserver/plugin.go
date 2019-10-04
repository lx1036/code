package httpserver

import (
	"github.com/mholt/certmagic"
	caddy "k8s-lx1036/routing-go/app/framework/caddyserver"
	"k8s-lx1036/routing-go/app/framework/caddyserver/caddytls"
)

// GetConfig gets the SiteConfig that corresponds to c.
// If none exist (should only happen in tests), then a
// new, empty one will be created.
func GetConfig(c *caddy.Controller) *SiteConfig {
	ctx := c.Context().(*httpContext)
	key := normalizedKey(c.Key)
	if cfg, ok := ctx.keysToSiteConfigs[key]; ok {
		return cfg
	}

	// we should only get here during tests because directive
	// actions typically skip the server blocks where we make
	// the configs
	cfg := &SiteConfig{
		Root:       Root,
		TLS:        &caddytls.Config{Manager: certmagic.NewDefault()},
		IndexPages: staticfiles.DefaultIndexPages,
	}
	ctx.saveConfig(key, cfg)
	return cfg
}

