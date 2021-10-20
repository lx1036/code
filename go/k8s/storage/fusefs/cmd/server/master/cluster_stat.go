package master

import "k8s.io/klog/v2"

type nodeStatInfo struct {
	TotalGB     uint64
	UsedGB      uint64
	IncreasedGB int64
	UsedRatio   string
}

// Check the total space, available space, and daily-used space in data nodes,  meta nodes, and volumes
func (cluster *Cluster) updateStatInfo() {
	defer func() {
		if err := recover(); err != nil {
			klog.Errorf("updateStatInfo occurred panic,err[%v]", err)
		}
	}()

	//cluster.updateMetaNodeStatInfo()
	//cluster.updateVolStatInfo()
}
