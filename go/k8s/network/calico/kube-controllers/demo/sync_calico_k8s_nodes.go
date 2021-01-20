package main

import (
	"context"
	"flag"
	apisv3 "github.com/projectcalico/libcalico-go/lib/apis/v3"
	"github.com/projectcalico/libcalico-go/lib/errors"
	"github.com/projectcalico/libcalico-go/lib/options"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/projectcalico/libcalico-go/lib/apiconfig"
	client "github.com/projectcalico/libcalico-go/lib/clientv3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// 同步calico nodes和k8s nodes
// 删除不在k8s nodes内但在calico nodes中的node，并删除该node的相关calico资源
func main() {
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.JSONFormatter{})

	var kubeconfig *string
	if home, _ := os.UserHomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "absolute path to kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to kubeconfig file")
	}

	datastoreType := flag.String("datastore-type", string(apiconfig.Kubernetes), "calico datastore type")

	flag.Parse()

	calicoClient := getCalicoClientOrDie(*kubeconfig, *datastoreType)
	k8sClientset := getKubernetesClientOrDie(*kubeconfig)

	k8sNodes, err := k8sClientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.WithError(err).Error("Error listing K8s nodes")
		os.Exit(1)
	}

	calicoNodes, err := calicoClient.Nodes().List(context.TODO(), options.ListOptions{})
	if err != nil {
		log.WithError(err).Error("Error listing Calico nodes")
		os.Exit(1)
	}

	var k8sNodeNames []string
	for _, k8sNode := range k8sNodes.Items {
		k8sNodeNames = append(k8sNodeNames, k8sNode.Name)
	}

	k8sNodeNamesSet := sets.NewString(k8sNodeNames...)

	for _, calicoNode := range calicoNodes.Items {
		k8sNodeName := toK8sNodeName(calicoNode)
		if len(k8sNodeName) != 0 && !k8sNodeNamesSet.Has(k8sNodeName) { // 在calico nodes里但不在k8s nodes集群里
			// 从calico中删除该node相关所有资源
			_, err := calicoClient.Nodes().Delete(context.TODO(), calicoNode.Name, options.DeleteOptions{})
			calicoClient.BGPPeers()
			if err != nil {
				_, notExist := err.(errors.ErrorResourceDoesNotExist)
				if notExist {
					log.Warnf("calico node resource does not exist with: %v", err)
				} else {
					log.WithError(err).Error("error to delete calico node")
				}
			}
		}
	}
}

func getCalicoClientOrDie(kubeconfig, datastoreType string) client.Interface {
	c, err := client.New(apiconfig.CalicoAPIConfig{
		Spec: apiconfig.CalicoAPIConfigSpec{
			DatastoreType: apiconfig.DatastoreType(datastoreType),
			KubeConfig: apiconfig.KubeConfig{
				Kubeconfig: kubeconfig,
			},
		},
	})
	if err != nil {
		panic(err)
	}

	return c
}

func getKubernetesClientOrDie(kubeconfig string) *kubernetes.Clientset {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err)
	}

	return kubernetes.NewForConfigOrDie(config)
}

func toK8sNodeName(calicoNode apisv3.Node) string {
	for _, ref := range calicoNode.Spec.OrchRefs {
		if ref.Orchestrator == "k8s" {
			return ref.NodeName
		}
	}

	return ""
}
