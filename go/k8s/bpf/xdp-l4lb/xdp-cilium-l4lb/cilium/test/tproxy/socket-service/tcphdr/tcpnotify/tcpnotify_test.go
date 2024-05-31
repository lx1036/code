package main

import (
	"os/exec"
	"strconv"
	"testing"

	"github.com/sirupsen/logrus"
)

// go test -v -run ^TestTcpNotify$ .
func TestTcpNotify(test *testing.T) {
	output, err := exec.Command("nc", "127.0.0.1", strconv.Itoa(TESTPORT), "-v").Output()
	if err != nil {
		logrus.Fatal(err)
	} else {
		logrus.Infof("exec output: %s", string(output))
	}
}
