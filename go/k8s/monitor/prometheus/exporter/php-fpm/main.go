package main

import (
	"flag"
	"go.uber.org/zap"
	"os"
)

func main() {
	var (
		addr            = flag.String("addr", getEnv("LISTEN_ADDR", "127.0.0.1:8080"), "listen address for metrics handler")
		endpoint        = flag.String("endpoint", getEnv("ENDPOINT_URL", "http://127.0.0.1:9000/status"), "url for php-fpm status")
		fcgiEndpoint    = flag.String("fastcgi", getEnv("FASTCGI_URL", ""), "url for php-fpm status")
		metricsEndpoint = flag.String("web.telemetry-path", getEnv("TELEMETRY_PATH", "/metrics"), "url for php-fpm status")
	)
	flag.Parse()

	logger, err := NewLogger()
	if err != nil {
		panic(err)
	}

	exporter, err := New(
		SetAddress(*addr),
		SetEndpoint(*endpoint),
		SetFastcgi(*fcgiEndpoint),
		SetLogger(logger),
		SetMetricsEndpoint(*metricsEndpoint),
	)

	if err != nil {
		logger.Fatal("failed to create exporter", zap.Error(err))
	}

	if err := exporter.Run(); err != nil {
		logger.Fatal("failed to run exporter", zap.Error(err))
	}
}

func getEnv(key string, defaultVal string) string {
	if envVal, ok := os.LookupEnv(key); ok {
		return envVal
	}
	return defaultVal
}
