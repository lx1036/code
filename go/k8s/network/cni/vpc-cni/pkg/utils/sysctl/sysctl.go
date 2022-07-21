package sysctl

import (
	"bytes"
	"io/ioutil"
)

// WriteProcSys takes the sysctl path and a string value to set i.e. "0" or "1" and sets the sysctl.
func WriteProcSys(path, value string) error {
	if content, err := ioutil.ReadFile(path); err == nil {
		if bytes.Equal(bytes.TrimSpace(content), []byte(value)) {
			return nil
		}
	}
	return ioutil.WriteFile(path, []byte(value), 0644)
}
