package hostpath

import (
	csicommon "k8s-lx1036/k8s/storage/csi/csi-drivers/pkg/csi-common"
)

type identityServer struct {
	*csicommon.DefaultIdentityServer
}

func NewIdentityServer(d *csicommon.CSIDriver) *identityServer {
	return &identityServer{
		DefaultIdentityServer: csicommon.NewDefaultIdentityServer(d),
	}
}
