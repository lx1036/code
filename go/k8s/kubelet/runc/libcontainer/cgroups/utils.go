package cgroups

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"k8s-lx1036/k8s/kubelet/runc/libcontainer/configs"

	securejoin "github.com/cyphar/filepath-securejoin"
	"golang.org/x/sys/unix"
	"k8s.io/klog/v2"
)

var (
	ErrNotValidFormat = errors.New("line is not a valid key value format")
)

// The absolute path to the root of the cgroup hierarchies.
var cgroupRootLock sync.Mutex

type cgroupData struct {
	root      string
	innerPath string
	config    *configs.Cgroup
	pid       int
}

func (raw *cgroupData) path(subsystem string) (string, error) {
	// If the cgroup name/path is absolute do not look relative to the cgroup of the init process.
	if filepath.IsAbs(raw.innerPath) {
		mnt, err := FindCgroupMountpoint(raw.root, subsystem)
		// If we didn't mount the subsystem, there is no point we make the path.
		if err != nil {
			return "", err
		}

		// Sometimes subsystems can be mounted together as 'cpu,cpuacct'.
		return filepath.Join(raw.root, filepath.Base(mnt), raw.innerPath), nil
	}

	// Use GetOwnCgroupPath instead of GetInitCgroupPath, because the creating
	// process could in container and shared pid namespace with host, and
	// /proc/1/cgroup could point to whole other world of cgroups.
	parentPath, err := GetOwnCgroupPath(subsystem)
	if err != nil {
		return "", err
	}

	return filepath.Join(parentPath, raw.innerPath), nil
}

func getCgroupData(c *configs.Cgroup, pid int) (*cgroupData, error) {
	root, err := getCgroupRoot() // "fixtures"
	if err != nil {
		return nil, err
	}

	if (c.Name != "" || c.Parent != "") && c.Path != "" {
		return nil, errors.New("cgroup: either Path or Name and Parent should be used")
	}

	cgPath := filepath.Clean(c.Path)
	cgParent := filepath.Clean(c.Parent)
	cgName := filepath.Clean(c.Name)
	innerPath := cgPath
	if innerPath == "" {
		innerPath = filepath.Join(cgParent, cgName)
	}

	return &cgroupData{
		root:      root,
		innerPath: innerPath,
		config:    c,
		pid:       pid,
	}, nil
}

type Mount struct {
	Mountpoint string
	Root       string
	Subsystems []string
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

	mountinfoFile, err := filepath.Abs(GetFixturesMountInfoPath())
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
	cgroupFile, err := filepath.Abs("fixtures/proc/self/cgroup")
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
	cgroupFile, err := filepath.Abs("fixtures/proc/self/cgroup")
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

// GetCgroupMounts returns the mounts for the cgroup subsystems.
// all indicates whether to return just the first instance or all the mounts.
// This function should not be used from cgroupv2 code, as in this case
// all the controllers are available under the constant unifiedMountpoint.
func GetCgroupMounts(all bool) ([]Mount, error) {
	return getCgroupMountsV1(all)
}

var (
	fixturesMountInfoPath = "fixtures/proc/self/mountinfo"
	fixturesCgroupPath    = "fixtures/proc/self/cgroup"
)

func SetFixturesMountInfoPath(path string) {
	fixturesMountInfoPath = path
}
func GetFixturesMountInfoPath() string {
	return fixturesMountInfoPath
}
func SetFixturesCgroupPath(path string) {
	fixturesCgroupPath = path
}
func GetFixturesCgroupPath() string {
	return fixturesCgroupPath
}
func getCgroupMountsV1(all bool) ([]Mount, error) {
	path, err := filepath.Abs(GetFixturesMountInfoPath())
	if err != nil {
		panic(err)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	path, err = filepath.Abs(GetFixturesCgroupPath())
	if err != nil {
		panic(err)
	}
	allSubsystems, err := ParseCgroupFile(path)
	if err != nil {
		return nil, err
	}

	allMap := make(map[string]bool)
	for s := range allSubsystems {
		allMap[s] = false
	}
	return getCgroupMountsHelper(allMap, f, all)
}

func getCgroupMountsHelper(ss map[string]bool, mi io.Reader, all bool) ([]Mount, error) {
	res := make([]Mount, 0, len(ss))
	scanner := bufio.NewScanner(mi)
	numFound := 0
	for scanner.Scan() && numFound < len(ss) {
		txt := scanner.Text()
		sepIdx := strings.Index(txt, " - ")
		if sepIdx == -1 {
			return nil, fmt.Errorf("invalid mountinfo format")
		}
		if txt[sepIdx+3:sepIdx+10] == "cgroup2" || txt[sepIdx+3:sepIdx+9] != "cgroup" {
			continue
		}
		fields := strings.Split(txt, " ")
		m := Mount{
			Mountpoint: fields[4],
			Root:       fields[3],
		}
		for _, opt := range strings.Split(fields[len(fields)-1], ",") {
			seen, known := ss[opt]
			if !known || (!all && seen) {
				continue
			}
			ss[opt] = true
			opt = strings.TrimPrefix(opt, CgroupNamePrefix)
			m.Subsystems = append(m.Subsystems, opt)
			numFound++
		}
		if len(m.Subsystems) > 0 || all {
			res = append(res, m)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

// GetCgroupParamString Gets a string value from the specified cgroup file
func GetCgroupParamString(cgroupPath, cgroupFile string) (string, error) {
	contents, err := ioutil.ReadFile(filepath.Join(cgroupPath, cgroupFile))
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(contents)), nil
}

func WriteFile(dir, file, data string) error {
	if dir == "" {
		return fmt.Errorf("no directory specified for %s", file)
	}
	path, err := securejoin.SecureJoin(dir, file)
	if err != nil {
		return err
	}
	if err := retryingWriteFile(path, []byte(data), 0700); err != nil {
		return fmt.Errorf("failed to write %q to %q: %q", data, path, err)
	}
	return nil
}

func retryingWriteFile(filename string, data []byte, perm os.FileMode) error {
	for {
		err := ioutil.WriteFile(filename, data, perm)
		if errors.Is(err, unix.EINTR) {
			klog.Infof("interrupted while writing %s to %s", string(data), filename)
			continue
		}
		return err
	}
}

func cleanPath(path string) string {
	if path == "" {
		return ""
	}
	path = filepath.Clean(path)
	if !filepath.IsAbs(path) {
		path, _ = filepath.Rel(string(os.PathSeparator), filepath.Clean(string(os.PathSeparator)+path))
	}
	return filepath.Clean(path)
}

func PathExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		return false
	}
	return true
}

// GetCgroupParamKeyValue parses a space-separated "name value" kind of cgroup
// parameter and returns its components. For example, "io_service_bytes 1234"
// will return as "io_service_bytes", 1234.
func GetCgroupParamKeyValue(t string) (string, uint64, error) {
	parts := strings.Fields(t)
	switch len(parts) {
	case 2:
		value, err := ParseUint(parts[1], 10, 64)
		if err != nil {
			return "", 0, fmt.Errorf("unable to convert to uint64: %v", err)
		}

		return parts[0], value, nil
	default:
		return "", 0, ErrNotValidFormat
	}
}

// ParseUint converts a string to an uint64 integer.
// Negative values are returned at zero as, due to kernel bugs,
// some of the memory cgroup stats can be negative.
func ParseUint(s string, base, bitSize int) (uint64, error) {
	value, err := strconv.ParseUint(s, base, bitSize)
	if err != nil {
		intValue, intErr := strconv.ParseInt(s, base, bitSize)
		// 1. Handle negative values greater than MinInt64 (and)
		// 2. Handle negative values lesser than MinInt64
		if intErr == nil && intValue < 0 {
			return 0, nil
		} else if intErr != nil && intErr.(*strconv.NumError).Err == strconv.ErrRange && intValue < 0 {
			return 0, nil
		}

		return value, err
	}

	return value, nil
}

// Gets a single uint64 value from the specified cgroup file.
func GetCgroupParamUint(cgroupPath, cgroupFile string) (uint64, error) {
	fileName := filepath.Join(cgroupPath, cgroupFile)
	contents, err := ioutil.ReadFile(fileName)
	if err != nil {
		return 0, err
	}
	trimmed := strings.TrimSpace(string(contents))
	if trimmed == "max" {
		return math.MaxUint64, nil
	}

	res, err := ParseUint(trimmed, 10, 64)
	if err != nil {
		return res, fmt.Errorf("unable to parse %q as a uint from Cgroup file %q", string(contents), fileName)
	}
	return res, nil
}
