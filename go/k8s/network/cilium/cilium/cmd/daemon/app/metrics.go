package app

import "github.com/cilium/cilium/pkg/metrics"

func InitMetrics() <-chan error {
	var errs <-chan error

	if option.Config.PrometheusServeAddr != "" {
		log.Infof("Serving prometheus metrics on %s", option.Config.PrometheusServeAddr)
		errs = metrics.Enable(option.Config.PrometheusServeAddr)
	}

	return errs
}
