package httpserver



// SiteConfig contains information about a site
// (also known as a virtual host).
type SiteConfig struct {
	// The address of the site
	Addr Address

	// The list of viable index page names of the site
	IndexPages []string

	// The hostname to bind listener to;
	// defaults to Addr.Host
	ListenHost string

	// TLS configuration
	TLS *caddytls.Config

	// If true, the Host header in the HTTP request must
	// match the SNI value in the TLS handshake (if any).
	// This should be enabled whenever a site relies on
	// TLS client authentication, for example; or any time
	// you want to enforce that THIS site's TLS config
	// is used and not the TLS config of any other site
	// on the same listener. TODO: Check how relevant this
	// is with TLS 1.3.
	StrictHostMatching bool

	// Uncompiled middleware stack
	middleware []Middleware

	// Compiled middleware stack
	middlewareChain Handler

	// listener middleware stack
	listenerMiddleware []ListenerMiddleware

	// Directory from which to serve files
	Root string

	// A list of files to hide (for example, the
	// source Caddyfile). TODO: Enforcing this
	// should be centralized, for example, a
	// standardized way of loading files from disk
	// for a request.
	HiddenFiles []string

	// Max request's header/body size
	Limits Limits

	// The path to the Caddyfile used to generate this site config
	originCaddyfile string

	// These timeout values are used, in conjunction with other
	// site configs on the same server instance, to set the
	// respective timeout values on the http.Server that
	// is created. Sensible values will mitigate slowloris
	// attacks and overcome faulty networks, while still
	// preserving functionality needed for proxying,
	// websockets, etc.
	Timeouts Timeouts

	// If true, any requests not matching other site definitions
	// may be served by this site.
	FallbackSite bool
}





