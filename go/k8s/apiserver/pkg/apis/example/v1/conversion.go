package v1

import "k8s.io/apimachinery/pkg/runtime"

func addConversionFuncs(scheme *runtime.Scheme) error {
	// Add non-generated conversion functions here. Currently there are none.
	return nil
}
