package csicommon

import (
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
)

type CSIDriver struct {
	name    string
	nodeID  string
	version string
	cap     []*csi.ControllerServiceCapability
	vc      []*csi.VolumeCapability_AccessMode
}

// Creates a NewCSIDriver object. Assumes vendor version is equal to driver version &
// does not support optional driver plugin info manifest field. Refer to CSI spec for more details.
func NewCSIDriver(name string, v string, nodeID string) *CSIDriver {
	if name == "" {
		glog.Errorf("Driver name missing")
		return nil
	}

	if nodeID == "" {
		glog.Errorf("NodeID missing")
		return nil
	}
	// TODO version format and validation
	if len(v) == 0 {
		glog.Errorf("Version argument missing")
		return nil
	}

	driver := CSIDriver{
		name:    name,
		version: v,
		nodeID:  nodeID,
	}

	return &driver
}

func NewControllerServiceCapability(cap csi.ControllerServiceCapability_RPC_Type) *csi.ControllerServiceCapability {
	return &csi.ControllerServiceCapability{
		Type: &csi.ControllerServiceCapability_Rpc{
			Rpc: &csi.ControllerServiceCapability_RPC{
				Type: cap,
			},
		},
	}
}

func (d *CSIDriver) AddControllerServiceCapabilities(cl []csi.ControllerServiceCapability_RPC_Type) {
	var csc []*csi.ControllerServiceCapability

	for _, c := range cl {
		glog.Infof("Enabling controller service capability: %v", c.String())
		csc = append(csc, NewControllerServiceCapability(c))
	}

	d.cap = csc

	return
}

func NewVolumeCapabilityAccessMode(mode csi.VolumeCapability_AccessMode_Mode) *csi.VolumeCapability_AccessMode {
	return &csi.VolumeCapability_AccessMode{Mode: mode}
}

func (d *CSIDriver) AddVolumeCapabilityAccessModes(vc []csi.VolumeCapability_AccessMode_Mode) []*csi.VolumeCapability_AccessMode {
	var vca []*csi.VolumeCapability_AccessMode
	for _, c := range vc {
		glog.Infof("Enabling volume access mode: %v", c.String())
		vca = append(vca, NewVolumeCapabilityAccessMode(c))
	}
	d.vc = vca
	return vca
}
