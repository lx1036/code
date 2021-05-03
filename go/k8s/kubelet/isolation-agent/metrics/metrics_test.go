package main

import (
	"fmt"
	"testing"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog/v2"
)

func TestQuantity(test *testing.T) {
	totalUsageResource := v1.ResourceList{
		v1.ResourceCPU:              *resource.NewQuantity(1, resource.DecimalSI),
		v1.ResourceMemory:           *resource.NewQuantity(0, resource.BinarySI),
		v1.ResourceEphemeralStorage: *resource.NewQuantity(0, resource.BinarySI),
	}

	cpu := totalUsageResource[v1.ResourceCPU]
	klog.Info(fmt.Sprintf("cpu: %s", cpu.String())) // 1
}

func TestRatio(test *testing.T) {
	fmt.Println(float64(3044220) / float64(65798180)) // node_usage / node_total = 0.04626602133979998
	fmt.Println(float64(649224) / float64(65798180))  // pods_usage / node_total = 0.009866899054046783
}
