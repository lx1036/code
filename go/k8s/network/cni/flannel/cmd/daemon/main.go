package main

import (
	"flag"
	"fmt"

	"net"
	"os"
	"path/filepath"

	"k8s-lx1036/k8s/network/cni/flannel/pkg/backend"
	_ "k8s-lx1036/k8s/network/cni/flannel/pkg/backend/vxlan"
	"k8s-lx1036/k8s/network/cni/flannel/pkg/ip"
	"k8s-lx1036/k8s/network/cni/flannel/pkg/iptables"
	"k8s-lx1036/k8s/network/cni/flannel/pkg/subnet"

	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/component-base/logs"
	"k8s.io/klog/v2"
)

var (
	kubeconfig           = flag.String("kubeconfig", "", "absolute path to the kubeconfig file (only needed when running outside of k8s)")
	subnetFile           = flag.String("subnet-file", "/run/flannel/subnet.env", "filename where env variables (subnet, MTU, ... ) will be written to")
	kubeAnnotationPrefix = flag.String("kube-annotation-prefix", "flannel.alpha.coreos.com", `Kubernetes annotation prefix. Can contain single slash "/", otherwise it will be appended at the end.`)
	netConfPath          = flag.String("net-config-path", "/etc/kube-flannel/net-conf.json", "path to the network configuration file")

	ifaceName                 = flag.String("iface", "eth0", "interface to use (IP or name) for inter-host communication. Can be specified multiple times to check each option in order. Returns the first match found.")
	ipMasq                    = flag.Bool("ip-masq", false, "setup IP masquerade rule for traffic destined outside of overlay network")
	iptablesForwardRules      = flag.Bool("iptables-forward-rules", true, "add default accept rules to FORWARD chain in iptables")
	iptablesResyncSeconds     = flag.Int("iptables-resync", 5, "resync period for iptables rules, in seconds")
	setNodeNetworkUnavailable = flag.Bool("set-node-network-unavailable", true, "set NodeNetworkUnavailable after ready")
)

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()

	flag.Parse()

	ctx := genericapiserver.SetupSignalContext()

	subnetMgr, err := subnet.NewSubnetManager(ctx, *kubeconfig, *kubeAnnotationPrefix, *netConfPath, *setNodeNetworkUnavailable)
	if err != nil {
		klog.Fatal(err)
	}
	config, err := subnetMgr.GetNetworkConfig(ctx)
	if err != nil {
		klog.Fatal(err)
	}

	// get eth0 interface
	iface, ifaceAddr := getInterfaceAndAddr(*ifaceName)
	extIface := &backend.ExternalInterface{
		Iface:     iface,
		IfaceAddr: ifaceAddr,
		ExtAddr:   ifaceAddr, // 这里配置的是 publicIP，也就是 nodeIP
	}
	// Create a backend manager then use it to create the backend and register the network with it.
	backendMgr := backend.NewManager(ctx, subnetMgr, extIface)
	vxlanBackend, err := backendMgr.GetBackend(config.BackendType)
	if err != nil {
		klog.Fatal(err)
	}
	vxlanNetwork, err := vxlanBackend.RegisterNetwork(ctx, config)
	if err != nil {
		klog.Fatal(err)
	}
	go vxlanNetwork.Run(ctx)

	// Set up ipMasq if needed
	if *ipMasq {
		if config.EnableIPv4 {
			if err = recycleIPTables(config.Network, vxlanNetwork.Lease()); err != nil {
				klog.Fatal(err)
			}

			go iptables.SetupAndEnsureIP4Tables(iptables.MasqRules(config.Network, vxlanNetwork.Lease()), *iptablesResyncSeconds)
		}
	}

	// Always enables forwarding rules. This is needed for Docker versions >1.13 (https://docs.docker.com/engine/userguide/networking/default_network/container-communication/#container-communication-between-hosts)
	// In Docker 1.12 and earlier, the default FORWARD chain policy was ACCEPT.
	// In Docker 1.13 and later, Docker sets the default policy of the FORWARD chain to DROP.
	if *iptablesForwardRules {
		if config.EnableIPv4 {
			klog.Infof("Changing default FORWARD chain policy to ACCEPT")
			go iptables.SetupAndEnsureIP4Tables(iptables.ForwardRules(config.Network.String()), *iptablesResyncSeconds)
		}
	}

	if err := WriteSubnetFile(*subnetFile, config, *ipMasq, vxlanNetwork); err != nil {
		// Continue, even though it failed.
		klog.Warningf(fmt.Sprintf("Failed to write subnet file: %v", err))
	} else {
		klog.Infof(fmt.Sprintf("Wrote subnet file to %s", *subnetFile))
	}

	<-ctx.Done()
}

func getInterfaceAndAddr(ifname string) (*net.Interface, net.IP) {
	iface, err := net.InterfaceByName(ifname)
	ifaceAddrs, err := ip.GetInterfaceIP4Addrs(iface)
	if err != nil || len(ifaceAddrs) == 0 {
		klog.Fatal(fmt.Sprintf("failed to find IPv4 address for interface %s", iface.Name))
	}

	return iface, ifaceAddrs[0]
}

// WriteSubnetFile 写 env 文件 /run/flannel/subnet.env
func WriteSubnetFile(path string, config *subnet.Config, ipMasq bool, vxlanNetwork backend.Network) error {
	dir, file := filepath.Split(path)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}
	tempFile := filepath.Join(dir, "."+file)
	f, err := os.Create(tempFile)
	if err != nil {
		return err
	}
	if config.EnableIPv4 {
		network := config.Network
		sn := vxlanNetwork.Lease().Subnet
		// Write out the first usable IP by incrementing sn.IP by one
		sn.IncrementIP() // INFO: 因为第一个IP被 vxlan device 使用了，@see vxlanBackend.RegisterNetwork()
		fmt.Fprintf(f, "FLANNEL_NETWORK=%s\n", network)
		fmt.Fprintf(f, "FLANNEL_SUBNET=%s\n", sn)
	}

	fmt.Fprintf(f, "FLANNEL_MTU=%d\n", vxlanNetwork.MTU())
	_, err = fmt.Fprintf(f, "FLANNEL_IPMASQ=%v\n", ipMasq)
	f.Close()
	if err != nil {
		return err
	}

	// rename(2) the temporary file to the desired location so that it becomes
	// atomically visible with the contents
	return os.Rename(tempFile, path)
}
