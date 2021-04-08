package evictions

import (
	"k8s-lx1036/k8s/scheduler/descheduler/pkg/client"
	"os"
	"path/filepath"
	"testing"
)

func TestSupportEviction(test *testing.T) {
	home, _ := os.UserHomeDir()
	kubeconfig := filepath.Join(home, ".kube", "config")
	kubeClient, err := client.CreateClient(kubeconfig)
	if err != nil {
		panic(err)
	}

	_, err = SupportEviction(kubeClient)
	if err != nil {
		panic(err)
	}
}
