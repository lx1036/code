package httpserver

import (
	"fmt"
	"github.com/mholt/certmagic"
	caddy "k8s-lx1036/app/k8s/dns/caddyserver"
	"k8s-lx1036/app/k8s/dns/caddyserver/caddyfile"
	"time"
)

const serverType = "http"

const (
	// DefaultHost is the default host.
	DefaultHost = ""
	// DefaultPort is the default port.
	DefaultPort = "2015"
	// DefaultRoot is the default root folder.
	DefaultRoot = "."
	// DefaultHTTPPort is the default port for HTTP.
	DefaultHTTPPort = "80"
	// DefaultHTTPSPort is the default port for HTTPS.
	DefaultHTTPSPort = "443"
)

// These "soft defaults" are configurable by
// command line flags, etc.
var (
	// Root is the site root
	Root = DefaultRoot

	// Host is the site host
	Host = DefaultHost

	// Port is the site port
	Port = DefaultPort

	// GracefulTimeout is the maximum duration of a graceful shutdown.
	GracefulTimeout time.Duration

	// HTTP2 indicates whether HTTP2 is enabled or not.
	HTTP2 bool

	// QUIC indicates whether QUIC is enabled or not.
	QUIC bool
)

// directives is the list of all directives known to exist for the
// http server type, including non-standard (3rd-party) directives.
// The ordering of this list is important.
var directives = []string{
	// primitive actions that set up the fundamental vitals of each config
	"root",
	"index",
	"bind",
	"limits",
	"timeouts",
	"tls",

	// services/utilities, or other directives that don't necessarily inject handlers
	"startup",  // TODO: Deprecate this directive
	"shutdown", // TODO: Deprecate this directive
	"on",
	"supervisor", // github.com/lucaslorentz/caddy-supervisor
	"request_id",
	"realip", // github.com/captncraig/caddy-realip
	"git",    // github.com/abiosoft/caddy-git

	// directives that add listener middleware to the stack
	"proxyprotocol", // github.com/mastercactapus/caddy-proxyprotocol

	// directives that add middleware to the stack
	"locale", // github.com/simia-tech/caddy-locale
	"log",
	"cache", // github.com/nicolasazrak/caddy-cache
	"rewrite",
	"ext",
	"minify", // github.com/hacdias/caddy-minify
	"gzip",
	"header",
	"geoip", // github.com/kodnaplakal/caddy-geoip
	"errors",
	"authz",        // github.com/casbin/caddy-authz
	"filter",       // github.com/echocat/caddy-filter
	"ipfilter",     // github.com/pyed/ipfilter
	"ratelimit",    // github.com/xuqingfeng/caddy-rate-limit
	"recaptcha",    // github.com/defund/caddy-recaptcha
	"expires",      // github.com/epicagency/caddy-expires
	"forwardproxy", // github.com/caddyserver/forwardproxy
	"basicauth",
	"redir",
	"status",
	"cors",      // github.com/captncraig/cors/caddy
	"s3browser", // github.com/techknowlogick/caddy-s3browser
	"nobots",    // github.com/Xumeiquer/nobots
	"mime",
	"login",      // github.com/tarent/loginsrv/caddy
	"reauth",     // github.com/freman/caddy-reauth
	"extauth",    // github.com/BTBurke/caddy-extauth
	"jwt",        // github.com/BTBurke/caddy-jwt
	"permission", // github.com/dhaavi/caddy-permission
	"jsonp",      // github.com/pschlump/caddy-jsonp
	"upload",     // blitznote.com/src/caddy.upload
	"multipass",  // github.com/namsral/multipass/caddy
	"internal",
	"pprof",
	"expvar",
	"push",
	"datadog",    // github.com/payintech/caddy-datadog
	"prometheus", // github.com/miekg/caddy-prometheus
	"templates",
	"proxy",
	"pubsub", // github.com/jung-kurt/caddy-pubsub
	"fastcgi",
	"cgi", // github.com/jung-kurt/caddy-cgi
	"websocket",
	"filebrowser", // github.com/filebrowser/caddy
	"webdav",      // github.com/hacdias/caddy-webdav
	"markdown",
	"browse",
	"mailout",   // github.com/SchumacherFM/mailout
	"awses",     // github.com/miquella/caddy-awses
	"awslambda", // github.com/coopernurse/caddy-awslambda
	"grpc",      // github.com/pieterlouw/caddy-grpc
	"gopkg",     // github.com/zikes/gopkg
	"restic",    // github.com/restic/caddy
	"wkd",       // github.com/emersion/caddy-wkd
	"dyndns",    // github.com/linkonoid/caddy-dyndns
}

func init() {

	// Write a Server Type Plugin: https://dengxiaolong.com/caddy/zh/wiki.Writing-a-Plugin%3A-Server-Type.html
	caddy.RegisterServerType(serverType, caddy.ServerType{
		Directives: func() []string { return directives },
		DefaultInput: func() caddy.Input {
			if Port == DefaultPort && Host != "" {
				// by leaving the port blank in this case we give auto HTTPS
				// a chance to set the port to 443 for us
				return caddy.CaddyfileInput{
					Contents:       []byte(fmt.Sprintf("%s\nroot %s", Host, Root)),
					ServerTypeName: serverType,
				}
			}

			return caddy.CaddyfileInput{
				Contents:       []byte(fmt.Sprintf("%s:%s\nroot %s", Host, Port, Root)),
				ServerTypeName: serverType,
			}
		},
		NewContext: newContext,
	})

}

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

type httpContext struct {
	instance *caddy.Instance

	// keysToSiteConfigs maps an address at the top of a
	// server block (a "key") to its SiteConfig. Not all
	// SiteConfigs will be represented here, only ones
	// that appeared in the Caddyfile.
	keysToSiteConfigs map[string]*SiteConfig

	// siteConfigs is the master list of all site configs.
	siteConfigs []*SiteConfig
}

func (h httpContext) InspectServerBlocks(string, []caddyfile.ServerBlock) ([]caddyfile.ServerBlock, error) {
	panic("implement me")
}

func (h httpContext) MakeServers() ([]caddy.Server, error) {
	panic("implement me")
}

func newContext(inst *caddy.Instance) caddy.Context {
	return &httpContext{instance: inst, keysToSiteConfigs: make(map[string]*SiteConfig)}
}
