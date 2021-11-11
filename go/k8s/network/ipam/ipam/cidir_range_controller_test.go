package ipam

import (
	"context"
	"net"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/kubernetes/pkg/controller/testutil"
)

type testCase struct {
	description     string
	fakeNodeHandler *testutil.FakeNodeHandler
	allocatorParams CIDRAllocatorParams
	// key is index of the cidr allocated
	expectedAllocatedCIDR map[int]string
	allocatedCIDRs        map[int][]string
	// should controller creation fail?
	ctrlCreateFail bool
}

func TestAllocateOrOccupyCIDR(test *testing.T) {

	// all tests operate on a single node
	testCases := []testCase{
		{
			description: "When there's no ServiceCIDR return first CIDR in range",
			fakeNodeHandler: &testutil.FakeNodeHandler{
				Existing: []*corev1.Node{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "node0",
						},
					},
				},
				Clientset: fake.NewSimpleClientset(),
			},
			allocatorParams: CIDRAllocatorParams{
				ClusterCIDRs: func() []*net.IPNet {
					_, clusterCIDR, _ := net.ParseCIDR("127.123.234.0/24")
					return []*net.IPNet{clusterCIDR}
				}(),
				ServiceCIDR:          nil,
				SecondaryServiceCIDR: nil,
				NodeCIDRMaskSizes:    []int{30},
			},
			expectedAllocatedCIDR: map[int]string{
				0: "127.123.234.0/30",
			},
		},
	}

	for _, tc := range testCases {
		test.Run(tc.description, func(t *testing.T) {
			// Initialize the range allocator.
			fakeNodeInformer := getFakeNodeInformer(tc.fakeNodeHandler)
			nodeList, _ := tc.fakeNodeHandler.List(context.TODO(), metav1.ListOptions{})
			_, err := NewCIDRRangeAllocator(tc.fakeNodeHandler, fakeNodeInformer, nodeList, tc.allocatorParams)
			if err == nil && tc.ctrlCreateFail {
				t.Fatalf("creating range allocator was expected to fail, but it did not")
			}
			if err != nil && !tc.ctrlCreateFail {
				t.Fatalf("creating range allocator was expected to succeed, but it did not")
			}
		})
	}
}

// Creates a fakeNodeInformer using the provided fakeNodeHandler.
func getFakeNodeInformer(fakeNodeHandler *testutil.FakeNodeHandler) coreinformers.NodeInformer {
	fakeClient := &fake.Clientset{}
	fakeInformerFactory := informers.NewSharedInformerFactory(fakeClient, 0)
	fakeNodeInformer := fakeInformerFactory.Core().V1().Nodes()

	for _, node := range fakeNodeHandler.Existing {
		fakeNodeInformer.Informer().GetStore().Add(node)
	}

	return fakeNodeInformer
}
