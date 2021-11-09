package routing

import (
	"errors"
	"fmt"

	"github.com/coreos/go-iptables/iptables"
	"k8s.io/klog/v2"
)

// INFO: set up MASQUERADE rule so that egress traffic from the pods gets masqueraded to node's IP
//  意思是从 pod 出去的流量可以伪装成 node ip，就是 snat
//  `iptables-t nat -A POSTROUTING -s 10.8.0.0/255.255.255.0 -o eth0 -j MASQUERADE`: 如此配置的话，不用指定SNAT的目标ip了，不管现在eth0的出口获得了怎样的动态ip，
//   MASQUERADE会自动读取eth0现在的ip地址然后做SNAT出去，这样就实现了很好的动态SNAT地址转换。

const (
	PodSubnetsIPSetName = "kube-router-pod-subnets"
	NodeAddrsIPSetName  = "kube-router-node-ips"
)

var (
	podEgressArgs4 = []string{"-m", "set", "--match-set", PodSubnetsIPSetName, "src",
		"-m", "set", "!", "--match-set", PodSubnetsIPSetName, "dst",
		"-m", "set", "!", "--match-set", NodeAddrsIPSetName, "dst",
		"-j", "MASQUERADE"}
	podEgressArgs6 = []string{"-m", "set", "--match-set", "inet6:" + PodSubnetsIPSetName, "src",
		"-m", "set", "!", "--match-set", "inet6:" + PodSubnetsIPSetName, "dst",
		"-m", "set", "!", "--match-set", "inet6:" + NodeAddrsIPSetName, "dst",
		"-j", "MASQUERADE"}
	podEgressArgsBad4 = [][]string{{"-m", "set", "--match-set", PodSubnetsIPSetName, "src",
		"-m", "set", "!", "--match-set", PodSubnetsIPSetName, "dst",
		"-j", "MASQUERADE"}}
	podEgressArgsBad6 = [][]string{{"-m", "set", "--match-set", "inet6:" + PodSubnetsIPSetName, "src",
		"-m", "set", "!", "--match-set", "inet6:" + PodSubnetsIPSetName, "dst",
		"-j", "MASQUERADE"}}
)

func (controller *NetworkRoutingController) newIptablesCmdHandler() (*iptables.IPTables, error) {
	if controller.isIpv6 {
		return iptables.NewWithProtocol(iptables.ProtocolIPv6)
	}

	return iptables.NewWithProtocol(iptables.ProtocolIPv4)
}

// `iptables -t nat -A POSTROUTING -m set --match-set kube-router-pod-subnets src -m set ! --match-set kube-router-pod-subnets dst -m set ! --match-set kube-router-node-ips dst -j MASQUERADE`
func (controller *NetworkRoutingController) createPodEgressRule() error {
	iptablesCmdHandler, err := controller.newIptablesCmdHandler()
	if err != nil {
		return errors.New("Failed create iptables handler:" + err.Error())
	}

	podEgressArgs := podEgressArgs4
	if controller.isIpv6 {
		podEgressArgs = podEgressArgs6
	}
	if iptablesCmdHandler.HasRandomFully() {
		podEgressArgs = append(podEgressArgs, "--random-fully")
	}

	err = iptablesCmdHandler.AppendUnique("nat", "POSTROUTING", podEgressArgs...)
	if err != nil {
		return errors.New("Failed to add iptables rule to masquerade outbound traffic from pods: " +
			err.Error() + "External connectivity will not work.")

	}

	klog.V(1).Infof("Added iptables rule to masquerade outbound traffic from pods.")
	return nil
}

func (controller *NetworkRoutingController) deletePodEgressRule() error {
	iptablesCmdHandler, err := controller.newIptablesCmdHandler()
	if err != nil {
		return errors.New("Failed create iptables handler:" + err.Error())
	}

	podEgressArgs := podEgressArgs4
	if controller.isIpv6 {
		podEgressArgs = podEgressArgs6
	}
	if iptablesCmdHandler.HasRandomFully() {
		podEgressArgs = append(podEgressArgs, "--random-fully")
	}

	exists, err := iptablesCmdHandler.Exists("nat", "POSTROUTING", podEgressArgs...)
	if err != nil {
		return errors.New("Failed to lookup iptables rule to masquerade outbound traffic from pods: " + err.Error())
	}

	if exists {
		err = iptablesCmdHandler.Delete("nat", "POSTROUTING", podEgressArgs...)
		if err != nil {
			return errors.New("Failed to delete iptables rule to masquerade outbound traffic from pods: " +
				err.Error() + ". Pod egress might still work...")
		}
		klog.Infof("Deleted iptables rule to masquerade outbound traffic from pods.")
	}

	return nil
}

func (controller *NetworkRoutingController) deleteBadPodEgressRules() error {
	iptablesCmdHandler, err := controller.newIptablesCmdHandler()
	if err != nil {
		return errors.New("Failed create iptables handler:" + err.Error())
	}
	podEgressArgsBad := podEgressArgsBad4
	if controller.isIpv6 {
		podEgressArgsBad = podEgressArgsBad6
	}

	// If random fully is supported remove the original rule as well
	if iptablesCmdHandler.HasRandomFully() {
		if !controller.isIpv6 {
			podEgressArgsBad = append(podEgressArgsBad, podEgressArgs4)
		} else {
			podEgressArgsBad = append(podEgressArgsBad, podEgressArgs6)
		}
	}

	for _, args := range podEgressArgsBad {
		exists, err := iptablesCmdHandler.Exists("nat", "POSTROUTING", args...)
		if err != nil {
			return fmt.Errorf("failed to lookup iptables rule: %s", err.Error())
		}

		if exists {
			err = iptablesCmdHandler.Delete("nat", "POSTROUTING", args...)
			if err != nil {
				return fmt.Errorf("failed to delete old/bad iptables rule to masquerade outbound traffic "+
					"from pods: %s. Pod egress might still work, or bugs may persist after upgrade", err)
			}
			klog.Infof("Deleted old/bad iptables rule to masquerade outbound traffic from pods.")
		}
	}

	return nil
}
