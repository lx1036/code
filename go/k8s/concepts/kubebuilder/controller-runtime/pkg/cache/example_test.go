package cache

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"k8s-lx1036/k8s/concepts/kubebuilder/controller-runtime/pkg/client"
	corev1 "k8s.io/api/core/v1"
	"os"
	"testing"
)

func TestList(test *testing.T) {
	log.SetLevel(log.DebugLevel)

	informerCache, err := New(client.GetConfigOrDie(), Options{})
	if err != nil {
		fmt.Println("failed to create client")
		os.Exit(1)
	}

	serviceList := &corev1.Service{}
	err = informerCache.List(context.Background(), serviceList)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Debug("[err]")
	}

}
