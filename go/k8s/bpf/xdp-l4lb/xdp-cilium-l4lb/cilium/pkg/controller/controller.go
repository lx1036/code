package controller

import (
	"context"
	"github.com/sirupsen/logrus"
	"time"

	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/lock"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/logging"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/logging/logfields"
)

const (
	// fieldControllerName is the name of the controller
	fieldControllerName = "name"

	// fieldUUID is the UUID of the controller
	fieldUUID = "uuid"

	// fieldConsecutiveErrors is the number of consecutive errors of a controller
	fieldConsecutiveErrors = "consecutiveErrors"
)

var (
	// log is the controller package logger object.
	log = logging.DefaultLogger.WithField(logfields.LogSubsys, "controller")
)

// ControllerFunc is a function that the controller runs. This type is used for
// DoFunc and StopFunc.
type ControllerFunc func(ctx context.Context) error

type ControllerParams struct {
	// DoFunc is the function that will be run until it succeeds and/or
	// using the interval RunInterval if not 0.
	// An unset DoFunc is an error and will be logged as one.
	DoFunc ControllerFunc

	// StopFunc is called when the controller stops. It is intended to run any
	// clean-up tasks for the controller (e.g. deallocate/release resources)
	// It is guaranteed that DoFunc is called at least once before StopFunc is
	// called.
	// An unset StopFunc is not an error (and will be a no-op)
	// Note: Since this occurs on controller exit, error counts and tracking may
	// not be checked after StopFunc is run.
	StopFunc ControllerFunc

	// If set to any other value than 0, will cause DoFunc to be run in the
	// specified interval. The interval starts from when the DoFunc has
	// returned last
	RunInterval time.Duration

	// ErrorRetryBaseDuration is the initial time to wait to run DoFunc
	// again on return of an error. On each consecutive error, this value
	// is multiplied by the number of consecutive errors to provide a
	// constant back off. The default is 1s.
	ErrorRetryBaseDuration time.Duration

	// NoErrorRetry when set to true, disabled retries on errors
	NoErrorRetry bool

	Context context.Context
}

// Controller is a simple pattern that allows to perform the following
// tasks:
//   - Run an operation in the background and retry until it succeeds
//   - Perform a regular sync operation in the background
type Controller struct {
	mutex             lock.RWMutex
	name              string
	params            ControllerParams
	successCount      int
	lastSuccessStamp  time.Time
	failureCount      int
	consecutiveErrors int
	lastError         error
	lastErrorStamp    time.Time
	lastDuration      time.Duration
	uuid              string
	stop              chan struct{}
	update            chan struct{}
	trigger           chan struct{}
	ctxDoFunc         context.Context
	cancelDoFunc      context.CancelFunc

	// terminated is closed after the controller has been terminated
	terminated chan struct{}
}

func (c *Controller) stopController() {
	if c.cancelDoFunc != nil {
		c.cancelDoFunc()
	}

	close(c.stop)
	close(c.update)
}

func (c *Controller) getLogger() *logrus.Entry {
	return log.WithFields(logrus.Fields{
		fieldControllerName: c.name,
		fieldUUID:           c.uuid,
	})
}
