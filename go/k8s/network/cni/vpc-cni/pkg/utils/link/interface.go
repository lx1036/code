package link

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/vishvananda/netlink"
)

// GetDeviceNumberByMac get interface device number by mac address
func GetDeviceNumberByMac(mac string) (int32, error) {
	linkList, err := netlink.LinkList()
	if err != nil {
		return 0, errors.Wrapf(err, "error get link list from netlink")
	}

	for _, link := range linkList {
		// ignore virtual nic type. eg. ipvlan veth bridge
		if _, ok := link.(*netlink.Device); !ok {
			continue
		}
		if link.Attrs().HardwareAddr.String() == mac {
			return int32(link.Attrs().Index), nil
		}
	}

	return 0, fmt.Errorf("can't found dev by mac %s", mac)
}
