package raft

import (
	"context"
	"encoding/json"
	"fmt"
	"k8s.io/klog/v2"
)

type Volume struct {
	Name     string `json:"name"`
	Capacity string `json:"capacity"`
}

type Cluster struct {
	node *RaftNode
}

func NewCluster(node *RaftNode) *Cluster {
	return &Cluster{
		node: node,
	}
}

func (cluster *Cluster) start() {
	go func() {
		for {
			select {
			case apply := <-cluster.node.Apply():
				klog.Infof(fmt.Sprintf("start entries: %d", len(apply.entries)))
			case commit := <-cluster.node.Commit():
				cluster.Apply(commit)
			}
		}
	}()
}

func (cluster *Cluster) createVol(name, capacity string) (*Volume, error) {
	vol := &Volume{
		Name:     name,
		Capacity: capacity,
	}

	cluster.submitVol(vol)

	return vol, nil
}

func (cluster *Cluster) submitVol(vol *Volume) {
	data, _ := json.Marshal(vol)
	cluster.Propose(string(data))
}

type RaftCmd struct {
	Op string `json:"op"`
	K  string `json:"k"`
	V  string `json:"v"`
}

func (cluster *Cluster) Propose(data string) {
	cmd := new(RaftCmd)
	cmd.Op = "CREATE"
	cmd.K = fmt.Sprintf("%s%d", "#vol#", 1)
	cmd.V = data
	data2, _ := json.Marshal(cmd)
	cluster.node.Propose(context.TODO(), data2)
}

func (cluster *Cluster) Apply(commit []string) {
	for _, value := range commit {
		cmd := new(RaftCmd)
		err := json.Unmarshal([]byte(value), cmd)
		if err != nil {
			klog.Error(err)
			continue
		}

		klog.Infof(fmt.Sprintf("%+v", *cmd))
	}
}
