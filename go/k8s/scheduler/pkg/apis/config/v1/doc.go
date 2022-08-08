// INFO: 可以使用可选选项 // +groupGoName= 使用 Golang 驼峰标识符来为 group 指定别名, 解决冲突
//  @see https://github.com/kubernetes/community/blob/master/contributors/devel/sig-api-machinery/generating-clientset.md

// +k8s:deepcopy-gen=package
// +groupName=kubescheduler.config.k8s.io
// +groupGoName=Config
// +kubebuilder:storageversion
// +versionName=v1

package v1
