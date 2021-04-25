package cgroups

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// defaults returns all known groups
func defaults(root string) ([]Subsystem, error) {
	s := []Subsystem{
		NewCpuset(root),
	}

	return s, nil
}

func pathers(subystems []Subsystem) []pather {
	var out []pather
	for _, s := range subystems {
		if p, ok := s.(pather); ok {
			out = append(out, p)
		}
	}
	return out
}

// remove will remove a cgroup path handling EAGAIN and EBUSY errors and
// retrying the remove after a exp timeout
func remove(path string) error {
	delay := 10 * time.Millisecond
	for i := 0; i < 5; i++ {
		if i != 0 {
			time.Sleep(delay)
			delay *= 2
		}
		if err := os.RemoveAll(path); err == nil {
			return nil
		}
	}
	return fmt.Errorf("cgroups: unable to remove path %q", path)
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
