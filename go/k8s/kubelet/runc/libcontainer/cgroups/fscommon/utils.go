package fscommon

import (
	"golang.org/x/sys/unix"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	securejoin "github.com/cyphar/filepath-securejoin"
	"github.com/pkg/errors"

	"k8s.io/klog/v2"
)

// Gets a string value from the specified cgroup file
func GetCgroupParamString(cgroupPath, cgroupFile string) (string, error) {
	contents, err := ioutil.ReadFile(filepath.Join(cgroupPath, cgroupFile))
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(contents)), nil
}

func WriteFile(dir, file, data string) error {
	if dir == "" {
		return errors.Errorf("no directory specified for %s", file)
	}
	path, err := securejoin.SecureJoin(dir, file)
	if err != nil {
		return err
	}
	if err := retryingWriteFile(path, []byte(data), 0700); err != nil {
		return errors.Wrapf(err, "failed to write %q to %q", data, path)
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
