// +build linux

package iptables

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
	"os/exec"
	"strconv"
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

	Drop   Policy = "DROP"
	Accept Policy = "ACCEPT"
)

var (
	iptabelsOnce sync.Once
	iptablesPath string
)

type Table string
type Action string
type Chain string
type Policy string

// define iptables chain
type ChainInfo struct {
	Name        string
	Table       Table
	HairpinMode bool
	Mu          *sync.Mutex
}

// 在 'table' 表内新建 'name' chain
func NewChain(name string, table Table, haripinMode bool) (*ChainInfo, error) {
	chain := &ChainInfo{
		Name:        name,
		Table:       table,
		HairpinMode: haripinMode,
		Mu:          &sync.Mutex{},
	}
	if len(chain.Table) == 0 { // default is Filter Table
		chain.Table = Filter
	}

	// create a chain.Table/chain.Name chain if not exists
	/*
		sudo iptables -t nat -n -L lx1036
		# Warning: iptables-legacy tables present, use iptables-legacy to see them
		iptables: No chain/target/match by that name.
	*/
	if _, err := IptablesCmd("-t", string(chain.Table), "-n", string(List), chain.Name); err != nil {
		// -N is --new: Create a new user-defined chain
		/*
			sudo iptables -t nat -N lx1036
			sudo iptables-save -t nat
		*/
		if output, err := IptablesCmd("-t", string(chain.Table), string(New), chain.Name); err != nil {
			return nil, err
		} else if len(string(output)) != 0 {
			return nil, fmt.Errorf("can't create %s/%s chain because of %s", chain.Table, chain.Name, output)
		}
	}

	return chain, nil
}

// iptables -t table -P chain DROP/ACCEPT
func SetPolicy(table Table, chain Chain, policy Policy) error {
	args := []string{"-t", string(table), "-P", string(chain), string(policy)}
	var mu = &sync.Mutex{}
	mu.Lock()
	defer mu.Unlock()
	output, err := IptablesCmd(args...)
	if err != nil {
		return err
	} else if len(output) != 0 {
		log.Infof("iptables cmd[iptables %s] output: %s", strings.Join(args, " "), string(output))
		return fmt.Errorf("can't add %s/%s policy because of %s", table, chain, output)
	}

	return nil
}

// check/add/delete rule to nat/PREROUTING chain
// iptables -t nat -C/A/D PREROUTING/OUTPUT -m addrtype --dst-type LOCAL -j lx1036
func (chainInfo *ChainInfo) Rule(action Action, chain Chain, rule ...string) error {
	args := []string{string(action), string(chain)}
	args = append(args, rule...)
	chainInfo.Mu.Lock()
	defer chainInfo.Mu.Unlock()
	output, err := IptablesCmd(args...)
	if err != nil {
		return err
	} else if len(output) != 0 {
		log.Infof("iptables cmd[iptables %s] output: %s", strings.Join(args, " "), string(output))
		return fmt.Errorf("can't add %s/%s rule because of %s", chainInfo.Table, chainInfo.Name, output)
	}

	return nil
}

// add forwarding rule to 'filter' table and corresponding nat rule to 'nat' table
// forward "tcp://192.168.1.1:1234" -> "tcp://172.17.0.1:4321"
func (chainInfo *ChainInfo) Forward(action Action, protocal string, bridgeName string, ip net.IP, port int, dstAddr string, dstPort int) error {
	if strings.ToLower(protocal) != "tcp" || strings.ToLower(protocal) != "udp" {
		return fmt.Errorf("unsupported protocal %s", protocal)
	}

	// 1.
	args := []string{"-t", string(Nat), "-p", ptotocal, "-d", dstAddr, "-dport", strconv.Itoa(dstPort), "-j", "DNAT", "--to-destination"}
	chainInfo.Rule(action, chain, args...)

	return nil
}

// add rule to a chain if rule not exits in the chain
func Rule(chainInfo *ChainInfo, bridgeName string, harbinMode bool, enable bool) error {
	switch chainInfo.Table {
	case Nat:
		chains := []Chain{Prerouting, Output}
		for _, chain := range chains {
			ruleArgs := []string{"-t", string(chainInfo.Table), "-m", "addrtype", "--dst-type", "LOCAL", "-j", chainInfo.Name}
			/*
				sudo iptables -C PREROUTING/OUTPUT -t nat -m addrtype --dst-type LOCAL -j lx1036
				iptables: Bad rule (does a matching rule exist in that chain?).
			*/
			if enable && chainInfo.Rule(Check, chain, ruleArgs...) != nil {
				/*
					sudo iptables -A PREROUTING/OUTPUT -t nat -m addrtype --dst-type LOCAL -j lx1036
					iptables: Bad rule (does a matching rule exist in that chain?).
				*/
				if err := chainInfo.Rule(Append, chain, ruleArgs...); err != nil {
					return fmt.Errorf("failed to %s %s rule in %s chain: %v", Append, chainInfo.Name, chain, err)
				}
			} else if !enable && chainInfo.Rule(Check, chain, ruleArgs...) == nil {
				/*
					sudo iptables -D PREROUTING/OUTPUT -t nat -m addrtype --dst-type LOCAL -j lx1036
					iptables: Bad rule (does a matching rule exist in that chain?).
				*/
				if err := chainInfo.Rule(Delete, chain, ruleArgs...); err != nil {
					return fmt.Errorf("failed to %s %s rule in %s chain: %v", Delete, chainInfo.Name, chain, err)
				}
			}
		}
	case Filter:
		if len(bridgeName) == 0 {
			return fmt.Errorf("missing bridge name in %s/%s", chainInfo.Table, chainInfo.Name)
		}
		link := []string{"-o", bridgeName, "-j", chainInfo.Name}
		establish := []string{"-o", bridgeName, "-m", "conntrack", "--ctstate", "RELATED,ESTABLISHED", "-j", "ACCEPT"}

		rules := [][]string{link, establish}
		for _, rule := range rules {
			if enable && chainInfo.Rule(Check, Forward, rule...) != nil {
				/*
					sudo iptables -I FORWARD -o lo -j lx1036
					sudo iptables -I FORWARD -o lo -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT
				*/
				if err := chainInfo.Rule(Insert, Forward, rule...); err != nil {
					return fmt.Errorf("failed to %s %s rule in %s chain: %v", Insert, strings.Join(rule, " "), Forward, err)
				}
			} else if !enable && chainInfo.Rule(Check, Forward, rule...) == nil {
				if err := chainInfo.Rule(Delete, Forward, rule...); err != nil {
					return fmt.Errorf("failed to %s %s rule in %s chain: %v", Delete, strings.Join(rule, " "), Forward, err)
				}
			}
		}
		return nil
	default:
		return fmt.Errorf("chain %s is not supported", chainInfo.Table)
	}

	return nil
}

// syscall 'iptables' command
func IptablesCmd(args ...string) ([]byte, error) {
	// check iptables command
	if err := checkIptables(); err != nil {
		return nil, err
	}
	var mu = &sync.Mutex{}
	mu.Lock()
	defer mu.Unlock()
	output, err := exec.Command(iptablesPath, args...).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("iptables failed to exec command [%s], output is: %s, error is %v", iptablesPath+" "+strings.Join(args, " "), string(output), err)
	}

	return output, nil
}

func checkIptables() error {
	iptabelsOnce.Do(initDependencies)

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
}
