package main

import (
	"github.com/sirupsen/logrus"
	"os/exec"
	"strconv"
	"testing"
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
