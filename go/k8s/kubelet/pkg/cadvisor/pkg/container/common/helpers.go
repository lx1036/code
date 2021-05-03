package common

import "path"

func MakeCgroupPaths(mountPoints map[string]string, name string) map[string]string {
	cgroupPaths := make(map[string]string, len(mountPoints))
	for key, val := range mountPoints {
		cgroupPaths[key] = path.Join(val, name)
	}

	return cgroupPaths
}
