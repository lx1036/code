package controller

import (
	"context"
	"errors"
	"fmt"
	"github.com/moby/ipvs"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"net"
	"os"
	"sync"
	"syscall"
	"time"
)

// INFO: @see https://github.com/cloudnativelabs/kube-router/blob/master/pkg/controllers/proxy/network_services_controller.go
//  https://github.com/kubernetes/kubernetes/blob/master/pkg/proxy/ipvs/proxier.go

var (
	NodeIP net.IP
)

type NetworkServiceController struct {
	sync.Mutex

	ln *linuxNetworking

	nodeIP   net.IP
	nodeName string

	svcLister   cache.Indexer
	epLister    cache.Indexer
	serviceMap  serviceInfoMap
	endpointMap endpointInfoMap

	syncChan chan struct{}
	stopCh   chan struct{}
}

func NewNetworkPolicyController(
	clientset kubernetes.Interface,
	svcInformer cache.SharedIndexInformer,
	epInformer cache.SharedIndexInformer,
) (*NetworkServiceController, error) {
	ln, err := newLinuxNetworking()
	if err != nil {
		return nil, err
	}

	c := &NetworkServiceController{
		ln: ln,

		svcLister: svcInformer.GetIndexer(),
		epLister:  epInformer.GetIndexer(),

		syncChan: make(chan struct{}, 2), // buffer chan，因为 service 互不影响，channel item 可以多个, @see NetworkPolicyController
	}

	nodeName := os.Getenv("NODE_NAME")
	if nodeName != "" {
		node, err := clientset.CoreV1().Nodes().Get(context.Background(), nodeName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		nodeIP, err := GetNodeIP(node)
		if err != nil {
			return nil, err
		}
		c.nodeIP = nodeIP
		c.nodeName = node.Name
	} else {
		return nil, fmt.Errorf("NODE_NAME env is empty")
	}

	svcInformer.AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			service, ok := obj.(*corev1.Service)
			if !ok {
				return false
			}
			if IsHeadlessService(service) || IsExternalNameService(service) { // skip headless service
				return false
			}
			return true
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				service, ok := obj.(*corev1.Service)
				if !ok {
					return
				}
				c.onServiceUpdate(service)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				service, ok := newObj.(*corev1.Service)
				if !ok {
					return
				}
				c.onServiceUpdate(service)
			},
			DeleteFunc: func(obj interface{}) {
				service, ok := obj.(*corev1.Service)
				if !ok {
					tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
					if !ok {
						klog.Errorf("unexpected object type: %v", obj)
						return
					}
					if service, ok = tombstone.Obj.(*corev1.Service); !ok {
						klog.Errorf("unexpected object type: %v", obj)
						return
					}
				}

				c.onServiceUpdate(service)
			},
		},
	})

	epInformer.AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			ep, ok := obj.(*corev1.Endpoints)
			if !ok {
				return false
			}
			if isEndpointsForLeaderElection(ep) {
				return false
			}

			return true
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				ep, ok := obj.(*corev1.Endpoints)
				if !ok {
					return
				}

				c.onEndpointUpdate(ep)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				ep, ok := newObj.(*corev1.Endpoints)
				if !ok {
					return
				}
				c.onEndpointUpdate(ep)
			},
			DeleteFunc: func(obj interface{}) {
				ep, ok := obj.(*corev1.Endpoints)
				if !ok {
					tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
					if !ok {
						klog.Errorf("unexpected object type: %v", obj)
						return
					}
					if ep, ok = tombstone.Obj.(*corev1.Endpoints); !ok {
						klog.Errorf("unexpected object type: %v", obj)
						return
					}
				}

				c.onEndpointUpdate(ep)
			},
		},
	})

	return c, nil
}

func (controller *NetworkServiceController) Run() {
	t := time.NewTicker(controller.syncPeriod)
	defer t.Stop()

	for {
		select {
		case <-controller.stopCh:
			return

		case <-controller.syncChan:
			// We call the component pieces of doSync() here because for methods that send this on the channel they
			// have already done expensive pieces of the doSync() method like building service and endpoint info
			// and we don't want to duplicate the effort, so this is a slimmer version of doSync()
			controller.Lock()
			controller.syncService(controller.serviceMap, controller.endpointMap)
			controller.syncHairpinIptablesRules()

			controller.Unlock()

		case <-t.C:
			controller.syncAll()
		}
	}
}

func (controller *NetworkServiceController) sync() {
	select {
	case controller.syncChan <- struct{}{}:
	default:
		klog.Infof("Already pending sync, dropping request")
	}
}

// INFO: 因为 delete 事件也是 onServiceUpdate() 使用的 svcLister.List()，没有考虑 delete 事件，所以需要整体全部捞一遍来 cleanup
func (controller *NetworkServiceController) syncService(serviceMap serviceInfoMap, endpointMap endpointInfoMap) {
	err := controller.syncClusterIPService(serviceMap, endpointMap)
	controller.syncNodePortService(serviceMap, endpointMap)

	//controller.cleanup

}

// INFO: (1) add ipvs virtual server
//  (2) assign clusterIP to dummy interface, add route in local table
//  (3) add/delete ipvs real server
func (controller *NetworkServiceController) syncClusterIPService(serviceMap serviceInfoMap, endpointMap endpointInfoMap) error {
	// (2) assign clusterIP to dummy interface, add route in local table
	dummyVipInterface, err := controller.ln.EnsureDummyDevice()
	if err != nil {
		return errors.New("Failed creating dummy interface: " + err.Error())
	}

	for svcID, svcInfo := range serviceMap {
		endpoints := endpointMap[svcID]

		// (1)add ipvs service virtual server
		if err := controller.ln.AddOrUpdateVirtualServer(*svcInfo); err != nil {
			klog.Errorf(fmt.Sprintf("[syncClusterIPService]AddOrUpdateVirtualServer for %s/%s err:%v", svcInfo.namespace, svcInfo.name, err))
			continue
		}
		ipvsSvc, err := controller.ln.GetVirtualServer(*svcInfo)
		if err != nil {
			klog.Errorf(fmt.Sprintf("failed to get ipvs service %s/%s err: %v", svcInfo.namespace, svcInfo.name, err))
			continue
		}

		// (2) assign clusterIP to dummy interface, add route in local table
		err = controller.ln.EnsureAddressBind(dummyVipInterface, svcInfo.address.String(), true)
		if err != nil {
			continue
		}

		// (3) add/delete ipvs real server
		destinations, err := controller.ln.ListRealServer(ipvsSvc)
		if err != nil {
			continue
		}
		oldDst := make(map[string]*ipvs.Destination)
		for _, destination := range destinations {
			key := fmt.Sprintf("%s:%d", destination.Address.String(), destination.Port)
			oldDst[key] = destination
		}
		newDst := make(map[string]endpointInfo)
		for _, endpoint := range endpoints {
			if svcInfo.isLocal && !endpoint.isLocal {
				continue
			}
			key := fmt.Sprintf("%s:%d", endpoint.address.String(), endpoint.port)
			newDst[key] = endpoint
		}
		for key, info := range newDst { // add new ipvs real server
			_, ok := oldDst[key]
			if !ok {
				dst := ipvs.Destination{
					Address:       info.address,
					AddressFamily: syscall.AF_INET,
					Port:          uint16(info.port),
					Weight:        1,
				}
				if err = controller.ln.AddRealServer(ipvsSvc, &dst); err != nil {
					klog.Errorf(fmt.Sprintf("failed to add new ipvs real server err: %v", err))
					continue
				}
			}
		}
		for key, destination := range oldDst { // delete old ipvs real server
			_, ok := newDst[key]
			if !ok {
				if err = controller.ln.DelRealServer(ipvsSvc, destination); err != nil {
					klog.Errorf(fmt.Sprintf("failed to add new ipvs real server err: %v", err))
					continue
				}
			}
		}
	}

	return nil
}

// GetNodeIP returns the most valid external facing IP address for a node.
// Order of preference:
// 1. NodeInternalIP
// 2. NodeExternalIP (Only set on cloud providers usually)
func GetNodeIP(node *corev1.Node) (net.IP, error) {
	addresses := node.Status.Addresses
	addressMap := make(map[corev1.NodeAddressType][]corev1.NodeAddress)
	for i := range addresses {
		addressMap[addresses[i].Type] = append(addressMap[addresses[i].Type], addresses[i])
	}
	if addresses, ok := addressMap[corev1.NodeInternalIP]; ok {
		return net.ParseIP(addresses[0].Address), nil
	}
	if addresses, ok := addressMap[corev1.NodeExternalIP]; ok {
		return net.ParseIP(addresses[0].Address), nil
	}
	return nil, fmt.Errorf("host IP unknown")
}
