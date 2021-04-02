package podautoscaler

import (
	"testing"
	"time"

	autoscalingv1 "k8s.io/api/autoscaling/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2beta2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func TestUnsafeConvertToVersionVia(test *testing.T) {
	now := metav1.Time{Time: time.Now().Add(-time.Hour)}
	lastScaleTime := &now
	minReplicas := int32(2)
	maxReplicas := int32(6)
	specReplicas := int32(2)
	CPUCurrent := int32(40)
	quantity := resource.MustParse("400m")

	obj := &autoscalingv2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-hpa",
			Namespace: "test-namespace",
			SelfLink:  "experimental/v1/namespaces/test-namespace/horizontalpodautoscalers/test-hpa",
		},
		Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
				Kind:       "ReplicationController",
				Name:       "test-rc",
				APIVersion: "v1",
			},
			MinReplicas: &minReplicas,
			MaxReplicas: maxReplicas,
		},
		Status: autoscalingv2.HorizontalPodAutoscalerStatus{
			CurrentReplicas: specReplicas,
			DesiredReplicas: specReplicas,
			LastScaleTime:   lastScaleTime,
			CurrentMetrics: []autoscalingv2.MetricStatus{
				{
					Type: autoscalingv2.ResourceMetricSourceType,
					Resource: &autoscalingv2.ResourceMetricStatus{
						Name: v1.ResourceCPU,
						Current: autoscalingv2.MetricValueStatus{
							AverageValue:       &quantity,
							AverageUtilization: &CPUCurrent,
						},
					},
				},
			},
			Conditions: []autoscalingv2.HorizontalPodAutoscalerCondition{
				{
					Type:               autoscalingv2.AbleToScale,
					Status:             v1.ConditionTrue,
					LastTransitionTime: *lastScaleTime,
					Reason:             "ReadyForNewScale",
					Message:            "recommended size matches current size",
				},
				{
					Type:               autoscalingv2.ScalingActive,
					Status:             v1.ConditionTrue,
					LastTransitionTime: *lastScaleTime,
					Reason:             "ValidMetricFound",
					Message:            "the HPA was able to successfully calculate a replica count from cpu resource utilization (percentage of request)",
				},
				{
					Type:               autoscalingv2.ScalingLimited,
					Status:             v1.ConditionTrue,
					LastTransitionTime: *lastScaleTime,
					Reason:             "TooFewReplicas",
					Message:            "the desired replica count is less than the minimum replica count",
				},
			},
		},
	}
	// convert to autoscaling v1
	objv1, err := UnsafeConvertToVersionVia(obj, autoscalingv1.SchemeGroupVersion)
	if err != nil {
		panic(err)
	}

	klog.Infof("%v", objv1)
}

// TestStoreScaleEvents tests events storage and usage
func TestStoreScaleEvents(t *testing.T) {
	type TestCase struct {
		name                   string
		key                    string
		replicaChange          int32
		prevScaleEvents        []timestampedScaleEvent
		newScaleEvents         []timestampedScaleEvent
		scalingRules           *autoscalingv2.HPAScalingRules
		expectedReplicasChange int32
	}
	tests := []TestCase{
		{
			name:                   "empty entries with default behavior",
			replicaChange:          5,
			prevScaleEvents:        []timestampedScaleEvent{}, // no history -> 0 replica change
			newScaleEvents:         []timestampedScaleEvent{}, // no behavior -> no events are stored
			expectedReplicasChange: 0,
		},
		{
			name:                   "empty entries with two-policy-behavior",
			replicaChange:          5,
			prevScaleEvents:        []timestampedScaleEvent{}, // no history -> 0 replica change
			newScaleEvents:         []timestampedScaleEvent{{5, time.Now(), false}},
			scalingRules:           generateScalingRules(10, 60, 100, 60, 0),
			expectedReplicasChange: 0,
		},
		{
			name:          "one outdated entry to be kept untouched without behavior",
			replicaChange: 5,
			prevScaleEvents: []timestampedScaleEvent{
				{7, time.Now().Add(-time.Second * time.Duration(61)), false}, // outdated event, should be replaced
			},
			newScaleEvents: []timestampedScaleEvent{
				{7, time.Now(), false}, // no behavior -> we don't touch stored events
			},
			expectedReplicasChange: 0,
		},
		{
			name:          "one outdated entry to be replaced with behavior",
			replicaChange: 5,
			prevScaleEvents: []timestampedScaleEvent{
				{7, time.Now().Add(-time.Second * time.Duration(61)), false}, // outdated event, should be replaced
			},
			newScaleEvents: []timestampedScaleEvent{
				{5, time.Now(), false},
			},
			scalingRules:           generateScalingRules(10, 60, 100, 60, 0),
			expectedReplicasChange: 0,
		},
		{
			name:          "one actual entry to be not touched with behavior",
			replicaChange: 5,
			prevScaleEvents: []timestampedScaleEvent{
				{7, time.Now().Add(-time.Second * time.Duration(58)), false},
			},
			newScaleEvents: []timestampedScaleEvent{
				{7, time.Now(), false},
				{5, time.Now(), false},
			},
			scalingRules:           generateScalingRules(10, 60, 100, 60, 0),
			expectedReplicasChange: 7,
		},
		{
			name:          "two entries, one of them to be replaced",
			replicaChange: 5,
			prevScaleEvents: []timestampedScaleEvent{
				{7, time.Now().Add(-time.Second * time.Duration(61)), false}, // outdated event, should be replaced
				{6, time.Now().Add(-time.Second * time.Duration(59)), false},
			},
			newScaleEvents: []timestampedScaleEvent{
				{5, time.Now(), false},
				{6, time.Now(), false},
			},
			scalingRules:           generateScalingRules(10, 60, 0, 0, 0),
			expectedReplicasChange: 6,
		},
		{
			name:          "replace one entry, use policies with different periods",
			replicaChange: 5,
			prevScaleEvents: []timestampedScaleEvent{
				{8, time.Now().Add(-time.Second * time.Duration(29)), false},
				{6, time.Now().Add(-time.Second * time.Duration(59)), false},
				{7, time.Now().Add(-time.Second * time.Duration(61)), false}, // outdated event, should be marked as outdated
				{9, time.Now().Add(-time.Second * time.Duration(61)), false}, // outdated event, should be replaced
			},
			newScaleEvents: []timestampedScaleEvent{
				{8, time.Now(), false},
				{6, time.Now(), false},
				{7, time.Now(), true},
				{5, time.Now(), false},
			},
			scalingRules:           generateScalingRules(10, 60, 100, 30, 0),
			expectedReplicasChange: 14,
		},
		{
			name:          "two entries, both actual",
			replicaChange: 5,
			prevScaleEvents: []timestampedScaleEvent{
				{7, time.Now().Add(-time.Second * time.Duration(58)), false},
				{6, time.Now().Add(-time.Second * time.Duration(59)), false},
			},
			newScaleEvents: []timestampedScaleEvent{
				{7, time.Now(), false},
				{6, time.Now(), false},
				{5, time.Now(), false},
			},
			scalingRules:           generateScalingRules(10, 120, 100, 30, 0),
			expectedReplicasChange: 13,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

		})
	}
}
