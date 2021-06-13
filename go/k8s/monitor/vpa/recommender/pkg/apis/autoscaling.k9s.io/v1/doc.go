// overrides enabling or disabling deepcopy generation for this type https://cloudnative.to/kubebuilder/reference/markers/object.html
// https://github.com/kubernetes/code-generator/blob/master/cmd/deepcopy-gen/main.go

// +k8s:deepcopy-gen=package,register
// +groupName=autoscaling.k9s.io

package v1
