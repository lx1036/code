package cors

import (
	"log"
	"net/http"
	"os"
	"strings"
)

type Logger interface {
	Printf(string, ...interface{})
}

type wildcard struct {
	prefix string
	suffix string
}

type Cors struct {
	Log                    Logger
	allowedOrigins         []string
	allowedWOrigins        []wildcard
	allowOriginFunc        func(origin string) bool
	allowOriginRequestFunc func(r *http.Request, origin string) bool
	allowedHeaders         []string
	allowedMethods         []string
	exposedHeaders         []string
	maxAge                 int
	allowedOriginsAll      bool
	allowedHeadersAll      bool
	allowCredentials       bool
	optionPassthrough      bool
}

type Options struct {
	AllowedOrigins         []string
	AllowOriginFunc        func(origin string) bool
	AllowOriginRequestFunc func(r *http.Request, origin string) bool
	AllowedMethods         []string
	AllowedHeaders         []string
	ExposedHeaders         []string
	MaxAge                 int
	AllowCredentials       bool
	OptionsPassthrough     bool
	Debug                  bool
}

func Default() *Cors {
	return New(Options{})
}

func New(options Options) *Cors {
	cors := &Cors{
		exposedHeaders:         convert(options.ExposedHeaders, http.CanonicalHeaderKey),
		allowOriginFunc:        options.AllowOriginFunc,
		allowOriginRequestFunc: options.AllowOriginRequestFunc,
		allowCredentials:       options.AllowCredentials,
		maxAge:                 options.MaxAge,
		optionPassthrough:      options.OptionsPassthrough,
	}
	if options.Debug && cors.Log == nil {
		cors.Log = log.New(os.Stdout, "[cors] ", log.LstdFlags)
	}

	// AllowedOrigins
	if len(options.AllowedOrigins) == 0 {
		if options.AllowOriginFunc == nil && options.AllowOriginRequestFunc == nil {
			// Default is all origins
			cors.allowedOriginsAll = true
		}
	} else {
		cors.allowedOrigins = []string{}
		cors.allowedWOrigins = []wildcard{}
		for _, origin := range options.AllowedOrigins {
			origin = strings.ToLower(origin)
			if origin == "*" {
				cors.allowedOriginsAll = true
				cors.allowedOrigins = nil
				cors.allowedWOrigins = nil
				break
			} else if index := strings.IndexByte(origin, '*'); index >= 0 { // '*' is byte
				w := wildcard{origin[0:index], origin[index+1:]}
				cors.allowedWOrigins = append(cors.allowedWOrigins, w)
			} else {
				cors.allowedOrigins = append(cors.allowedOrigins, origin)
			}
		}
	}

	// AllowedHeaders
	if len(options.AllowedHeaders) == 0 {
		cors.allowedHeaders = []string{"Origin", "Accept", "Content-Type", "X-Requested-With"}
	} else {
		cors.allowedHeaders = convert(append(options.AllowedHeaders, "Origin"), http.CanonicalHeaderKey)
		for _, header := range options.AllowedHeaders {
			if header == "*" {
				cors.allowedHeadersAll = true
				cors.allowedHeaders = nil
				break
			}
		}
	}

	// AllowedMethods
	if len(options.AllowedMethods) == 0 {
		cors.allowedMethods = []string{http.MethodGet, http.MethodPost, http.MethodHead}
	} else {
		cors.allowedMethods = convert(options.AllowedMethods, strings.ToUpper)
	}

	return cors
}

func AllowAll() *Cors {
	return New(Options{
		AllowedOrigins:         []string{"*"},
		AllowOriginFunc:        nil,
		AllowOriginRequestFunc: nil,
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodHead,
			http.MethodPatch,
			http.MethodDelete,
		},
		AllowedHeaders:     []string{"*"},
		ExposedHeaders:     nil,
		MaxAge:             0,
		AllowCredentials:   false,
		OptionsPassthrough: false,
		Debug:              false,
	})
}

func (cors *Cors) HandlerFunc(writer http.ResponseWriter, request *http.Request) {
	if request.Method == http.MethodOptions && request.Header.Get("Access-Control-Request-Method") != "" {
		cors.logf("Preflight request:")
		cors.HandlePreflightRequest(writer, request)
	} else {
		cors.logf("Actual request:")
		cors.HandleActualRequest(writer, request)
	}
}

func (cors *Cors) HandlePreflightRequest(response http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodOptions {
		cors.logf("Preflight request method must be %s", http.MethodOptions)
		return
	}

	headers := response.Header()
	headers.Add("Vary", "Origin")
	headers.Add("Vary", "Access-Control-Request-Method")
	headers.Add("Vary", "Access-Control-Request-Headers")

	origin := request.Header.Get("Origin")
	if origin == "" {
		cors.logf("Preflight request aborted: empty origin")
		return
	}

	allowedMethod := request.Header.Get("Access-Control-Request-Method")
	if !cors.isMethodAllowed(allowedMethod) {
		cors.logf("Preflight request aborted: Access-Control-Request-Method [%s] is not allowed", allowedMethod)
		return
	}

}

func (cors *Cors) isMethodAllowed(allowedMethod string) bool {
	if len(allowedMethod) == 0 {
		return false
	}

	allowedMethod = strings.ToUpper(allowedMethod)

	if allowedMethod == http.MethodOptions {
		return true
	}

	for _, method := range cors.allowedMethods {
		if method == allowedMethod {
			return true
		}
	}

	return false
}

func (cors *Cors) HandleActualRequest(response http.ResponseWriter, request *http.Request) {
	origin := request.Header.Get("Origin")
	if origin == "" {
		cors.logf("Preflight request aborted: empty origin")
		return
	}
	headers := response.Header()
	headers.Add("Vary", "Origin")
	if cors.allowedOriginsAll {
		headers.Set("Access-Control-Allow-Origin", "*")
	} else {
		headers.Set("Access-Control-Allow-Origin", origin)
	}
	
	if len(cors.exposedHeaders) > 0 {
		headers.Set("Access-Control-Expose-Headers", strings.Join(cors.exposedHeaders, ", "))
	}
	
	if cors.allowCredentials {
		headers.Set("Access-Control-Allow-Credentials", "true")
	}
}

func (cors *Cors) logf(format string, v ...interface{}) {
	if cors.Log != nil {
		cors.Log.Printf(format, v...)
	}
}

type converter func(string) string

func convert(s []string, c converter) []string {
	out := []string{}
	for _, i := range s {
		out = append(out, c(i))
	}
	return out
}
