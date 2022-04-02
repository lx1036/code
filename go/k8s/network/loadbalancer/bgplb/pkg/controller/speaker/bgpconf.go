package speaker

import (
	"context"

	v1 "k8s-lx1036/k8s/network/loadbalancer/bgplb/pkg/apis/bgplb.k9s.io/v1"

	gobgpapi "github.com/osrg/gobgp/v3/api"
	"k8s.io/klog/v2"
)

func (c *SpeakerController) onBGPConfAdd(obj interface{}) {
	bgpConf := obj.(*v1.BgpConf)
	klog.Infof("bgpConf %s/%s was added, enqueuing it for submission", bgpConf.Namespace, bgpConf.Name)

	global, err := bgpConf.Spec.ConvertToGoBGPGlobalConf()
	if err != nil {
		klog.Error(err)
		return
	}

	// stop bgp server firstly
	c.bgpServer.StopBgp(context.TODO(), &gobgpapi.StopBgpRequest{})
	err = c.bgpServer.StartBgp(context.TODO(), &gobgpapi.StartBgpRequest{
		Global: global,
	})
	if err != nil {
		klog.Error(err)
		return
	}
}

func (c *SpeakerController) onBGPConfUpdate(oldObj, newObj interface{}) {

}

func (c *SpeakerController) onBGPConfDelete(obj interface{}) {

}
