package fs

import (
	"path/filepath"
	"strings"
	"testing"

	"k8s-lx1036/k8s/kubelet/runc/libcontainer/configs"
)

func TestInvalidCgroupPath(t *testing.T) {
	root, err := getCgroupRoot()
	if err != nil {
		t.Fatalf("couldn't get cgroup root: %v", err)
	}

	testCases := []struct {
		test               string
		path, name, parent string
	}{
		{
			test: "invalid cgroup path",
			path: "../../../../../../../../../../some/path",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.test, func(t *testing.T) {
			config := &configs.Cgroup{Path: tc.path, Name: tc.name, Parent: tc.parent}
			data, err := getCgroupData(config, 0)
			if err != nil {
				t.Fatalf("couldn't get cgroup data: %v", err)
			}

			// Make sure the final innerPath doesn't go outside the cgroup mountpoint.
			if strings.HasPrefix(data.innerPath, "..") {
				t.Errorf("SECURITY: cgroup innerPath is outside cgroup mountpoint!")
			}

			// Double-check, using an actual cgroup.
			deviceRoot := filepath.Join(root, "devices")
			devicePath, err := data.path("devices")
			if err != nil {
				t.Fatalf("couldn't get cgroup path: %v", err)
			}
			if !strings.HasPrefix(devicePath, deviceRoot) {
				t.Errorf("SECURITY: cgroup path() is outside cgroup mountpoint!")
			}
		})
	}

}
