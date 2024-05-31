package cmd

import (
	"os"
	"os/signal"

	"golang.org/x/sys/unix"
)

var cleaner = &daemonCleanup{}

type daemonCleanup struct {
}

func (d *daemonCleanup) registerSigHandler() <-chan struct{} {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, unix.SIGQUIT, unix.SIGINT, unix.SIGHUP, unix.SIGTERM)
	interrupt := make(chan struct{})
	go func() {
		for s := range sig {
			log.WithField("signal", s).Info("Exiting due to signal")

			// cleanup...

			break
		}
		close(interrupt)
	}()
	return interrupt
}
