package deployment

import (
	"k8s-lx1036/dashboard/backend/api"
	"k8s-lx1036/dashboard/backend/errors"
	metricapi "k8s-lx1036/dashboard/backend/integration/metric/api"
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	apps "k8s.io/api/apps/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

func TestGetDeploymentListFromChannels(test *testing.T) {
	cases := []struct {
		k8sDeployment      apps.DeploymentList
		k8sDeploymentError error
		pods               *v1.PodList
		expected           *DeploymentList
		expectedError      error
	}{
		{
			apps.DeploymentList{},
			nil,
			&v1.PodList{},
			&DeploymentList{
				ListMeta:          api.ListMeta{},
				Deployments:       []Deployment{},
				CumulativeMetrics: make([]metricapi.Metric, 0),
				Errors:            []error{},
			},
			nil,
		},
		{
			apps.DeploymentList{},
			errors.NewInvalid("MyCustomError"),
			&v1.PodList{},
			nil,
			errors.NewInvalid("MyCustomError"),
		},
		{
			apps.DeploymentList{},
			&k8serrors.StatusError{},
			&v1.PodList{},
			nil,
			&k8serrors.StatusError{},
		},
		{
			apps.DeploymentList{},
			&k8serrors.StatusError{ErrStatus: metaV1.Status{}},
			&v1.PodList{},
			nil,
			&k8serrors.StatusError{ErrStatus: metaV1.Status{}},
		},
		{
			apps.DeploymentList{},
			&k8serrors.StatusError{ErrStatus: metaV1.Status{Reason: "foo-bar"}},
			&v1.PodList{},
			nil,
			&k8serrors.StatusError{ErrStatus: metaV1.Status{Reason: "foo-bar"}},
		},
		{
			apps.DeploymentList{
				Items: []apps.Deployment{{
					ObjectMeta: metaV1.ObjectMeta{
						Name:              "rs-name",
						Namespace:         "rs-namespace",
						Labels:            map[string]string{"key": "value"},
						CreationTimestamp: metaV1.Unix(111, 222),
					},
					Spec: apps.DeploymentSpec{
						Selector: &metaV1.LabelSelector{MatchLabels: map[string]string{"foo": "bar"}},
						Replicas: getReplicasPointer(21),
					},
					Status: apps.DeploymentStatus{
						Replicas: 7,
					},
				}},
			},
			nil,
			&v1.PodList{},
			&DeploymentList{
				ListMeta:          api.ListMeta{TotalItems: 1},
				CumulativeMetrics: make([]metricapi.Metric, 0),
				Status:            common.ResourceStatus{Running: 1},
				Deployments: []Deployment{{
					ObjectMeta: api.ObjectMeta{
						Name:              "rs-name",
						Namespace:         "rs-namespace",
						Labels:            map[string]string{"key": "value"},
						CreationTimestamp: metaV1.Unix(111, 222),
					},
					TypeMeta: api.TypeMeta{
						Kind:     api.ResourceKindDeployment,
						Scalable: true,
					},
					Pods: common.PodInfo{
						Current:  7,
						Desired:  getReplicasPointer(21),
						Failed:   0,
						Warnings: []common.Event{},
					},
				}},
				Errors: []error{},
			},
			nil,
		},
	}

	for _, c := range cases {
		channels := &common.ResourceChannels{
			DeploymentList: common.DeploymentListChannel{
				List:  make(chan *apps.DeploymentList, 1),
				Error: make(chan error, 1),
			},
			NodeList: common.NodeListChannel{
				List:  make(chan *v1.NodeList, 1),
				Error: make(chan error, 1),
			},
			ServiceList: common.ServiceListChannel{
				List:  make(chan *v1.ServiceList, 1),
				Error: make(chan error, 1),
			},
			PodList: common.PodListChannel{
				List:  make(chan *v1.PodList, 1),
				Error: make(chan error, 1),
			},
			EventList: common.EventListChannel{
				List:  make(chan *v1.EventList, 1),
				Error: make(chan error, 1),
			},
			ReplicaSetList: common.ReplicaSetListChannel{
				List:  make(chan *apps.ReplicaSetList, 1),
				Error: make(chan error, 1),
			},
		}
	}
}
