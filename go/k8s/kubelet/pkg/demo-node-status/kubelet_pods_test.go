package demo_node_status

import v1 "k8s.io/api/core/v1"

type testServiceLister struct {
	services []*v1.Service
}
