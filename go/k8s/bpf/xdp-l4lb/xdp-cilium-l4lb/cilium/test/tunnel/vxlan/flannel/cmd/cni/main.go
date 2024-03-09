package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/containernetworking/cni/pkg/invoke"
	"k8s.io/klog/v2"
	"runtime"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/version"
	bv "github.com/containernetworking/plugins/pkg/utils/buildversion"
)

func init() {
	// this ensures that main runs only on main thread (thread group leader).
	// since namespace ops (unshare, setns) are done for a single thread, we
	// must ensure that the goroutine does not jump from OS thread to thread
	runtime.LockOSThread()
}

func main() {
	skel.PluginMain(cmdAdd, cmdCheck, cmdDel, version.All, bv.BuildString("flannel-cni"))
}

/*
NetConf

	{
	      "name": "cbr0",
	      "cniVersion": "0.3.1",
	      "plugins": [
	        {
	          "type": "flannel",
	          "delegate": {
	            "hairpinMode": true,
	            "isDefaultGateway": true
	          }
	        },
	        {
	          "type": "portmap",
	          "capabilities": {
	            "portMappings": true
	          }
	        }
	      ]
	    }
*/
type NetConf struct {
	types.NetConf

	// IPAM field "replaces" that of types.NetConf which is incomplete
	IPAM          map[string]interface{} `json:"ipam,omitempty"`
	SubnetFile    string                 `json:"subnetFile"`
	DataDir       string                 `json:"dataDir"`
	Delegate      map[string]interface{} `json:"delegate"`
	RuntimeConfig map[string]interface{} `json:"runtimeConfig,omitempty"`
}

const (
	/* /run/flannel/subnet.env file that looks like this
	FLANNEL_NETWORK=10.1.0.0/16
	FLANNEL_SUBNET=10.1.17.1/24
	FLANNEL_MTU=1472
	FLANNEL_IPMASQ=true
	*/
	defaultSubnetFile = "/run/flannel/subnet.env" // INFO: 该文件由 daemon 去写

	defaultDataDir = "/var/lib/cni/flannel"
)

func loadConf(bytes []byte) (*NetConf, error) {
	n := &NetConf{
		SubnetFile: defaultSubnetFile,
		DataDir:    defaultDataDir,
	}
	if err := json.Unmarshal(bytes, n); err != nil {
		return nil, fmt.Errorf("failed to load netconf: %v", err)
	}

	return n, nil
}

func cmdAdd(args *skel.CmdArgs) error {
	conf, err := loadConf(args.StdinData)
	if err != nil {
		return err
	}
	subnetEnvVar, err := loadFlannelSubnetEnv(conf.SubnetFile)
	if err != nil {
		return err
	}

	if conf.Delegate == nil {
		conf.Delegate = make(map[string]interface{})
	} else {
		if hasKey(conf.Delegate, "type") && !isString(conf.Delegate["type"]) {
			return fmt.Errorf("'delegate' dictionary, if present, must have (string) 'type' field")
		}
		if hasKey(conf.Delegate, "name") {
			return fmt.Errorf("'delegate' dictionary must not have 'name' field, it'll be set by flannel daemon")
		}
		if hasKey(conf.Delegate, "ipam") {
			return fmt.Errorf("'delegate' dictionary must not have 'ipam' field, it'll be set by flannel daemon")
		}
	}

	if conf.RuntimeConfig != nil {
		conf.Delegate["runtimeConfig"] = conf.RuntimeConfig
	}
	conf.Delegate["name"] = conf.Name
	if !hasKey(conf.Delegate, "type") {
		conf.Delegate["type"] = "bridge" // 默认是 bridge cni
	}
	if !hasKey(conf.Delegate, "ipMasq") {
		// if flannel is not doing ipmasq, we should
		ipmasq := !*subnetEnvVar.ipmasq
		conf.Delegate["ipMasq"] = ipmasq
	}
	if !hasKey(conf.Delegate, "mtu") {
		mtu := subnetEnvVar.mtu
		conf.Delegate["mtu"] = mtu
	}
	if conf.Delegate["type"].(string) == "bridge" {
		if !hasKey(conf.Delegate, "isGateway") {
			conf.Delegate["isGateway"] = true
		}
	}
	if conf.CNIVersion != "" {
		conf.Delegate["cniVersion"] = conf.CNIVersion
	}
	ipam, err := getDelegateIPAM(conf, subnetEnvVar)
	if err != nil {
		return fmt.Errorf("failed to assemble Delegate IPAM: %w", err)
	}
	conf.Delegate["ipam"] = ipam
	klog.Infof(fmt.Sprintf("%+v", conf.Delegate))

	return delegateAdd(args.ContainerID, conf.DataDir, conf.Delegate)
}

func getDelegateIPAM(n *NetConf, subnetEnvVar *subnetEnv) (map[string]interface{}, error) {
	ipam := n.IPAM
	if ipam == nil {
		ipam = map[string]interface{}{}
	}

	if !hasKey(ipam, "type") {
		ipam["type"] = "host-local"
	}

	// subnet
	var rangesSlice [][]map[string]interface{}
	if subnetEnvVar.subnet != nil && subnetEnvVar.subnet.String() != "" {
		rangesSlice = append(rangesSlice, []map[string]interface{}{
			{"subnet": subnetEnvVar.subnet.String()},
		},
		)
	}
	ipam["ranges"] = rangesSlice

	//routes
	routes, err := getIPAMRoutes(n)
	if err != nil {
		return nil, fmt.Errorf("failed to read IPAM routes: %w", err)
	}
	if subnetEnvVar.network != nil {
		routes = append(routes, types.Route{Dst: *subnetEnvVar.network})
	}
	ipam["routes"] = routes

	return ipam, nil
}

func getIPAMRoutes(n *NetConf) ([]types.Route, error) {
	routes := []types.Route{}
	if n.IPAM != nil && hasKey(n.IPAM, "routes") {
		buf, _ := json.Marshal(n.IPAM["routes"])
		if err := json.Unmarshal(buf, &routes); err != nil {
			return routes, fmt.Errorf("failed to parse ipam.routes: %w", err)
		}
	}
	return routes, nil
}

func delegateAdd(containerID, dataDir string, netconf map[string]interface{}) error {
	netconfBytes, err := json.Marshal(netconf)
	if err != nil {
		return fmt.Errorf("error serializing delegate netconf: %v", err)
	}
	klog.Infof(fmt.Sprintf("delegate netconf: %s", netconfBytes))

	// save the rendered netconf for cmdDel
	if err = saveScratchNetConf(containerID, dataDir, netconfBytes); err != nil {
		return err
	}

	// netconf["type"].(string) = bridge, 用的 bridge cni plugin
	result, err := invoke.DelegateAdd(context.TODO(), netconf["type"].(string), netconfBytes, nil)
	if err != nil {
		err = fmt.Errorf("failed to delegate add: %w", err)
		return err
	}

	return result.Print()
}

func cmdDel(args *skel.CmdArgs) error {

}

func cmdCheck(args *skel.CmdArgs) error {
	// TODO: implement
	return nil
}
