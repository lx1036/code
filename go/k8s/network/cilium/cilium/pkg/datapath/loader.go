package datapath

import (
	"context"

	"github.com/cilium/cilium/pkg/datapath/loader/metrics"
)

// Loader is an interface to abstract out loading of datapath programs.
type Loader interface {
	CallsMapPath(id uint16) string
	CompileAndLoad(ctx context.Context, ep Endpoint, stats *metrics.SpanStat) error
	CompileOrLoad(ctx context.Context, ep Endpoint, stats *metrics.SpanStat) error
	ReloadDatapath(ctx context.Context, ep Endpoint, stats *metrics.SpanStat) error
	EndpointHash(cfg EndpointConfiguration) (string, error)
	DeleteDatapath(ctx context.Context, ifName, direction string) error
	Unload(ep Endpoint)
	Reinitialize(ctx context.Context, o BaseProgramOwner, deviceMTU int, iptMgr IptablesManager, p Proxy, r RouteReserver) error
}
