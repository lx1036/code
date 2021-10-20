package master

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
)

func (cluster *Cluster) scheduleToLoadMetaPartitions() {
	go wait.UntilWithContext(context.TODO(), func(ctx context.Context) {
		//check vols after switching leader two minutes
		if cluster.partition != nil && cluster.partition.IsRaftLeader() {
			if cluster.vols != nil {
				//cluster.checkLoadMetaPartitions()
			}
		}
	}, time.Second*defaultIntervalToCheckMetaPartition)
}
