package sysctl

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	subsystem = "sysctl"

	procFsDefault = "/proc"
)

var (
	parameterElemRx = regexp.MustCompile(`\A[-0-9_a-z]+\z`)
)

type ErrInvalidSysctlParameter string

func (e ErrInvalidSysctlParameter) Error() string {
	return fmt.Sprintf("invalid sysctl parameter: %q", string(e))
}

func Enable(name string) error {
	return writeSysctl(name, "1")
}

func writeSysctl(name string, value string) error {
	path, err := parameterPath(name)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("could not open the sysctl file %s: %s",
			path, err)
	}
	defer f.Close()
	if _, err := io.WriteString(f, value); err != nil {
		return fmt.Errorf("could not write to the systctl file %s: %s",
			path, err)
	}
	return nil
}

func parameterPath(name string) (string, error) {
	elems := strings.Split(name, ".")
	for _, elem := range elems {
		if !parameterElemRx.MatchString(elem) {
			return "", ErrInvalidSysctlParameter(name)
		}
	}
	return filepath.Join(append([]string{procFsDefault, "sys"}, elems...)...), nil
}
