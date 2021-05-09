package v1

import "k8s.io/apimachinery/pkg/runtime"

func addDefaultingFuncs(scheme *runtime.Scheme) error {
	// return RegisterDefaults(scheme)
	return nil
}
