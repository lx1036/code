package cmd

import (
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/spanstat"
)

type bootstrapStatistics struct {
	overall         spanstat.SpanStat
	earlyInit       spanstat.SpanStat
	k8sInit         spanstat.SpanStat
	restore         spanstat.SpanStat
	healthCheck     spanstat.SpanStat
	ingressIPAM     spanstat.SpanStat
	initAPI         spanstat.SpanStat
	initDaemon      spanstat.SpanStat
	cleanup         spanstat.SpanStat
	bpfBase         spanstat.SpanStat
	clusterMeshInit spanstat.SpanStat
	ipam            spanstat.SpanStat
	daemonInit      spanstat.SpanStat
	mapsInit        spanstat.SpanStat
	workloadsInit   spanstat.SpanStat
	proxyStart      spanstat.SpanStat
	fqdn            spanstat.SpanStat
	enableConntrack spanstat.SpanStat
	kvstore         spanstat.SpanStat
}
