package devicemanager

import (
	"io/ioutil"
	"os"
	"testing"

	"k8s-lx1036/k8s/kubelet/pkg/cm/topologymanager"

	"github.com/stretchr/testify/require"

	v1 "k8s.io/api/core/v1"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

const (
	testResourceName = "fake-domain/resource"
)

// socketName: /tmp/device_plugin/server.sock, pluginSocketName: /tmp/device_plugin/device-plugin.sock
func tmpSocketDir() (socketDir, socketName, pluginSocketName string, err error) {
	socketDir, err = ioutil.TempDir("", "device_plugin")
	if err != nil {
		return
	}
	socketName = socketDir + "/server.sock"
	pluginSocketName = socketDir + "/device-plugin.sock"
	os.MkdirAll(socketDir, 0755)
	return
}

func setupDeviceManager(t *testing.T, devs []*pluginapi.Device, callback monitorCallback,
	socketName string) (Manager, <-chan interface{}) {
	topologyStore := topologymanager.NewFakeManager()
	m, err := newManagerImpl(socketName, nil, topologyStore)
	require.NoError(t, err)
	updateChan := make(chan interface{})

	if callback != nil {
		m.callback = callback
	}

	originalCallback := m.callback
	m.callback = func(resourceName string, devices []pluginapi.Device) {
		originalCallback(resourceName, devices)
		updateChan <- new(interface{})
	}
	activePods := func() []*v1.Pod {
		return []*v1.Pod{}
	}

	err = m.Start(activePods, &sourcesReadyStub{})
	require.NoError(t, err)

	return m, updateChan
}

func setup(t *testing.T, devs []*pluginapi.Device, callback monitorCallback,
	socketName string, pluginSocketName string) (Manager, <-chan interface{}, *DevicePlugin) {
	m, updateChan := setupDeviceManager(t, devs, callback, socketName)
	p := setupDevicePlugin(t, devs, pluginSocketName)
	return m, updateChan, p
}

func setupDevicePlugin(t *testing.T, devs []*pluginapi.Device, pluginSocketName string) *DevicePlugin {
	p := NewDevicePlugin(devs, pluginSocketName, testResourceName, false, false)
	err := p.Start()
	require.NoError(t, err)
	return p
}

func cleanup(t *testing.T, m Manager, p *DevicePlugin) {
	p.Stop()
	m.Stop()
}
