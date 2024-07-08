package features

import (
	"k8s.io/apimachinery/pkg/util/runtime"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/component-base/featuregate"
)

const (
	// WorkLoadSupport can cache and operate **K8s native resource**, Deployment/Replicas/ReplicationController/StatefulSet resources currently.
	WorkLoadSupport featuregate.Feature = "WorkLoadSupport"

	// PodDisruptionBudgetsSupport can cache and support PodDisruptionBudgets
	PodDisruptionBudgetsSupport featuregate.Feature = "PodDisruptionBudgetsSupport"
)

func init() {
	runtime.Must(utilfeature.DefaultMutableFeatureGate.Add(defaultVolcanoFeatureGates))
}

var defaultVolcanoFeatureGates = map[featuregate.Feature]featuregate.FeatureSpec{
	WorkLoadSupport:             {Default: true, PreRelease: featuregate.Alpha},
	PodDisruptionBudgetsSupport: {Default: true, PreRelease: featuregate.Alpha},
}
