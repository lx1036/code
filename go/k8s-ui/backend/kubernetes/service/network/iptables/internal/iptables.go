package internal

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
	"os/exec"
	"strings"
	"sync"
)

const (
	Filter Table = "filter"
	Nat    Table = "nat"
	Mangle Table = "mangle"
	Raw    Table = "raw"

	Prerouting  Chain = "PREROUTING"
	Postrouting Chain = "POSTROUTING"
	Forward     Chain = "FORWARD"
	Input       Chain = "INPUT"
	Output      Chain = "OUTPUT"

	Append  Action = "-A"
	Check   Action = "-C"
	Delete  Action = "-D"
	Insert  Action = "-I"
	Replace Action = "-R"
	List    Action = "-L"
	Flush   Action = "-F"
	New     Action = "-N"
)

var (
	once          sync.Once
	iptablesPath  string
	supportsXlock = false
)

type Table string
type Action string
type Chain string

// define iptables chain
type ChainInfo struct {
	Name        string
	Table       Table
	HairpinMode bool
}

// add rule to nat/PREROUTING chain
// iptables -t nat -A PREROUTING -m addrtype --dst-type LOCAL -j lx1036
func (chainInfo *ChainInfo) Rule(table Table, chain Chain, action Action, args ...string) error {
	a := []string{"-t", string(table), string(action), string(chain)}
	if len(args) != 0 {
		a = append(a, args...)
	}
	if output, err := IptableCmd(a...); err != nil {
		return err
	} else if len(output) != 0 {
		return fmt.Errorf("")
	}

	return nil
}

// add forwarding rule to 'filter' table and corresponding nat rule to 'nat' table
func (chainInfo *ChainInfo) Forward(action Action, protocal string, bridgeName string, ip net.IP, port int, dstAddr string, destPort int) error {

}

func NewChain(name string, table Table, haripinMode bool) (*ChainInfo, error) {
	chain := &ChainInfo{
		Name:        name,
		Table:       table,
		HairpinMode: haripinMode,
	}
	if len(chain.Table) == 0 { // default is Filter Table
		chain.Table = Filter
	}

	// create a chain.Table/chain.Name chain if not exists
	if _, err := IptableCmd("-t", string(chain.Table), "-n", string(List), chain.Name); err != nil {
		// -N is --new: Create a new user-defined chain
		if output, err := IptableCmd("-t", string(chain.Table), string(New), chain.Name); err != nil {
			return nil, err
		} else if len(string(output)) != 0 {
			return nil, fmt.Errorf("can't create %s/%s chain because of %s", chain.Table, chain.Name, output)
		}
	}

	return chain, nil
}

// ???
func Rule(chain *ChainInfo, bridgeName string, harbinMode bool, enable bool) error {
	switch chain.Table {
	case Nat:
		prerouting := []string{"-m", "addrtype", "--dst-type", "LOCAL", "-j", chain.Name}
		if enable && !Exists(chain.Table, Prerouting, prerouting...) {
			if err := chain.Rule(chain.Table, Prerouting, Append, prerouting...); err != nil {
				return fmt.Errorf("failed to create %s rule in %s chain: %v", chain.Name, Prerouting, err)
			}
		} else if !enable && Exists(chain.Table, Prerouting, prerouting...) {
			if err := chain.Rule(chain.Table, Prerouting, Delete, prerouting...); err != nil {
				return fmt.Errorf("failed to delete %s rule in %s chain: %v", chain.Name, Prerouting, err)
			}
		}

		output := []string{"-m", "addrtype", "--dst-type", "LOCAL", "-j", chain.Name}
		if enable && !Exists(chain.Table, Output, output...) {
			if err := chain.Rule(chain.Table, Output, Append, output...); err != nil {
				return fmt.Errorf("failed to create %s rule in %s chain: %v", chain.Name, Output, err)
			}
		} else if !enable && Exists(chain.Table, Output, output...) {
			if err := chain.Rule(chain.Table, Output, Delete, output...); err != nil {
				return fmt.Errorf("failed to delete %s rule in %s chain: %v", chain.Name, Output, err)
			}
		}
	case Filter:
		link := []string{"-o", bridgeName, "-j", chain.Name}
		if enable && !Exists(chain.Table, Forward, link...) {
			// iptables -I FORWARD -o lo -j lx1036
			if err := chain.Rule(chain.Table, Forward, Insert, link...); err != nil {
				return fmt.Errorf("failed to create %s rule in %s chain: %v", chain.Name, Forward, err)
			}
		} else if !enable && Exists(chain.Table, Forward, link...) {
			// iptables -D FORWARD -o lo -j lx1036
			if err := chain.Rule(chain.Table, Forward, Delete, link...); err != nil {
				return fmt.Errorf("failed to delete %s rule in %s chain: %v", chain.Name, Forward, err)
			}
		}
		return nil
	default:
		return fmt.Errorf("chain %s is not supported", chain.Table)
	}

	return nil
}

// syscall 'iptables' command
func IptableCmd(args ...string) ([]byte, error) {
	if firewallRunning {

	}

	return iptableCmd(args...)
}

func iptableCmd(args ...string) ([]byte, error) {
	// check iptables command
	if err := checkIptables(); err != nil {
		return nil, err
	}

	if supportsXlock {
		args = append([]string{"--wait"}, args...)
	} else {

	}

	output, err := exec.Command(iptablesPath, args...).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("iptables failed to exec command [%s], output is: %s, error is %v", iptablesPath+" "+strings.Join(args, " "), string(output), err)
	}

	return output, nil
}

func checkIptables() error {
	once.Do(initDependencies)

	if len(iptablesPath) == 0 {
		return fmt.Errorf("iptables not found")
	}

	return nil
}

func initDependencies() {
	path, err := exec.LookPath("iptables")
	if err != nil {
		log.Warnf("failed to find iptables: %v", err)
		return
	}
	if out, err := exec.Command(path, "--wait", "-t", "nat", "-L", "-n").CombinedOutput(); err != nil {
		log.Warnf("failed to run `iptables --wait -t nat -L -n` with messages: %s, error: %v", strings.TrimSpace(string(out)), err)
	}

	if err := FirewallInit(); err != nil {
		log.Debugf("failed to initialize firewall:%v, using raw iptables instead", err)
	}

	iptablesPath = path
	supportsXlock = exec.Command(iptablesPath, "--wait", "-L", "-n").Run() == nil

}

func supportsCheckOption() bool {
	return true
}

func Exists(table Table, chain Chain, rule ...string) bool {
	return exists(false, table, chain, rule...)
}

func exists(native bool, table Table, chain Chain, rule ...string) bool {

	if supportsCheckOption() {
		// -C --check: Check for the existence of a rule
		// 检查 rule 是否存在: iptables -t nat -C PREROUTING -m addrtype --dst-type LOCAL -j lx1036
		rules := append([]string{"-t", string(table), "-C", string(chain)}, rule...)
		_, err := iptableCmd(rules...)
		return err == nil
	}

	return true
}
