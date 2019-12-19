package whoami

import (
	"context"
	"fmt"
	"github.com/miekg/dns"
)

type Whoami struct {}

func (w Whoami) ServeDNS(ctx context.Context, writer dns.ResponseWriter, msg *dns.Msg) (int, error) {
	fmt.Println("asdfdsf")

	return 0, nil
}

func (w Whoami) Name() string {
	return "whoami"
}



