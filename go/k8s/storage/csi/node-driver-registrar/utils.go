package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sys/unix"
	"k8s.io/klog/v2"
)

func removeRegSocket(csiDriverName string) {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGTERM)
	<-sigc
	socketPath := buildSocketPath(csiDriverName)
	err := os.Remove(socketPath)
	if err != nil && !os.IsNotExist(err) {
		klog.Errorf("failed to remove socket: %s with error: %+v", socketPath, err)
		os.Exit(1)
	}
	os.Exit(0)
}

func healthzServer(socketPath string, httpEndpoint string) {
	if httpEndpoint == "" {
		klog.Infof("Skipping healthz server because HTTP endpoint is set to: %q", httpEndpoint)
		return
	}
	klog.Infof("Starting healthz server at HTTP endpoint: %v\n", httpEndpoint)

	http.HandleFunc("/healthz", func(w http.ResponseWriter, req *http.Request) {
		socketExists, err := DoesSocketExist(socketPath)
		if err == nil && socketExists {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`ok`))
			klog.Infof("health check succeeded")
		} else if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			klog.Errorf("health check failed: %+v", err)
		} else if !socketExists {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("registration socket does not exist"))
			klog.Errorf("health check failed, registration socket does not exist")
		}
	})

	klog.Fatal(http.ListenAndServe(httpEndpoint, nil))
}

func Umask(mask int) (int, error) {
	return unix.Umask(mask), nil
}

func CleanupSocketFile(socketPath string) error {
	socketExists, err := DoesSocketExist(socketPath)
	if err != nil {
		return err
	}
	if socketExists {
		if err := os.Remove(socketPath); err != nil {
			return fmt.Errorf("failed to remove stale socket %s with error: %+v", socketPath, err)
		}
	}
	return nil
}

func DoesSocketExist(socketPath string) (bool, error) {
	fi, err := os.Stat(socketPath)
	if err == nil {
		if isSocket := fi.Mode()&os.ModeSocket != 0; isSocket {
			return true, nil
		}
		return false, fmt.Errorf("file exists in socketPath %s but it's not a socket.: %+v", socketPath, fi)
	}
	if !os.IsNotExist(err) {
		return false, fmt.Errorf("failed to stat the socket %s with error: %+v", socketPath, err)
	}
	return false, nil
}

func buildSocketPath(csiDriverName string) string {
	return fmt.Sprintf("%s/%s-reg.sock", *pluginRegistrationPath, csiDriverName)
}
