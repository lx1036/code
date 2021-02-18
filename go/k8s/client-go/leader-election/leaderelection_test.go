package main

import (
	"encoding/json"
	"testing"
	"time"

	"k8s-lx1036/k8s/client-go/leader-election/pkg/resourcelock"

	coordinationv1 "k8s.io/api/coordination/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	fakeclient "k8s.io/client-go/testing"
)

type Reactor struct {
	verb       string
	objectType string
	reaction   fakeclient.ReactionFunc
}

func createLockObject(t *testing.T, objectType, namespace, name string, record *resourcelock.LeaderElectionRecord) (obj runtime.Object) {
	objectMeta := metav1.ObjectMeta{
		Namespace: namespace,
		Name:      name,
	}
	if record != nil {
		recordBytes, _ := json.Marshal(record)
		objectMeta.Annotations = map[string]string{
			resourcelock.LeaderElectionRecordAnnotationKey: string(recordBytes),
		}
	}
	switch objectType {
	case "endpoints":
		obj = &corev1.Endpoints{ObjectMeta: objectMeta}
	case "configmaps":
		obj = &corev1.ConfigMap{ObjectMeta: objectMeta}
	case "leases":
		var spec coordinationv1.LeaseSpec
		if record != nil {
			spec = resourcelock.LeaderElectionRecordToLeaseSpec(record)
		}
		obj = &coordinationv1.Lease{ObjectMeta: objectMeta, Spec: spec}
	default:
		t.Fatal("unexpected objType:" + objectType)
	}
	return
}

func testTryAcquireOrRenew(t *testing.T, objectType string) {
	future := time.Now().Add(1000 * time.Hour)
	past := time.Now().Add(-1000 * time.Hour)

	tests := []struct {
		name           string
		observedRecord resourcelock.LeaderElectionRecord
		observedTime   time.Time
		reactors       []Reactor

		expectSuccess    bool
		transitionLeader bool
		outHolder        string
	}{
		{
			name: "acquire from no object",
			reactors: []Reactor{
				{
					verb: "get",
					reaction: func(action fakeclient.Action) (handled bool, ret runtime.Object, err error) {
						return true, nil, errors.NewNotFound(action.(fakeclient.GetAction).GetResource().GroupResource(), action.(fakeclient.GetAction).GetName())
					},
				},
				{
					verb: "create",
					reaction: func(action fakeclient.Action) (handled bool, ret runtime.Object, err error) {
						return true, action.(fakeclient.CreateAction).GetObject(), nil
					},
				},
			},
			expectSuccess: true,
			outHolder:     "baz",
		},
		{
			name: "acquire from object without annotations",
			reactors: []Reactor{
				{
					verb: "get",
					reaction: func(action fakeclient.Action) (handled bool, ret runtime.Object, err error) {
						return true, createLockObject(t, objectType, action.GetNamespace(), action.(fakeclient.GetAction).GetName(), nil), nil
					},
				},
				{
					verb: "update",
					reaction: func(action fakeclient.Action) (handled bool, ret runtime.Object, err error) {
						return true, action.(fakeclient.CreateAction).GetObject(), nil
					},
				},
			},
			expectSuccess:    true,
			transitionLeader: true,
			outHolder:        "baz",
		},
		{
			name: "acquire from unled object",
			reactors: []Reactor{
				{
					verb: "get",
					reaction: func(action fakeclient.Action) (handled bool, ret runtime.Object, err error) {
						return true, createLockObject(t, objectType, action.GetNamespace(), action.(fakeclient.GetAction).GetName(), &resourcelock.LeaderElectionRecord{}), nil
					},
				},
				{
					verb: "update",
					reaction: func(action fakeclient.Action) (handled bool, ret runtime.Object, err error) {
						return true, action.(fakeclient.CreateAction).GetObject(), nil
					},
				},
			},

			expectSuccess:    true,
			transitionLeader: true,
			outHolder:        "baz",
		},
		{
			name: "acquire from led, unacked object",
			reactors: []Reactor{
				{
					verb: "get",
					reaction: func(action fakeclient.Action) (handled bool, ret runtime.Object, err error) {
						return true, createLockObject(t, objectType, action.GetNamespace(), action.(fakeclient.GetAction).GetName(), &resourcelock.LeaderElectionRecord{HolderIdentity: "bing"}), nil
					},
				},
				{
					verb: "update",
					reaction: func(action fakeclient.Action) (handled bool, ret runtime.Object, err error) {
						return true, action.(fakeclient.CreateAction).GetObject(), nil
					},
				},
			},
			observedRecord: resourcelock.LeaderElectionRecord{HolderIdentity: "bing"},
			observedTime:   past,

			expectSuccess:    true,
			transitionLeader: true,
			outHolder:        "baz",
		},
		{
			name: "acquire from empty led, acked object",
			reactors: []Reactor{
				{
					verb: "get",
					reaction: func(action fakeclient.Action) (handled bool, ret runtime.Object, err error) {
						return true, createLockObject(t, objectType, action.GetNamespace(), action.(fakeclient.GetAction).GetName(), &resourcelock.LeaderElectionRecord{HolderIdentity: ""}), nil
					},
				},
				{
					verb: "update",
					reaction: func(action fakeclient.Action) (handled bool, ret runtime.Object, err error) {
						return true, action.(fakeclient.CreateAction).GetObject(), nil
					},
				},
			},
			observedTime: future,

			expectSuccess:    true,
			transitionLeader: true,
			outHolder:        "baz",
		},
		{
			name: "don't acquire from led, acked object",
			reactors: []Reactor{
				{
					verb: "get",
					reaction: func(action fakeclient.Action) (handled bool, ret runtime.Object, err error) {
						return true, createLockObject(t, objectType, action.GetNamespace(), action.(fakeclient.GetAction).GetName(), &resourcelock.LeaderElectionRecord{HolderIdentity: "bing"}), nil
					},
				},
			},
			observedTime: future,

			expectSuccess: false,
			outHolder:     "bing",
		},
		{
			name: "renew already acquired object",
			reactors: []Reactor{
				{
					verb: "get",
					reaction: func(action fakeclient.Action) (handled bool, ret runtime.Object, err error) {
						return true, createLockObject(t, objectType, action.GetNamespace(), action.(fakeclient.GetAction).GetName(), &resourcelock.LeaderElectionRecord{HolderIdentity: "baz"}), nil
					},
				},
				{
					verb: "update",
					reaction: func(action fakeclient.Action) (handled bool, ret runtime.Object, err error) {
						return true, action.(fakeclient.CreateAction).GetObject(), nil
					},
				},
			},
			observedTime:   future,
			observedRecord: resourcelock.LeaderElectionRecord{HolderIdentity: "baz"},

			expectSuccess: true,
			outHolder:     "baz",
		},
	}

}

// Will test leader election using lease as the resource
func TestTryAcquireOrRenewLeases(t *testing.T) {
	testTryAcquireOrRenew(t, "leases")
}
