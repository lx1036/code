package main

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestHostLocal(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "plugins/ipam/host-local")
}
