package main

import (
	"k8s.io/klog/v2"
	"math/rand"
	"testing"
)

func TestRand(test *testing.T) {
	klog.Info(rand.Float64()) // [0-1)
}
