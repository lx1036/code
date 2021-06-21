


groups:
	bash "../../../../apiserver/generate-groups.sh" deepcopy,client,lister,informer \
	k8s-lx1036/k8s/scheduler/pkg/scheduler/client k8s-lx1036/k8s/scheduler/pkg/scheduler/apis \
	"scheduling:v1alpha1 podgroup:v1" \
	--output-base "." \
	--go-header-file "../../../../apiserver/hack/boilerplate/boilerplate.go.txt"

