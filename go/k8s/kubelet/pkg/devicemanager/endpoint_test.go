package devicemanager

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/klog/v2"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

func TestNewEndpoint(test *testing.T) {
	socketDir, socketName, pluginSocketName, err := tmpSocketDir()
	if err != nil {
		panic(err)
	}
	klog.Infof("[TestNewEndpoint]socketDir %s, socketName %s, pluginSocketName %s", socketDir, socketName, pluginSocketName)

	devices := []*pluginapi.Device{
		{ID: "deviceID", Health: pluginapi.Healthy},
	}

	// listen这个socket(pluginSocketName: /tmp/device_plugin/device-plugin.sock)
	devicePlugin := NewDevicePlugin(devices, pluginSocketName, testResourceName, false, false)
	err = devicePlugin.Start()
	defer devicePlugin.Stop()
	require.NoError(test, err)

	// call这个socket
	callback := func(resourceName string, devices []pluginapi.Device) {}
	endpoint, err := newEndpointImpl(pluginSocketName, "mock", callback)
	require.NoError(test, err)
	defer endpoint.stop()
}

func TestRun(test *testing.T) {
	socketDir, socketName, pluginSocketName, err := tmpSocketDir()
	if err != nil {
		panic(err)
	}
	klog.Infof("[TestNewEndpoint]socketDir %s, socketName %s, pluginSocketName %s", socketDir, socketName, pluginSocketName)

	devices := []*pluginapi.Device{
		{ID: "deviceID1", Health: pluginapi.Healthy},
		{ID: "deviceID2", Health: pluginapi.Healthy},
		{ID: "deviceID3", Health: pluginapi.Unhealthy},
	}

	updated := []*pluginapi.Device{
		{ID: "deviceID1", Health: pluginapi.Unhealthy},
		{ID: "deviceID3", Health: pluginapi.Healthy},
		{ID: "deviceID4", Health: pluginapi.Healthy},
	}

	// listen这个socket(pluginSocketName: /tmp/device_plugin/device-plugin.sock)
	devicePlugin := NewDevicePlugin(devices, pluginSocketName, testResourceName, false, false)
	err = devicePlugin.Start()
	defer devicePlugin.Stop()
	require.NoError(test, err)

	// call这个socket
	callbackChan := make(chan int)
	callbackCount := 0
	callback := func(resourceName string, devices []pluginapi.Device) {
		// Should be called twice:
		// one for plugin registration, one for plugin update.
		if callbackCount > 2 {
			test.FailNow()
		}

		// 首次是注册 devices{deviceID1,deviceID2,deviceID3}
		if callbackCount == 0 {
			require.Len(test, devices, 3)
			require.Equal(test, devices[0].ID, devices[0].ID)
			require.Equal(test, devices[1].ID, devices[1].ID)
			require.Equal(test, devices[2].ID, devices[2].ID)
			require.Equal(test, devices[0].Health, devices[0].Health)
			require.Equal(test, devices[1].Health, devices[1].Health)
			require.Equal(test, devices[2].Health, devices[2].Health)
		}

		// Check plugin update
		if callbackCount == 1 {
			require.Len(test, devices, 3)
			require.Equal(test, devices[0].ID, updated[0].ID)
			require.Equal(test, devices[1].ID, updated[1].ID)
			require.Equal(test, devices[2].ID, updated[2].ID)
			require.Equal(test, devices[0].Health, updated[0].Health)
			require.Equal(test, devices[1].Health, updated[1].Health)
			require.Equal(test, devices[2].Health, updated[2].Health)
		}

		callbackCount++
		callbackChan <- callbackCount
	}
	endpoint, err := newEndpointImpl(pluginSocketName, testResourceName, callback)
	require.NoError(test, err)
	defer endpoint.stop()

	// 同步 devices
	go endpoint.run()

	pluginRegistration := <-callbackChan
	// Wait for the first callback to be issued.
	klog.Infof("plugin registration: %d", pluginRegistration)

	devicePlugin.Update(updated)
	pluginUpdate := <-callbackChan
	// Wait for the second callback to be issued.
	klog.Infof("plugin update: %d", pluginUpdate)

	require.Equal(test, callbackCount, 2)
}

// INFO: 测试 Allocate()，很重要的函数
func TestAllocate(test *testing.T) {
	socketDir, socketName, pluginSocketName, err := tmpSocketDir()
	if err != nil {
		panic(err)
	}
	klog.Infof("[TestNewEndpoint]socketDir %s, socketName %s, pluginSocketName %s", socketDir, socketName, pluginSocketName)

	devices := []*pluginapi.Device{
		{ID: "deviceID1", Health: pluginapi.Healthy},
	}
	callbackCount := 0
	callbackChan := make(chan int)
	callback := func(resourceName string, devices []pluginapi.Device) {
		callbackCount++
		callbackChan <- callbackCount
	}

	// listen这个socket(pluginSocketName: /tmp/device_plugin/device-plugin.sock)
	devicePlugin := NewDevicePlugin(devices, pluginSocketName, testResourceName, false, false)
	err = devicePlugin.Start()
	defer devicePlugin.Stop()
	require.NoError(test, err)

	endpoint, err := newEndpointImpl(pluginSocketName, "mock", callback)
	require.NoError(test, err)
	defer endpoint.stop()

	resp := new(pluginapi.AllocateResponse)
	containerResp := new(pluginapi.ContainerAllocateResponse)
	containerResp.Devices = append(containerResp.Devices,
		&pluginapi.DeviceSpec{
			ContainerPath: "/dev/aaa",
			HostPath:      "/dev/aaa",
			Permissions:   "mrw",
		},
		&pluginapi.DeviceSpec{
			ContainerPath: "/dev/bbb",
			HostPath:      "/dev/bbb",
			Permissions:   "mrw",
		},
	)
	containerResp.Mounts = append(containerResp.Mounts, &pluginapi.Mount{
		ContainerPath: "/container_dir1/file1",
		HostPath:      "host_dir1/file1",
		ReadOnly:      true,
	})
	resp.ContainerResponses = append(resp.ContainerResponses, containerResp)
	devicePlugin.SetAllocFunc(func(r *pluginapi.AllocateRequest, devs map[string]pluginapi.Device) (*pluginapi.AllocateResponse, error) {
		return resp, nil
	})

	go endpoint.run()

	// Wait for the callback to be issued.
	select {
	case <-callbackChan:
		break
	case <-time.After(time.Second):
		test.FailNow()
	}

	respOut, err := endpoint.allocate([]string{"deviceID1"})
	require.NoError(test, err)
	require.Equal(test, resp, respOut)
}
