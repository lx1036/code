package devicemanager

import (
	"os"
	"testing"

	"k8s-lx1036/k8s/kubelet/pkg/cm/topologymanager"

	"github.com/stretchr/testify/require"
	"k8s.io/klog/v2"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

func TestNewManagerImpl(t *testing.T) {
	socketDir, socketName, _, err := tmpSocketDir()
	topologyStore := topologymanager.NewFakeManager()
	require.NoError(t, err)
	defer os.RemoveAll(socketDir)
	_, err = newManagerImpl(socketName, nil, topologyStore)
	require.NoError(t, err)
}

func TestNewManagerImplStart(t *testing.T) {
	socketDir, socketName, pluginSocketName, err := tmpSocketDir()
	require.NoError(t, err)
	defer os.RemoveAll(socketDir)

	klog.Infof("[TestNewManagerImplStart]socketDir %s, socketName %s, pluginSocketName %s", socketDir, socketName, pluginSocketName)
	manager, _, p := setup(t, []*pluginapi.Device{}, func(n string, d []pluginapi.Device) {}, socketName, pluginSocketName)
	cleanup(t, manager, p)
	// Stop should tolerate being called more than once.
	cleanup(t, manager, p)
}
