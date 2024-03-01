package main

import (
    "errors"
    "fmt"
    "github.com/cilium/ebpf"
    "github.com/sirupsen/logrus"
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
    "inet.af/netaddr"
    "strings"
    "unsafe"
)

var (
    bindLabel    string
    bindProtocol string
    bindIpPrefix string
    bindPort     int
)

func init() {
    rootCmd.AddCommand(bindCmd)

    flags := bindCmd.PersistentFlags()
    flags.StringVarP(&bindLabel, "label", "", "foo", "label")
    viper.BindPFlag("label", flags.Lookup("label"))
    flags.StringVarP(&bindProtocol, "protocol", "", "tcp", "protocol")
    viper.BindPFlag("protocol", flags.Lookup("protocol"))
    flags.StringVarP(&bindIpPrefix, "ip-prefix", "", "127.0.0.0/24", "ip-prefix")
    viper.BindPFlag("ip-prefix", flags.Lookup("ip-prefix"))
    flags.IntVarP(&bindPort, "port", "", 0, "port")
    viper.BindPFlag("port", flags.Lookup("port"))

}

var bindCmd = &cobra.Command{
    Use: "bind",
    Example: `
        bind --label=foo --protocol=udp --ip-prefix=127.0.0.1 --port=0
        bind --label=bar --protocol=tcp --ip-prefix=127.0.0.1/32 --port=80
    `,
    Run: func(cmd *cobra.Command, args []string) {
        dispatcher, err := OpenDispatcher()
        if err != nil {
            logrus.Errorf("[bind]err: %v", err)
            return
        }

        binding, err := NewBinding(bindLabel, ConvertProtocol(bindProtocol), bindIpPrefix, uint16(bindPort))
        if err != nil {
            logrus.Errorf("[bind]err: %v", err)
            return
        }

        dispatcher.AddBinding(binding)
    },
}

type Binding struct {
    Label    string
    Protocol Protocol
    Prefix   netaddr.IPPrefix
    Port     uint16
}

func NewBinding(label string, proto Protocol, prefix string, port uint16) (*Binding, error) {
    cidr, err := ParsePrefix(prefix)
    if err != nil {
        return nil, err
    }

    return &Binding{
        label,
        proto,
        netaddr.IPPrefixFrom(cidr.IP(), cidr.Bits()).Masked(),
        port,
    }, nil
}

// ParsePrefix 127.0.0.1 or 127.0.0.1/24
func ParsePrefix(prefix string) (netaddr.IPPrefix, error) {
    if strings.ContainsRune(prefix, '/') {
        return netaddr.ParseIPPrefix(prefix)
    }

    ip, err := netaddr.ParseIP(prefix)
    if err != nil {
        return netaddr.IPPrefix{}, err
    }

    var prefixBits uint8
    if ip.Is4() {
        prefixBits = 32
    } else {
        prefixBits = 128
    }

    return netaddr.IPPrefixFrom(ip, prefixBits), nil
}

type bindingKey struct {
    PrefixLen uint32
    Protocol  Protocol
    Port      uint16
    IP        [16]byte
}

const bindingKeyHeaderBits = uint8(unsafe.Sizeof(bindingKey{}.Protocol)+unsafe.Sizeof(bindingKey{}.Port)) * 8

func newBindingKey(bind *Binding) *bindingKey {
    // Get the length of the prefix
    prefixLen := bind.Prefix.Bits()

    // If the prefix is v4, offset it by 96
    if bind.Prefix.IP().Is4() { // ???
        prefixLen += 96
    }

    key := bindingKey{
        PrefixLen: uint32(bindingKeyHeaderBits + prefixLen), // ???
        Protocol:  bind.Protocol,
        Port:      bind.Port,
        IP:        bind.Prefix.IP().As16(),
    }

    return &key
}

type bindingValue struct {
    ID        DestinationID
    PrefixLen uint32
}

// AddBinding bindings 和 destinations bpf map 要一一对应
func (dispatcher *Dispatcher) AddBinding(binding *Binding) error {

    key := newBindingKey(binding)

    var old bindingValue
    var releaseOldID bool
    if err := dispatcher.bindings.Lookup(key, &old); err == nil {
        // Since the LPM trie will return the "best" match we have to make sure
        // that the prefix length matches to ensure that we're replacing a binding,
        // not just installing a more specific one.
        releaseOldID = old.PrefixLen == key.PrefixLen
    } else if !errors.Is(err, ebpf.ErrKeyNotExist) {
        return fmt.Errorf("lookup binding: %s", err)
    }

    destination := newDestinationFromBinding(binding)
    id, err := dispatcher.destinations.Acquire(destination) // -> destinations map
    if err != nil {
        return fmt.Errorf("acquire destination: %s", err)
    }

    value := bindingValue{ID: id, PrefixLen: key.PrefixLen}
    err = dispatcher.bindings.Put(key, &value) // -> bindings map
    if err != nil {
        return err
    }

    if releaseOldID {
        dispatcher.destinations.ReleaseByID(old.ID)
    }

    return nil
}
