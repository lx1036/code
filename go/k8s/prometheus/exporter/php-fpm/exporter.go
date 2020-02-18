package main

import (
	"context"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func NewLogger() (*zap.Logger, error) {
	config := zap.Config{
		Development:       false,
		DisableCaller:     true,
		DisableStacktrace: true,
		EncoderConfig:     zap.NewProductionEncoderConfig(),
		Encoding:          "json",
		ErrorOutputPaths:  []string{"stdout"},
		Level:             zap.NewAtomicLevel(),
		OutputPaths:       []string{"stdout"},
	}
	l, err := config.Build()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create logger")
	}
	return l, nil
}

type Exporter struct {
	addr            string
	endpoint        *url.URL
	fcgiEndpoint    *url.URL
	logger          *zap.Logger
	metricsEndpoint string
}

type OptionsFunc func(*Exporter) error

func New(options ...OptionsFunc) (*Exporter, error) {
	exporter := &Exporter{
		addr: ":9090",
	}

	for _, f := range options {
		if err := f(exporter); err != nil {
			return nil, errors.Wrap(err, "failed to set options")
		}
	}

	return exporter, nil
}

func SetAddress(addr string) OptionsFunc {
	return func(exporter *Exporter) error {
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return errors.Wrap(err, "invalid address")
		}
		exporter.addr = net.JoinHostPort(host, port)
		return nil
	}
}

func SetLogger(logger *zap.Logger) OptionsFunc {
	return func(exporter *Exporter) error {
		exporter.logger = logger
		return nil
	}
}

func SetEndpoint(endpoint string) OptionsFunc {
	return func(exporter *Exporter) error {
		uri, err := url.Parse(endpoint)
		if err != nil {
			return errors.Wrap(err, "failed to parse url")
		}
		exporter.endpoint = uri
		return nil
	}
}

func SetFastcgi(rawurl string) func(*Exporter) error {
	return func(e *Exporter) error {
		if rawurl == "" {
			return nil
		}
		u, err := url.Parse(rawurl)
		if err != nil {
			return errors.Wrap(err, "failed to parse url")
		}
		e.fcgiEndpoint = u
		return nil
	}
}
func SetMetricsEndpoint(path string) func(*Exporter) error {
	return func(e *Exporter) error {
		if path == "" || path == "/" {
			return nil
		}
		e.metricsEndpoint = path
		return nil
	}
}

// Run starts the http server and collecting metrics. It generally does not return.
func (exporter *Exporter) Run() error {
	collector := exporter.newCollector()
	if err := prometheus.Register(collector); err != nil {
		return errors.Wrap(err, "failed to register metrics")
	}
	prometheus.Unregister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{
		Namespace: MetricsNamespace,
	}))
	prometheus.Unregister(prometheus.NewGoCollector())

	http.HandleFunc("/healthz", exporter.healthz)
	http.Handle(exporter.metricsEndpoint, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<html>
			<head><title>php-fpm exporter</title></head>
			<body>
			<h1>php-fpm exporter</h1>
			<p><a href="` + exporter.metricsEndpoint + `">Metrics</a></p>
			</body>
			</html>`))
	})

	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	server := &http.Server{Addr: exporter.addr}
	var group errgroup.Group
	group.Go(func() error {
		return server.ListenAndServe()
	})
	group.Go(func() error {
		<-stopChan
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
		return nil
	})

	if err := group.Wait(); err != http.ErrServerClosed {
		return errors.Wrap(err, "failed to run server")
	}
	return nil
}

func (exporter *Exporter) healthz(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte(`ok\n`))
}
