package daemon

import "fmt"

const (
	podNetworkTypeENIMultiIP = "ENIMultiIP"
)

func podInfoKey(namespace, name string) string {
	return fmt.Sprintf("%s/%s", namespace, name)
}
