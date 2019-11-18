package main

import (
	"fmt"
	"github.com/go-redis/redis/v7"
	"os"
)

func main() {
	manualSetup()
	os.Exit(0)

	rdb := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: []string{":7000", ":7001", ":7002", ":7003", ":7004", ":7005"},
	})
	fmt.Println(rdb.Ping().Val())
}

func manualSetup()  {
	// clusterSlots returns cluster slots information.
	// It can use service like ZooKeeper to maintain configuration information
	// and Cluster.ReloadState to manually trigger state reloading.
	clusterSlots := func() ([]redis.ClusterSlot, error) {
		slots := []redis.ClusterSlot{
			// First node with 1 master and 1 slave.
			{
				Start: 0,
				End:   8191,
				Nodes: []redis.ClusterNode{
					{
						Addr: ":7000", // master
					},
					{
						Addr: ":7003", // 1st slave
					},
				},
			},
			// Second node with 1 master and 1 slave.
			{
				Start: 8192,
				End:   16383,
				Nodes: []redis.ClusterNode{
					{
						Addr: ":7001", // master
					},
					{
						Addr: ":7004", // 1st slave
					},
				},
			},
		}

		return slots, nil
	}

	rdb := redis.NewClusterClient(&redis.ClusterOptions{
		ClusterSlots:  clusterSlots,
		RouteRandomly: true,
	})

	fmt.Println(rdb.Ping().Val())

	// ReloadState reloads cluster state. It calls ClusterSlots func
	// to get cluster slots information.
	err := rdb.ReloadState()
	if err != nil {
		panic(err)
	}

	fmt.Println(rdb.ClusterInfo().Val())
}
