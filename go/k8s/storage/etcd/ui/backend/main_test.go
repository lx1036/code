package main

import (
	"fmt"
	"github.com/spf13/viper"
	"k8s-lx1036/k8s/storage/etcd/ui/backend/common"
	"testing"
)

func TestViper(test *testing.T) {
	viper.AutomaticEnv()
	viper.SetConfigFile("./etcd.conf")
	if err := viper.ReadInConfig(); err != nil {
		panic(err.Error())
	}

	fmt.Println(viper.AllSettings())
	var etcdServer common.EtcdServer
	err := viper.Unmarshal(&etcdServer)
	if err != nil {
		panic(err.Error())
	}

	fmt.Println(etcdServer)
}
