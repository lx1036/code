package recommender

import (
	"k8s-lx1036/k8s/monitor/vpa/recommender/pkg/input"
	"k8s-lx1036/k8s/monitor/vpa/recommender/pkg/types"
)

type Recommender struct {
	clusterState 	                  *types.ClusterState
	
	
	clusterStateFeeder            input.ClusterStateFeeder
	
}






func (r *Recommender) RunOnce() {
	
	r.clusterStateFeeder.LoadVPAs()
	
	r.clusterStateFeeder.LoadPods()
	
	r.clusterStateFeeder.LoadRealTimeMetrics()
	
	r.UpdateVPAs()
	
}

// Updates VPA CRD objects' statuses.
func (r *Recommender) UpdateVPAs() {
	
	
	for _, observedVpa := range r.clusterState.ObservedVpas {
		key := types.VpaID{
			Namespace: observedVpa.Namespace,
			VpaName:   observedVpa.Name,
		}
		vpa, found := r.clusterState.Vpas[key]
		if !found {
			continue
		}
		
		
		vpa.UpdateRecommendation(getCappedRecommendation(vpa.ID, resources, observedVpa.Spec.ResourcePolicy))
		
		
		
		_, err := vpa_utils.UpdateVpaStatusIfNeeded(
			r.vpaClient.VerticalPodAutoscalers(vpa.ID.Namespace), vpa.ID.VpaName, vpa.AsStatus(), &observedVpa.Status)
		if err != nil {
			klog.Errorf(
				"Cannot update VPA %v object. Reason: %+v", vpa.ID.VpaName, err)
		}
	}
}

