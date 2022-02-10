package calico

import (
	"os"

	"github.com/projectcalico/calico/libcalico-go/lib/apiconfig"
	//client "github.com/projectcalico/calico/libcalico-go/lib/clientv3"
	"github.com/spf13/viper"
)

func GetCalicoClientOrDie() client.Interface {
	var kubeConfig = viper.GetString("kubeconfig")
	if len(kubeConfig) == 0 {
		kubeConfig = os.Getenv("KUBECONFIG")
		if len(kubeConfig) == 0 {
			kubeConfig = os.Getenv("HOME") + "/.kube/config"
		}
	}

	c, err := client.New(apiconfig.CalicoAPIConfig{
		Spec: apiconfig.CalicoAPIConfigSpec{
			DatastoreType: apiconfig.DatastoreType(viper.GetString("datastore-type")),
			KubeConfig: apiconfig.KubeConfig{
				Kubeconfig: kubeConfig,
			},
		},
	})
	if err != nil {
		panic(err)
	}

	return c
}
