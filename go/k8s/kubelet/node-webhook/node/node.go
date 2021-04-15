package node

import (
	"encoding/json"
	"strings"

	"github.com/mattbaird/jsonpatch"

	v1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
)

const (
	admissionWebhookAnnotationStatusKey = "lxcfs-admission-webhook.aliyun.com/status"
)

// -v /var/lib/lxcfs/proc/cpuinfo:/proc/cpuinfo:rw
// -v /var/lib/lxcfs/proc/diskstats:/proc/diskstats:rw
// -v /var/lib/lxcfs/proc/meminfo:/proc/meminfo:rw
// -v /var/lib/lxcfs/proc/stat:/proc/stat:rw
// -v /var/lib/lxcfs/proc/swaps:/proc/swaps:rw
// -v /var/lib/lxcfs/proc/uptime:/proc/uptime:rw
// -v /var/lib/lxcfs/proc/loadavg:/proc/loadavg:rw
var volumeMounts = []corev1.VolumeMount{
	{
		Name:      "lxcfs-proc-cpuinfo",
		MountPath: "/proc/cpuinfo",
	},
	{
		Name:      "lxcfs-proc-meminfo",
		MountPath: "/proc/meminfo",
	},
	{
		Name:      "lxcfs-proc-diskstats",
		MountPath: "/proc/diskstats",
	},
	{
		Name:      "lxcfs-proc-stat",
		MountPath: "/proc/stat",
	},
	{
		Name:      "lxcfs-proc-swaps",
		MountPath: "/proc/swaps",
	},
	{
		Name:      "lxcfs-proc-uptime",
		MountPath: "/proc/uptime",
	},
	{
		Name:      "lxcfs-proc-loadavg",
		MountPath: "/proc/loadavg",
	},
	{
		Name:      "lxcfs-sys-devices-system-cpu-online",
		MountPath: "/sys/devices/system/cpu/online",
	},
}
var volumes = []corev1.Volume{
	{
		Name: "lxcfs-proc-cpuinfo",
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: "/var/lib/lxcfs/proc/cpuinfo",
			},
		},
	},
	{
		Name: "lxcfs-proc-diskstats",
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: "/var/lib/lxcfs/proc/diskstats",
			},
		},
	},
	{
		Name: "lxcfs-proc-meminfo",
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: "/var/lib/lxcfs/proc/meminfo",
			},
		},
	},
	{
		Name: "lxcfs-proc-stat",
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: "/var/lib/lxcfs/proc/stat",
			},
		},
	},
	{
		Name: "lxcfs-proc-swaps",
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: "/var/lib/lxcfs/proc/swaps",
			},
		},
	},
	{
		Name: "lxcfs-proc-uptime",
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: "/var/lib/lxcfs/proc/uptime",
			},
		},
	},
	{
		Name: "lxcfs-proc-loadavg",
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: "/var/lib/lxcfs/proc/loadavg",
			},
		},
	},
	{
		Name: "lxcfs-sys-devices-system-cpu-online",
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: "/var/lib/lxcfs/sys/devices/system/cpu/online",
			},
		},
	},
}

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
	//podResource := metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "nodes"}
	//if ar.Request.Resource != podResource {
	//	err := fmt.Errorf("expect resource to be %s", podResource)
	//	klog.Error(err)
	//	return toAdmissionResponse(err)
	//}

	klog.Infof("")
	raw := ar.Request.Object.Raw
	pod := corev1.Node{}
	deserializer := codecs.UniversalDeserializer()
	if _, _, err := deserializer.Decode(raw, nil, &pod); err != nil {
		klog.Error(err)
		return toAdmissionResponse(err)
	}

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
