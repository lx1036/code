package v1

import "k8s.io/apimachinery/pkg/runtime/schema"

// GroupName is the group name used in this package
const GroupName = "kubescheduler.config.k8s.io"

// SchemeGroupVersion is group version used to register these objects
var SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1"}
