package client

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

type patch struct {
	patchType types.PatchType
	data      []byte
}

// RawPatch constructs a new Patch with the given PatchType and data.
func RawPatch(patchType types.PatchType, data []byte) Patch {
	return &patch{patchType, data}
}
func (s *patch) Type() types.PatchType {
	return s.patchType
}

// Data implements Patch.
func (s *patch) Data(obj runtime.Object) ([]byte, error) {
	return s.data, nil
}
