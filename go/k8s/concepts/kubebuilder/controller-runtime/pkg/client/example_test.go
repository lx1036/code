package client

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"os"
	"strings"
	"testing"
)

// go test -v -run ^TestExampleNew$ .
func TestExampleNew(test *testing.T) {
	client, err := New(GetConfigOrDie(), Options{})
	if err != nil {
		fmt.Println("failed to create client")
		os.Exit(1)
	}

	podList := &corev1.PodList{}
	err = client.List(context.Background(), podList, InNamespace("kube-system"))
	if err != nil {
		fmt.Printf("failed to list pods in namespace default: %v\n", err)
		os.Exit(1)
	}

	pods := podList.Items
	for _, pod := range pods {
		var ips []string
		for _, podIP := range pod.Status.PodIPs {
			ips = append(ips, podIP.IP)
		}

		log.WithFields(log.Fields{
			"kind":      pod.Kind,
			"version":   pod.APIVersion,
			"namespace": pod.Namespace,
			"pod":       fmt.Sprintf("%s/%s", pod.Name, strings.Join(ips, "/")),
		}).Info("[pod]")
	}
}
