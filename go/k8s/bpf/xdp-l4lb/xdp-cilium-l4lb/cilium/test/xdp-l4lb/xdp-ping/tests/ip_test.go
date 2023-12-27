package tests

import (
	"encoding/binary"
	"fmt"
	"github.com/sirupsen/logrus"
	"net"
	"os/exec"
	"testing"
)

func TestIPToInt(test *testing.T) {
	ip := "10.1.1.100"
	netip := net.ParseIP(ip)
	ip4 := netip.To4()
	ui32 := uint32(ip4[0])<<24 | uint32(ip4[1])<<16 | uint32(ip4[2])<<8 | uint32(ip4[3])
	fmt.Println(ui32) // 167838052

	ui32_2 := binary.BigEndian.Uint32(ip4)
	fmt.Println(ui32_2) // 167838052

	ui32_3 := binary.LittleEndian.Uint32(ip4)
	fmt.Println(ui32_3) // 1677787402
}

func TestPing(test *testing.T) {
	err := exec.Command("ping", "-c", "10", "-I", "veth1", "10.1.1.100").Run()
	if err != nil {
		logrus.Fatal(err)
	}
}
