package nodeipam

import (
	"github.com/projectcalico/calico/libcalico-go/lib/selector"
	"github.com/projectcalico/calico/libcalico-go/lib/selector/tokenizer"
	"k8s.io/klog/v2"
	"testing"
)

func TestNodeSelector(test *testing.T) {
	sel, _ := selector.Parse("group/network=='default'")
	klog.Info(sel.String(), sel.UniqueID())

	tokens, _ := tokenizer.Tokenize("group/network=='default'")
	for _, token := range tokens {
		klog.Info(token.Kind, token.Value)
	}

}
