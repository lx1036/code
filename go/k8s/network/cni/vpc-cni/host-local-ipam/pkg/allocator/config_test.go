package allocator

import (
	"net"

	"github.com/containernetworking/cni/pkg/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("IPAM config", func() {
	It("Should parse a new-style config", func() {
		input := `{
			"cniVersion": "0.3.1",
			"name": "mynet",
			"type": "ipvlan",
			"master": "foo0",
			"ipam": {
				"type": "host-local",
				"ranges": [
					{
						"subnet": "10.1.2.0/24",
						"rangeStart": "10.1.2.9",
						"rangeEnd": "10.1.2.20",
						"gateway": "10.1.2.30"
					},
					{
						"subnet": "10.1.4.0/24"
					}
				]
			}
		}`

		conf, version, err := LoadIPAMConfig([]byte(input), "")
		Expect(err).NotTo(HaveOccurred())
		Expect(version).Should(Equal("0.3.1"))

		Expect(conf).To(Equal(&IPAMConfig{
			Name: "mynet",
			Type: "host-local",
			Ranges: RangeSet{
				{
					RangeStart: net.IP{10, 1, 2, 9},
					RangeEnd:   net.IP{10, 1, 2, 20},
					Gateway:    net.IP{10, 1, 2, 30},
					Subnet: types.IPNet{
						IP:   net.IP{10, 1, 2, 0},
						Mask: net.CIDRMask(24, 32),
					},
				},
				{
					RangeStart: net.IP{10, 1, 4, 1},
					RangeEnd:   net.IP{10, 1, 4, 254},
					Gateway:    net.IP{10, 1, 4, 1},
					Subnet: types.IPNet{
						IP:   net.IP{10, 1, 4, 0},
						Mask: net.CIDRMask(24, 32),
					},
				},
			},
		}))
	})

	Context("Should parse CNI_ARGS env", func() {
		It("without prefix", func() {
			input := `{
				"cniVersion": "0.3.1",
				"name": "mynet",
				"type": "ipvlan",
				"master": "foo0",
				"ipam": {
					"type": "host-local",
					"ranges": [
						{
							"subnet": "10.1.2.0/24",
							"rangeStart": "10.1.2.9",
							"rangeEnd": "10.1.2.20",
							"gateway": "10.1.2.30"
						}
					]
				}
			}`

			envArgs := "IP=10.1.2.10"

			conf, _, err := LoadIPAMConfig([]byte(input), envArgs)
			Expect(err).NotTo(HaveOccurred())
			Expect(conf.IPArgs).To(Equal([]net.IP{{10, 1, 2, 10}}))
		})

		It("with prefix", func() {
			input := `{
				"cniVersion": "0.3.1",
				"name": "mynet",
				"type": "ipvlan",
				"master": "foo0",
				"ipam": {
					"type": "host-local",
					"ranges": [
						{
							"subnet": "10.1.2.0/24",
							"rangeStart": "10.1.2.9",
							"rangeEnd": "10.1.2.20",
							"gateway": "10.1.2.30"
						}
					]
				}
			}`

			envArgs := "IP=10.1.2.11/24"

			conf, _, err := LoadIPAMConfig([]byte(input), envArgs)
			Expect(err).NotTo(HaveOccurred())
			Expect(conf.IPArgs).To(Equal([]net.IP{{10, 1, 2, 11}}))
		})
	})
})
