package calico

import (
	client "github.com/projectcalico/libcalico-go/lib/clientv3"
)

func GetCalicoClientOrDie() client.Interface {
	c, err := client.NewFromEnv()
	if err != nil {
		panic(err)
	}

	return c
}
