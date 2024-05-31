package mac

import (
    "fmt"
    "net"
)

// MAC address is an net.HardwareAddr encapsulation to force cilium to only use MAC-48.
type MAC net.HardwareAddr

// ParseMAC parses s only as an IEEE 802 MAC-48.
func ParseMAC(s string) (MAC, error) {
    ha, err := net.ParseMAC(s)
    if err != nil {
        return nil, err
    }
    if len(ha) != 6 {
        return nil, fmt.Errorf("invalid MAC address %s", s)
    }

    return MAC(ha), nil
}
