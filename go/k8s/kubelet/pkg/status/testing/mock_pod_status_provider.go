package testing

import (
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

// MockStatusProvider mocks a PodStatusProvider.
type MockStatusProvider struct {
	mock.Mock
}

func (m *MockStatusProvider) GetPodStatus(uid types.UID) (v1.PodStatus, bool) {
	args := m.Called(uid)
	return args.Get(0).(v1.PodStatus), args.Bool(1)
}
