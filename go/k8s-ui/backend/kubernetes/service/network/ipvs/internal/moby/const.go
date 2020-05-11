// +build linux

package moby

const (
	genlCtrlID = 0x10
)

// GENL control commands
const (
	genlCtrlCmdUnspec uint8 = iota
	genlCtrlCmdNewFamily
	genlCtrlCmdDelFamily
	genlCtrlCmdGetFamily
)

// GENL family attributes
const (
	genlCtrlAttrUnspec int = iota
	genlCtrlAttrFamilyID
	genlCtrlAttrFamilyName
)

// IPVS genl commands
const (
	ipvsCmdUnspec uint8 = iota
	ipvsCmdNewService
	ipvsCmdSetService
	ipvsCmdDelService
	ipvsCmdGetService
	ipvsCmdNewDestination
	ipvsCmdSetDestination
	ipvsCmdDelDestination
	ipvsCmdGetDestination
	ipvsCmdNewDaemon
	ipvsCmdDelDaemon
	ipvsCmdGetDaemon
	ipvsCmdSetConfig
	ipvsCmdGetConfig
	ipvsCmdSetInfo
	ipvsCmdGetInfo
	ipvsCmdZero
	ipvsCmdFlush
)

// Attributes used in the first level of commands
const (
	ipvsCmdAttrUnspec int = iota
	ipvsCmdAttrService
	ipvsCmdAttrDest
	ipvsCmdAttrDaemon
	ipvsCmdAttrTimeoutTCP
	ipvsCmdAttrTimeoutTCPFin
	ipvsCmdAttrTimeoutUDP
)

// Attributes used to describe a service. Used inside nested attribute
// ipvsCmdAttrService
const (
	ipvsSvcAttrUnspec int = iota
	ipvsSvcAttrAddressFamily
	ipvsSvcAttrProtocol
	ipvsSvcAttrAddress
	ipvsSvcAttrPort
	ipvsSvcAttrFWMark
	ipvsSvcAttrSchedName
	ipvsSvcAttrFlags
	ipvsSvcAttrTimeout
	ipvsSvcAttrNetmask
	ipvsSvcAttrStats
	ipvsSvcAttrPEName
)
