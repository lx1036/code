package main

import (
	"errors"
	"fmt"

	"golang.org/x/net/context"
	"k8s.io/klog/v2"
	registerapi "k8s.io/kubelet/pkg/apis/pluginregistration/v1"
)

// registrationServer is a sample plugin to work with plugin watcher
type registrationServer struct {
	driverName string
	endpoint   string
	version    []string
}

func (e *registrationServer) GetInfo(c context.Context, req *registerapi.InfoRequest) (*registerapi.PluginInfo, error) {
	klog.Infof("Received GetInfo call: %+v", req)
	return &registerapi.PluginInfo{
		Type:              registerapi.CSIPlugin,
		Name:              e.driverName,
		Endpoint:          e.endpoint,
		SupportedVersions: e.version,
	}, nil
}

func (e *registrationServer) NotifyRegistrationStatus(c context.Context, status *registerapi.RegistrationStatus) (*registerapi.RegistrationStatusResponse, error) {
	klog.Infof("Received NotifyRegistrationStatus call: %+v", status)
	if !status.PluginRegistered {
		errmsg := fmt.Sprintf("Registration process failed with error: %+v, restarting registration container.", status.Error)
		klog.Error(errmsg)
		return nil, errors.New(errmsg)
	}

	return &registerapi.RegistrationStatusResponse{}, nil
}

// NewregistrationServer returns an initialized registrationServer instance
func newRegistrationServer(driverName string, endpoint string, versions []string) registerapi.RegistrationServer {
	return &registrationServer{
		driverName: driverName,
		endpoint:   endpoint,
		version:    versions,
	}
}
