package controller

import (
	"fmt"
	"strings"

	v1 "k8s-lx1036/k8s/storage/etcd/etcd-operator/pkg/apis/etcd.k9s.io/v1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// EtcdClientPort is the client port on client service and etcd nodes.
	EtcdClientPort = 2379

	etcdVolumeMountDir = "/var/etcd"
	dataDir            = etcdVolumeMountDir + "/data"

	peerTLSDir            = "/etc/etcdtls/member/peer-tls"
	peerTLSVolume         = "member-peer-tls"
	serverTLSDir          = "/etc/etcdtls/member/server-tls"
	serverTLSVolume       = "member-server-tls"
	operatorEtcdTLSDir    = "/etc/etcdtls/operator/etcd-tls"
	operatorEtcdTLSVolume = "etcd-client-tls"
)

/*
INFO:
  /usr/local/bin/etcd \
      --name=$(name) \
      --data-dir= /var/etcd/data\
      # List of this member's peer URLs to advertise to the rest of the cluster
      --initial-advertise-peer-urls=https://127.0.0.1:52380 \
      # List of this member's client URLs to advertise to the public.
      # The client URLs advertised should be accessible to machines that talk to etcd cluster.
      # etcd client libraries parse these URLs to connect to the cluster
      --advertise-client-urls=https://127.0.0.1:52379 \
      # List of URLs to listen on for peer traffic
      --listen-peer-urls=https://127.0.0.1:52380 \
      # List of URLs to listen on for client traffic
      --listen-client-urls=https://127.0.0.1:52379 \
      # --initial-cluster Initial cluster configuration for bootstrapping
      --initial-cluster 'infra1=https://127.0.0.1:42380,infra2=https://127.0.0.1:52380,infra3=https://127.0.0.1:62380' \
      # Initial cluster state ('new' or 'existing')
      --initial-cluster-state=new \
      # Initial cluster token for the etcd cluster during bootstrap.
      # Specifying this can protect you from unintended cross-cluster interaction when running multiple clusters
      # 只有 state=new 才需要 --initial-cluster-token
      --initial-cluster-token=etcd-cluster-0 \
      # securePeer, peer 之间是否是 HTTPS
      --peer-client-cert-auth=true \
      --peer-trusted-ca-file=$(PWD)/tls/ca.pem \
      --peer-cert-file=$(PWD)/tls/etcd.pem \
      --peer-key-file=$(PWD)/tls/etcd-key.pem \
      # secureClient, client 是否是 HTTPS
      --client-cert-auth=true \
      --trusted-ca-file=$(PWD)/tls/ca.pem \
      --cert-file=$(PWD)/tls/etcd.pem \
      --key-file=$(PWD)/tls/etcd-key.pem \
*/
func newEtcdPod(member *Member, initialCluster []string, clusterName, state, token string, etcdClusterSpec v1.EtcdClusterSpec) *corev1.Pod {
	commands := fmt.Sprintf(`/usr/local/bin/etcd --name=%s --data-dir=%s
		--initial-advertise-peer-urls=%s --advertise-client-urls=%s
		--listen-peer-urls=%s --listen-client-urls=%s
		--initial-cluster=%s --initial-cluster-state=%s`,
		member.Name, dataDir, member.PeerURL(), member.ClientURL(), member.ListenPeerURL(), member.ListenClientURL(),
		strings.Join(initialCluster, ","), state)
	if member.SecurePeer {
		commands += fmt.Sprintf(" --peer-client-cert-auth=true --peer-trusted-ca-file=%[1]s/peer-ca.crt --peer-cert-file=%[1]s/peer.crt --peer-key-file=%[1]s/peer.key", peerTLSDir)
	}
	if member.SecureClient {
		commands += fmt.Sprintf(" --client-cert-auth=true --trusted-ca-file=%[1]s/server-ca.crt --cert-file=%[1]s/server.crt --key-file=%[1]s/server.key", serverTLSDir)
	}
	if state == "new" {
		commands = fmt.Sprintf("%s --initial-cluster-token=%s", commands, token)
	}

	container := corev1.Container{
		Command: strings.Split(commands, " "),
		Name:    "etcd",
		Image:   ImageName(etcdClusterSpec.Repository, etcdClusterSpec.Version),
		Ports: []corev1.ContainerPort{
			{
				Name:          "server",
				ContainerPort: int32(2380),
				Protocol:      corev1.ProtocolTCP,
			},
			{
				Name:          "client",
				ContainerPort: int32(EtcdClientPort),
				Protocol:      corev1.ProtocolTCP,
			},
		},
		VolumeMounts: etcdVolumeMounts(),
	}
	labels := map[string]string{
		"app":          "etcd",
		"etcd_node":    member.Name,
		"etcd_cluster": clusterName,
	}
	volumes := []corev1.Volume{}
	if member.SecurePeer {
		container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
			MountPath: peerTLSDir,
			Name:      peerTLSVolume,
		})
		volumes = append(volumes, corev1.Volume{
			Name: peerTLSVolume,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: etcdClusterSpec.TLS.Static.Member.PeerSecret,
				},
			},
		})
	}
	if member.SecureClient {
		container.VolumeMounts = append(container.VolumeMounts,
			corev1.VolumeMount{
				MountPath: serverTLSDir,
				Name:      serverTLSVolume,
			},
			corev1.VolumeMount{
				MountPath: operatorEtcdTLSDir,
				Name:      operatorEtcdTLSVolume,
			},
		)
		volumes = append(volumes,
			corev1.Volume{
				Name: serverTLSVolume,
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: etcdClusterSpec.TLS.Static.Member.ServerSecret,
					},
				},
			},
			corev1.Volume{
				Name: operatorEtcdTLSVolume,
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: etcdClusterSpec.TLS.Static.OperatorSecret,
					},
				},
			},
		)
	}
	var securityContext *corev1.PodSecurityContext
	if etcdClusterSpec.Pod != nil {
		securityContext = etcdClusterSpec.Pod.SecurityContext
	}
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        member.Name,
			Labels:      labels,
			Annotations: map[string]string{},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				container,
			},
			RestartPolicy: corev1.RestartPolicyNever,
			Volumes:       volumes,
			// DNS A record: `[m.Name].[clusterName].Namespace.svc`
			// For example, etcd-795649v9kq in default namesapce will have DNS name
			// `etcd-795649v9kq.etcd.default.svc`.
			Hostname:                     member.Name,
			Subdomain:                    clusterName,
			AutomountServiceAccountToken: func(b bool) *bool { return &b }(false),
			SecurityContext:              securityContext,
		},
	}
	SetEtcdVersion(pod, etcdClusterSpec.Version)

	return pod
}

func NewEtcdPod(member *Member, initialCluster []string, clusterName, state, token string, etcdClusterSpec v1.EtcdClusterSpec,
	owner metav1.OwnerReference) *corev1.Pod {
	pod := newEtcdPod(member, initialCluster, clusterName, state, token, etcdClusterSpec)
	//applyPodPolicy(clusterName, pod, etcdClusterSpec.Pod)
	addOwnerRefToObject(pod.GetObjectMeta(), owner)

	return pod
}

func ImageName(repo, version string) string {
	return fmt.Sprintf("%s:v%v", repo, version)
}

const (
	etcdVolumeName = "etcd-data"

	etcdVersionAnnotationKey = "etcd.version"
)

func etcdVolumeMounts() []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{Name: etcdVolumeName, MountPath: etcdVolumeMountDir},
	}
}

func SetEtcdVersion(pod *corev1.Pod, version string) {
	pod.Annotations[etcdVersionAnnotationKey] = version
}

// 这会直接修改 pod ownerReferences 字段
func addOwnerRefToObject(obj metav1.Object, owner metav1.OwnerReference) {
	obj.SetOwnerReferences(append(obj.GetOwnerReferences(), owner))
}
