


groups:
	bash "../../../apiserver/generate-groups.sh" deepcopy,client,lister,informer \
	k8s-lx1036/k8s/scheduler/pkg/client k8s-lx1036/k8s/scheduler/pkg/apis \
	"scheduling:v1alpha1 podgroup:v1" \
	--output-base "." \
	--go-header-file "../../../apiserver/hack/boilerplate/boilerplate.go.txt"

scheduler-config:
	bash "../../../apiserver/generate-groups.sh" deepcopy \
	k8s-lx1036/k8s/scheduler/pkg/client k8s-lx1036/k8s/scheduler/pkg/apis \
	"config:v1" \
	--output-base "." \
	--go-header-file "../../../apiserver/hack/boilerplate/boilerplate.go.txt"
