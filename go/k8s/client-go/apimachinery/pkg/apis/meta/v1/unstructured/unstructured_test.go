package unstructured

import (
	"fmt"
	"os"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/klog/v2"
)

func TestYamlUnstructured(test *testing.T) {
	const dsManifest = `
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: example
  namespace: default
spec:
  selector:
    matchLabels:
      name: nginx-ds
  template:
    metadata:
      labels:
        name: nginx-ds
    spec:
      containers:
      - name: nginx
        image: nginx:latest
`

	obj := &unstructured.Unstructured{}
	// decode YAML into unstructured.Unstructured
	dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	_, gvk, err := dec.Decode([]byte(dsManifest), nil, obj)
	if err != nil {
		panic(err)
	}
	// Get the common metadata, and show GVK
	klog.Info(gvk.String())
	klog.Info(fmt.Sprintf("%s/%s", obj.GetNamespace(), obj.GetName()))
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ") // 这个函数可以 perfect 打印
	encoder.Encode(obj)
}
