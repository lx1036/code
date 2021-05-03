package node

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mattbaird/jsonpatch"

	v1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
)

const (
	admissionWebhookAnnotationStatusKey = "lxcfs-admission-webhook.aliyun.com/status"
)

func mutationRequired(pod corev1.Pod) bool {
	if pod.Annotations == nil {
		return true
	}

	status := pod.Annotations[admissionWebhookAnnotationStatusKey]

	if strings.ToLower(status) == "mutated" {
		return false
	}

	return true
}

func MutatePod(ar v1.AdmissionReview) *v1.AdmissionResponse {
	nodeResource := metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "nodes"}
	if ar.Request.Resource != nodeResource {
		err := fmt.Errorf("expect resource to be %s", nodeResource.String())
		klog.Error(err)
		return toAdmissionResponse(err)
	}

	raw := ar.Request.Object.Raw
	node := corev1.Node{}
	deserializer := codecs.UniversalDeserializer()
	if _, _, err := deserializer.Decode(raw, nil, &node); err != nil {
		klog.Error(err)
		return toAdmissionResponse(err)
	}

	klog.Info(fmt.Sprintf("nodeName: %s, node.Status.Allocatable: %v", node.Name, node.Status.Allocatable))

	return &v1.AdmissionResponse{Allowed: true}
}

func createPodPatch(raw []byte, mutated runtime.Object) ([]byte, error) {
	mu, err := json.Marshal(mutated)
	if err != nil {
		return nil, err
	}
	patch, err := jsonpatch.CreatePatch(raw, mu)
	if err != nil {
		return nil, err
	}
	if len(patch) > 0 {
		return json.Marshal(patch)
	}
	return nil, nil
}
