package client

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"os"
	"strings"
	"testing"
	"time"
)

// go test -v -run ^TestExampleNew$ .
func TestExampleNew(test *testing.T) {
	log.SetLevel(log.DebugLevel)

	client, err := New(GetConfigOrDie(), Options{})
	if err != nil {
		fmt.Println("failed to create client")
		os.Exit(1)
	}

	podList := &corev1.PodList{}
	err = client.List(context.Background(), podList, InNamespace("default"))
	if err != nil {
		fmt.Printf("failed to list pods in namespace default: %v\n", err)
		os.Exit(1)
	}

	pods := podList.Items
	for _, pod := range pods {
		logPod(pod)
	}
}

// go test -v -run ^TestClientGet$ .
func TestClientGet(test *testing.T) {
	log.SetLevel(log.DebugLevel)

	client, err := New(GetConfigOrDie(), Options{})
	if err != nil {
		fmt.Println("failed to create client")
		os.Exit(1)
	}

	pod := &corev1.Pod{}
	err = client.Get(context.Background(), ObjectKey{
		Namespace: "default",
		Name:      "nginx-demo-1",
	}, pod)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Debug("[err]")
	} else {
		logPod(*pod)
	}

	// Using a unstructured object.
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "apps",
		Kind:    "Deployment",
		Version: "v1",
	})
	err = client.Get(context.Background(), ObjectKey{
		Namespace: "default",
		Name:      "nginx-demo-1",
	}, u)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Debug("[err]")
	}
}

// go test -v -run ^TestClientCreate$ .
func TestClientCreate(test *testing.T) {
	log.SetLevel(log.DebugLevel)

	client, err := New(GetConfigOrDie(), Options{})
	if err != nil {
		fmt.Println("failed to create client")
		os.Exit(1)
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "test-client-create",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Image: "nginx:1.17.8",
					Name:  "nginx",
				},
			},
		},
	}
	err = client.Create(context.Background(), pod)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Debug("[err]")
	}

	// Using a unstructured object
	u := &unstructured.Unstructured{}
	u.Object = map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":      "test-client-create-unstructured",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"replicas": 1,
			"selector": map[string]interface{}{
				"matchLabels": map[string]interface{}{
					"foo": "bar",
				},
			},
			"template": map[string]interface{}{
				"metadata": map[string]interface{}{
					"name": "test-client-create-unstructured",
					"labels": map[string]interface{}{
						"foo": "bar",
					},
				},
				"spec": map[string]interface{}{
					"containers": []map[string]interface{}{
						{
							"name":  "nginx",
							"image": "nginx:1.17.8",
						},
					},
				},
			},
		},
	}
	u.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "apps",
		Kind:    "Deployment",
		Version: "v1",
	})
	err = client.Create(context.Background(), u)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Debug("[err]")
	}

	// Using a typed object.
	pod2 := &corev1.PodList{}
	// c is a created client.
	_ = client.List(context.Background(), pod2)
	pods := pod2.Items
	for _, pod := range pods {
		logPod(pod)
	}

	// Using a unstructured object.
	u2 := &unstructured.UnstructuredList{}
	u2.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "apps",
		Kind:    "DeploymentList",
		Version: "v1",
	})
	_ = client.List(context.Background(), u2)
}

// go test -v -run ^TestClientUpdate$ .
func TestClientUpdate(test *testing.T) {
	log.SetLevel(log.DebugLevel)

	client, err := New(GetConfigOrDie(), Options{})
	if err != nil {
		fmt.Println("failed to create client")
		os.Exit(1)
	}

	// Using a typed object.
	pod := &corev1.Pod{}
	// c is a created client.
	err = client.Get(context.Background(), ObjectKey{
		Namespace: "default",
		Name:      "nginx-demo-1",
	}, pod)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Debug("[err]")
	}
	pod.SetFinalizers(append(pod.GetFinalizers(), "finalizer2"))
	err = client.Update(context.Background(), pod)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Debug("[err]")
	}

	// Using a unstructured object.
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "apps",
		Kind:    "Deployment",
		Version: "v1",
	})
	err = client.Get(context.Background(), ObjectKey{
		Namespace: "default",
		Name:      "nginx-demo-1",
	}, u)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Debug("[err]")
	}
	log.WithFields(log.Fields{
		"GetFinalizers": strings.Join(u.GetFinalizers(), "/"),
	}).Info("[GetFinalizers]")
	u.SetFinalizers(append(u.GetFinalizers(), "finalizer2"))
	err = client.Update(context.Background(), u)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Debug("[err]")
	}
}

// go test -v -run ^TestClientPatch$ .
func TestClientPatch(test *testing.T) {
	log.SetLevel(log.DebugLevel)

	client, err := New(GetConfigOrDie(), Options{})
	if err != nil {
		fmt.Println("failed to create client")
		os.Exit(1)
	}

	patch := []byte(`{"metadata":{"annotations":{"version": "v2"}}}`)
	err = client.Patch(context.Background(), &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "nginx-demo-1",
		},
	}, RawPatch(types.StrategicMergePatchType, patch))
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Debug("[err]")
	}

	// This example shows how to use the client with typed and unstructured objects to patch objects' status
	u := &unstructured.Unstructured{}
	u.Object = map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":      "nginx-demo-1-7f67f8bdd8-cvqsg",
			"namespace": "default",
		},
	}
	u.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "",
		Version: "v1",
		Kind:    "Pod",
	})
	patch2 := []byte(fmt.Sprintf(`{"status":{"startTime":"%s"}}`, time.Now().Format(time.RFC3339)))
	err = client.Status().Patch(context.Background(), u, RawPatch(types.StrategicMergePatchType, patch2))
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Debug("[err]")
	}
}

// go test -v -run ^TestClientDelete$ .
func TestClientDelete(test *testing.T) {
	log.SetLevel(log.DebugLevel)

	// Using a typed object
	client, err := New(GetConfigOrDie(), Options{})
	if err != nil {
		fmt.Println("failed to create client")
		os.Exit(1)
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "test-client-create",
		},
	}
	err = client.Delete(context.Background(), pod)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Debug("[err]")
	}

	// Using a unstructured object.
	u := &unstructured.Unstructured{}
	u.SetName("test-client-create-unstructured")
	u.SetNamespace("default")
	u.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "apps",
		Kind:    "Deployment",
		Version: "v1",
	})
	err = client.Delete(context.Background(), u)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Debug("[err]")
	}
}

// go test -v -run ^TestClientDeleteAll$ .
func TestClientDeleteAll(test *testing.T) {
	log.SetLevel(log.DebugLevel)

	// Using a typed object
	client, err := New(GetConfigOrDie(), Options{})
	if err != nil {
		fmt.Println("failed to create client")
		os.Exit(1)
	}

	err = client.DeleteAllOf(context.Background(), &corev1.Pod{}, InNamespace("default"), MatchingLabels{"app": "nginx-demo-1"})
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Debug("[err]")
	}

	// Using an unstructured Object
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "apps",
		Kind:    "Deployment",
		Version: "v1",
	})
	err = client.DeleteAllOf(context.Background(), u, InNamespace("default"), MatchingLabels{"app": "nginx-demo-1"})
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Debug("[err]")
	}
}

func logPod(pod corev1.Pod) {
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
