package watchers

import (
	"errors"
	"fmt"
	"github.com/cilium/cilium/pkg/annotation"
	"github.com/cilium/cilium/pkg/identity"
	"github.com/cilium/cilium/pkg/ipcache"
	"github.com/cilium/cilium/pkg/logging/logfields"
	"github.com/cilium/cilium/pkg/source"
	"github.com/cilium/cilium/pkg/u8proto"
	log "github.com/sirupsen/logrus"
	"net"
	"reflect"

	"k8s-lx1036/k8s/network/cilium/cilium/pkg/bpf/endpoint"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/bpf/endpoint/regeneration"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/k8s"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/k8s/loadbalancer"
	nodeTypes "k8s-lx1036/k8s/network/cilium/cilium/pkg/k8s/node/types"

	"sort"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

var (
	errIPCacheOwnedByNonK8s = fmt.Errorf("ipcache entry owned by kvstore or agent")
)

func (k *K8sWatcher) watchK8sPod(k8sClient kubernetes.Interface) {

	podStore, podController := cache.NewTransformingInformer(
		cache.NewListWatchFromClient(k8sClient.CoreV1().RESTClient(),
			"pods", corev1.NamespaceAll, fields.ParseSelectorOrDie("spec.nodeName="+nodeTypes.GetName())),
		&corev1.Pod{},
		0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				var valid, equal bool
				defer func() { k.K8sEventReceived(metricPod, metricCreate, valid, equal) }()

				k8sPod, ok := obj.(*corev1.Pod)
				if !ok {
					return
				}

				err := k.addK8sPodV1(k8sPod, swgSvcs)
				k.K8sEventProcessed(metricPod, metricCreate, err == nil)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				var valid, equal bool
				defer func() { k.K8sEventReceived(metricPod, metricUpdate, valid, equal) }()

				oldK8sPod, ok := oldObj.(*corev1.Pod)
				if !ok {
					return
				}
				newK8sPod, ok := newObj.(*corev1.Pod)
				if !ok {
					return
				}
				if EqualPod(oldK8sPod, newK8sPod) {
					equal = true
					return
				}

				err := k.updateK8sPodV1(oldK8sPod, newK8sPod)
				k.K8sEventProcessed(metricPod, metricUpdate, err == nil)
			},
			DeleteFunc: func(obj interface{}) {
				var valid, equal bool
				defer func() { k.K8sEventReceived(metricPod, metricDelete, valid, equal) }()

				newK8sPod, ok := obj.(*corev1.Pod)
				if !ok {
					return
				}

				valid = true
				err := k.deleteK8sPodV1(newK8sPod)
				k.K8sEventProcessed(metricPod, metricDelete, err == nil)
			},
		},
		nil,
	)

	go podController.Run(wait.NeverStop)
	k.podStore = podStore

}

// watch pod 只处理 frontend HostIP:HostPot -> backend podIP:ContainerPort BPF maps 数据
func (k *K8sWatcher) addK8sPodV1(pod *corev1.Pod) error {
	logger := log.WithFields(log.Fields{
		logfields.K8sPodName:   pod.ObjectMeta.Name,
		logfields.K8sNamespace: pod.ObjectMeta.Namespace,
		"podIP":                pod.Status.PodIP,
		"podIPs":               pod.Status.PodIPs,
		"hostIP":               pod.Status.PodIP,
	})

	skipped, err := k.updatePodHostData(pod)
	switch {
	case skipped:
		logger.WithError(err).Debug("Skipped ipcache map update on pod add")
		return nil
	case err != nil:
		msg := "Unable to update ipcache map entry on pod add"
		if err == errIPCacheOwnedByNonK8s {
			logger.WithError(err).Debug(msg)
		} else {
			logger.WithError(err).Warning(msg)
		}
	default:
		logger.Debug("Updated ipcache map entry on pod add")
	}

	return err
}

func (k *K8sWatcher) updateK8sPodV1(oldK8sPod, newK8sPod *corev1.Pod) error {
	if oldK8sPod == nil || newK8sPod == nil {
		return nil
	}

	// The pod IP can never change, it can only switch from unassigned to
	// assigned
	// Process IP updates
	k.addK8sPodV1(newK8sPod)

	// Check annotation updates.
	oldAnno := oldK8sPod.ObjectMeta.Annotations
	newAnno := newK8sPod.ObjectMeta.Annotations
	annotationsChanged := !k8s.AnnotationsEqual([]string{annotation.ProxyVisibility}, oldAnno, newAnno)
	// Check label updates too.
	oldPodLabels := oldK8sPod.ObjectMeta.Labels
	newPodLabels := newK8sPod.ObjectMeta.Labels
	labelsChanged := !reflect.DeepEqual(oldPodLabels, newPodLabels)
	// Nothing changed.
	if !annotationsChanged && !labelsChanged {
		return nil
	}

	podKey := fmt.Sprintf("%s/%s", newK8sPod.Namespace, newK8sPod.Name)
	podEndpoint := k.endpointManager.LookupPodName(podKey)
	if podEndpoint == nil {
		log.WithField("pod", podKey).Debugf("Endpoint not found running for the given pod")
		return nil
	}

	if labelsChanged {
		err := updateEndpointLabels(podEndpoint, oldPodLabels, newPodLabels)
		if err != nil {
			return err
		}
	}

	if annotationsChanged {
		podEndpoint.UpdateVisibilityPolicy(func(ns, podName string) (proxyVisibility string, err error) {
			p, err := k.GetCachedPod(ns, podName)
			if err != nil {
				return "", nil
			}
			return p.ObjectMeta.Annotations[annotation.ProxyVisibility], nil
		})

		realizePodAnnotationUpdate(podEndpoint)
	}
	return nil
}

func realizePodAnnotationUpdate(podEndpoint *endpoint.Endpoint) {
	regenMetadata := &regeneration.ExternalRegenerationMetadata{
		Reason:            "annotations updated",
		RegenerationLevel: regeneration.RegenerateWithoutDatapath,
	}
	// No need to log an error if the state transition didn't succeed,
	// if it didn't succeed that means the endpoint is being deleted, or
	// another regeneration has already been queued up for this endpoint.
	regen, _ := podEndpoint.SetRegenerateStateIfAlive(regenMetadata)
	if regen {
		podEndpoint.Regenerate(regenMetadata)
	}
}

func (k *K8sWatcher) deleteK8sPodV1(pod *corev1.Pod) error {
	logger := log.WithFields(log.Fields{
		logfields.K8sPodName:   pod.ObjectMeta.Name,
		logfields.K8sNamespace: pod.ObjectMeta.Namespace,
		"podIP":                pod.Status.PodIP,
		"podIPs":               pod.Status.PodIPs,
		"hostIP":               pod.Status.HostIP,
	})

	skipped, err := k.deletePodHostData(pod)
	switch {
	case skipped:
		logger.WithError(err).Debug("Skipped ipcache map delete on pod delete")
	case err != nil:
		logger.WithError(err).Warning("Unable to delete ipcache map entry on pod delete")
	default:
		logger.Debug("Deleted ipcache map entry on pod delete")
	}
	return err
}

func (k *K8sWatcher) updatePodHostData(pod *corev1.Pod) (bool, error) {
	if pod.Spec.HostNetwork {
		return true, fmt.Errorf("pod is using host networking")
	}

	podIPs, err := validIPs(pod.Status)
	if err != nil {
		return true, err
	}

	err = k.UpdateOrInsertHostPortMapping(pod, podIPs)
	if err != nil {
		return true, fmt.Errorf("cannot upsert hostPort for PodIPs: %s", podIPs)
	}

	k8sMeta := &ipcache.K8sMetadata{
		Namespace: pod.Namespace,
		PodName:   pod.Name,
	}
	// Store Named ports, if any.
	for _, container := range pod.Spec.Containers {
		for _, port := range container.Ports {
			if port.Name == "" {
				continue
			}
			p, err := u8proto.ParseProtocol(string(port.Protocol))
			if err != nil {
				return true, fmt.Errorf("ContainerPort: invalid protocol: %s", port.Protocol)
			}
			if port.ContainerPort < 1 || port.ContainerPort > 65535 {
				return true, fmt.Errorf("ContainerPort: invalid port: %d", port.ContainerPort)
			}
			if k8sMeta.NamedPorts == nil {
				k8sMeta.NamedPorts = make(policy.NamedPortsMap)
			}
			k8sMeta.NamedPorts[port.Name] = policy.NamedPort{
				Proto: uint8(p),
				Port:  uint16(port.ContainerPort),
			}
		}
	}

	var errs []string
	for _, podIP := range podIPs {
		// Initial mapping of podIP <-> hostIP <-> identity. The mapping is
		// later updated once the allocator has determined the real identity.
		// If the endpoint remains unmanaged, the identity remains untouched.
		selfOwned, namedPortsChanged := ipcache.IPIdentityCache.Upsert(podIP, hostIP, hostKey, k8sMeta, ipcache.Identity{
			ID:     identity.ReservedIdentityUnmanaged,
			Source: source.Kubernetes,
		})
		// This happens at most once due to k8sMeta being the same for all podIPs in this loop
		if namedPortsChanged {
			k.policyManager.TriggerPolicyUpdates(true, "Named ports added or updated")
		}
		if !selfOwned {
			errs = append(errs, fmt.Sprintf("ipcache entry for podIP %s owned by kvstore or agent", podIP))
		}
	}
	if len(errs) != 0 {
		return true, errors.New(strings.Join(errs, ", "))
	}

	return false, nil
}

// UpdateOrInsertHostPortMapping INFO: 处理 HostPort，和 NodePort 类似
func (k *K8sWatcher) UpdateOrInsertHostPortMapping(pod *corev1.Pod, podIPs []string) error {
	if option.Config.DisableK8sServices || !option.Config.EnableHostPort {
		return nil
	}

	svcs := genServiceMappings(pod, podIPs)
	if len(svcs) == 0 {
		return nil
	}

	logger := log.WithFields(log.Fields{
		logfields.K8sPodName:   pod.ObjectMeta.Name,
		logfields.K8sNamespace: pod.ObjectMeta.Namespace,
		"podIPs":               podIPs,
		"hostIP":               pod.Status.HostIP,
	})

	hostIP := net.ParseIP(pod.Status.HostIP)
	if hostIP == nil {
		logger.Error("Cannot upsert HostPort service for the podIP due to missing hostIP")
		return fmt.Errorf("no/invalid HostIP: %s", pod.Status.HostIP)
	}

	for _, dpSvc := range svcs {
		if _, _, err := k.serviceBPFManager.UpdateOrInsertService(dpSvc.Frontend, dpSvc.Backends, dpSvc.Type,
			dpSvc.TrafficPolicy, false, 0, dpSvc.HealthCheckNodePort,
			fmt.Sprintf("%s/host-port/%d", pod.ObjectMeta.Name, dpSvc.Frontend.L3n4Addr.Port),
			pod.ObjectMeta.Namespace); err != nil {
			logger.WithError(err).Error("Error while inserting service in LB map")
			return err
		}
	}

	return nil
}

func (k *K8sWatcher) deletePodHostData(pod *corev1.Pod) (bool, error) {
	if pod.Spec.HostNetwork {
		return true, fmt.Errorf("pod is using host networking")
	}

	podIPs, err := validIPs(pod.Status)
	if err != nil {
		return true, err
	}

	k.DeleteHostPortMapping(pod, podIPs)

	var (
		errs    []string
		skipped bool
	)
	for _, podIP := range podIPs {
		// a small race condition exists here as deletion could occur in
		// parallel based on another event but it doesn't matter as the
		// identity is going away
		id, exists := ipcache.IPIdentityCache.LookupByIP(podIP)
		if !exists {
			skipped = true
			errs = append(errs, fmt.Sprintf("identity for IP %s does not exist in case", podIP))
			continue
		}

		if id.Source != source.Kubernetes {
			skipped = true
			errs = append(errs, fmt.Sprintf("ipcache entry for IP %s not owned by kubernetes source", podIP))
			continue
		}

		ipcache.IPIdentityCache.Delete(podIP, source.Kubernetes)
	}
	if len(errs) != 0 {
		return skipped, errors.New(strings.Join(errs, ", "))
	}

	return skipped, nil
}

func (k *K8sWatcher) DeleteHostPortMapping(pod *corev1.Pod, podIPs []string) error {
	if option.Config.DisableK8sServices || !option.Config.EnableHostPort {
		return nil
	}

	svcs := genServiceMappings(pod, podIPs)
	if len(svcs) == 0 {
		return nil
	}

	logger := log.WithFields(log.Fields{
		logfields.K8sPodName:   pod.ObjectMeta.Name,
		logfields.K8sNamespace: pod.ObjectMeta.Namespace,
		"podIPs":               podIPs,
		"hostIP":               pod.Status.HostIP,
	})

	hostIP := net.ParseIP(pod.Status.HostIP)
	if hostIP == nil {
		logger.Error("Cannot delete HostPort service for the podIP due to missing hostIP")
		return fmt.Errorf("no/invalid HostIP: %s", pod.Status.HostIP)
	}

	for _, dpSvc := range svcs {
		if _, err := k.serviceBPFManager.DeleteService(dpSvc.Frontend.L3n4Addr); err != nil {
			logger.WithError(err).Error("Error while deleting service in LB map")
			return err
		}
	}

	return nil
}

// get frontend HostIP:HostPot -> backend podIP:ContainerPort
func genServiceMappings(pod *corev1.Pod, podIPs []string) []loadbalancer.SVC {
	var svcs []loadbalancer.SVC
	for _, c := range pod.Spec.Containers {
		for _, p := range c.Ports {
			if p.HostPort == 0 {
				continue
			}

			feIP := net.ParseIP(p.HostIP)
			if feIP == nil {
				feIP = net.ParseIP(pod.Status.HostIP)
			}
			proto, err := loadbalancer.NewL4Type(string(p.Protocol))
			if err != nil {
				continue
			}

			fe := loadbalancer.L3n4AddrID{
				L3n4Addr: loadbalancer.L3n4Addr{
					IP: feIP,
					L4Addr: loadbalancer.L4Addr{
						Protocol: proto,
						Port:     uint16(p.HostPort),
					},
				},
				ID: loadbalancer.ID(0),
			}

			bes := make([]loadbalancer.Backend, 0, len(podIPs))
			for _, podIP := range podIPs {
				be := loadbalancer.Backend{
					L3n4Addr: loadbalancer.L3n4Addr{
						IP: net.ParseIP(podIP),
						L4Addr: loadbalancer.L4Addr{
							Protocol: proto,
							Port:     uint16(p.ContainerPort),
						},
					},
				}
				bes = append(bes, be)
			}

			svcs = append(svcs,
				loadbalancer.SVC{
					Frontend: fe,
					Backends: bes,
					Type:     loadbalancer.SVCTypeHostPort,
					// We don't have the node name available here, but in
					// any case in the BPF data path we drop any potential
					// non-local backends anyway (which should never exist).
					TrafficPolicy: loadbalancer.SVCTrafficPolicyCluster,
				})
		}
	}

	return svcs
}

// get pod ips from pod.status
func validIPs(podStatus corev1.PodStatus) ([]string, error) {
	if len(podStatus.PodIPs) == 0 && len(podStatus.PodIP) == 0 {
		return nil, fmt.Errorf("empty PodIPs")
	}

	// make it a set first to avoid repeated IP addresses
	ipsMap := make(map[string]struct{}, 1+len(podStatus.PodIPs))
	if podStatus.PodIP != "" {
		ipsMap[podStatus.PodIP] = struct{}{}
	}
	for _, podIP := range podStatus.PodIPs {
		if podIP.IP != "" {
			ipsMap[podIP.IP] = struct{}{}
		}
	}

	ips := make([]string, 0, len(ipsMap))
	for ipStr := range ipsMap {
		ips = append(ips, ipStr)
	}
	sort.Strings(ips)
	return ips, nil
}
