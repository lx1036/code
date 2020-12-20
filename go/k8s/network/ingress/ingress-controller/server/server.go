package server

import (
	"context"
	"sync"
	"sync/atomic"
)

type config struct {
	host    string
	port    int
	tlsPort int
}

type Option func(*config)

func defaultConfig() *config {
	return &config{
		host:    "0.0.0.0",
		port:    80,
		tlsPort: 443,
	}
}

type Event struct {
	once sync.Once
	C    chan struct{}
}

func (event Event) Set() {
	event.once.Do(func() {
		close(event.C)
	})
}

// NewEvent creates a new Event.
func NewEvent() *Event {
	return &Event{
		C: make(chan struct{}),
	}
}

type Server struct {
	config       *config
	routingTable atomic.Value

	ready *Event
}

func New(options ...Option) *Server {
	config := defaultConfig()
	for _, option := range options {
		option(config)
	}
	server := &Server{
		config: config,
		ready:  NewEvent(),
	}
	server.routingTable.Store(NewRoutingTable(nil))
	return server
}

func (s *Server) Update(payload *watcher.Payload) {
	s.routingTable.Store(NewRoutingTable(payload))
	s.ready.Set()
}

func WithHost(host string) Option {
	return func(cfg *config) {
		cfg.host = host
	}
}
func WithPort(port int) Option {
	return func(cfg *config) {
		cfg.port = port
	}
}

// WithTLSPort sets the TLS port in the config.
func WithTLSPort(port int) Option {
	return func(cfg *config) {
		cfg.tlsPort = port
	}
}

// Run 启动服务器.
func (s *Server) Run(ctx context.Context) error {

}
