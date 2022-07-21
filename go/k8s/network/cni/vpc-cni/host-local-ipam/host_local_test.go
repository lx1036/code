package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"

	"k8s-lx1036/k8s/network/cni/vpc-cni/host-local-ipam/pkg/store/disk"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	types100 "github.com/containernetworking/cni/pkg/types/100"
	"github.com/containernetworking/plugins/pkg/testutils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("host-local Operations", func() {
	var tmpDir string
	const (
		ifname string = "eth0"
		nspath string = "/some/where"
	)

	BeforeEach(func() {
		var err error
		tmpDir, err = ioutil.TempDir(".", "host-local_test")
		Expect(err).NotTo(HaveOccurred())
		tmpDir = filepath.ToSlash(tmpDir)
	})

	AfterEach(func() {
		os.RemoveAll(tmpDir)
	})

	for _, ver := range testutils.AllSpecVersions {
		// Redefine ver inside for scope so real value is picked up by each dynamically defined It()
		// See Gingkgo's "Patterns for dynamically generating tests" documentation.
		ver := ver

		It(fmt.Sprintf("[%s] allocates and releases addresses with ADD/DEL", ver), func() {
			err := ioutil.WriteFile(filepath.Join(tmpDir, "resolv.conf"), []byte("nameserver 192.0.2.3"), 0644)
			Expect(err).NotTo(HaveOccurred())

			conf := fmt.Sprintf(`{
				"cniVersion": "%s",
				"name": "mynet",
				"type": "ipvlan",
				"master": "foo0",
				"ipam": {
					"type": "host-local",
					"dataDir": "%s",
					"resolvConf": "%s/resolv.conf",
					"ranges": [
						{"subnet": "10.1.2.0/24"}, 
						{"subnet": "10.2.2.0/24"}
					],
					"routes": [
						{"dst": "0.0.0.0/0"},
						{"dst": "192.168.0.0/16", "gw": "1.1.1.1"}
					]
				}
			}`, ver, tmpDir, tmpDir)

			args := &skel.CmdArgs{
				ContainerID: "dummy",
				Netns:       nspath,
				IfName:      ifname,
				StdinData:   []byte(conf),
			}

			// Allocate the IP
			r, raw, err := testutils.CmdAddWithArgs(args, func() error {
				return cmdAdd(args)
			})
			Expect(err).NotTo(HaveOccurred())
			if testutils.SpecVersionHasIPVersion(ver) {
				Expect(strings.Index(string(raw), "\"version\":")).Should(BeNumerically(">", 0))
			}

			result, err := types100.GetResult(r)
			Expect(err).NotTo(HaveOccurred())

			// Gomega is cranky about slices with different caps
			Expect(*result.IPs[0]).To(Equal(
				types100.IPConfig{
					Address: mustCIDR("10.1.2.2/24"),
					Gateway: net.ParseIP("10.1.2.1"),
				}))
			Expect(len(result.IPs)).To(Equal(1))

			for _, expectedRoute := range []*types.Route{
				{Dst: mustCIDR("0.0.0.0/0"), GW: nil},
				{Dst: mustCIDR("192.168.0.0/16"), GW: net.ParseIP("1.1.1.1")},
			} {
				found := false
				for _, foundRoute := range result.Routes {
					if foundRoute.String() == expectedRoute.String() {
						found = true
						break
					}
				}
				Expect(found).To(BeTrue())
			}

			ipFilePath1 := filepath.Join(tmpDir, "mynet", "10.1.2.2")
			contents, err := ioutil.ReadFile(ipFilePath1)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents)).To(Equal(args.ContainerID + disk.LineBreak + ifname))

			lastFilePath1 := filepath.Join(tmpDir, "mynet", "last_reserved_ip.0")
			contents, err = ioutil.ReadFile(lastFilePath1)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents)).To(Equal("10.1.2.2"))

			// Release the IP
			err = testutils.CmdDelWithArgs(args, func() error {
				return cmdDel(args)
			})
			Expect(err).NotTo(HaveOccurred())

			_, err = os.Stat(ipFilePath1)
			Expect(err).To(HaveOccurred())
		})
	}
})

func mustCIDR(s string) net.IPNet {
	ip, n, err := net.ParseCIDR(s)
	n.IP = ip
	if err != nil {
		Fail(err.Error())
	}

	return *n
}
