package client

import (
	"flag"
	"fmt"
	"io/ioutil"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestMergoSemantics(test *testing.T) {

}

func TestInsecureOverridesCA(test *testing.T) {

}

func TestCAOverridesCAData(test *testing.T) {

}

func TestMergeContext(test *testing.T) {

}

func TestModifyContext(test *testing.T) {

}

func TestCertificateData(test *testing.T) {

}

func TestBasicAuthData(test *testing.T) {

}

func TestBasicTokenFile(test *testing.T) {

}

func TestPrecedenceTokenFile(test *testing.T) {

}

func TestCreateClean(t *testing.T) {

}

func TestCreateCleanWithPrefix(t *testing.T) {

}

func TestCreateCleanDefault(t *testing.T) {

}

func TestCreateCleanDefaultCluster(t *testing.T) {

}

func TestCreateMissingContextNoDefault(t *testing.T) {

}

func TestCreateMissingContext(t *testing.T) {

}

func TestInClusterClientConfigPrecedence(t *testing.T) {

}

func TestNamespaceOverride(t *testing.T) {

}

func TestAuthConfigMerge(test *testing.T) {
	content := `
apiVersion: v1
clusters:
- cluster:
    server: https://localhost:8080
  name: minikube
contexts:
- context:
    cluster: minikube
    user: minikube
    namespace: bar
  name: minikube
current-context: minikube
kind: Config
users:
- name: minikube
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1alpha1
      args:
      - arg-1
      - arg-2
      command: foo-command
`
	tmpfile, err := ioutil.TempFile("", "kubeconfig")
	if err != nil {
		test.Error(err)
	}
	defer os.Remove(tmpfile.Name())
	if err := ioutil.WriteFile(tmpfile.Name(), []byte(content), 0666); err != nil {
		test.Error(err)
	}

	config, err := clientcmd.BuildConfigFromFlags("", tmpfile.Name())
	if err != nil {
		test.Error(err)
	}

	if !reflect.DeepEqual(config.ExecProvider.Args, []string{"arg-1", "arg-2"}) {
		test.Errorf("Got args %v when they should be %v\n", config.ExecProvider.Args, []string{"arg-1", "arg-2"})
	}

	fmt.Println(config)
}

func GetClientSet() *kubernetes.Clientset {
	var kubeconfig *string
	if home, _ := os.UserHomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "absolute path to kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to kubeconfig file")
	}

	fmt.Println("kube config path: " + *kubeconfig)

	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	return clientSet
}

func TestRestClient(test *testing.T) {
	restClient := GetClientSet().CoreV1().RESTClient()
	restClient.Post().Resource("pods").SetHeader("Content-Type", "application/json").Body()
}
