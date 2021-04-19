package cgroups

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/go-units"
)

var HugePageSizeUnitList = []string{"B", "KB", "MB", "GB", "TB", "PB"}

type Mount struct {
	Mountpoint string
	Root       string
	Subsystems []string
}

func GetHugePageSize() ([]string, error) {
	// INFO: 这里mock使用本地目录文件, linux上是 /sys/kernel/mm/hugepages
	files, err := ioutil.ReadDir("./mock/sys/kernel/mm/hugepages")
	if err != nil {
		return []string{}, err
	}
	var fileNames []string
	for _, st := range files {
		fileNames = append(fileNames, st.Name())
	}
	return getHugePageSizeFromFilenames(fileNames)
}

func getHugePageSizeFromFilenames(fileNames []string) ([]string, error) {
	var pageSizes []string
	for _, fileName := range fileNames {
		nameArray := strings.Split(fileName, "-")
		pageSize, err := units.RAMInBytes(nameArray[1])
		if err != nil {
			return []string{}, err
		}
		sizeString := units.CustomSize("%g%s", float64(pageSize), 1024.0, HugePageSizeUnitList)
		pageSizes = append(pageSizes, sizeString)
	}

	return pageSizes, nil
}

// https://www.kernel.org/doc/Documentation/cgroup-v1/cgroups.txt
func FindCgroupMountpoint(cgroupPath, subsystem string) (string, error) {
	mnt, _, err := FindCgroupMountpointAndRoot(cgroupPath, subsystem)
	return mnt, err
}

func FindCgroupMountpointAndRoot(cgroupPath, subsystem string) (string, string, error) {
	// We are not using mount.GetMounts() because it's super-inefficient,
	// parsing it directly sped up x10 times because of not using Sscanf.
	// It was one of two major performance drawbacks in container start.
	if !isSubsystemAvailable(subsystem) {
		return "", "", fmt.Errorf("mountpoint for %s not found", subsystem)
	}

	mountinfoFile, err := filepath.Abs("mock/proc/self/mountinfo")
	if err != nil {
		panic(err)
	}
	f, err := os.Open(mountinfoFile)
	if err != nil {
		return "", "", err
	}
	defer f.Close()

	return findCgroupMountpointAndRootFromReader(f, cgroupPath, subsystem)
}

func findCgroupMountpointAndRootFromReader(reader io.Reader, cgroupPath, subsystem string) (string, string, error) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		txt := scanner.Text()
		fields := strings.Fields(txt)
		if len(fields) < 9 {
			continue
		}
		if strings.HasPrefix(fields[4], cgroupPath) {
			for _, opt := range strings.Split(fields[len(fields)-1], ",") {
				if opt == subsystem {
					return fields[4], fields[3], nil
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return "", "", err
	}

	return "", "", fmt.Errorf("mountpoint for %s not found", subsystem)
}

func isSubsystemAvailable(subsystem string) bool {
	cgroupFile, err := filepath.Abs("mock/proc/self/cgroup")
	if err != nil {
		panic(err)
	}
	cgroups, err := ParseCgroupFile(cgroupFile)
	if err != nil {
		return false
	}
	_, avail := cgroups[subsystem]
	return avail
}

// ParseCgroupFile parses the given cgroup file, typically /proc/self/cgroup
// or /proc/<pid>/cgroup, into a map of subsystems to cgroup paths, e.g.
//   "cpu": "/user.slice/user-1000.slice"
//   "pids": "/user.slice/user-1000.slice"
// etc.
//
// Note that for cgroup v2 unified hierarchy, there are no per-controller
// cgroup paths, so the resulting map will have a single element where the key
// is empty string ("") and the value is the cgroup path the <pid> is in.
func ParseCgroupFile(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return parseCgroupFromReader(f)
}

// helper function for ParseCgroupFile to make testing easier
func parseCgroupFromReader(r io.Reader) (map[string]string, error) {
	s := bufio.NewScanner(r)
	cgroups := make(map[string]string)

	for s.Scan() {
		text := s.Text()
		// from cgroups(7):
		// /proc/[pid]/cgroup
		// ...
		// For each cgroup hierarchy ... there is one entry
		// containing three colon-separated fields of the form:
		//     hierarchy-ID:subsystem-list:cgroup-path
		if len(strings.Trim(text, " ")) == 0 {
			continue
		}
		parts := strings.SplitN(text, ":", 3)
		if len(parts) < 3 {
			return nil, fmt.Errorf("invalid cgroup entry: must contain at least two colons: %v", text)
		}

		for _, subs := range strings.Split(parts[1], ",") {
			cgroups[subs] = parts[2]
		}
	}
	if err := s.Err(); err != nil {
		return nil, err
	}

	return cgroups, nil
}

func GetOwnCgroupPath(subsystem string) (string, error) {
	cgroup, err := GetOwnCgroup(subsystem)
	if err != nil {
		return "", err
	}

	return getCgroupPathHelper(subsystem, cgroup)
}

// GetOwnCgroup returns the relative path to the cgroup docker is running in.
func GetOwnCgroup(subsystem string) (string, error) {
	cgroupFile, err := filepath.Abs("mock/proc/self/cgroup")
	if err != nil {
		panic(err)
	}
	cgroups, err := ParseCgroupFile(cgroupFile)
	if err != nil {
		return "", err
	}

	return getControllerPath(subsystem, cgroups)
}

const (
	CgroupNamePrefix = "name="
)

func getControllerPath(subsystem string, cgroups map[string]string) (string, error) {
	if p, ok := cgroups[subsystem]; ok {
		return p, nil
	}

	if p, ok := cgroups[CgroupNamePrefix+subsystem]; ok {
		return p, nil
	}

	return "", fmt.Errorf("mountpoint for %s not found", subsystem)
}

func getCgroupPathHelper(subsystem, cgroup string) (string, error) {
	mnt, root, err := FindCgroupMountpointAndRoot("", subsystem)
	if err != nil {
		return "", err
	}

	// This is needed for nested containers, because in /proc/self/cgroup we
	// see paths from host, which don't exist in container.
	relCgroup, err := filepath.Rel(root, cgroup)
	if err != nil {
		return "", err
	}

	return filepath.Join(mnt, relCgroup), nil
}

type NotFoundError struct {
	Subsystem string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("mountpoint for %s not found", e.Subsystem)
}

func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*NotFoundError)
	return ok
}
