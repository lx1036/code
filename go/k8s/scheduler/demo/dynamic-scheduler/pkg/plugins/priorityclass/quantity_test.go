package priorityclass

import (
	"fmt"
	"testing"

	"k8s.io/apimachinery/pkg/api/resource"
)

func TestFormat(test *testing.T) {
	memorySize := resource.NewQuantity(5*1024*1024*1024, resource.BinarySI)
	fmt.Printf("memorySize = %v\n", memorySize)

	diskSize := resource.NewQuantity(5*1000*1000*1000, resource.DecimalSI)
	fmt.Printf("diskSize = %v\n", diskSize)

	cores := resource.NewMilliQuantity(5300, resource.DecimalSI)
	fmt.Printf("cores = %v\n", cores)

	// Output:
	// memorySize = 5Gi
	// diskSize = 5G
	// cores = 5300m
}
