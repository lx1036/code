package controller

import (
	"crypto/sha256"
	"encoding/base32"
	"strconv"
)

const (
	kubeNetworkPolicyChainPrefix = "KUBE-NWPLCY-"
	kubeDestinationIPSetPrefix   = "KUBE-DST-"
	kubeSourceIPSetPrefix        = "KUBE-SRC-"
)

func networkPolicyChainName(namespace, policyName string) string {
	data := namespace + policyName
	return encode(kubeNetworkPolicyChainPrefix, data)
}

func policyDstPodIPSetName(namespace, policyName string) string {
	data := namespace + policyName
	return encode(kubeDestinationIPSetPrefix, data)
}

func policySrcPodIPSetName(namespace, policyName string) string {
	data := namespace + policyName
	return encode(kubeSourceIPSetPrefix, data)
}

func policyIndexedSrcPodIPSetName(namespace, policyName string, ingressRuleNo int) string {
	data := namespace + policyName + "ingressrule" + strconv.Itoa(ingressRuleNo) + "pod"
	return encode(kubeSourceIPSetPrefix, data)
}

func policyIndexedDstPodIPSetName(namespace, policyName string, egressRuleNo int) string {
	data := namespace + policyName + "egressrule" + strconv.Itoa(egressRuleNo) + "pod"
	return encode(kubeDestinationIPSetPrefix, data)
}

func policyIndexedIngressNamedPortIPSetName(namespace, policyName string, ingressRuleNo, namedPortNo int) string {
	data := namespace + policyName + "ingressrule" + strconv.Itoa(ingressRuleNo) + strconv.Itoa(namedPortNo) + "namedport"
	return encode(kubeDestinationIPSetPrefix, data)
}

func policyIndexedEgressNamedPortIPSetName(namespace, policyName string, egressRuleNo, namedPortNo int) string {
	data := namespace + policyName + "egressrule" + strconv.Itoa(egressRuleNo) + strconv.Itoa(namedPortNo) + "namedport"
	return encode(kubeDestinationIPSetPrefix, data)
}

func policyIndexedSourceIPBlockIPSetName(namespace, policyName string, ingressRuleNo int) string {
	data := namespace + policyName + "ingressrule" + strconv.Itoa(ingressRuleNo) + "ipblock"
	return encode(kubeSourceIPSetPrefix, data)
}

func policyIndexedDstIPBlockIPSetName(namespace, policyName string, egressRuleNo int) string {
	data := namespace + policyName + "egressrule" + strconv.Itoa(egressRuleNo) + "ipblock"
	return encode(kubeDestinationIPSetPrefix, data)
}

func encode(prefix, data string) string {
	hash := sha256.Sum256([]byte(data))
	encoded := base32.StdEncoding.EncodeToString(hash[:])
	return prefix + encoded[:16]
}

func getIPsFromPods(pods []podInfo) []string {
	ips := make([]string, len(pods))
	for idx, pod := range pods {
		ips[idx] = pod.ip
	}
	return ips
}
