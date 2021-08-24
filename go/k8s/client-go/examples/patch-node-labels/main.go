package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	jsonpatch "github.com/evanphx/json-patch"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

var (
	nodeName = flag.String("node", "", "")
)

// go run . --node=xxx
func main() {
	flag.Parse()
	if len(*nodeName) == 0 {
		klog.Fatal("--node is needed")
	}

	var kubeconfig *string
	if home, _ := os.UserHomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "absolute path to kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to kubeconfig file")
	}

	fmt.Println("kube config path: " + *kubeconfig)

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	klog.Info("===add custom lables===")

	node, err := clientSet.CoreV1().Nodes().Get(context.TODO(), *nodeName, metav1.GetOptions{})
	if err != nil {
		panic(err)
	}
	klog.Infof(fmt.Sprintf("node %s/%s", node.Namespace, node.Name))
	for key, value := range node.Labels {
		klog.Infof(fmt.Sprintf("%s=%s", key, value))
	}

	newNode := node.DeepCopy()
	newNode.Labels["test"] = "liuxiang"
	originalJSON, err := json.Marshal(node)
	if err != nil {
		panic(err)
	}
	modifiedJSON, err := json.Marshal(newNode)
	if err != nil {
		panic(err)
	}
	patch, err := jsonpatch.CreateMergePatch(originalJSON, modifiedJSON)
	if err != nil {
		panic(err)
	}

	patchNode, err := clientSet.CoreV1().Nodes().Patch(context.TODO(), *nodeName, types.StrategicMergePatchType, patch, metav1.PatchOptions{})
	if err != nil {
		panic(err)
	}
	for key, value := range patchNode.Labels {
		klog.Infof(fmt.Sprintf("%s=%s", key, value))
	}

	klog.Info("===delete custom lables===")
	// delete labels
	node, err = clientSet.CoreV1().Nodes().Get(context.TODO(), *nodeName, metav1.GetOptions{})
	if err != nil {
		panic(err)
	}
	newNode = node.DeepCopy()
	patchLabels := map[string]string{}
	for key, value := range newNode.Labels {
		if key == "test" {
			continue
		}
		patchLabels[key] = value
	}
	newNode.Labels = patchLabels
	originalJSON, err = json.Marshal(node)
	if err != nil {
		panic(err)
	}
	modifiedJSON, err = json.Marshal(newNode)
	if err != nil {
		panic(err)
	}
	patch, err = jsonpatch.CreateMergePatch(originalJSON, modifiedJSON)
	if err != nil {
		panic(err)
	}
	patchNode, err = clientSet.CoreV1().Nodes().Patch(context.TODO(), *nodeName, types.StrategicMergePatchType, patch, metav1.PatchOptions{})
	if err != nil {
		panic(err)
	}
	for key, value := range patchNode.Labels {
		klog.Infof(fmt.Sprintf("%s=%s", key, value))
	}
}
